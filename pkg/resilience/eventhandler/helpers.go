package eventhandler

import (
	"context"
	"time"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/nexus/pkg/resilience/idempotency"
)

// IsEventStale checks if an event is older than the specified age.
func IsEventStale(event events.Event, maxAge time.Duration) bool {
	return time.Since(event.OccurredAt()) > maxAge
}

// CheckAndMarkProcessed checks if an event was processed and marks it if not.
// Returns true if the event was already processed.
func CheckAndMarkProcessed(
	ctx context.Context,
	store idempotency.Store,
	event events.Event,
	ttl time.Duration,
	log logger.Logger,
) bool {
	// Check if already processed
	exists, err := store.Exists(ctx, event.EventID())
	if err != nil {
		log.Error("Failed to check idempotency",
			logger.Err(err),
			logger.String("event_id", event.EventID()),
		)
		return false // Fail open - allow processing
	}

	if exists {
		return true // Already processed
	}

	// Mark as processed
	if err := store.Save(ctx, event.EventID(), nil, ttl); err != nil {
		log.Error("Failed to mark event as processed",
			logger.Err(err),
			logger.String("event_id", event.EventID()),
		)
	}

	return false
}

// SkipIfStale returns an error if the event is stale, nil otherwise.
func SkipIfStale(event events.Event, maxAge time.Duration, log logger.Logger) error {
	if IsEventStale(event, maxAge) {
		eventAge := time.Since(event.OccurredAt())
		log.Debug("Skipping stale event",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.Duration("event_age", eventAge),
			logger.Duration("max_age", maxAge),
		)
		return nil // Not an error, just skip
	}
	return nil
}
