package domain

import (
	"context"
	"time"

	"github.com/0xsj/result"
)

// SessionRepository defines operations for session persistence.
type SessionRepository interface {
	// Create persists a new session.
	Create(ctx context.Context, session *Session) result.Result[*Session]

	// FindByID retrieves a session by its ID.
	FindByID(ctx context.Context, id string) result.Result[*Session]

	// FindByUserID retrieves all sessions for a user.
	FindByUserID(ctx context.Context, userID string) result.Result[[]*Session]

	// Update updates an existing session.
	Update(ctx context.Context, session *Session) result.Result[*Session]

	// Delete deletes a session by ID.
	Delete(ctx context.Context, id string) result.Result[struct{}]

	// DeleteByUserID deletes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID string) result.Result[struct{}]

	// DeleteExpired deletes all expired sessions.
	DeleteExpired(ctx context.Context) result.Result[int64]

	// UpdateLastActivity updates the last activity timestamp for a session.
	UpdateLastActivity(ctx context.Context, id string) result.Result[struct{}]
}

// MagicLinkRepository defines operations for magic link persistence.
// Note: This is typically stored in Redis, not PostgreSQL.
type MagicLinkRepository interface {
	// Store stores a magic link token.
	Store(ctx context.Context, link *MagicLink) result.Result[*MagicLink]

	// FindByToken retrieves a magic link by its token.
	FindByToken(ctx context.Context, token string) result.Result[*MagicLink]

	// MarkAsUsed marks a magic link as used.
	MarkAsUsed(ctx context.Context, token string) result.Result[struct{}]

	// Delete deletes a magic link by token.
	Delete(ctx context.Context, token string) result.Result[struct{}]

	// DeleteByEmail deletes all magic links for an email.
	DeleteByEmail(ctx context.Context, email string) result.Result[struct{}]
}

// OAuthRepository defines operations for OAuth connection persistence.
// Note: Requires a migration to be created for oauth_connections table.
type OAuthRepository interface {
	// Create persists a new OAuth connection.
	Create(ctx context.Context, conn *OAuthConnection) result.Result[*OAuthConnection]

	// FindByID retrieves an OAuth connection by its ID.
	FindByID(ctx context.Context, id string) result.Result[*OAuthConnection]

	// FindByUserID retrieves all OAuth connections for a user.
	FindByUserID(ctx context.Context, userID string) result.Result[[]*OAuthConnection]

	// FindByProvider retrieves an OAuth connection by provider and provider user ID.
	FindByProvider(ctx context.Context, provider OAuthProvider, providerUserID string) result.Result[*OAuthConnection]

	// Update updates an existing OAuth connection.
	Update(ctx context.Context, conn *OAuthConnection) result.Result[*OAuthConnection]

	// Delete deletes an OAuth connection by ID.
	Delete(ctx context.Context, id string) result.Result[struct{}]
}

// OAuthStateRepository defines operations for OAuth state persistence.
// Note: This is typically stored in Redis with TTL.
type OAuthStateRepository interface {
	// Store stores an OAuth state.
	Store(ctx context.Context, state *OAuthState, ttl time.Duration) result.Result[*OAuthState]

	// FindByState retrieves an OAuth state by its state value.
	FindByState(ctx context.Context, state string) result.Result[*OAuthState]

	// Delete deletes an OAuth state.
	Delete(ctx context.Context, state string) result.Result[struct{}]
}

// RefreshTokenRepository defines operations for refresh token persistence.
// Note: This is typically stored in Redis with TTL.
type RefreshTokenRepository interface {
	// Store stores a refresh token for a session.
	Store(ctx context.Context, sessionID, refreshToken string, ttl time.Duration) result.Result[struct{}]

	// FindByToken retrieves a session ID by refresh token.
	FindByToken(ctx context.Context, refreshToken string) result.Result[string]

	// Delete deletes a refresh token.
	Delete(ctx context.Context, refreshToken string) result.Result[struct{}]

	// DeleteBySessionID deletes all refresh tokens for a session.
	DeleteBySessionID(ctx context.Context, sessionID string) result.Result[struct{}]
}
