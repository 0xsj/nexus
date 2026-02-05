package eventbus

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Adapter bridges pkg/eventbus.Subscriber to ledger domain.EventSubscriber.
// It converts EventEnvelope to the ledger's DomainEvent interface.
type Adapter struct {
	subscriber eventbus.Subscriber
}

// Compile-time check.
var _ domain.EventSubscriber = (*Adapter)(nil)

// NewAdapter creates a new eventbus adapter for Ledger.
func NewAdapter(subscriber eventbus.Subscriber) *Adapter {
	return &Adapter{subscriber: subscriber}
}

// Subscribe registers a handler for the given event patterns.
// For each pattern, it creates a pkg/eventbus subscription that converts
// EventEnvelope → envelopeDomainEvent before passing to the ledger handler.
func (a *Adapter) Subscribe(ctx context.Context, patterns []string, handler domain.EventHandler) error {
	for _, pattern := range patterns {
		_, err := a.subscriber.Subscribe(ctx, pattern, func(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
			adapted := &envelopeDomainEvent{envelope: envelope}
			return handler(ctx, adapted)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Unsubscribe is a no-op for now. Full unsubscription tracking would require
// storing subscription handles per pattern.
func (a *Adapter) Unsubscribe(_ context.Context, _ []string) error {
	return nil
}

// envelopeDomainEvent adapts an EventEnvelope to implement ledger domain.DomainEvent.
type envelopeDomainEvent struct {
	envelope *eventsourcing.EventEnvelope
}

var _ domain.DomainEvent = (*envelopeDomainEvent)(nil)

func (e *envelopeDomainEvent) EventType() string {
	return e.envelope.Type
}

func (e *envelopeDomainEvent) OccurredAt() int64 {
	return e.envelope.Timestamp.UnixMilli()
}

func (e *envelopeDomainEvent) AggregateID() string {
	return e.envelope.AggregateID
}

func (e *envelopeDomainEvent) AggregateType() string {
	return e.envelope.AggregateType
}

// Envelope returns the underlying EventEnvelope for metadata extraction.
func (e *envelopeDomainEvent) Envelope() *eventsourcing.EventEnvelope {
	return e.envelope
}
