package domain

import (
	"context"
)

// OrganizationRepository defines the persistence operations for Organization aggregates.
type OrganizationRepository interface {
	// Save persists an organization aggregate (appends new events).
	Save(ctx context.Context, org *Organization) error

	// Get retrieves an organization by ID (replays events to rebuild state).
	Get(ctx context.Context, id OrganizationID) (*Organization, error)

	// Exists checks if an organization with the given ID exists.
	Exists(ctx context.Context, id OrganizationID) (bool, error)
}
