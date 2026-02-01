package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher defines the interface for publishing domain events.
// This is an outbound port — implementations handle delivery to message brokers,
// event buses, or other bounded contexts (e.g., Ledger for audit trails).
//
// Schema emits events such as:
//   - Schema.Registered
//   - Schema.VersionAdded
//   - Schema.Deprecated
//   - Schema.Activated
//   - Schema.MetadataUpdated
type EventPublisher interface {
	// Publish publishes one or more domain events synchronously.
	// Returns an error if any event fails to publish.
	// Use for critical events that must be confirmed before proceeding.
	Publish(ctx context.Context, events ...eventsourcing.Event) error

	// PublishAsync publishes events asynchronously.
	// Returns immediately without waiting for confirmation.
	// Use for non-critical events where eventual delivery is acceptable.
	PublishAsync(ctx context.Context, events ...eventsourcing.Event) error
}

// ============================================================================
// Issuer Reader Port (Cross-Context)
// ============================================================================

// IssuerInfo contains information about an issuer from the Organization context.
type IssuerInfo struct {
	// ID is the issuer's unique identifier.
	ID types.ID

	// Name is the issuer's display name.
	Name string

	// Active indicates whether the issuer is currently active.
	Active bool
}

// IssuerReader provides read access to the Organization context.
// Used to validate issuer existence when registering custom schemas.
type IssuerReader interface {
	// GetIssuer retrieves issuer information by ID.
	// Returns ErrIssuerNotFound if the issuer does not exist.
	GetIssuer(ctx context.Context, issuerID types.ID) (*IssuerInfo, error)

	// IssuerExists checks if an issuer exists and is active.
	IssuerExists(ctx context.Context, issuerID types.ID) (bool, error)
}

// ============================================================================
// Schema Lookup Port
// ============================================================================

// SchemaLookup provides read-optimized queries for validation.
// Used by command handlers to check uniqueness constraints
// without loading full aggregates.
type SchemaLookup interface {
	// ExistsByType checks if a schema with the given type exists.
	ExistsByType(ctx context.Context, schemaType string) (bool, error)

	// GetIDByType returns the schema ID for a given type, if it exists.
	GetIDByType(ctx context.Context, schemaType string) (types.ID, error)
}

// ============================================================================
// Null Implementations (for testing)
// ============================================================================

// NullEventPublisher is a no-op implementation of EventPublisher.
type NullEventPublisher struct{}

// NewNullEventPublisher creates a new NullEventPublisher.
func NewNullEventPublisher() *NullEventPublisher {
	return &NullEventPublisher{}
}

// Publish does nothing and returns nil.
func (p *NullEventPublisher) Publish(ctx context.Context, events ...eventsourcing.Event) error {
	return nil
}

// PublishAsync does nothing and returns nil.
func (p *NullEventPublisher) PublishAsync(ctx context.Context, events ...eventsourcing.Event) error {
	return nil
}

// Ensure NullEventPublisher implements EventPublisher.
var _ EventPublisher = (*NullEventPublisher)(nil)

// NullIssuerReader is a no-op implementation that always returns not found.
type NullIssuerReader struct{}

// NewNullIssuerReader creates a new NullIssuerReader.
func NewNullIssuerReader() *NullIssuerReader {
	return &NullIssuerReader{}
}

// GetIssuer always returns ErrIssuerNotFound.
func (r *NullIssuerReader) GetIssuer(ctx context.Context, issuerID types.ID) (*IssuerInfo, error) {
	return nil, ErrIssuerNotFound
}

// IssuerExists always returns false.
func (r *NullIssuerReader) IssuerExists(ctx context.Context, issuerID types.ID) (bool, error) {
	return false, nil
}

// Ensure NullIssuerReader implements IssuerReader.
var _ IssuerReader = (*NullIssuerReader)(nil)
