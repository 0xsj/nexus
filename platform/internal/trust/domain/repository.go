package domain

import (
	"context"
)

// ============================================================================
// Vouch Repository
// ============================================================================

// VouchRepository defines the persistence operations for Vouch aggregates.
type VouchRepository interface {
	// Save persists a vouch aggregate (appends new events).
	Save(ctx context.Context, vouch *Vouch) error

	// Get retrieves a vouch by ID (replays events to rebuild state).
	Get(ctx context.Context, id VouchID) (*Vouch, error)

	// Exists checks if a vouch with the given ID exists.
	Exists(ctx context.Context, id VouchID) (bool, error)
}
