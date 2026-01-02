package domain

import (
	"context"
	"time"
)

// ============================================================================
// User Repository
// ============================================================================

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	// Save persists a user (create or update).
	Save(ctx context.Context, user *User) error

	// FindByID finds a user by ID.
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByDID finds a user by their primary DID.
	FindByDID(ctx context.Context, did string) (*User, error)

	// FindByWallet finds a user by a linked wallet address.
	FindByWallet(ctx context.Context, address string, chain Chain) (*User, error)

	// FindByEmail finds a user by a linked email.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// ExistsByDID checks if a user with the given DID exists.
	ExistsByDID(ctx context.Context, did string) (bool, error)

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByWallet checks if a user with the given wallet exists.
	ExistsByWallet(ctx context.Context, address string, chain Chain) (bool, error)

	// Delete removes a user.
	Delete(ctx context.Context, id string) error

	// List returns users with pagination.
	List(ctx context.Context, opts UserListOptions) ([]*User, int, error)
}

// UserListOptions defines options for listing users.
type UserListOptions struct {
	Limit  int
	Offset int

	// Filters
	Status *UserStatus

	// Sorting
	SortBy    string // "created_at", "updated_at", "last_login_at"
	SortOrder string // "asc", "desc"
}

// DefaultUserListOptions returns default list options.
func DefaultUserListOptions() UserListOptions {
	return UserListOptions{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// ============================================================================
// Session Repository
// ============================================================================

// SessionRepository defines the interface for session persistence.
type SessionRepository interface {
	// Save persists a session (create or update).
	Save(ctx context.Context, session *Session) error

	// FindByID finds a session by ID.
	FindByID(ctx context.Context, id string) (*Session, error)

	// FindByTokenHash finds a session by token hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)

	// FindByUserID finds all sessions for a user.
	FindByUserID(ctx context.Context, userID string) ([]*Session, error)

	// FindActiveByUserID finds all active sessions for a user.
	FindActiveByUserID(ctx context.Context, userID string) ([]*Session, error)

	// Delete removes a session.
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpired removes all expired sessions.
	DeleteExpired(ctx context.Context) (int64, error)

	// CountByUserID counts sessions for a user.
	CountByUserID(ctx context.Context, userID string) (int, error)

	// CountActiveByUserID counts active sessions for a user.
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
}

// ============================================================================
// Connection Repository
// ============================================================================

// ConnectionRepository defines the interface for connection persistence.
type ConnectionRepository interface {
	// Save persists a connection (create or update).
	Save(ctx context.Context, connection *Connection) error

	// FindByID finds a connection by ID.
	FindByID(ctx context.Context, id string) (*Connection, error)

	// FindByUserAndProvider finds a connection by user and provider.
	FindByUserAndProvider(ctx context.Context, userID string, provider OAuthProvider) (*Connection, error)

	// FindByProviderUserID finds a connection by provider user ID.
	FindByProviderUserID(ctx context.Context, provider OAuthProvider, providerUserID string) (*Connection, error)

	// FindByUserID finds all connections for a user.
	FindByUserID(ctx context.Context, userID string) ([]*Connection, error)

	// FindActiveByUserID finds all active connections for a user.
	FindActiveByUserID(ctx context.Context, userID string) ([]*Connection, error)

	// Delete removes a connection.
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all connections for a user.
	DeleteByUserID(ctx context.Context, userID string) error

	// ExistsByUserAndProvider checks if a connection exists.
	ExistsByUserAndProvider(ctx context.Context, userID string, provider OAuthProvider) (bool, error)

	// FindStale finds connections that need syncing.
	FindStale(ctx context.Context, staleDuration time.Duration, limit int) ([]*Connection, error)
}

// ============================================================================
// API Key Repository
// ============================================================================

// APIKeyRepository defines the interface for API key persistence.
type APIKeyRepository interface {
	// Save persists an API key (create or update).
	Save(ctx context.Context, apiKey *APIKey) error

	// FindByID finds an API key by ID.
	FindByID(ctx context.Context, id string) (*APIKey, error)

	// FindByKeyHash finds an API key by key hash.
	FindByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)

	// FindByKeyPrefix finds API keys by prefix (for listing).
	FindByKeyPrefix(ctx context.Context, prefix string) ([]*APIKey, error)

	// FindByUserID finds all API keys for a user.
	FindByUserID(ctx context.Context, userID string) ([]*APIKey, error)

	// FindActiveByUserID finds all active API keys for a user.
	FindActiveByUserID(ctx context.Context, userID string) ([]*APIKey, error)

	// Delete removes an API key.
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all API keys for a user.
	DeleteByUserID(ctx context.Context, userID string) error

	// CountByUserID counts API keys for a user.
	CountByUserID(ctx context.Context, userID string) (int, error)

	// CountActiveByUserID counts active API keys for a user.
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
}

// ============================================================================
// Challenge Repository
// ============================================================================

// ChallengeRepository defines the interface for auth challenge persistence.
// Challenges are short-lived, typically stored in Redis.
type ChallengeRepository interface {
	// Save stores a challenge with TTL.
	Save(ctx context.Context, challenge *Challenge) error

	// FindByNonce finds a challenge by nonce.
	FindByNonce(ctx context.Context, nonce string) (*Challenge, error)

	// Delete removes a challenge.
	Delete(ctx context.Context, nonce string) error

	// DeleteExpired removes all expired challenges.
	DeleteExpired(ctx context.Context) error
}

// ============================================================================
// Token Repository
// ============================================================================

// TokenRepository defines the interface for refresh token persistence.
// May be backed by Redis for fast access.
type TokenRepository interface {
	// Save stores a refresh token.
	Save(ctx context.Context, token *RefreshToken) error

	// FindByTokenHash finds a refresh token by hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// FindByUserID finds all refresh tokens for a user.
	FindByUserID(ctx context.Context, userID string) ([]*RefreshToken, error)

	// Delete removes a refresh token.
	Delete(ctx context.Context, id string) error

	// DeleteByTokenHash removes a refresh token by hash.
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// DeleteByUserID removes all refresh tokens for a user.
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpired removes all expired refresh tokens.
	DeleteExpired(ctx context.Context) (int64, error)
}

// RefreshToken represents a refresh token entity.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	SessionID string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
	RevokedAt time.Time
}

// NewRefreshToken creates a new refresh token.
func NewRefreshToken(id, userID, tokenHash, sessionID string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}
}

// IsExpired returns true if the token has expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid returns true if the token is valid.
func (t *RefreshToken) IsValid() bool {
	return !t.Revoked && !t.IsExpired()
}

// Revoke revokes the token.
func (t *RefreshToken) Revoke() {
	t.Revoked = true
	t.RevokedAt = time.Now()
}

// ============================================================================
// Unit of Work (Optional)
// ============================================================================

// UnitOfWork provides transactional operations across repositories.
type UnitOfWork interface {
	// Users returns the user repository.
	Users() UserRepository

	// Sessions returns the session repository.
	Sessions() SessionRepository

	// Connections returns the connection repository.
	Connections() ConnectionRepository

	// APIKeys returns the API key repository.
	APIKeys() APIKeyRepository

	// Commit commits the transaction.
	Commit() error

	// Rollback rolls back the transaction.
	Rollback() error
}

// UnitOfWorkFactory creates units of work.
type UnitOfWorkFactory interface {
	// Begin starts a new unit of work.
	Begin(ctx context.Context) (UnitOfWork, error)
}
