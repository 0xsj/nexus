package bridge

import (
	"context"
	"encoding/json"

	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
	ws "github.com/0xsj/nexus/platform/pkg/websocket"
)

// RouteMapping maps an event bus pattern to a WebSocket room.
type RouteMapping struct {
	// Pattern is the event bus subscription pattern (e.g., "identity.*").
	Pattern string

	// Room is the WebSocket room to broadcast to. If empty, broadcasts to all connections.
	Room string
}

// Bridge subscribes to event bus patterns and forwards matching events to a WebSocket hub.
type Bridge struct {
	bus           eventbus.Subscriber
	hub           ws.Hub
	logger        log.Logger
	mappings      []RouteMapping
	subscriptions []eventbus.Subscription
}

// New creates a new EventBus-to-WebSocket bridge.
func New(bus eventbus.Subscriber, hub ws.Hub, logger log.Logger, mappings ...RouteMapping) *Bridge {
	if logger == nil {
		logger = log.New()
	}

	return &Bridge{
		bus:      bus,
		hub:      hub,
		logger:   logger.With(log.Component("websocket.bridge")),
		mappings: mappings,
	}
}

// Start subscribes to all configured patterns and begins forwarding events.
func (b *Bridge) Start(ctx context.Context) error {
	for _, mapping := range b.mappings {
		m := mapping // capture loop variable
		sub, err := b.bus.Subscribe(ctx, m.Pattern, func(ctx context.Context, event *eventsourcing.EventEnvelope) error {
			return b.forward(m, event)
		})
		if err != nil {
			b.Close()
			return err
		}
		b.subscriptions = append(b.subscriptions, sub)

		b.logger.Debug("bridge route active",
			log.String("pattern", m.Pattern),
			log.String("room", m.Room),
		)
	}

	b.logger.Info("eventbus-websocket bridge started",
		log.Int("routes", len(b.mappings)),
	)
	return nil
}

// Close unsubscribes all active subscriptions.
func (b *Bridge) Close() error {
	for _, sub := range b.subscriptions {
		sub.Unsubscribe()
	}
	b.subscriptions = nil
	return nil
}

// forward converts an EventEnvelope into a WebSocket Message and sends it to the hub.
func (b *Bridge) forward(mapping RouteMapping, event *eventsourcing.EventEnvelope) error {
	payload, err := json.Marshal(event)
	if err != nil {
		b.logger.Error("failed to marshal event for websocket",
			log.String("event_type", event.Type),
			log.Err(err),
		)
		return nil // don't fail the handler, just skip
	}

	msg := ws.Message{
		Type:      ws.MessageTypeEvent,
		Topic:     event.Type,
		Payload:   payload,
		Timestamp: event.Timestamp,
	}

	if mapping.Room != "" {
		b.hub.BroadcastToRoom(mapping.Room, msg)
	} else {
		b.hub.Broadcast(msg)
	}

	return nil
}
