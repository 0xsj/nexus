package session

import (
	"context"
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

// Manager handles session lifecycle operations.
type Manager interface {
	// Create creates a new session for a user.
	Create(ctx context.Context, userID, refreshToken, ipAddress, userAgent string) result.Result[*Session]

	// Get retrieves a session by ID.
	Get(ctx context.Context, sessionID string) result.Result[*Session]

	// GetByRefreshToken retrieves a session by refresh token.
	GetByRefreshToken(ctx context.Context, refreshToken string) result.Result[*Session]

	// Refresh updates a session with a new refresh token and extends expiry.
	Refresh(ctx context.Context, sessionID, newRefreshToken string) result.Result[*Session]

	// Delete deletes a specific session.
	Delete(ctx context.Context, sessionID string) result.Result[struct{}]

	// DeleteByRefreshToken deletes a session by refresh token.
	DeleteByRefreshToken(ctx context.Context, refreshToken string) result.Result[struct{}]

	// DeleteAllForUser deletes all sessions for a user (logout from all devices).
	DeleteAllForUser(ctx context.Context, userID string) result.Result[int]

	// ListUserSessions lists all active sessions for a user.
	ListUserSessions(ctx context.Context, userID string) result.Result[[]*Session]

	// Validate checks if a session is valid (exists and not expired).
	Validate(ctx context.Context, sessionID string) result.Result[*Session]
}

// Config holds session manager configuration.
type Config struct {
	// SessionTTL is the time-to-live for sessions (default: 7 days)
	SessionTTL time.Duration

	// MaxSessionsPerUser is the maximum number of concurrent sessions per user (default: 5)
	// Set to 0 for unlimited
	MaxSessionsPerUser int
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		SessionTTL:         7 * 24 * time.Hour, // 7 days
		MaxSessionsPerUser: 5,
	}
}

// WithSessionTTL sets the session TTL and returns the config (builder pattern).
func (c Config) WithSessionTTL(ttl time.Duration) Config {
	c.SessionTTL = ttl
	return c
}

// WithMaxSessionsPerUser sets the max sessions per user and returns the config (builder pattern).
func (c Config) WithMaxSessionsPerUser(max int) Config {
	c.MaxSessionsPerUser = max
	return c
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.SessionTTL <= 0 {
		return ErrInvalidTTL{TTL: c.SessionTTL}
	}

	if c.MaxSessionsPerUser < 0 {
		return ErrInvalidMaxSessions{Max: c.MaxSessionsPerUser}
	}

	return nil
}

type manager struct {
	config Config
	store  Store
}

// NewManager creates a new session manager.
func NewManager(config Config, store Store) result.Result[Manager] {
	// Validate config
	if err := config.Validate(); err != nil {
		return result.Err[Manager](err)
	}

	if store == nil {
		return result.Err[Manager](ErrNilStore)
	}

	return result.Ok[Manager](&manager{
		config: config,
		store:  store,
	})
}

// Create creates a new session for a user.
func (m *manager) Create(ctx context.Context, userID, refreshToken, ipAddress, userAgent string) result.Result[*Session] {
	// Validate inputs
	if userID == "" {
		return result.Err[*Session](ErrInvalidUserID)
	}

	if refreshToken == "" {
		return result.Err[*Session](ErrInvalidToken)
	}

	// Check max sessions limit
	if m.config.MaxSessionsPerUser > 0 {
		countResult := m.store.CountByUser(ctx, userID)
		if countResult.IsErr() {
			return result.Err[*Session](countResult.UnwrapErr())
		}

		count := countResult.Unwrap()
		if count >= m.config.MaxSessionsPerUser {
			// Delete oldest session to make room
			deleteResult := m.deleteOldestSession(ctx, userID)
			if deleteResult.IsErr() {
				return result.Err[*Session](deleteResult.UnwrapErr())
			}
		}
	}

	// Create session
	now := time.Now()
	session := &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    now.Add(m.config.SessionTTL),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return m.store.Create(ctx, session)
}

// Get retrieves a session by ID.
func (m *manager) Get(ctx context.Context, sessionID string) result.Result[*Session] {
	if sessionID == "" {
		return result.Err[*Session](ErrEmptySessionID)
	}

	return m.store.GetByID(ctx, sessionID)
}

// GetByRefreshToken retrieves a session by refresh token.
func (m *manager) GetByRefreshToken(ctx context.Context, refreshToken string) result.Result[*Session] {
	if refreshToken == "" {
		return result.Err[*Session](ErrInvalidToken)
	}

	return m.store.GetByRefreshToken(ctx, refreshToken)
}

// Refresh updates a session with a new refresh token and extends expiry.
func (m *manager) Refresh(ctx context.Context, sessionID, newRefreshToken string) result.Result[*Session] {
	if sessionID == "" {
		return result.Err[*Session](ErrEmptySessionID)
	}

	if newRefreshToken == "" {
		return result.Err[*Session](ErrInvalidToken)
	}

	// Get existing session
	sessionResult := m.store.GetByID(ctx, sessionID)
	if sessionResult.IsErr() {
		return sessionResult
	}

	session := sessionResult.Unwrap()

	// Check if expired
	if session.IsExpired() {
		return result.Err[*Session](ErrSessionExpired)
	}

	// Update session
	session.Refresh(newRefreshToken, m.config.SessionTTL)

	return m.store.Update(ctx, session)
}

// Delete deletes a specific session.
func (m *manager) Delete(ctx context.Context, sessionID string) result.Result[struct{}] {
	if sessionID == "" {
		return result.Err[struct{}](ErrEmptySessionID)
	}

	return m.store.Delete(ctx, sessionID)
}

// DeleteByRefreshToken deletes a session by refresh token.
func (m *manager) DeleteByRefreshToken(ctx context.Context, refreshToken string) result.Result[struct{}] {
	if refreshToken == "" {
		return result.Err[struct{}](ErrInvalidToken)
	}

	return m.store.DeleteByRefreshToken(ctx, refreshToken)
}

// DeleteAllForUser deletes all sessions for a user.
func (m *manager) DeleteAllForUser(ctx context.Context, userID string) result.Result[int] {
	if userID == "" {
		return result.Err[int](ErrInvalidUserID)
	}

	return m.store.DeleteAllForUser(ctx, userID)
}

// ListUserSessions lists all active sessions for a user.
func (m *manager) ListUserSessions(ctx context.Context, userID string) result.Result[[]*Session] {
	if userID == "" {
		return result.Err[[]*Session](ErrInvalidUserID)
	}

	return m.store.ListByUser(ctx, userID)
}

// Validate checks if a session is valid (exists and not expired).
func (m *manager) Validate(ctx context.Context, sessionID string) result.Result[*Session] {
	sessionResult := m.Get(ctx, sessionID)
	if sessionResult.IsErr() {
		return sessionResult
	}

	session := sessionResult.Unwrap()

	// Check if expired
	if session.IsExpired() {
		return result.Err[*Session](ErrSessionExpired)
	}

	return result.Ok(session)
}

// deleteOldestSession deletes the oldest session for a user.
func (m *manager) deleteOldestSession(ctx context.Context, userID string) result.Result[struct{}] {
	sessionsResult := m.store.ListByUser(ctx, userID)
	if sessionsResult.IsErr() {
		return result.Err[struct{}](sessionsResult.UnwrapErr())
	}

	sessions := sessionsResult.Unwrap()
	if len(sessions) == 0 {
		return result.Ok(struct{}{})
	}

	// Find oldest session
	oldest := sessions[0]
	for _, session := range sessions {
		if session.CreatedAt.Before(oldest.CreatedAt) {
			oldest = session
		}
	}

	return m.store.Delete(ctx, oldest.ID)
}
