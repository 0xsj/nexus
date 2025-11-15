package magic

import (
	"context"

	"github.com/0xsj/result"
)

// Store defines how magic link tokens are persisted.
// Implementation will be Redis-backed for transient storage with automatic TTL.
type Store interface {
	// Save stores a token with automatic TTL based on ExpiresAt.
	// Returns the saved token with any server-assigned fields.
	Save(ctx context.Context, token *Token) result.Result[*Token]

	// GetByValue retrieves a token by its value and purpose.
	// Returns ErrTokenNotFound if token doesn't exist.
	GetByValue(ctx context.Context, value string, purpose Purpose) result.Result[*Token]

	// GetByID retrieves a token by its ID.
	// Returns ErrTokenNotFound if token doesn't exist.
	GetByID(ctx context.Context, tokenID string) result.Result[*Token]

	// MarkAsUsed marks a token as used, preventing reuse.
	// This is idempotent - calling multiple times is safe.
	MarkAsUsed(ctx context.Context, tokenID string) result.Result[struct{}]

	// Delete removes a token from storage.
	// This is idempotent - deleting non-existent token returns success.
	Delete(ctx context.Context, tokenID string) result.Result[struct{}]

	// DeleteByValue removes a token by its value.
	// This is idempotent - deleting non-existent token returns success.
	DeleteByValue(ctx context.Context, value string, purpose Purpose) result.Result[struct{}]

	// DeleteExpired removes all expired tokens (cleanup job).
	// Returns the number of tokens deleted.
	DeleteExpired(ctx context.Context) result.Result[int]

	// DeleteAllForUser removes all tokens for a specific user.
	// Useful when user changes password or logs out from all devices.
	// Returns the number of tokens deleted.
	DeleteAllForUser(ctx context.Context, userID string) result.Result[int]

	// CountByUser counts active (not expired, not used) tokens for a user and purpose.
	// Useful for rate limiting (e.g., max 3 login tokens per user).
	CountByUser(ctx context.Context, userID string, purpose Purpose) result.Result[int]

	// ListByUser retrieves all active tokens for a user and purpose.
	// Ordered by creation time (newest first).
	ListByUser(ctx context.Context, userID string, purpose Purpose) result.Result[[]*Token]
}
