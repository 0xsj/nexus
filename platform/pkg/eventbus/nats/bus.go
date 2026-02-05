package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// Bus is a NATS JetStream-backed event bus for production use.
type Bus struct {
	config Config
	logger log.Logger

	conn   *nats.Conn
	js     jetstream.JetStream
	stream jetstream.Stream

	mu            sync.Mutex
	subscriptions []*natsSubscription
	closed        bool
}

// New creates a new NATS event bus. Call Start to connect and initialize.
func New(config Config) *Bus {
	logger := config.Logger
	if logger == nil {
		logger = log.New()
	}

	return &Bus{
		config:        config,
		logger:        logger.With(log.Component("eventbus.nats")),
		subscriptions: make([]*natsSubscription, 0),
	}
}

// Start connects to NATS and ensures the JetStream stream exists.
func (b *Bus) Start(ctx context.Context) error {
	opts := []nats.Option{
		nats.MaxReconnects(b.config.MaxReconnects),
		nats.ReconnectWait(b.config.ReconnectWait),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			b.logger.Warn("nats disconnected", log.Err(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			b.logger.Info("nats reconnected", log.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			b.logger.Info("nats connection closed")
		}),
	}

	if b.config.CredentialsFile != "" {
		opts = append(opts, nats.UserCredentials(b.config.CredentialsFile))
	}

	conn, err := nats.Connect(b.config.URL, opts...)
	if err != nil {
		return eventbus.ErrConnectionFailed("nats.Bus.Start", err)
	}
	b.conn = conn

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return eventbus.ErrConnectionFailed("nats.Bus.Start", err)
	}
	b.js = js

	subjects := b.config.StreamSubjects
	if len(subjects) == 0 {
		subjects = []string{"events.>"}
	}

	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     b.config.StreamName,
		Subjects: subjects,
	})
	if err != nil {
		conn.Close()
		return eventbus.ErrConnectionFailed("nats.Bus.Start",
			fmt.Errorf("create stream %s: %w", b.config.StreamName, err))
	}
	b.stream = stream

	b.logger.Info("nats event bus started",
		log.String("url", b.config.URL),
		log.String("stream", b.config.StreamName),
	)
	return nil
}

// Close drains the NATS connection and cleans up subscriptions.
func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	for _, sub := range b.subscriptions {
		sub.cancel()
	}
	b.subscriptions = nil

	if b.conn != nil {
		b.conn.Drain()
	}

	b.logger.Info("nats event bus closed")
	return nil
}

// Publish sends events to the NATS JetStream stream synchronously.
func (b *Bus) Publish(ctx context.Context, events ...*eventsourcing.EventEnvelope) error {
	if b.isClosed() {
		return eventbus.ErrBusClosed("nats.Bus.Publish")
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return eventbus.ErrSerializationFailed("nats.Bus.Publish", err)
		}

		subject := toNATSSubject(event.Type)
		if _, err := b.js.Publish(ctx, subject, data); err != nil {
			return eventbus.ErrPublishFailed("nats.Bus.Publish", err)
		}
	}
	return nil
}

// PublishAsync sends events to NATS asynchronously using PublishAsync.
func (b *Bus) PublishAsync(ctx context.Context, events ...*eventsourcing.EventEnvelope) error {
	if b.isClosed() {
		return eventbus.ErrBusClosed("nats.Bus.PublishAsync")
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return eventbus.ErrSerializationFailed("nats.Bus.PublishAsync", err)
		}

		subject := toNATSSubject(event.Type)
		if _, err := b.js.PublishAsync(subject, data); err != nil {
			return eventbus.ErrPublishFailed("nats.Bus.PublishAsync", err)
		}
	}
	return nil
}

// Subscribe creates a JetStream consumer for events matching the pattern.
func (b *Bus) Subscribe(ctx context.Context, pattern string, handler eventbus.Handler, opts ...eventbus.SubscribeOption) (eventbus.Subscription, error) {
	if b.isClosed() {
		return nil, eventbus.ErrBusClosed("nats.Bus.Subscribe")
	}

	options := eventbus.ApplyOptions(opts...)

	// Apply configured middleware chain.
	wrapped := handler
	for i := len(b.config.Middlewares) - 1; i >= 0; i-- {
		wrapped = b.config.Middlewares[i](wrapped)
	}

	subject := toNATSSubject(pattern)

	consumerCfg := jetstream.ConsumerConfig{
		FilterSubject: subject,
		AckWait:       b.config.AckWait,
		MaxDeliver:    b.config.MaxDeliver,
	}

	if options.QueueGroup != "" {
		consumerCfg.Durable = options.QueueGroup
	}

	if options.AckTimeout > 0 {
		consumerCfg.AckWait = options.AckTimeout
	}

	consumer, err := b.stream.CreateOrUpdateConsumer(ctx, consumerCfg)
	if err != nil {
		return nil, eventbus.ErrSubscribeFailed("nats.Bus.Subscribe", pattern,
			fmt.Errorf("create consumer: %w", err))
	}

	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		var envelope eventsourcing.EventEnvelope
		if err := json.Unmarshal(msg.Data(), &envelope); err != nil {
			b.logger.Error("failed to unmarshal event",
				log.String("subject", msg.Subject()),
				log.Err(err),
			)
			msg.Term()
			return
		}

		if err := wrapped(ctx, &envelope); err != nil {
			b.logger.Error("handler failed, nacking",
				log.String("event_type", envelope.Type),
				log.Err(err),
			)
			msg.Nak()
			return
		}

		msg.Ack()
	})
	if err != nil {
		return nil, eventbus.ErrSubscribeFailed("nats.Bus.Subscribe", pattern,
			fmt.Errorf("consume: %w", err))
	}

	sub := &natsSubscription{
		pattern:    pattern,
		consumeCtx: consumeCtx,
	}

	b.mu.Lock()
	b.subscriptions = append(b.subscriptions, sub)
	b.mu.Unlock()

	b.logger.Debug("subscription added",
		log.String("pattern", pattern),
		log.String("subject", subject),
	)

	return sub, nil
}

// isClosed returns true if the bus has been closed.
func (b *Bus) isClosed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.closed
}

// toNATSSubject converts a dot-delimited event pattern to a NATS subject.
// Wildcards: '*' stays as '*', '>' stays as '>'.
// Event types like "User.Registered" become "events.User.Registered".
func toNATSSubject(pattern string) string {
	if strings.HasPrefix(pattern, "events.") {
		return pattern
	}
	return "events." + pattern
}

// natsSubscription implements eventbus.Subscription backed by a JetStream consumer.
type natsSubscription struct {
	pattern    string
	consumeCtx jetstream.ConsumeContext
}

func (s *natsSubscription) Unsubscribe() error {
	s.cancel()
	return nil
}

func (s *natsSubscription) Topic() string {
	return s.pattern
}

func (s *natsSubscription) cancel() {
	if s.consumeCtx != nil {
		s.consumeCtx.Stop()
	}
}
