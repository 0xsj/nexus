package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/nats-io/nats.go"
)

// Bus is a NATS implementation of EventBus.
type Bus struct {
	conn         *nats.Conn
	js           nats.JetStreamContext
	config       *Config
	logger       logger.Logger
	subs         map[string]*subscription
	mu           sync.RWMutex
	closed       bool
	useJetStream bool
}

// New creates a new NATS event bus.
func New(cfg *Config, log logger.Logger) (*Bus, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Build NATS options
	opts := buildNATSOptions(cfg, log)

	// Connect to NATS
	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	bus := &Bus{
		conn:         nc,
		config:       cfg,
		logger:       log,
		subs:         make(map[string]*subscription),
		useJetStream: cfg.UseJetStream,
	}

	// Initialize JetStream if enabled
	if cfg.UseJetStream {
		js, err := nc.JetStream()
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to initialize JetStream: %w", err)
		}
		bus.js = js

		// Create stream if it doesn't exist
		if err := bus.ensureStream(); err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to create stream: %w", err)
		}
	}

	log.Info("NATS event bus connected",
		logger.String("url", cfg.URL),
		logger.Bool("jetstream", cfg.UseJetStream),
	)

	return bus, nil
}

// Publish publishes an event to NATS.
func (b *Bus) Publish(ctx context.Context, event events.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("event bus is closed")
	}

	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Build subject from event type
	subject := eventTypeToSubject(event.EventType())

	b.logger.Debug("Publishing event",
		logger.String("event_type", event.EventType()),
		logger.String("event_id", event.EventID()),
		logger.String("subject", subject),
	)

	// Publish with JetStream or core NATS
	if b.useJetStream {
		_, err = b.js.Publish(subject, data)
	} else {
		err = b.conn.Publish(subject, data)
	}

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple events.
func (b *Bus) PublishBatch(ctx context.Context, eventList []events.Event) error {
	for _, event := range eventList {
		if err := b.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers a handler for events matching the pattern.
func (b *Bus) Subscribe(ctx context.Context, pattern string, handler events.Handler) (events.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("event bus is closed")
	}

	subject := patternToSubject(pattern)

	sub := &subscription{
		pattern: pattern,
		subject: subject,
		handler: handler,
		bus:     b,
		active:  true,
	}

	// Create NATS subscription
	var err error
	if b.useJetStream {
		// JetStream subscription (durable, persistent)
		sub.jsSub, err = b.js.Subscribe(subject, sub.handleMessage, nats.Durable(handler.HandlerName()))
	} else {
		// Core NATS subscription (ephemeral)
		sub.natsSub, err = b.conn.Subscribe(subject, sub.handleMessage)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	b.subs[pattern] = sub

	b.logger.Info("Subscribed to events",
		logger.String("pattern", pattern),
		logger.String("subject", subject),
		logger.String("handler", handler.HandlerName()),
	)

	return sub, nil
}

// SubscribeGroup registers a handler as part of a consumer group.
func (b *Bus) SubscribeGroup(ctx context.Context, group, pattern string, handler events.Handler) (events.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("event bus is closed")
	}

	subject := patternToSubject(pattern)

	sub := &subscription{
		pattern: pattern,
		subject: subject,
		handler: handler,
		bus:     b,
		active:  true,
		group:   group,
	}

	var err error
	if b.useJetStream {
		// JetStream queue subscription with durable name
		durableName := fmt.Sprintf("%s-%s", group, handler.HandlerName())
		sub.jsSub, err = b.js.QueueSubscribe(subject, group, sub.handleMessage, nats.Durable(durableName))
	} else {
		// Core NATS queue subscription
		sub.natsSub, err = b.conn.QueueSubscribe(subject, group, sub.handleMessage)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	key := fmt.Sprintf("%s:%s", group, pattern)
	b.subs[key] = sub

	b.logger.Info("Subscribed to events with group",
		logger.String("pattern", pattern),
		logger.String("subject", subject),
		logger.String("group", group),
		logger.String("handler", handler.HandlerName()),
	)

	return sub, nil
}

// Close closes the NATS connection and all subscriptions.
func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Unsubscribe all
	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}

	// Close NATS connection
	b.conn.Close()

	b.logger.Info("NATS event bus closed")
	return nil
}

// Health checks if the NATS connection is healthy.
func (b *Bus) Health(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("event bus is closed")
	}

	if !b.conn.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	return nil
}

// ensureStream creates the JetStream stream if it doesn't exist.
func (b *Bus) ensureStream() error {
	streamName := b.config.StreamName
	subjects := []string{streamName + ".>"}

	// Check if stream exists
	streamInfo, err := b.js.StreamInfo(streamName)
	if err == nil {
		// Stream exists - check if config needs updating
		if streamInfo.Config.MaxAge != b.config.StreamMaxAge {
			b.logger.Info("Updating JetStream stream configuration",
				logger.String("stream", streamName),
				logger.Duration("old_max_age", streamInfo.Config.MaxAge),
				logger.Duration("new_max_age", b.config.StreamMaxAge),
			)

			// Update stream configuration
			streamInfo.Config.MaxAge = b.config.StreamMaxAge
			_, err = b.js.UpdateStream(&streamInfo.Config)
			if err != nil {
				return fmt.Errorf("failed to update stream: %w", err)
			}

			b.logger.Info("Updated JetStream stream",
				logger.String("stream", streamName),
				logger.Duration("max_age", b.config.StreamMaxAge),
			)
		}
		return nil
	}

	// Create stream with configurable retention
	_, err = b.js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
		Storage:  nats.FileStorage,
		MaxAge:   b.config.StreamMaxAge, // ✅ Use configurable value instead of hardcoded 7 days
	})

	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	b.logger.Info("Created JetStream stream",
		logger.String("stream", streamName),
		logger.Duration("max_age", b.config.StreamMaxAge),
	)

	return nil
}

// buildNATSOptions builds NATS connection options from config.
func buildNATSOptions(cfg *Config, log logger.Logger) []nats.Option {
	opts := []nats.Option{
		nats.Name("nexus-event-bus"),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.Timeout(cfg.ConnectionTimeout),
		nats.PingInterval(cfg.PingInterval),
		nats.MaxPingsOutstanding(cfg.MaxPingsOut),
	}

	// Authentication
	if cfg.CredsFile != "" {
		opts = append(opts, nats.UserCredentials(cfg.CredsFile))
	} else if cfg.Token != "" {
		opts = append(opts, nats.Token(cfg.Token))
	} else if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.Username, cfg.Password))
	}

	// Connection event handlers
	opts = append(opts,
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Warn("NATS disconnected", logger.Err(err))
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS reconnected",
				logger.String("url", nc.ConnectedUrl()),
			)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS connection closed")
		}),
	)

	return opts
}

// eventTypeToSubject converts an event type to a NATS subject.
// Example: "user.created" -> "events.user.created"
func eventTypeToSubject(eventType string) string {
	return "events." + eventType
}

// patternToSubject converts a subscription pattern to a NATS subject.
// Examples:
//   - "user.created" -> "events.user.created"
//   - "user.*" -> "events.user.*"
func patternToSubject(pattern string) string {
	return "events." + pattern
}

// subscription represents an active NATS subscription.
type subscription struct {
	pattern string
	subject string
	handler events.Handler
	bus     *Bus
	active  bool
	group   string

	// NATS subscriptions
	natsSub *nats.Subscription
	jsSub   *nats.Subscription
}

// handleMessage handles incoming NATS messages.
func (s *subscription) handleMessage(msg *nats.Msg) {
	if !s.active {
		return
	}

	// Deserialize event
	var baseEvent events.BaseEvent
	if err := json.Unmarshal(msg.Data, &baseEvent); err != nil {
		s.bus.logger.Error("Failed to unmarshal event",
			logger.Err(err),
			logger.String("subject", msg.Subject),
		)
		return
	}

	// Handle the event
	ctx := context.Background()
	if err := s.handler.Handle(ctx, baseEvent); err != nil {
		s.bus.logger.Error("Handler failed to process event",
			logger.String("handler", s.handler.HandlerName()),
			logger.String("event_type", baseEvent.EventType()),
			logger.Err(err),
		)

		// NACK if using JetStream
		if s.jsSub != nil {
			_ = msg.Nak()
		}
		return
	}

	// ACK if using JetStream
	if s.jsSub != nil {
		_ = msg.Ack()
	}
}

// Unsubscribe unsubscribes from NATS.
func (s *subscription) Unsubscribe() error {
	s.active = false

	if s.natsSub != nil {
		return s.natsSub.Unsubscribe()
	}
	if s.jsSub != nil {
		return s.jsSub.Unsubscribe()
	}
	return nil
}

// IsActive returns true if the subscription is active.
func (s *subscription) IsActive() bool {
	return s.active
}
