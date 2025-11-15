package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")

	result := NewManager(config)

	require.True(t, result.IsOk(), "should create manager successfully")
	manager := result.Unwrap()
	assert.NotNil(t, manager)
}

func TestNewManager_InvalidConfig(t *testing.T) {
	t.Run("missing secret key", func(t *testing.T) {
		config := DefaultConfig().WithSecretKey("")

		result := NewManager(config)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrInvalidSigningKey, result.UnwrapErr())
	})

	t.Run("weak secret key", func(t *testing.T) {
		config := DefaultConfig().WithSecretKey("short")

		result := NewManager(config)

		assert.True(t, result.IsErr())
		_, ok := result.UnwrapErr().(ErrWeakSecretKey)
		assert.True(t, ok)
	})

	t.Run("missing issuer", func(t *testing.T) {
		config := DefaultConfig().
			WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!").
			WithIssuer("")

		result := NewManager(config)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrMissingIssuer, result.UnwrapErr())
	})

	t.Run("missing audience", func(t *testing.T) {
		config := DefaultConfig().
			WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!").
			WithAudience("")

		result := NewManager(config)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrMissingAudience, result.UnwrapErr())
	})

	t.Run("invalid access token TTL", func(t *testing.T) {
		config := DefaultConfig().
			WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!").
			WithAccessTokenTTL(0)

		result := NewManager(config)

		assert.True(t, result.IsErr())
		_, ok := result.UnwrapErr().(ErrInvalidTTL)
		assert.True(t, ok)
	})

	t.Run("refresh TTL shorter than access TTL", func(t *testing.T) {
		config := DefaultConfig().
			WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!").
			WithAccessTokenTTL(1 * time.Hour).
			WithRefreshTokenTTL(30 * time.Minute)

		result := NewManager(config)

		assert.True(t, result.IsErr())
		_, ok := result.UnwrapErr().(ErrRefreshTTLTooShort)
		assert.True(t, ok)
	})
}

func TestGenerateAccessToken_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	userID := "user-123"
	email := "test@example.com"
	username := "testuser"
	roles := []string{"admin", "user"}

	result := manager.GenerateAccessToken(userID, email, username, roles)

	require.True(t, result.IsOk(), "should generate access token successfully")
	token := result.Unwrap()

	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".") // JWT format has dots
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	userID := "user-123"

	result := manager.GenerateRefreshToken(userID)

	require.True(t, result.IsOk(), "should generate refresh token successfully")
	token := result.Unwrap()

	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".")
}

func TestGenerateTokenPair_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	userID := "user-123"
	email := "test@example.com"
	username := "testuser"
	roles := []string{"admin"}

	result := manager.GenerateTokenPair(userID, email, username, roles)

	require.True(t, result.IsOk(), "should generate token pair successfully")
	pair := result.Unwrap()

	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, "Bearer", pair.TokenType)
	assert.Greater(t, pair.ExpiresIn, int64(0))
	assert.False(t, pair.ExpiresAt.IsZero())
}

func TestValidateToken_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	userID := "user-123"
	email := "test@example.com"
	username := "testuser"
	roles := []string{"admin", "user"}

	// Generate token
	tokenResult := manager.GenerateAccessToken(userID, email, username, roles)
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	// Validate token
	claimsResult := manager.ValidateToken(token)

	require.True(t, claimsResult.IsOk(), "should validate token successfully")
	claims := claimsResult.Unwrap()

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
	assert.Equal(t, config.Issuer, claims.Issuer)
	assert.Equal(t, userID, claims.Subject)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	invalidTokens := []string{
		"invalid.token.format",
		"",
		"not-a-jwt",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
	}

	for _, token := range invalidTokens {
		t.Run(token, func(t *testing.T) {
			result := manager.ValidateToken(token)
			assert.True(t, result.IsErr(), "should fail to validate invalid token")
		})
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"admin"})
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	claimsResult := manager.ValidateAccessToken(token)

	require.True(t, claimsResult.IsOk())
	claims := claimsResult.Unwrap()
	assert.True(t, claims.IsAccessToken())
}

func TestValidateAccessToken_WrongTokenType(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	// Generate refresh token
	tokenResult := manager.GenerateRefreshToken("user-123")
	require.True(t, tokenResult.IsOk())
	refreshToken := tokenResult.Unwrap()

	// Try to validate as access token
	claimsResult := manager.ValidateAccessToken(refreshToken)

	assert.True(t, claimsResult.IsErr())
	err, ok := claimsResult.UnwrapErr().(ErrInvalidTokenType)
	assert.True(t, ok)
	assert.Equal(t, string(TokenTypeAccess), err.Expected)
	assert.Equal(t, string(TokenTypeRefresh), err.Actual)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	tokenResult := manager.GenerateRefreshToken("user-123")
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	claimsResult := manager.ValidateRefreshToken(token)

	require.True(t, claimsResult.IsOk())
	claims := claimsResult.Unwrap()
	assert.True(t, claims.IsRefreshToken())
}

func TestValidateRefreshToken_WrongTokenType(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	// Generate access token
	tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, tokenResult.IsOk())
	accessToken := tokenResult.Unwrap()

	// Try to validate as refresh token
	claimsResult := manager.ValidateRefreshToken(accessToken)

	assert.True(t, claimsResult.IsErr())
	err, ok := claimsResult.UnwrapErr().(ErrInvalidTokenType)
	assert.True(t, ok)
	assert.Equal(t, string(TokenTypeRefresh), err.Expected)
	assert.Equal(t, string(TokenTypeAccess), err.Actual)
}

func TestRefreshAccessToken_Success(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	// Generate initial token pair
	pairResult := manager.GenerateTokenPair("user-123", "test@example.com", "testuser", []string{"admin"})
	require.True(t, pairResult.IsOk())
	initialPair := pairResult.Unwrap()

	// Refresh using refresh token
	newPairResult := manager.RefreshAccessToken(
		initialPair.RefreshToken,
		"test@example.com",
		"testuser",
		[]string{"admin"},
	)

	require.True(t, newPairResult.IsOk())
	newPair := newPairResult.Unwrap()

	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)
	assert.NotEqual(t, initialPair.AccessToken, newPair.AccessToken, "new access token should be different")
}

func TestRefreshAccessToken_InvalidRefreshToken(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	result := manager.RefreshAccessToken(
		"invalid.refresh.token",
		"test@example.com",
		"testuser",
		[]string{"user"},
	)

	assert.True(t, result.IsErr())
}

func TestRefreshAccessToken_UsingAccessToken(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	// Generate access token
	accessTokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, accessTokenResult.IsOk())
	accessToken := accessTokenResult.Unwrap()

	// Try to refresh using access token (should fail)
	result := manager.RefreshAccessToken(accessToken, "test@example.com", "testuser", []string{"user"})

	assert.True(t, result.IsErr())
	_, ok := result.UnwrapErr().(ErrInvalidTokenType)
	assert.True(t, ok)
}

func TestTokenExpiration(t *testing.T) {
	// Create config with very short TTL for testing
	config := DefaultConfig().
		WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!").
		WithAccessTokenTTL(100 * time.Millisecond).
		WithRefreshTokenTTL(1 * time.Second)

	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	// Generate token
	tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	// Token should be valid immediately
	claimsResult := manager.ValidateToken(token)
	assert.True(t, claimsResult.IsOk(), "token should be valid initially")

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Token should be expired now
	claimsResult = manager.ValidateToken(token)
	assert.True(t, claimsResult.IsErr(), "token should be expired")
	assert.Equal(t, ErrTokenExpired, claimsResult.UnwrapErr())
}

func TestClaims_HasRole(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	roles := []string{"admin", "editor", "viewer"}
	tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", roles)
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	claimsResult := manager.ValidateToken(token)
	require.True(t, claimsResult.IsOk())
	claims := claimsResult.Unwrap()

	assert.True(t, claims.HasRole("admin"))
	assert.True(t, claims.HasRole("editor"))
	assert.True(t, claims.HasRole("viewer"))
	assert.False(t, claims.HasRole("superadmin"))
}

func TestClaims_HasAnyRole(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	roles := []string{"editor", "viewer"}
	tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", roles)
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	claimsResult := manager.ValidateToken(token)
	require.True(t, claimsResult.IsOk())
	claims := claimsResult.Unwrap()

	assert.True(t, claims.HasAnyRole("admin", "editor"))
	assert.True(t, claims.HasAnyRole("viewer", "superadmin"))
	assert.False(t, claims.HasAnyRole("admin", "superadmin"))
}

func TestClaims_HasAllRoles(t *testing.T) {
	config := DefaultConfig().WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")
	managerResult := NewManager(config)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	roles := []string{"admin", "editor", "viewer"}
	tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", roles)
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	claimsResult := manager.ValidateToken(token)
	require.True(t, claimsResult.IsOk())
	claims := claimsResult.Unwrap()

	assert.True(t, claims.HasAllRoles("admin", "editor"))
	assert.True(t, claims.HasAllRoles("viewer"))
	assert.False(t, claims.HasAllRoles("admin", "superadmin"))
}

func TestDifferentAlgorithms(t *testing.T) {
	algorithms := []Algorithm{
		AlgorithmHS256,
		AlgorithmHS384,
		AlgorithmHS512,
	}

	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			config := DefaultConfig().
				WithAlgorithm(algo).
				WithSecretKey("this-is-a-very-secure-secret-key-with-32-chars!")

			managerResult := NewManager(config)
			require.True(t, managerResult.IsOk())
			manager := managerResult.Unwrap()

			// Generate and validate token
			tokenResult := manager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
			require.True(t, tokenResult.IsOk())
			token := tokenResult.Unwrap()

			claimsResult := manager.ValidateToken(token)
			assert.True(t, claimsResult.IsOk(), "should work with %s algorithm", algo)
		})
	}
}

func TestConfig_BuilderPattern(t *testing.T) {
	config := DefaultConfig().
		WithAlgorithm(AlgorithmHS512).
		WithSecretKey("my-super-secret-key-that-is-long-enough!!").
		WithIssuer("my-app").
		WithAudience("my-api").
		WithAccessTokenTTL(30 * time.Minute).
		WithRefreshTokenTTL(14 * 24 * time.Hour)

	assert.Equal(t, AlgorithmHS512, config.Algorithm)
	assert.Equal(t, "my-app", config.Issuer)
	assert.Equal(t, "my-api", config.Audience)
	assert.Equal(t, 30*time.Minute, config.AccessTokenTTL)
	assert.Equal(t, 14*24*time.Hour, config.RefreshTokenTTL)
}

func TestAlgorithm_IsSymmetric(t *testing.T) {
	assert.True(t, AlgorithmHS256.IsSymmetric())
	assert.True(t, AlgorithmHS384.IsSymmetric())
	assert.True(t, AlgorithmHS512.IsSymmetric())

	assert.False(t, AlgorithmRS256.IsSymmetric())
	assert.False(t, AlgorithmES256.IsSymmetric())
}

func TestAlgorithm_IsAsymmetric(t *testing.T) {
	assert.False(t, AlgorithmHS256.IsAsymmetric())

	assert.True(t, AlgorithmRS256.IsAsymmetric())
	assert.True(t, AlgorithmRS384.IsAsymmetric())
	assert.True(t, AlgorithmRS512.IsAsymmetric())
	assert.True(t, AlgorithmES256.IsAsymmetric())
}
