package domain

import (
	"context"
)

// ============================================================================
// Issuer Repository
// ============================================================================

// IssuerRepository defines the persistence operations for Issuer aggregates.
type IssuerRepository interface {
	// Save persists an issuer aggregate (appends new events).
	Save(ctx context.Context, issuer *Issuer) error

	// Get retrieves an issuer by ID (replays events to rebuild state).
	Get(ctx context.Context, id IssuerID) (*Issuer, error)

	// Exists checks if an issuer with the given ID exists.
	Exists(ctx context.Context, id IssuerID) (bool, error)
}

// ============================================================================
// Template Repository
// ============================================================================

// TemplateRepository defines the persistence operations for Template aggregates.
type TemplateRepository interface {
	// Save persists a template aggregate (appends new events).
	Save(ctx context.Context, template *Template) error

	// Get retrieves a template by ID (replays events to rebuild state).
	Get(ctx context.Context, id TemplateID) (*Template, error)

	// Exists checks if a template with the given ID exists.
	Exists(ctx context.Context, id TemplateID) (bool, error)
}
