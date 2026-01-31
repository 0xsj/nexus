package domain

import (
	"context"
)

// ============================================================================
// User Repository
// ============================================================================

// UserRepository defines the persistence operations for User aggregates.
type UserRepository interface {
	// Save persists a user aggregate (appends new events).
	Save(ctx context.Context, user *User) error

	// Get retrieves a user by ID (replays events to rebuild state).
	Get(ctx context.Context, id UserID) (*User, error)

	// Exists checks if a user with the given ID exists.
	Exists(ctx context.Context, id UserID) (bool, error)
}

// ============================================================================
// Session Repository
// ============================================================================

// SessionRepository defines the persistence operations for Session aggregates.
type SessionRepository interface {
	// Save persists a session aggregate (appends new events).
	Save(ctx context.Context, session *Session) error

	// Get retrieves a session by ID (replays events to rebuild state).
	Get(ctx context.Context, id SessionID) (*Session, error)

	// Exists checks if a session with the given ID exists.
	Exists(ctx context.Context, id SessionID) (bool, error)
}

// ============================================================================
// Magic Link Repository
// ============================================================================

// MagicLinkRecord represents a stored magic link for persistence.
// Magic links are not event-sourced; they are short-lived tokens stored directly.
type MagicLinkRecord struct {
	TokenHash string
	Email     string
	ExpiresAt int64 // Unix timestamp
	Used      bool
	CreatedAt int64 // Unix timestamp
}

// MagicLinkRepository defines persistence operations for magic links.
type MagicLinkRepository interface {
	// Save stores a magic link record.
	Save(ctx context.Context, record *MagicLinkRecord) error

	// GetByTokenHash retrieves a magic link by its token hash.
	GetByTokenHash(ctx context.Context, tokenHash string) (*MagicLinkRecord, error)

	// MarkUsed marks a magic link as used.
	MarkUsed(ctx context.Context, tokenHash string) error

	// DeleteExpired removes expired magic links.
	DeleteExpired(ctx context.Context) (int64, error)
}

// ============================================================================
// OAuth State Repository
// ============================================================================

// OAuthStateRecord represents a stored OAuth state for CSRF protection.
// OAuth states are short-lived and not event-sourced.
type OAuthStateRecord struct {
	State       string
	Provider    string
	RedirectURL string
	ExpiresAt   int64 // Unix timestamp
	CreatedAt   int64 // Unix timestamp
}

// OAuthStateRepository defines persistence operations for OAuth states.
type OAuthStateRepository interface {
	// Save stores an OAuth state record.
	Save(ctx context.Context, record *OAuthStateRecord) error

	// GetByState retrieves an OAuth state by its state value.
	GetByState(ctx context.Context, state string) (*OAuthStateRecord, error)

	// Delete removes an OAuth state (after use or expiration).
	Delete(ctx context.Context, state string) error

	// DeleteExpired removes expired OAuth states.
	DeleteExpired(ctx context.Context) (int64, error)
}
