package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// Chain composes multiple middleware into a single middleware.
// Middleware are applied in order: Chain(A, B, C) wraps as A(B(C(handler))).
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// WithLogging returns middleware that logs event receipt and handling duration.
func WithLogging(logger log.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event *eventsourcing.EventEnvelope) error {
			start := time.Now()
			logger.Debug("handling event",
				log.String("event_type", event.Type),
				log.String("aggregate_id", event.AggregateID),
				log.String("event_id", event.ID),
			)

			err := next(ctx, event)

			fields := []log.Field{
				log.String("event_type", event.Type),
				log.String("aggregate_id", event.AggregateID),
				log.Duration("duration", time.Since(start)),
			}

			if err != nil {
				logger.Error("event handler failed", append(fields, log.Err(err))...)
			} else {
				logger.Debug("event handled", fields...)
			}

			return err
		}
	}
}

// WithRecovery returns middleware that recovers from panics in handlers.
func WithRecovery(logger log.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event *eventsourcing.EventEnvelope) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("event handler panicked",
						log.String("event_type", event.Type),
						log.String("aggregate_id", event.AggregateID),
						log.String("panic", fmt.Sprintf("%v", r)),
					)
					err = ErrHandlerFailed("eventbus.WithRecovery", event.Type, fmt.Errorf("panic: %v", r))
				}
			}()
			return next(ctx, event)
		}
	}
}

// WithRetry returns middleware that retries failed handlers with a fixed delay.
func WithRetry(maxRetries int, delay time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event *eventsourcing.EventEnvelope) error {
			var err error
			for attempt := 0; attempt <= maxRetries; attempt++ {
				err = next(ctx, event)
				if err == nil {
					return nil
				}

				if attempt < maxRetries {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(delay):
					}
				}
			}
			return err
		}
	}
}
