package magic

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerifier(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	assert.NotNil(t, verifier)
}

func TestVerify_Success(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create a valid token in store
	token := &Token{
		ID:        "token-123",
		Value:     "valid-token-value",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	result := verifier.Verify(ctx, token.Value, PurposeLogin)

	require.True(t, result.IsOk(), "should verify token successfully")
	verifiedToken := result.Unwrap()

	assert.Equal(t, token.ID, verifiedToken.ID)
	assert.Equal(t, token.UserID, verifiedToken.UserID)
	assert.Equal(t, token.Email, verifiedToken.Email)
	assert.Equal(t, token.Purpose, verifiedToken.Purpose)
	assert.False(t, verifiedToken.IsUsed(), "token should not be marked as used")
}

func TestVerify_EmptyTokenValue(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	ctx := context.Background()

	result := verifier.Verify(ctx, "", PurposeLogin)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrTokenNotFound, result.UnwrapErr())
}

func TestVerify_InvalidPurpose(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	ctx := context.Background()

	result := verifier.Verify(ctx, "some-token", Purpose("invalid"))

	assert.True(t, result.IsErr())
	_, ok := result.UnwrapErr().(ErrInvalidPurpose)
	assert.True(t, ok)
}

func TestVerify_TokenNotFound(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	ctx := context.Background()

	result := verifier.Verify(ctx, "non-existent-token", PurposeLogin)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrTokenNotFound, result.UnwrapErr())
}

func TestVerify_ExpiredToken(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create an expired token
	token := &Token{
		ID:        "token-123",
		Value:     "expired-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		UsedAt:    nil,
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	result := verifier.Verify(ctx, token.Value, PurposeLogin)

	assert.True(t, result.IsErr())
	err, ok := result.UnwrapErr().(ErrTokenExpired)
	assert.True(t, ok)
	assert.Equal(t, token.ID, err.TokenID)
}

func TestVerify_UsedToken(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create a used token
	usedAt := time.Now().Add(-30 * time.Minute)
	token := &Token{
		ID:        "token-123",
		Value:     "used-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    &usedAt,
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	result := verifier.Verify(ctx, token.Value, PurposeLogin)

	assert.True(t, result.IsErr())
	err, ok := result.UnwrapErr().(ErrTokenAlreadyUsed)
	assert.True(t, ok)
	assert.Equal(t, token.ID, err.TokenID)
}

func TestVerify_WrongPurpose(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create a login token
	token := &Token{
		ID:        "token-123",
		Value:     "login-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	// Try to verify as email verification token
	result := verifier.Verify(ctx, token.Value, PurposeEmailVerify)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrTokenNotFound, result.UnwrapErr())
}

func TestVerifyAndConsume_Success(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create a valid token
	token := &Token{
		ID:        "token-123",
		Value:     "valid-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	result := verifier.VerifyAndConsume(ctx, token.Value, PurposeLogin)

	require.True(t, result.IsOk(), "should verify and consume token successfully")
	consumedToken := result.Unwrap()

	assert.True(t, consumedToken.IsUsed(), "token should be marked as used")
	assert.NotNil(t, consumedToken.UsedAt)
}

func TestVerifyAndConsume_ExpiredToken(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create expired token
	token := &Token{
		ID:        "token-123",
		Value:     "expired-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		UsedAt:    nil,
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	result := verifier.VerifyAndConsume(ctx, token.Value, PurposeLogin)

	assert.True(t, result.IsErr())
	_, ok := result.UnwrapErr().(ErrTokenExpired)
	assert.True(t, ok)
}

func TestVerifyAndConsume_AlreadyUsed(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create used token
	usedAt := time.Now().Add(-10 * time.Minute)
	token := &Token{
		ID:        "token-123",
		Value:     "used-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    &usedAt,
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	result := verifier.VerifyAndConsume(ctx, token.Value, PurposeLogin)

	assert.True(t, result.IsErr())
	_, ok := result.UnwrapErr().(ErrTokenAlreadyUsed)
	assert.True(t, ok)
}

func TestVerifyAndConsume_Idempotent(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Create valid token
	token := &Token{
		ID:        "token-123",
		Value:     "valid-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	// First consumption - should succeed
	result1 := verifier.VerifyAndConsume(ctx, token.Value, PurposeLogin)
	require.True(t, result1.IsOk())

	// Update store to reflect the used token
	consumedToken := result1.Unwrap()
	store.tokens[token.Value] = consumedToken

	// Second consumption - should fail
	result2 := verifier.VerifyAndConsume(ctx, token.Value, PurposeLogin)
	assert.True(t, result2.IsErr())
	_, ok := result2.UnwrapErr().(ErrTokenAlreadyUsed)
	assert.True(t, ok)
}

func TestVerifyAndConsume_TokenNotFound(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	ctx := context.Background()

	result := verifier.VerifyAndConsume(ctx, "non-existent-token", PurposeLogin)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrTokenNotFound, result.UnwrapErr())
}

func TestVerify_DifferentPurposes(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	purposes := []Purpose{
		PurposeLogin,
		PurposeEmailVerify,
		PurposePasswordReset,
	}

	for _, purpose := range purposes {
		t.Run(string(purpose), func(t *testing.T) {
			token := &Token{
				ID:        "token-" + string(purpose),
				Value:     "token-value-" + string(purpose),
				UserID:    "user-123",
				Purpose:   purpose,
				Email:     "test@example.com",
				ExpiresAt: time.Now().Add(15 * time.Minute),
				UsedAt:    nil,
				CreatedAt: time.Now(),
			}
			store.tokens[token.Value] = token

			ctx := context.Background()

			result := verifier.Verify(ctx, token.Value, purpose)

			require.True(t, result.IsOk())
			verifiedToken := result.Unwrap()
			assert.Equal(t, purpose, verifiedToken.Purpose)
		})
	}
}

func TestVerifyAndConsume_CompleteFlow(t *testing.T) {
	store := newMockStore()
	verifier := NewVerifier(store)

	// Simulate complete magic link flow
	token := &Token{
		ID:        "token-123",
		Value:     "magic-link-token",
		UserID:    "user-123",
		Purpose:   PurposeLogin,
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
	store.tokens[token.Value] = token

	ctx := context.Background()

	// Step 1: Verify without consuming (preview/validation)
	verifyResult := verifier.Verify(ctx, token.Value, PurposeLogin)
	require.True(t, verifyResult.IsOk())
	assert.False(t, verifyResult.Unwrap().IsUsed())

	// Step 2: Actually consume (login)
	consumeResult := verifier.VerifyAndConsume(ctx, token.Value, PurposeLogin)
	require.True(t, consumeResult.IsOk())
	consumedToken := consumeResult.Unwrap()

	assert.True(t, consumedToken.IsUsed())
	assert.NotNil(t, consumedToken.UsedAt)
	assert.Equal(t, token.UserID, consumedToken.UserID)
	assert.Equal(t, token.Email, consumedToken.Email)
}
