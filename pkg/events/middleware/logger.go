// Package middleware provides event middleware.
package middleware

import (
	"context"
	"time"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
)

// Logger wraps a handler with logging.
func Logger(log logger.Logger) events.Middleware {
	return func(next events.Handler) events.Handler {
		return events.HandlerFunc(func(ctx context.Context, event events.Event) error {
			start := time.Now()

			log.Info("Event handler started",
				logger.String("handler", next.HandlerName()),
				logger.String("event_type", event.EventType()),
				logger.String("event_id", event.EventID()),
				logger.String("aggregate_id", event.AggregateID()),
			)

			err := next.Handle(ctx, event)

			duration := time.Since(start)

			if err != nil {
				log.Error("Event handler failed",
					logger.String("handler", next.HandlerName()),
					logger.String("event_type", event.EventType()),
					logger.Err(err),
					logger.Duration("duration", duration),
				)
			} else {
				log.Info("Event handler completed",
					logger.String("handler", next.HandlerName()),
					logger.String("event_type", event.EventType()),
					logger.Duration("duration", duration),
				)
			}

			return err
		})
	}
}
