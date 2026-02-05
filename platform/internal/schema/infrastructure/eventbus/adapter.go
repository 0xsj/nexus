package eventbus

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Adapter bridges schema domain.EventPublisher to pkg/eventbus.Publisher.
// Schema domain events are eventsourcing.Event directly.
type Adapter struct {
	publisher eventbus.Publisher
}

// Compile-time check.
var _ domain.EventPublisher = (*Adapter)(nil)

// NewAdapter creates a new eventbus adapter for Schema.
func NewAdapter(publisher eventbus.Publisher) *Adapter {
	return &Adapter{publisher: publisher}
}

// Publish converts schema domain events to EventEnvelopes and publishes synchronously.
func (a *Adapter) Publish(ctx context.Context, events ...eventsourcing.Event) error {
	envelopes, err := toEnvelopes(events)
	if err != nil {
		return err
	}
	return a.publisher.Publish(ctx, envelopes...)
}

// PublishAsync converts schema domain events to EventEnvelopes and publishes asynchronously.
func (a *Adapter) PublishAsync(ctx context.Context, events ...eventsourcing.Event) error {
	envelopes, err := toEnvelopes(events)
	if err != nil {
		return err
	}
	return a.publisher.PublishAsync(ctx, envelopes...)
}

// toEnvelopes converts eventsourcing.Event to EventEnvelopes.
func toEnvelopes(events []eventsourcing.Event) ([]*eventsourcing.EventEnvelope, error) {
	envelopes := make([]*eventsourcing.EventEnvelope, 0, len(events))
	for _, event := range events {
		envelope, err := eventsourcing.NewEventEnvelope(event, 0)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, envelope)
	}
	return envelopes, nil
}
