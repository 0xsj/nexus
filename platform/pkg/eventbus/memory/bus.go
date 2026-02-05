package memory

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// Bus is an in-memory event bus suitable for development and testing.
// It uses mutex-protected slices for subscription management and supports
// dot-delimited wildcard pattern matching.
type Bus struct {
	config eventbus.Config
	logger log.Logger

	mu            sync.RWMutex
	subscriptions []*subscription
	closed        atomic.Bool
}

// subscription is an internal representation of an active subscription.
type subscription struct {
	pattern  string
	segments []string
	handler  eventbus.Handler
	opts     eventbus.SubscribeOptions
	active   atomic.Bool
}

// New creates a new in-memory event bus.
func New(config eventbus.Config) *Bus {
	logger := config.Logger
	if logger == nil {
		logger = log.New()
	}

	return &Bus{
		config:        config,
		logger:        logger.With(log.Component("eventbus.memory")),
		subscriptions: make([]*subscription, 0),
	}
}

// Start initializes the in-memory bus. For this implementation it's a no-op
// since no background goroutines or connections are needed.
func (b *Bus) Start(_ context.Context) error {
	if b.closed.Load() {
		return eventbus.ErrBusClosed("memory.Bus.Start")
	}
	b.logger.Info("in-memory event bus started")
	return nil
}

// Close shuts down the bus, preventing further publishes and unsubscribing all active subscriptions.
func (b *Bus) Close() error {
	if b.closed.Swap(true) {
		return nil // already closed
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, sub := range b.subscriptions {
		sub.active.Store(false)
	}
	b.subscriptions = nil

	b.logger.Info("in-memory event bus closed")
	return nil
}

// Publish sends events synchronously to all matching subscribers.
// Each matching handler is invoked in sequence; the first error stops propagation.
func (b *Bus) Publish(ctx context.Context, events ...*eventsourcing.EventEnvelope) error {
	if b.closed.Load() {
		return eventbus.ErrBusClosed("memory.Bus.Publish")
	}

	for _, event := range events {
		if err := b.dispatch(ctx, event); err != nil {
			return eventbus.ErrPublishFailed("memory.Bus.Publish", err)
		}
	}
	return nil
}

// PublishAsync sends events asynchronously. Each event is dispatched in its own goroutine.
func (b *Bus) PublishAsync(ctx context.Context, events ...*eventsourcing.EventEnvelope) error {
	if b.closed.Load() {
		return eventbus.ErrBusClosed("memory.Bus.PublishAsync")
	}

	for _, event := range events {
		go func() {
			if err := b.dispatch(ctx, event); err != nil {
				b.logger.Error("async dispatch failed",
					log.String("event_type", event.Type),
					log.Err(err),
				)
			}
		}()
	}
	return nil
}

// Subscribe registers a handler for events matching the given pattern.
func (b *Bus) Subscribe(_ context.Context, pattern string, handler eventbus.Handler, opts ...eventbus.SubscribeOption) (eventbus.Subscription, error) {
	if b.closed.Load() {
		return nil, eventbus.ErrBusClosed("memory.Bus.Subscribe")
	}

	options := eventbus.ApplyOptions(opts...)

	// Apply configured middleware chain to the handler.
	wrapped := handler
	for i := len(b.config.Middlewares) - 1; i >= 0; i-- {
		wrapped = b.config.Middlewares[i](wrapped)
	}

	sub := &subscription{
		pattern:  pattern,
		segments: strings.Split(pattern, "."),
		handler:  wrapped,
		opts:     options,
	}
	sub.active.Store(true)

	b.mu.Lock()
	b.subscriptions = append(b.subscriptions, sub)
	b.mu.Unlock()

	b.logger.Debug("subscription added", log.String("pattern", pattern))

	return &memorySubscription{bus: b, sub: sub}, nil
}

// dispatch fans out a single event to all matching active subscribers.
func (b *Bus) dispatch(ctx context.Context, event *eventsourcing.EventEnvelope) error {
	b.mu.RLock()
	subs := make([]*subscription, len(b.subscriptions))
	copy(subs, b.subscriptions)
	b.mu.RUnlock()

	for _, sub := range subs {
		if !sub.active.Load() {
			continue
		}
		if !matchPattern(sub.segments, strings.Split(event.Type, ".")) {
			continue
		}
		if err := sub.handler(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// unsubscribe removes a subscription from the bus.
func (b *Bus) unsubscribe(target *subscription) {
	target.active.Store(false)

	b.mu.Lock()
	defer b.mu.Unlock()

	for i, sub := range b.subscriptions {
		if sub == target {
			b.subscriptions = append(b.subscriptions[:i], b.subscriptions[i+1:]...)
			return
		}
	}
}

// matchPattern matches event type segments against a subscription pattern.
// '*' matches exactly one segment. '>' matches one or more trailing segments.
func matchPattern(pattern, subject []string) bool {
	if len(pattern) == 0 && len(subject) == 0 {
		return true
	}
	if len(pattern) == 0 || len(subject) == 0 {
		// A trailing '>' in pattern matches remaining segments.
		if len(pattern) == 1 && pattern[0] == ">" {
			return len(subject) > 0
		}
		return false
	}

	if pattern[0] == ">" {
		return true // matches rest
	}

	if pattern[0] == "*" || pattern[0] == subject[0] {
		return matchPattern(pattern[1:], subject[1:])
	}

	return false
}

// memorySubscription implements eventbus.Subscription for the in-memory bus.
type memorySubscription struct {
	bus *Bus
	sub *subscription
}

func (s *memorySubscription) Unsubscribe() error {
	s.bus.unsubscribe(s.sub)
	return nil
}

func (s *memorySubscription) Topic() string {
	return s.sub.pattern
}
