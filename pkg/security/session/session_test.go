package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		assert.False(t, session.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		assert.True(t, session.IsExpired())
	})

	t.Run("expires exactly now", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now(),
		}

		time.Sleep(1 * time.Millisecond)

		assert.True(t, session.IsExpired())
	})
}

func TestSession_IsValid(t *testing.T) {
	t.Run("valid session", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		assert.True(t, session.IsValid())
	})

	t.Run("invalid session", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		assert.False(t, session.IsValid())
	})
}

func TestSession_TimeUntilExpiry(t *testing.T) {
	t.Run("time remaining", func(t *testing.T) {
		duration := 30 * time.Minute
		session := &Session{
			ExpiresAt: time.Now().Add(duration),
		}

		timeUntil := session.TimeUntilExpiry()

		assert.InDelta(t, duration.Seconds(), timeUntil.Seconds(), 1.0)
	})

	t.Run("already expired", func(t *testing.T) {
		session := &Session{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		timeUntil := session.TimeUntilExpiry()

		assert.Equal(t, time.Duration(0), timeUntil)
	})
}

func TestSession_UpdateRefreshToken(t *testing.T) {
	session := &Session{
		RefreshToken: "old-token",
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	}

	oldUpdatedAt := session.UpdatedAt
	newToken := "new-token"

	time.Sleep(10 * time.Millisecond)
	session.UpdateRefreshToken(newToken)

	assert.Equal(t, newToken, session.RefreshToken)
	assert.True(t, session.UpdatedAt.After(oldUpdatedAt))
}

func TestSession_Extend(t *testing.T) {
	session := &Session{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	oldExpiresAt := session.ExpiresAt
	oldUpdatedAt := session.UpdatedAt
	extension := 2 * time.Hour

	time.Sleep(10 * time.Millisecond)
	session.Extend(extension)

	assert.True(t, session.ExpiresAt.After(oldExpiresAt))
	assert.True(t, session.UpdatedAt.After(oldUpdatedAt))
	assert.InDelta(t, extension.Seconds(), time.Until(session.ExpiresAt).Seconds(), 1.0)
}

func TestSession_Refresh(t *testing.T) {
	session := &Session{
		RefreshToken: "old-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	}

	oldToken := session.RefreshToken
	oldExpiresAt := session.ExpiresAt
	oldUpdatedAt := session.UpdatedAt

	newToken := "new-token"
	extension := 2 * time.Hour

	time.Sleep(10 * time.Millisecond)
	session.Refresh(newToken, extension)

	assert.NotEqual(t, oldToken, session.RefreshToken)
	assert.Equal(t, newToken, session.RefreshToken)
	assert.True(t, session.ExpiresAt.After(oldExpiresAt))
	assert.True(t, session.UpdatedAt.After(oldUpdatedAt))
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultConfig()

		err := config.Validate()

		assert.NoError(t, err)
	})

	t.Run("invalid TTL", func(t *testing.T) {
		config := DefaultConfig().WithSessionTTL(0)

		err := config.Validate()

		assert.Error(t, err)
		_, ok := err.(ErrInvalidTTL)
		assert.True(t, ok)
	})

	t.Run("invalid max sessions", func(t *testing.T) {
		config := DefaultConfig().WithMaxSessionsPerUser(-1)

		err := config.Validate()

		assert.Error(t, err)
		_, ok := err.(ErrInvalidMaxSessions)
		assert.True(t, ok)
	})
}

func TestConfig_BuilderPattern(t *testing.T) {
	config := DefaultConfig().
		WithSessionTTL(14 * 24 * time.Hour).
		WithMaxSessionsPerUser(10)

	assert.Equal(t, 14*24*time.Hour, config.SessionTTL)
	assert.Equal(t, 10, config.MaxSessionsPerUser)
}
