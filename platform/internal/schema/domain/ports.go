package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
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
// Null Publisher (for testing/default)
// ============================================================================

// NullEventPublisher is a no-op implementation of EventPublisher.
// Useful for testing or when event publishing is not required.
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
