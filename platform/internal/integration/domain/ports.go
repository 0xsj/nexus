package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	// Publish publishes one or more domain events.
	Publish(ctx context.Context, events ...eventsourcing.Event) error
}

// ============================================================================
// Null Implementations
// ============================================================================

// NullEventPublisher is a no-op implementation of EventPublisher.
// Silently discards all events.
type NullEventPublisher struct{}

var _ EventPublisher = (*NullEventPublisher)(nil)

// Publish silently discards all events.
func (n *NullEventPublisher) Publish(_ context.Context, _ ...eventsourcing.Event) error {
	return nil
}
