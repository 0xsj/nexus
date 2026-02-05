package domain

import (
	"context"
)

// ============================================================================
// Verification Repository
// ============================================================================

// VerificationRepository defines the persistence operations for Verification aggregates.
type VerificationRepository interface {
	// Save persists a verification aggregate (appends new events).
	Save(ctx context.Context, verification *Verification) error

	// Get retrieves a verification by ID (replays events to rebuild state).
	Get(ctx context.Context, id VerificationID) (*Verification, error)

	// Exists checks if a verification with the given ID exists.
	Exists(ctx context.Context, id VerificationID) (bool, error)
}
