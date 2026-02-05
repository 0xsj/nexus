package domain

import (
	"context"
)

// ============================================================================
// Credential Repository
// ============================================================================

// CredentialRepository defines the persistence operations for Credential aggregates.
type CredentialRepository interface {
	// Save persists a credential aggregate (appends new events).
	Save(ctx context.Context, credential *Credential) error

	// Get retrieves a credential by ID (replays events to rebuild state).
	Get(ctx context.Context, id CredentialID) (*Credential, error)

	// Exists checks if a credential with the given ID exists.
	Exists(ctx context.Context, id CredentialID) (bool, error)
}
