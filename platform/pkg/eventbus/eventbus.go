package eventbus

import (
	"context"
	"io"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Publisher sends events to the bus.
type Publisher interface {
	// Publish sends events synchronously, blocking until all subscribers have processed them.
	Publish(ctx context.Context, events ...*eventsourcing.EventEnvelope) error

	// PublishAsync sends events asynchronously, returning immediately after dispatch.
	PublishAsync(ctx context.Context, events ...*eventsourcing.EventEnvelope) error
}

// Handler processes a single event envelope.
type Handler func(ctx context.Context, event *eventsourcing.EventEnvelope) error

// Subscription represents an active subscription that can be cancelled.
type Subscription interface {
	// Unsubscribe cancels this subscription.
	Unsubscribe() error

	// Topic returns the pattern this subscription is listening on.
	Topic() string
}

// Subscriber receives events from the bus.
type Subscriber interface {
	// Subscribe registers a handler for events matching the given pattern.
	// Pattern uses dot-delimited segments with '*' wildcard (e.g., "identity.*" matches "identity.user_created").
	Subscribe(ctx context.Context, pattern string, handler Handler, opts ...SubscribeOption) (Subscription, error)
}

// Middleware wraps a Handler, enabling cross-cutting concerns like logging and retries.
type Middleware func(Handler) Handler

// EventBus combines Publisher and Subscriber with lifecycle management.
type EventBus interface {
	Publisher
	Subscriber

	// Start initializes the event bus and begins processing events.
	Start(ctx context.Context) error

	io.Closer
}
