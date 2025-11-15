// Package memory provides an in-memory event bus implementation.
// Suitable for development and testing. Not for production use across multiple instances.
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
)

// Bus is an in-memory implementation of EventBus.
// Events are delivered synchronously in the same process.
type Bus struct {
	subscribers map[string][]*subscription
	mu          sync.RWMutex
	logger      logger.Logger
	closed      bool
}

// New creates a new in-memory event bus.
func New(log logger.Logger) *Bus {
	return &Bus{
		subscribers: make(map[string][]*subscription),
		logger:      log,
	}
}

// Publish publishes an event to all matching subscribers.
func (b *Bus) Publish(ctx context.Context, event events.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("event bus is closed")
	}

	eventType := event.EventType()

	b.logger.Debug("Publishing event",
		logger.String("event_type", eventType),
		logger.String("event_id", event.EventID()),
		logger.String("aggregate_id", event.AggregateID()),
	)

	// Find matching subscribers
	subscribers := b.findMatchingSubscribers(eventType)

	if len(subscribers) == 0 {
		b.logger.Debug("No subscribers for event",
			logger.String("event_type", eventType),
		)
		return nil
	}

	// Deliver to all subscribers
	var wg sync.WaitGroup
	errors := make(chan error, len(subscribers))

	for _, sub := range subscribers {
		if !sub.active {
			continue
		}

		wg.Add(1)
		go func(s *subscription) {
			defer wg.Done()

			if err := s.handler.Handle(ctx, event); err != nil {
				b.logger.Error("Handler failed to process event",
					logger.String("handler", s.handler.HandlerName()),
					logger.String("event_type", eventType),
					logger.Err(err),
				)
				errors <- err
			}
		}(sub)
	}

	wg.Wait()
	close(errors)

	// Check if any handlers failed
	var firstError error
	for err := range errors {
		if firstError == nil {
			firstError = err
		}
	}

	return firstError
}

// PublishBatch publishes multiple events.
func (b *Bus) PublishBatch(ctx context.Context, eventList []events.Event) error {
	for _, event := range eventList {
		if err := b.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers a handler for events matching the pattern.
func (b *Bus) Subscribe(ctx context.Context, pattern string, handler events.Handler) (events.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("event bus is closed")
	}

	sub := &subscription{
		pattern: pattern,
		handler: handler,
		active:  true,
		bus:     b,
	}

	b.subscribers[pattern] = append(b.subscribers[pattern], sub)

	b.logger.Info("Subscribed to events",
		logger.String("pattern", pattern),
		logger.String("handler", handler.HandlerName()),
	)

	return sub, nil
}

// SubscribeGroup registers a handler as part of a consumer group.
// For in-memory bus, this behaves the same as Subscribe.
func (b *Bus) SubscribeGroup(ctx context.Context, group, pattern string, handler events.Handler) (events.Subscription, error) {
	b.logger.Warn("Consumer groups not supported in memory bus, using regular subscription",
		logger.String("group", group),
		logger.String("pattern", pattern),
	)
	return b.Subscribe(ctx, pattern, handler)
}

// Close closes the event bus.
func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Unsubscribe all
	for _, subs := range b.subscribers {
		for _, sub := range subs {
			sub.active = false
		}
	}

	b.subscribers = make(map[string][]*subscription)

	b.logger.Info("Event bus closed")
	return nil
}

// Health checks if the event bus is healthy.
func (b *Bus) Health(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("event bus is closed")
	}

	return nil
}

// findMatchingSubscribers finds all subscribers matching the event type.
func (b *Bus) findMatchingSubscribers(eventType string) []*subscription {
	var matches []*subscription

	for pattern, subs := range b.subscribers {
		if matchesPattern(eventType, pattern) {
			matches = append(matches, subs...)
		}
	}

	return matches
}

// matchesPattern checks if an event type matches a subscription pattern.
// Supports:
//   - Exact match: "user.created"
//   - Wildcard: "user.*"
func matchesPattern(eventType, pattern string) bool {
	if pattern == eventType {
		return true
	}

	// Simple wildcard support: "user.*" matches "user.created", "user.updated", etc.
	if len(pattern) > 2 && pattern[len(pattern)-2:] == ".*" {
		prefix := pattern[:len(pattern)-2]
		return len(eventType) > len(prefix) && eventType[:len(prefix)] == prefix && eventType[len(prefix)] == '.'
	}

	return false
}

// subscription represents an active subscription.
type subscription struct {
	pattern string
	handler events.Handler
	active  bool
	bus     *Bus
}

// Unsubscribe unsubscribes the handler.
func (s *subscription) Unsubscribe() error {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()

	s.active = false

	// Remove from bus
	subs := s.bus.subscribers[s.pattern]
	for i, sub := range subs {
		if sub == s {
			s.bus.subscribers[s.pattern] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	return nil
}

// IsActive returns true if the subscription is active.
func (s *subscription) IsActive() bool {
	return s.active
}
