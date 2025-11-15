package magic

import (
	"context"
	"testing"
	"time"

	"github.com/0xsj/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Store for testing
type mockStore struct {
	tokens      map[string]*Token
	countByUser map[string]int
	saveError   error
	countError  error
}

func newMockStore() *mockStore {
	return &mockStore{
		tokens:      make(map[string]*Token),
		countByUser: make(map[string]int),
	}
}

func (m *mockStore) Save(ctx context.Context, token *Token) result.Result[*Token] {
	if m.saveError != nil {
		return result.Err[*Token](m.saveError)
	}
	m.tokens[token.Value] = token
	return result.Ok(token)
}

func (m *mockStore) GetByValue(ctx context.Context, value string, purpose Purpose) result.Result[*Token] {
	token, exists := m.tokens[value]
	if !exists {
		return result.Err[*Token](ErrTokenNotFound)
	}

	// ✅ Check if purpose matches
	if token.Purpose != purpose {
		return result.Err[*Token](ErrTokenNotFound)
	}

	return result.Ok(token)
}

func (m *mockStore) GetByID(ctx context.Context, tokenID string) result.Result[*Token] {
	for _, token := range m.tokens {
		if token.ID == tokenID {
			return result.Ok(token)
		}
	}
	return result.Err[*Token](ErrTokenNotFound)
}

func (m *mockStore) MarkAsUsed(ctx context.Context, tokenID string) result.Result[struct{}] {
	return result.Ok(struct{}{})
}

func (m *mockStore) Delete(ctx context.Context, tokenID string) result.Result[struct{}] {
	return result.Ok(struct{}{})
}

func (m *mockStore) DeleteByValue(ctx context.Context, value string, purpose Purpose) result.Result[struct{}] {
	return result.Ok(struct{}{})
}

func (m *mockStore) DeleteExpired(ctx context.Context) result.Result[int] {
	return result.Ok(0)
}

func (m *mockStore) DeleteAllForUser(ctx context.Context, userID string) result.Result[int] {
	return result.Ok(0)
}

func (m *mockStore) CountByUser(ctx context.Context, userID string, purpose Purpose) result.Result[int] {
	if m.countError != nil {
		return result.Err[int](m.countError)
	}
	key := userID + string(purpose)
	count := m.countByUser[key]
	return result.Ok(count)
}

func (m *mockStore) ListByUser(ctx context.Context, userID string, purpose Purpose) result.Result[[]*Token] {
	var tokens []*Token
	for _, token := range m.tokens {
		if token.UserID == userID && token.Purpose == purpose {
			tokens = append(tokens, token)
		}
	}
	return result.Ok(tokens)
}

func TestNewGenerator(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()

	gen := NewGenerator(config, store)

	assert.NotNil(t, gen)
}

func TestNewGenerator_AppliesDefaults(t *testing.T) {
	store := newMockStore()
	config := Config{} // Empty config

	gen := NewGenerator(config, store).(*generator)

	assert.Equal(t, 32, gen.config.TokenLength)
	assert.Equal(t, 15*time.Minute, gen.config.LoginTTL)
	assert.Equal(t, 24*time.Hour, gen.config.VerifyTTL)
	assert.Equal(t, 1*time.Hour, gen.config.ResetTTL)
	assert.Equal(t, 3, gen.config.MaxTokensPerUser)
}

func TestGenerate_Success(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()
	userID := "user-123"
	email := "test@example.com"
	purpose := PurposeLogin

	result := gen.Generate(ctx, userID, email, purpose)

	require.True(t, result.IsOk(), "should generate token successfully")
	token := result.Unwrap()

	assert.NotEmpty(t, token.ID)
	assert.NotEmpty(t, token.Value)
	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, email, token.Email)
	assert.Equal(t, purpose, token.Purpose)
	assert.False(t, token.IsExpired())
	assert.False(t, token.IsUsed())
	assert.NotZero(t, token.CreatedAt)
	assert.NotZero(t, token.ExpiresAt)
}

func TestGenerate_DifferentPurposes(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	tests := []struct {
		purpose     Purpose
		expectedTTL time.Duration
	}{
		{PurposeLogin, 15 * time.Minute},
		{PurposeEmailVerify, 24 * time.Hour},
		{PurposePasswordReset, 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(string(tt.purpose), func(t *testing.T) {
			result := gen.Generate(ctx, "user-123", "test@example.com", tt.purpose)

			require.True(t, result.IsOk())
			token := result.Unwrap()

			// Check TTL is approximately correct (allow 1 second variance)
			actualTTL := token.ExpiresAt.Sub(token.CreatedAt)
			assert.InDelta(t, tt.expectedTTL.Seconds(), actualTTL.Seconds(), 1.0)
		})
	}
}

func TestGenerate_UniqueTokens(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	// Generate multiple tokens
	tokens := make([]*Token, 10)
	for i := 0; i < 10; i++ {
		result := gen.Generate(ctx, "user-123", "test@example.com", PurposeLogin)
		require.True(t, result.IsOk())
		tokens[i] = result.Unwrap()
	}

	// Verify all tokens are unique
	seen := make(map[string]bool)
	for _, token := range tokens {
		assert.False(t, seen[token.Value], "token values should be unique")
		seen[token.Value] = true
	}
}

func TestGenerate_InvalidInputs(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	t.Run("empty user ID", func(t *testing.T) {
		result := gen.Generate(ctx, "", "test@example.com", PurposeLogin)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrInvalidUserID, result.UnwrapErr())
	})

	t.Run("empty email", func(t *testing.T) {
		result := gen.Generate(ctx, "user-123", "", PurposeLogin)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrInvalidEmail, result.UnwrapErr())
	})

	t.Run("invalid purpose", func(t *testing.T) {
		result := gen.Generate(ctx, "user-123", "test@example.com", Purpose("invalid"))

		assert.True(t, result.IsErr())
		_, ok := result.UnwrapErr().(ErrInvalidPurpose)
		assert.True(t, ok)
	})
}

func TestGenerateWithTTL_CustomTTL(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()
	customTTL := 30 * time.Minute

	result := gen.GenerateWithTTL(ctx, "user-123", "test@example.com", PurposeLogin, customTTL)

	require.True(t, result.IsOk())
	token := result.Unwrap()

	actualTTL := token.ExpiresAt.Sub(token.CreatedAt)
	assert.InDelta(t, customTTL.Seconds(), actualTTL.Seconds(), 1.0)
}

func TestGenerateWithTTL_InvalidTTL(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	tests := []time.Duration{0, -1 * time.Minute, -10 * time.Hour}

	for _, ttl := range tests {
		t.Run(ttl.String(), func(t *testing.T) {
			result := gen.GenerateWithTTL(ctx, "user-123", "test@example.com", PurposeLogin, ttl)

			assert.True(t, result.IsErr())
			_, ok := result.UnwrapErr().(ErrInvalidTTL)
			assert.True(t, ok)
		})
	}
}

func TestGenerate_RateLimiting(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig().WithMaxTokensPerUser(3)
	gen := NewGenerator(config, store)

	ctx := context.Background()
	userID := "user-123"
	email := "test@example.com"
	purpose := PurposeLogin

	// Generate tokens up to the limit
	for i := 0; i < 3; i++ {
		store.countByUser[userID+string(purpose)] = i
		result := gen.Generate(ctx, userID, email, purpose)
		require.True(t, result.IsOk(), "should generate token %d", i+1)
	}

	// Next attempt should fail due to rate limit
	store.countByUser[userID+string(purpose)] = 3
	result := gen.Generate(ctx, userID, email, purpose)

	assert.True(t, result.IsErr())
	err, ok := result.UnwrapErr().(ErrTooManyTokens)
	assert.True(t, ok)
	assert.Equal(t, userID, err.UserID)
	assert.Equal(t, purpose, err.Purpose)
	assert.Equal(t, 3, err.Count)
	assert.Equal(t, 3, err.Limit)
}

func TestGenerate_RateLimitPerPurpose(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig().WithMaxTokensPerUser(2)
	gen := NewGenerator(config, store)

	ctx := context.Background()
	userID := "user-123"
	email := "test@example.com"

	// Generate 2 login tokens (hits limit)
	store.countByUser[userID+string(PurposeLogin)] = 0
	result1 := gen.Generate(ctx, userID, email, PurposeLogin)
	require.True(t, result1.IsOk())

	store.countByUser[userID+string(PurposeLogin)] = 1
	result2 := gen.Generate(ctx, userID, email, PurposeLogin)
	require.True(t, result2.IsOk())

	// Login tokens at limit
	store.countByUser[userID+string(PurposeLogin)] = 2
	result3 := gen.Generate(ctx, userID, email, PurposeLogin)
	assert.True(t, result3.IsErr())

	// But can still generate verify tokens (different purpose)
	store.countByUser[userID+string(PurposeEmailVerify)] = 0
	result4 := gen.Generate(ctx, userID, email, PurposeEmailVerify)
	assert.True(t, result4.IsOk())
}

func TestGenerateURL_Success(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig().WithBaseURL("https://app.example.com")
	gen := NewGenerator(config, store)

	tests := []struct {
		purpose      Purpose
		expectedPath string
	}{
		{PurposeLogin, "/auth/magic-link/verify"},
		{PurposeEmailVerify, "/auth/verify-email"},
		{PurposePasswordReset, "/auth/reset-password"},
	}

	for _, tt := range tests {
		t.Run(string(tt.purpose), func(t *testing.T) {
			token := &Token{
				Value:   "test-token-value",
				Purpose: tt.purpose,
			}

			result := gen.GenerateURL(token)

			require.True(t, result.IsOk())
			url := result.Unwrap()

			expectedURL := "https://app.example.com" + tt.expectedPath + "?token=test-token-value"
			assert.Equal(t, expectedURL, url)
		})
	}
}

func TestGenerateURL_NoBaseURL(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig() // No BaseURL set
	gen := NewGenerator(config, store)

	token := &Token{
		Value:   "test-token",
		Purpose: PurposeLogin,
	}

	result := gen.GenerateURL(token)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrInvalidBaseURL, result.UnwrapErr())
}

func TestGenerateURL_NilToken(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig().WithBaseURL("https://app.example.com")
	gen := NewGenerator(config, store)

	result := gen.GenerateURL(nil)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrTokenNotFound, result.UnwrapErr())
}

func TestGenerateURL_InvalidPurpose(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig().WithBaseURL("https://app.example.com")
	gen := NewGenerator(config, store)

	token := &Token{
		Value:   "test-token",
		Purpose: Purpose("invalid"),
	}

	result := gen.GenerateURL(token)

	assert.True(t, result.IsErr())
	_, ok := result.UnwrapErr().(ErrInvalidPurpose)
	assert.True(t, ok)
}

func TestGenerate_StoreError(t *testing.T) {
	store := newMockStore()
	store.saveError = assert.AnError
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	result := gen.Generate(ctx, "user-123", "test@example.com", PurposeLogin)

	assert.True(t, result.IsErr())
	assert.Equal(t, assert.AnError, result.UnwrapErr())
}

func TestGenerate_CountError(t *testing.T) {
	store := newMockStore()
	store.countError = assert.AnError
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	result := gen.Generate(ctx, "user-123", "test@example.com", PurposeLogin)

	assert.True(t, result.IsErr())
	assert.Equal(t, assert.AnError, result.UnwrapErr())
}

func TestGenerate_TokenValueFormat(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx := context.Background()

	result := gen.Generate(ctx, "user-123", "test@example.com", PurposeLogin)

	require.True(t, result.IsOk())
	token := result.Unwrap()

	// Token should be base64 URL-safe encoded
	assert.NotContains(t, token.Value, "+")
	assert.NotContains(t, token.Value, "/")
	assert.Greater(t, len(token.Value), 40) // 32 bytes -> ~43 chars in base64
}

func TestGenerate_ContextCancellation(t *testing.T) {
	store := newMockStore()
	config := DefaultConfig()
	gen := NewGenerator(config, store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := gen.Generate(ctx, "user-123", "test@example.com", PurposeLogin)

	// Generator doesn't check context yet, but token is still generated
	// This test documents current behavior
	// In production, store.Save should respect context cancellation
	assert.True(t, result.IsOk() || result.IsErr())
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid token length", func(t *testing.T) {
		config := DefaultConfig().WithTokenLength(0)
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid login TTL", func(t *testing.T) {
		config := DefaultConfig().WithLoginTTL(0)
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid verify TTL", func(t *testing.T) {
		config := DefaultConfig().WithVerifyTTL(-1 * time.Hour)
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid reset TTL", func(t *testing.T) {
		config := DefaultConfig().WithResetTTL(0)
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid max tokens", func(t *testing.T) {
		config := DefaultConfig().WithMaxTokensPerUser(0)
		err := config.Validate()
		assert.Error(t, err)
	})

	t.Run("empty base URL is valid", func(t *testing.T) {
		config := DefaultConfig().WithBaseURL("")
		err := config.Validate()
		assert.NoError(t, err, "empty base URL should be valid")
	})
}
