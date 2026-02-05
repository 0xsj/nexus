package eventbus

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Adapter bridges identity domain.EventPublisher to pkg/eventbus.Publisher.
// Identity domain events implement domain.DomainEvent (EventType, AggregateID, AggregateType)
// and also embed eventsourcing.BaseEvent, so they implement eventsourcing.Event.
type Adapter struct {
	publisher eventbus.Publisher
}

// Compile-time check.
var _ domain.EventPublisher = (*Adapter)(nil)

// NewAdapter creates a new eventbus adapter for Identity.
func NewAdapter(publisher eventbus.Publisher) *Adapter {
	return &Adapter{publisher: publisher}
}

// Publish converts identity domain events to EventEnvelopes and publishes synchronously.
func (a *Adapter) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	envelopes, err := a.toEnvelopes(events)
	if err != nil {
		return err
	}
	return a.publisher.Publish(ctx, envelopes...)
}

// toEnvelopes converts domain events to event envelopes.
func (a *Adapter) toEnvelopes(events []domain.DomainEvent) ([]*eventsourcing.EventEnvelope, error) {
	envelopes := make([]*eventsourcing.EventEnvelope, 0, len(events))
	for _, event := range events {
		// Identity events embed eventsourcing.BaseEvent, so they implement eventsourcing.Event.
		esEvent, ok := event.(eventsourcing.Event)
		if !ok {
			continue
		}
		envelope, err := eventsourcing.NewEventEnvelope(esEvent, 0)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, envelope)
	}
	return envelopes, nil
}
