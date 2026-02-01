package projection

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/ledger/domain"
)

// ============================================================================
// Projection Handler
// ============================================================================

// Handler processes domain events and projects them into audit entries.
type Handler struct {
	writer    domain.Writer
	mapper    *Mapper
	extractor domain.MetadataExtractor
}

// NewHandler creates a new projection Handler.
func NewHandler(
	writer domain.Writer,
	mapper *Mapper,
	extractor domain.MetadataExtractor,
) *Handler {
	return &Handler{
		writer:    writer,
		mapper:    mapper,
		extractor: extractor,
	}
}

// HandleEvent processes a single domain event.
func (h *Handler) HandleEvent(ctx context.Context, event domain.DomainEvent) error {
	// Extract metadata from event
	var meta domain.EventMetadata
	if h.extractor != nil {
		meta = h.extractor.Extract(event)
	}

	// Map event to audit entry
	entry, err := h.mapper.MapEvent(event, meta)
	if err != nil {
		return err
	}

	// Append to ledger
	return h.writer.Append(ctx, entry)
}

// HandleEvents processes multiple domain events.
func (h *Handler) HandleEvents(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	entries := make([]*domain.AuditEntry, 0, len(events))

	for _, event := range events {
		var meta domain.EventMetadata
		if h.extractor != nil {
			meta = h.extractor.Extract(event)
		}

		entry, err := h.mapper.MapEvent(event, meta)
		if err != nil {
			return err
		}

		entries = append(entries, entry)
	}

	return h.writer.AppendBatch(ctx, entries)
}

// ============================================================================
// Event Handler Function
// ============================================================================

// AsEventHandler returns the handler as a domain.EventHandler function.
// This can be used to subscribe to the event bus.
func (h *Handler) AsEventHandler() domain.EventHandler {
	return func(ctx context.Context, event domain.DomainEvent) error {
		return h.HandleEvent(ctx, event)
	}
}

// ============================================================================
// Subscription Setup
// ============================================================================

// DefaultEventPatterns returns the default event patterns to subscribe to.
// These cover all major bounded contexts in the platform.
func DefaultEventPatterns() []string {
	return []string{
		"identity.*",
		"wallet.*",
		"verification.*",
		"credential.*",
		"presentation.*",
		"trust.*",
		"organization.*",
		"schema.*",
	}
}

// Subscribe sets up the handler to receive events from the subscriber.
func (h *Handler) Subscribe(ctx context.Context, subscriber domain.EventSubscriber) error {
	return subscriber.Subscribe(ctx, DefaultEventPatterns(), h.AsEventHandler())
}

// SubscribeToPatterns sets up the handler to receive specific event patterns.
func (h *Handler) SubscribeToPatterns(
	ctx context.Context,
	subscriber domain.EventSubscriber,
	patterns []string,
) error {
	return subscriber.Subscribe(ctx, patterns, h.AsEventHandler())
}
