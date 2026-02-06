package domain

import (
	"context"
)

// ============================================================================
// Integration Repository
// ============================================================================

// IntegrationRepository defines the persistence operations for Integration aggregates.
type IntegrationRepository interface {
	// Save persists an integration aggregate (appends new events).
	Save(ctx context.Context, integration *Integration) error

	// Get retrieves an integration by ID (replays events to rebuild state).
	Get(ctx context.Context, id IntegrationID) (*Integration, error)

	// Exists checks if an integration with the given ID exists.
	Exists(ctx context.Context, id IntegrationID) (bool, error)
}
