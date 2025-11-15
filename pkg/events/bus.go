package events

import (
	"context"
)

// EventBus publishes and subscribes to domain events.
// Implementations must be safe for concurrent use.
type EventBus interface {
	// Publish publishes an event to all subscribers.
	// Returns immediately after queuing the event.
	Publish(ctx context.Context, event Event) error

	// PublishBatch publishes multiple events atomically.
	// Either all events are published or none are.
	PublishBatch(ctx context.Context, events []Event) error

	// Subscribe registers a handler for events matching the pattern.
	// Pattern can be:
	//   - Exact match: "user.created"
	//   - Wildcard: "user.*" (all user events)
	//   - Multiple: "user.created,user.updated"
	//
	// Returns a Subscription that can be used to unsubscribe.
	Subscribe(ctx context.Context, pattern string, handler Handler) (Subscription, error)

	// SubscribeGroup registers a handler as part of a consumer group.
	// Only one handler in the group will receive each event.
	// Useful for load balancing event processing.
	SubscribeGroup(ctx context.Context, group, pattern string, handler Handler) (Subscription, error)

	// Close closes the event bus and all subscriptions.
	Close() error

	// Health checks if the event bus is healthy.
	Health(ctx context.Context) error
}

// Subscription represents an active event subscription.
type Subscription interface {
	// Unsubscribe stops receiving events and cleans up resources.
	Unsubscribe() error

	// IsActive returns true if the subscription is active.
	IsActive() bool
}

// PublishOptions configures event publishing behavior.
type PublishOptions struct {
	// Async publishes the event asynchronously without waiting for confirmation.
	Async bool

	// Headers are additional metadata to attach to the event.
	Headers map[string]string

	// RetryAttempts specifies how many times to retry on failure.
	RetryAttempts int
}

// SubscribeOptions configures event subscription behavior.
type SubscribeOptions struct {
	// BufferSize is the size of the event buffer for this subscription.
	BufferSize int

	// AutoAck automatically acknowledges events after successful processing.
	AutoAck bool

	// MaxConcurrent is the maximum number of events to process concurrently.
	MaxConcurrent int

	// RetryPolicy specifies how to handle failed event processing.
	RetryPolicy *RetryPolicy
}

// RetryPolicy defines how to retry failed event processing.
type RetryPolicy struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int

	// BackoffMultiplier multiplies the delay after each attempt.
	BackoffMultiplier float64

	// InitialInterval is the delay before the first retry.
	InitialInterval int64 // milliseconds

	// MaxInterval is the maximum delay between retries.
	MaxInterval int64 // milliseconds
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:       3,
		BackoffMultiplier: 2.0,
		InitialInterval:   100,  // 100ms
		MaxInterval:       5000, // 5s
	}
}
