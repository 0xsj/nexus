package eventhandler

import (
	"context"
	"time"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/nexus/pkg/resilience/idempotency"
)

// Middleware is a function that wraps an event handler.
type Middleware func(events.Handler) events.Handler

// Chain combines multiple middleware into one.
func Chain(middlewares ...Middleware) Middleware {
	return func(handler events.Handler) events.Handler {
		// Apply middleware in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// WithIdempotency adds idempotency checking to a handler.
func WithIdempotency(store idempotency.Store, ttl time.Duration, log logger.Logger) Middleware {
	return func(next events.Handler) events.Handler {
		return events.HandlerFunc(func(ctx context.Context, event events.Event) error {
			// Check if already processed
			if CheckAndMarkProcessed(ctx, store, event, ttl, log) {
				log.Debug("Event already processed",
					logger.String("event_id", event.EventID()),
				)
				return nil
			}

			// Process event
			return next.Handle(ctx, event)
		})
	}
}

// WithStaleEventFilter skips events older than maxAge.
func WithStaleEventFilter(maxAge time.Duration, log logger.Logger) Middleware {
	return func(next events.Handler) events.Handler {
		return events.HandlerFunc(func(ctx context.Context, event events.Event) error {
			// Skip stale events
			if IsEventStale(event, maxAge) {
				log.Debug("Skipping stale event",
					logger.String("event_id", event.EventID()),
					logger.Duration("event_age", time.Since(event.OccurredAt())),
				)
				return nil
			}

			// Process event
			return next.Handle(ctx, event)
		})
	}
}

// WithLogging adds logging to a handler.
func WithLogging(log logger.Logger) Middleware {
	return func(next events.Handler) events.Handler {
		return events.HandlerFunc(func(ctx context.Context, event events.Event) error {
			start := time.Now()

			log.Debug("Processing event",
				logger.String("event_id", event.EventID()),
				logger.String("event_type", event.EventType()),
				logger.String("handler", next.HandlerName()),
			)

			err := next.Handle(ctx, event)
			duration := time.Since(start)

			if err != nil {
				log.Error("Event processing failed",
					logger.String("event_id", event.EventID()),
					logger.String("event_type", event.EventType()),
					logger.Duration("duration", duration),
					logger.Err(err),
				)
			} else {
				log.Debug("Event processed successfully",
					logger.String("event_id", event.EventID()),
					logger.String("event_type", event.EventType()),
					logger.Duration("duration", duration),
				)
			}

			return err
		})
	}
}
