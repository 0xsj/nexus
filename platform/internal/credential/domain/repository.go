package domain

import (
	"context"
)

// ============================================================================
// Write Repository Interface
// ============================================================================

// CredentialRepository defines the interface for credential persistence (write side).
// This is used by command handlers to load and save aggregates.
type CredentialRepository interface {
	// Load loads a credential aggregate by ID.
	// Returns ErrAggregateNotFound if the credential does not exist.
	Load(ctx context.Context, id string) (*Credential, error)

	// Save saves a credential aggregate.
	// Persists all uncommitted events and clears the changes.
	// Returns ErrConcurrencyConflict if optimistic locking fails.
	Save(ctx context.Context, credential *Credential) error

	// Exists checks if a credential exists.
	Exists(ctx context.Context, id string) (bool, error)
}

// ============================================================================
// Aggregate Factory
// ============================================================================

// CredentialFactory creates new Credential aggregates.
// Implements eventsourcing.AggregateFactory.
type CredentialFactory struct{}

// NewCredentialFactory creates a new CredentialFactory.
func NewCredentialFactory() *CredentialFactory {
	return &CredentialFactory{}
}

// Create creates a new Credential aggregate with the given ID.
func (f *CredentialFactory) Create(aggregateID string) *Credential {
	return NewCredential(aggregateID)
}
