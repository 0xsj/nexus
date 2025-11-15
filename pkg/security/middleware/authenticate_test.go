package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xsj/nexus/pkg/security/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestJWTManager(t *testing.T) jwt.Manager {
	config := jwt.DefaultConfig().
		WithSecretKey("test-secret-key-that-is-long-enough-for-security!!")

	managerResult := jwt.NewManager(config)
	require.True(t, managerResult.IsOk())
	return managerResult.Unwrap()
}

func TestAuthenticate_Success(t *testing.T) {
	jwtManager := setupTestJWTManager(t)

	// Generate token
	tokenResult := jwtManager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	rec := httptest.NewRecorder()

	// Handler that checks context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "user-123", claims.UserID)
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	middleware := Authenticate(jwtManager)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticate_MissingToken(t *testing.T) {
	jwtManager := setupTestJWTManager(t)

	// Create request without token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := Authenticate(jwtManager)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	jwtManager := setupTestJWTManager(t)

	// Create request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := Authenticate(jwtManager)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticate_ExpiredToken(t *testing.T) {
	config := jwt.DefaultConfig().
		WithSecretKey("test-secret-key-that-is-long-enough-for-security!!").
		WithAccessTokenTTL(10 * time.Millisecond)

	managerResult := jwt.NewManager(config)
	require.True(t, managerResult.IsOk())
	jwtManager := managerResult.Unwrap()

	// Generate token
	tokenResult := jwtManager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := Authenticate(jwtManager)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestOptionalAuthenticate_WithToken(t *testing.T) {
	jwtManager := setupTestJWTManager(t)

	// Generate token
	tokenResult := jwtManager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		assert.True(t, ok, "should have claims in context")
		assert.Equal(t, "user-123", claims.UserID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuthenticate(jwtManager)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalAuthenticate_WithoutToken(t *testing.T) {
	jwtManager := setupTestJWTManager(t)

	// Create request without token
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := ClaimsFromContext(r.Context())
		assert.False(t, ok, "should not have claims in context")
		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuthenticate(jwtManager)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "lowercase bearer",
			header:   "bearer xyz789",
			expected: "xyz789",
		},
		{
			name:     "mixed case bearer",
			header:   "BeArEr token123",
			expected: "token123",
		},
		{
			name:     "missing bearer",
			header:   "abc123",
			expected: "",
		},
		{
			name:     "empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "wrong scheme",
			header:   "Basic abc123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token := extractBearerToken(req)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestExtractTokenFromQuery(t *testing.T) {
	extractor := extractTokenFromQuery("token")

	t.Run("token present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?token=abc123", nil)
		token := extractor(req)
		assert.Equal(t, "abc123", token)
	})

	t.Run("token missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		token := extractor(req)
		assert.Empty(t, token)
	})
}

func TestExtractTokenFromCookie(t *testing.T) {
	extractor := extractTokenFromCookie("auth_token")

	t.Run("cookie present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: "abc123",
		})

		token := extractor(req)
		assert.Equal(t, "abc123", token)
	})

	t.Run("cookie missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		token := extractor(req)
		assert.Empty(t, token)
	})
}

func TestChainTokenExtractors(t *testing.T) {
	extractor := ChainTokenExtractors(
		extractBearerToken,
		extractTokenFromQuery("token"),
		extractTokenFromCookie("auth_token"),
	)

	t.Run("found in header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")

		token := extractor(req)
		assert.Equal(t, "header-token", token)
	})

	t.Run("found in query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)

		token := extractor(req)
		assert.Equal(t, "query-token", token)
	})

	t.Run("found in cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: "cookie-token",
		})

		token := extractor(req)
		assert.Equal(t, "cookie-token", token)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		token := extractor(req)
		assert.Empty(t, token)
	})
}

func TestAuthenticateWithConfig_CustomExtractor(t *testing.T) {
	jwtManager := setupTestJWTManager(t)

	// Generate token
	tokenResult := jwtManager.GenerateAccessToken("user-123", "test@example.com", "testuser", []string{"user"})
	require.True(t, tokenResult.IsOk())
	token := tokenResult.Unwrap()

	// Create request with token in query
	req := httptest.NewRequest(http.MethodGet, "/?token="+token, nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "user-123", claims.UserID)
		w.WriteHeader(http.StatusOK)
	})

	config := DefaultAuthenticateConfig()
	config.TokenExtractor = extractTokenFromQuery("token")

	middleware := AuthenticateWithConfig(jwtManager, config)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
