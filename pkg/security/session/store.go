package session

import (
	"context"
	"time"

	"github.com/0xsj/result"
)

// Store defines how sessions are persisted.
// Implementations will typically use Redis for fast access with automatic TTL.
type Store interface {
	// Create creates a new session.
	Create(ctx context.Context, session *Session) result.Result[*Session]

	// GetByID retrieves a session by its ID.
	GetByID(ctx context.Context, sessionID string) result.Result[*Session]

	// GetByRefreshToken retrieves a session by its refresh token.
	GetByRefreshToken(ctx context.Context, refreshToken string) result.Result[*Session]

	// Update updates an existing session.
	Update(ctx context.Context, session *Session) result.Result[*Session]

	// Delete deletes a session by ID.
	Delete(ctx context.Context, sessionID string) result.Result[struct{}]

	// DeleteByRefreshToken deletes a session by its refresh token.
	DeleteByRefreshToken(ctx context.Context, refreshToken string) result.Result[struct{}]

	// DeleteAllForUser deletes all sessions for a specific user.
	// Useful for "logout from all devices" functionality.
	DeleteAllForUser(ctx context.Context, userID string) result.Result[int]

	// ListByUser retrieves all active sessions for a user.
	// Ordered by creation time (newest first).
	ListByUser(ctx context.Context, userID string) result.Result[[]*Session]

	// CountByUser counts active sessions for a user.
	CountByUser(ctx context.Context, userID string) result.Result[int]

	// DeleteExpired removes all expired sessions (cleanup job).
	// Returns the number of sessions deleted.
	DeleteExpired(ctx context.Context) result.Result[int]

	// ExtendExpiry extends a session's expiry time.
	ExtendExpiry(ctx context.Context, sessionID string, duration time.Duration) result.Result[*Session]
}
