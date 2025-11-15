package eventhandler

import (
	"context"
	"time"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/0xsj/nexus/pkg/resilience/circuitbreaker"
	"github.com/0xsj/nexus/pkg/resilience/idempotency"
	"github.com/0xsj/nexus/pkg/resilience/retry"
)

// ResilientHandler wraps an event handler with resilience features.
type ResilientHandler struct {
	handler events.Handler
	logger  logger.Logger

	// Resilience features
	idempotencyStore idempotency.Store
	retryPolicy      *retry.Policy
	circuitBreaker   *circuitbreaker.CircuitBreaker

	// Configuration
	idempotencyTTL  time.Duration
	skipStaleEvents bool
	staleEventAge   time.Duration
}

// Config holds configuration for resilient handlers.
type Config struct {
	// Idempotency
	EnableIdempotency bool
	IdempotencyStore  idempotency.Store
	IdempotencyTTL    time.Duration

	// Retry
	EnableRetry bool
	RetryPolicy *retry.Policy

	// Circuit Breaker
	EnableCircuitBreaker bool
	CircuitBreaker       *circuitbreaker.CircuitBreaker

	// Stale Events
	SkipStaleEvents bool
	StaleEventAge   time.Duration

	// Logger
	Logger logger.Logger
}

// DefaultConfig returns a default resilient handler configuration.
func DefaultConfig(log logger.Logger) Config {
	return Config{
		EnableIdempotency: true,
		IdempotencyTTL:    24 * time.Hour,

		EnableRetry: true,
		RetryPolicy: retry.NewPolicy(retry.DefaultConfig()),

		EnableCircuitBreaker: false,

		SkipStaleEvents: true,
		StaleEventAge:   5 * time.Minute,

		Logger: log,
	}
}

// NewResilientHandler creates a new resilient event handler wrapper.
func NewResilientHandler(handler events.Handler, cfg Config) *ResilientHandler {
	return &ResilientHandler{
		handler:          handler,
		logger:           cfg.Logger,
		idempotencyStore: cfg.IdempotencyStore,
		retryPolicy:      cfg.RetryPolicy,
		circuitBreaker:   cfg.CircuitBreaker,
		idempotencyTTL:   cfg.IdempotencyTTL,
		skipStaleEvents:  cfg.SkipStaleEvents,
		staleEventAge:    cfg.StaleEventAge,
	}
}

// Handle processes the event with resilience features.
func (h *ResilientHandler) Handle(ctx context.Context, event events.Event) error {
	// 1. Check if event is stale
	if h.skipStaleEvents && h.isStaleEvent(event) {
		h.logger.Debug("Skipping stale event",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.Duration("event_age", time.Since(event.OccurredAt())),
		)
		return nil
	}

	// 2. Check idempotency
	if h.idempotencyStore != nil {
		if h.isEventProcessed(ctx, event) {
			h.logger.Debug("Event already processed, skipping",
				logger.String("event_id", event.EventID()),
				logger.String("event_type", event.EventType()),
			)
			return nil
		}
	}

	// 3. Execute with circuit breaker and retry
	var err error

	if h.circuitBreaker != nil {
		// With circuit breaker
		err = h.executeWithCircuitBreaker(ctx, event)
	} else if h.retryPolicy != nil {
		// With retry only
		err = h.executeWithRetry(ctx, event)
	} else {
		// Direct execution
		err = h.handler.Handle(ctx, event)
	}

	// 4. Mark as processed if successful
	if err == nil && h.idempotencyStore != nil {
		h.markEventProcessed(ctx, event)
	}

	return err
}

// HandlerName returns the wrapped handler's name.
func (h *ResilientHandler) HandlerName() string {
	return h.handler.HandlerName()
}

// isStaleEvent checks if an event is too old.
func (h *ResilientHandler) isStaleEvent(event events.Event) bool {
	eventAge := time.Since(event.OccurredAt())
	return eventAge > h.staleEventAge
}

// isEventProcessed checks if the event has already been processed.
func (h *ResilientHandler) isEventProcessed(ctx context.Context, event events.Event) bool {
	exists, err := h.idempotencyStore.Exists(ctx, event.EventID())
	if err != nil {
		h.logger.Error("Failed to check idempotency",
			logger.Err(err),
			logger.String("event_id", event.EventID()),
		)
		return false // Fail open - allow processing
	}
	return exists
}

// markEventProcessed marks an event as processed.
func (h *ResilientHandler) markEventProcessed(ctx context.Context, event events.Event) {
	if err := h.idempotencyStore.Save(ctx, event.EventID(), nil, h.idempotencyTTL); err != nil {
		h.logger.Error("Failed to save idempotency key",
			logger.Err(err),
			logger.String("event_id", event.EventID()),
		)
	}
}

// executeWithRetry executes the handler with retry logic.
func (h *ResilientHandler) executeWithRetry(ctx context.Context, event events.Event) error {
	policy := h.retryPolicy.WithOnRetry(func(attempt int, err error, delay time.Duration) {
		h.logger.Warn("Retrying event handler",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.Int("attempt", attempt),
			logger.Duration("delay", delay),
			logger.Err(err),
		)
	})

	return retry.Do(ctx, policy, func() error {
		return h.handler.Handle(ctx, event)
	})
}

// executeWithCircuitBreaker executes the handler with circuit breaker protection.
func (h *ResilientHandler) executeWithCircuitBreaker(ctx context.Context, event events.Event) error {
	if h.retryPolicy != nil {
		// Circuit breaker + retry
		return h.circuitBreaker.Execute(ctx, func() error {
			return h.executeWithRetry(ctx, event)
		})
	}

	// Circuit breaker only
	return h.circuitBreaker.Execute(ctx, func() error {
		return h.handler.Handle(ctx, event)
	})
}
