package domain

import (
	"context"
)

// ============================================================================
// Event Subscriber Port
// ============================================================================

// DomainEvent represents an event from another bounded context.
// This is a simplified interface that Ledger expects from incoming events.
type DomainEvent interface {
	// EventType returns the event type (e.g., "credential.issued").
	EventType() string

	// OccurredAt returns when the event occurred as Unix timestamp (milliseconds).
	OccurredAt() int64

	// AggregateID returns the ID of the aggregate that emitted the event.
	AggregateID() string

	// AggregateType returns the type of aggregate (e.g., "Credential", "User").
	AggregateType() string
}

// EventHandler processes a domain event.
type EventHandler func(ctx context.Context, event DomainEvent) error

// EventSubscriber defines the interface for subscribing to domain events.
// This is a port — the actual implementation will be in infrastructure.
type EventSubscriber interface {
	// Subscribe registers a handler for the given event type patterns.
	// Patterns support wildcards (e.g., "credential.*" matches all credential events).
	Subscribe(ctx context.Context, patterns []string, handler EventHandler) error

	// Unsubscribe removes a subscription.
	Unsubscribe(ctx context.Context, patterns []string) error
}

// ============================================================================
// Event Metadata Extractor
// ============================================================================

// EventMetadata contains metadata extracted from a domain event.
type EventMetadata struct {
	CorrelationID string
	CausationID   string
	UserID        string
	TenantID      string
	TraceID       string
	SpanID        string
	Custom        map[string]any
}

// MetadataExtractor extracts metadata from domain events.
// Different event implementations may store metadata differently.
type MetadataExtractor interface {
	// Extract retrieves metadata from a domain event.
	Extract(event DomainEvent) EventMetadata
}
