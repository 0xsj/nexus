package domain

import (
	"context"
)

// ============================================================================
// Profile Repository
// ============================================================================

// ProfileRepository defines the persistence operations for Profile aggregates.
type ProfileRepository interface {
	// Save persists a profile aggregate (appends new events).
	Save(ctx context.Context, profile *Profile) error

	// Get retrieves a profile by ID (replays events to rebuild state).
	Get(ctx context.Context, id ProfileID) (*Profile, error)

	// Exists checks if a profile with the given ID exists.
	Exists(ctx context.Context, id ProfileID) (bool, error)
}
