package magic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurpose_String(t *testing.T) {
	tests := []struct {
		purpose  Purpose
		expected string
	}{
		{PurposeLogin, "login"},
		{PurposeEmailVerify, "email_verify"},
		{PurposePasswordReset, "password_reset"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.purpose.String())
		})
	}
}

func TestPurpose_IsValid(t *testing.T) {
	t.Run("valid purposes", func(t *testing.T) {
		validPurposes := []Purpose{
			PurposeLogin,
			PurposeEmailVerify,
			PurposePasswordReset,
		}

		for _, purpose := range validPurposes {
			assert.True(t, purpose.IsValid(), "purpose %s should be valid", purpose)
		}
	})

	t.Run("invalid purposes", func(t *testing.T) {
		invalidPurposes := []Purpose{
			Purpose("invalid"),
			Purpose(""),
			Purpose("random_string"),
		}

		for _, purpose := range invalidPurposes {
			assert.False(t, purpose.IsValid(), "purpose %s should be invalid", purpose)
		}
	})
}

func TestToken_IsExpired(t *testing.T) {
	t.Run("token not expired", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		assert.False(t, token.IsExpired(), "token should not be expired")
	})

	t.Run("token expired", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now().Add(-10 * time.Minute),
		}

		assert.True(t, token.IsExpired(), "token should be expired")
	})

	t.Run("token expires exactly now", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now(),
		}

		// Sleep a tiny bit to ensure time has passed
		time.Sleep(1 * time.Millisecond)

		assert.True(t, token.IsExpired(), "token should be expired")
	})
}

func TestToken_IsUsed(t *testing.T) {
	t.Run("token not used", func(t *testing.T) {
		token := &Token{
			UsedAt: nil,
		}

		assert.False(t, token.IsUsed(), "token should not be used")
	})

	t.Run("token used", func(t *testing.T) {
		now := time.Now()
		token := &Token{
			UsedAt: &now,
		}

		assert.True(t, token.IsUsed(), "token should be used")
	})
}

func TestToken_IsValid(t *testing.T) {
	t.Run("valid token - not expired, not used", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now().Add(10 * time.Minute),
			UsedAt:    nil,
		}

		assert.True(t, token.IsValid(), "token should be valid")
	})

	t.Run("invalid token - expired", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now().Add(-10 * time.Minute),
			UsedAt:    nil,
		}

		assert.False(t, token.IsValid(), "token should be invalid (expired)")
	})

	t.Run("invalid token - used", func(t *testing.T) {
		now := time.Now()
		token := &Token{
			ExpiresAt: time.Now().Add(10 * time.Minute),
			UsedAt:    &now,
		}

		assert.False(t, token.IsValid(), "token should be invalid (used)")
	})

	t.Run("invalid token - expired and used", func(t *testing.T) {
		now := time.Now()
		token := &Token{
			ExpiresAt: time.Now().Add(-10 * time.Minute),
			UsedAt:    &now,
		}

		assert.False(t, token.IsValid(), "token should be invalid (expired and used)")
	})
}

func TestToken_MarkAsUsed(t *testing.T) {
	t.Run("marks unused token as used", func(t *testing.T) {
		token := &Token{
			UsedAt: nil,
		}

		assert.False(t, token.IsUsed(), "token should not be used initially")

		beforeMark := time.Now()
		token.MarkAsUsed()
		afterMark := time.Now()

		assert.True(t, token.IsUsed(), "token should be used after marking")
		assert.NotNil(t, token.UsedAt, "UsedAt should be set")
		assert.True(t, token.UsedAt.After(beforeMark) || token.UsedAt.Equal(beforeMark),
			"UsedAt should be after or equal to beforeMark")
		assert.True(t, token.UsedAt.Before(afterMark) || token.UsedAt.Equal(afterMark),
			"UsedAt should be before or equal to afterMark")
	})

	t.Run("marks already used token again", func(t *testing.T) {
		firstUse := time.Now().Add(-1 * time.Hour)
		token := &Token{
			UsedAt: &firstUse,
		}

		assert.True(t, token.IsUsed(), "token should already be used")

		token.MarkAsUsed()

		assert.True(t, token.IsUsed(), "token should still be used")
		assert.NotEqual(t, &firstUse, token.UsedAt, "UsedAt should be updated")
		assert.True(t, token.UsedAt.After(firstUse), "UsedAt should be more recent")
	})
}

func TestToken_TimeUntilExpiry(t *testing.T) {
	t.Run("token not expired", func(t *testing.T) {
		duration := 30 * time.Minute
		token := &Token{
			ExpiresAt: time.Now().Add(duration),
		}

		timeUntil := token.TimeUntilExpiry()

		// Allow for small timing differences
		assert.InDelta(t, duration.Seconds(), timeUntil.Seconds(), 1.0,
			"time until expiry should be approximately 30 minutes")
		assert.Greater(t, timeUntil, time.Duration(0), "time until expiry should be positive")
	})

	t.Run("token expired", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now().Add(-10 * time.Minute),
		}

		timeUntil := token.TimeUntilExpiry()

		assert.Equal(t, time.Duration(0), timeUntil, "expired token should return 0")
	})

	t.Run("token expires in 1 second", func(t *testing.T) {
		token := &Token{
			ExpiresAt: time.Now().Add(1 * time.Second),
		}

		timeUntil := token.TimeUntilExpiry()

		assert.Greater(t, timeUntil, time.Duration(0), "should have time remaining")
		assert.LessOrEqual(t, timeUntil, 1*time.Second, "should be less than or equal to 1 second")
	})
}

func TestToken_CompleteLifecycle(t *testing.T) {
	t.Run("complete token lifecycle", func(t *testing.T) {
		// Create a fresh token
		token := &Token{
			ID:        "test-id",
			Value:     "test-token-value",
			UserID:    "user-123",
			Purpose:   PurposeLogin,
			Email:     "test@example.com",
			ExpiresAt: time.Now().Add(15 * time.Minute),
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}

		// Initially valid
		assert.True(t, token.IsValid(), "fresh token should be valid")
		assert.False(t, token.IsExpired(), "fresh token should not be expired")
		assert.False(t, token.IsUsed(), "fresh token should not be used")

		// Mark as used
		token.MarkAsUsed()

		// Now invalid because used
		assert.False(t, token.IsValid(), "used token should be invalid")
		assert.True(t, token.IsUsed(), "token should be used")
		assert.False(t, token.IsExpired(), "token should still not be expired")
	})

	t.Run("token expiration lifecycle", func(t *testing.T) {
		// Create token that expires in 50ms
		token := &Token{
			ExpiresAt: time.Now().Add(50 * time.Millisecond),
			UsedAt:    nil,
		}

		// Initially valid
		assert.True(t, token.IsValid(), "token should initially be valid")
		assert.Greater(t, token.TimeUntilExpiry(), time.Duration(0), "should have time until expiry")

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Now invalid because expired
		assert.False(t, token.IsValid(), "expired token should be invalid")
		assert.True(t, token.IsExpired(), "token should be expired")
		assert.Equal(t, time.Duration(0), token.TimeUntilExpiry(), "expired token should return 0")
	})
}

func TestErrors(t *testing.T) {
	t.Run("ErrInvalidTTL", func(t *testing.T) {
		err := ErrInvalidTTL{TTL: -5 * time.Minute}
		assert.Contains(t, err.Error(), "invalid TTL")
		assert.Contains(t, err.Error(), "-5m")
	})

	t.Run("ErrTokenGeneration", func(t *testing.T) {
		innerErr := assert.AnError
		err := ErrTokenGeneration{Err: innerErr}

		assert.Contains(t, err.Error(), "token generation failed")
		assert.Equal(t, innerErr, err.Unwrap())
	})

	t.Run("ErrInvalidPurpose", func(t *testing.T) {
		err := ErrInvalidPurpose{Purpose: "invalid_purpose"}
		assert.Contains(t, err.Error(), "invalid token purpose")
		assert.Contains(t, err.Error(), "invalid_purpose")
	})

	t.Run("ErrTokenExpired", func(t *testing.T) {
		expiredAt := time.Now().Add(-1 * time.Hour)
		err := ErrTokenExpired{
			TokenID:   "token-123",
			ExpiresAt: expiredAt,
		}
		assert.Contains(t, err.Error(), "token-123")
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("ErrTokenAlreadyUsed", func(t *testing.T) {
		usedAt := time.Now().Add(-30 * time.Minute)
		err := ErrTokenAlreadyUsed{
			TokenID: "token-456",
			UsedAt:  usedAt,
		}
		assert.Contains(t, err.Error(), "token-456")
		assert.Contains(t, err.Error(), "already used")
	})

	t.Run("ErrTooManyTokens", func(t *testing.T) {
		err := ErrTooManyTokens{
			UserID:  "user-789",
			Purpose: PurposeLogin,
			Count:   10,
			Limit:   5,
		}
		assert.Contains(t, err.Error(), "too many")
		assert.Contains(t, err.Error(), "user-789")
		assert.Contains(t, err.Error(), "login")
		assert.Contains(t, err.Error(), "10/5")
	})

	t.Run("global errors", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidUserID)
		assert.NotNil(t, ErrInvalidEmail)
		assert.NotNil(t, ErrInvalidBaseURL)
		assert.NotNil(t, ErrTokenNotFound)

		assert.Contains(t, ErrInvalidUserID.Error(), "user ID")
		assert.Contains(t, ErrInvalidEmail.Error(), "email")
		assert.Contains(t, ErrInvalidBaseURL.Error(), "base URL")
		assert.Contains(t, ErrTokenNotFound.Error(), "not found")
	})
}
