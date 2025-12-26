package eventsourcing

import (
	"context"
	"time"
)

// ============================================================================
// Event Store Interface
// ============================================================================

// EventStore is the interface for event persistence.
// Implementations can be in-memory, PostgreSQL, Kafka, etc.
type EventStore interface {
	// Save persists events for an aggregate.
	// Returns ErrConcurrencyConflict if expectedVersion doesn't match.
	Save(ctx context.Context, aggregateID string, events []*EventEnvelope, expectedVersion int) error

	// Load retrieves all events for an aggregate.
	// Returns empty slice if aggregate doesn't exist.
	Load(ctx context.Context, aggregateID string) ([]*EventEnvelope, error)

	// LoadFrom retrieves events starting from a specific version.
	LoadFrom(ctx context.Context, aggregateID string, fromVersion int) ([]*EventEnvelope, error)

	// LoadByType retrieves all events for an aggregate type.
	// Useful for rebuilding projections.
	LoadByType(ctx context.Context, aggregateType string, opts LoadOptions) ([]*EventEnvelope, error)
}

// LoadOptions configures event loading.
type LoadOptions struct {
	// FromVersion loads events from this version (inclusive).
	FromVersion int

	// ToVersion loads events up to this version (inclusive).
	// 0 means no upper limit.
	ToVersion int

	// FromTime loads events from this timestamp (inclusive).
	FromTime *time.Time

	// ToTime loads events up to this timestamp (inclusive).
	ToTime *time.Time

	// Limit is the maximum number of events to return.
	// 0 means no limit.
	Limit int

	// EventTypes filters by specific event types.
	// Empty means all types.
	EventTypes []string
}

// DefaultLoadOptions returns default load options.
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		FromVersion: 0,
		ToVersion:   0,
		Limit:       0,
	}
}

// WithFromVersion sets the from version.
func (o LoadOptions) WithFromVersion(v int) LoadOptions {
	o.FromVersion = v
	return o
}

// WithToVersion sets the to version.
func (o LoadOptions) WithToVersion(v int) LoadOptions {
	o.ToVersion = v
	return o
}

// WithFromTime sets the from time.
func (o LoadOptions) WithFromTime(t time.Time) LoadOptions {
	o.FromTime = &t
	return o
}

// WithToTime sets the to time.
func (o LoadOptions) WithToTime(t time.Time) LoadOptions {
	o.ToTime = &t
	return o
}

// WithLimit sets the limit.
func (o LoadOptions) WithLimit(limit int) LoadOptions {
	o.Limit = limit
	return o
}

// WithEventTypes sets the event type filter.
func (o LoadOptions) WithEventTypes(types ...string) LoadOptions {
	o.EventTypes = types
	return o
}

// ============================================================================
// Extended Event Store Interface
// ============================================================================

// ExtendedEventStore provides additional querying capabilities.
type ExtendedEventStore interface {
	EventStore

	// LoadAll retrieves all events across all aggregates.
	// Use with caution - can be expensive.
	LoadAll(ctx context.Context, opts LoadOptions) ([]*EventEnvelope, error)

	// Count returns the total number of events for an aggregate.
	Count(ctx context.Context, aggregateID string) (int, error)

	// CountByType returns the total number of events for an aggregate type.
	CountByType(ctx context.Context, aggregateType string) (int, error)

	// Delete removes all events for an aggregate.
	// Use with extreme caution - breaks event sourcing guarantees.
	Delete(ctx context.Context, aggregateID string) error
}

// ============================================================================
// Subscription Support
// ============================================================================

// EventSubscription allows subscribing to new events.
type EventSubscription interface {
	// Subscribe starts receiving events.
	// Returns a channel that receives events.
	Subscribe(ctx context.Context, opts SubscriptionOptions) (<-chan *EventEnvelope, error)

	// Unsubscribe stops receiving events.
	Unsubscribe() error
}

// SubscriptionOptions configures event subscription.
type SubscriptionOptions struct {
	// AggregateTypes filters by aggregate types.
	// Empty means all types.
	AggregateTypes []string

	// EventTypes filters by event types.
	// Empty means all types.
	EventTypes []string

	// FromPosition starts from this position.
	// Empty string means from the beginning.
	FromPosition string

	// BufferSize is the channel buffer size.
	BufferSize int
}

// DefaultSubscriptionOptions returns default subscription options.
func DefaultSubscriptionOptions() SubscriptionOptions {
	return SubscriptionOptions{
		BufferSize: 100,
	}
}

// SubscribableEventStore is an event store that supports subscriptions.
type SubscribableEventStore interface {
	EventStore
	EventSubscription
}

// ============================================================================
// Transactional Event Store
// ============================================================================

// TransactionalEventStore supports transactional operations.
type TransactionalEventStore interface {
	EventStore

	// BeginTx starts a transaction.
	BeginTx(ctx context.Context) (EventStoreTx, error)
}

// EventStoreTx is a transactional event store.
type EventStoreTx interface {
	EventStore

	// Commit commits the transaction.
	Commit() error

	// Rollback rolls back the transaction.
	Rollback() error
}

// ============================================================================
// Event Publisher
// ============================================================================

// EventPublisher publishes events to external systems.
type EventPublisher interface {
	// Publish publishes events.
	Publish(ctx context.Context, events []*EventEnvelope) error
}

// PublishingEventStore wraps an EventStore and publishes events after saving.
type PublishingEventStore struct {
	store     EventStore
	publisher EventPublisher
}

// NewPublishingEventStore creates a new publishing event store.
func NewPublishingEventStore(store EventStore, publisher EventPublisher) *PublishingEventStore {
	return &PublishingEventStore{
		store:     store,
		publisher: publisher,
	}
}

// Save saves events and publishes them.
func (s *PublishingEventStore) Save(ctx context.Context, aggregateID string, events []*EventEnvelope, expectedVersion int) error {
	// Save first
	if err := s.store.Save(ctx, aggregateID, events, expectedVersion); err != nil {
		return err
	}

	// Then publish (best effort)
	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, events); err != nil {
			// Log but don't fail - events are already saved
			// In production, use outbox pattern for guaranteed delivery
			_ = err
		}
	}

	return nil
}

// Load delegates to the underlying store.
func (s *PublishingEventStore) Load(ctx context.Context, aggregateID string) ([]*EventEnvelope, error) {
	return s.store.Load(ctx, aggregateID)
}

// LoadFrom delegates to the underlying store.
func (s *PublishingEventStore) LoadFrom(ctx context.Context, aggregateID string, fromVersion int) ([]*EventEnvelope, error) {
	return s.store.LoadFrom(ctx, aggregateID, fromVersion)
}

// LoadByType delegates to the underlying store.
func (s *PublishingEventStore) LoadByType(ctx context.Context, aggregateType string, opts LoadOptions) ([]*EventEnvelope, error) {
	return s.store.LoadByType(ctx, aggregateType, opts)
}

// ============================================================================
// Snapshot Store
// ============================================================================

// Snapshot represents a point-in-time state of an aggregate.
type Snapshot struct {
	// AggregateID is the aggregate identifier.
	AggregateID string

	// AggregateType is the type of aggregate.
	AggregateType string

	// Version is the aggregate version at snapshot time.
	Version int

	// Data is the serialized aggregate state.
	Data []byte

	// Timestamp is when the snapshot was taken.
	Timestamp time.Time
}

// SnapshotStore persists aggregate snapshots.
type SnapshotStore interface {
	// Save saves a snapshot.
	Save(ctx context.Context, snapshot *Snapshot) error

	// Load loads the latest snapshot for an aggregate.
	Load(ctx context.Context, aggregateID string) (*Snapshot, error)

	// Delete removes a snapshot.
	Delete(ctx context.Context, aggregateID string) error
}

// SnapshotStrategy determines when to take snapshots.
type SnapshotStrategy interface {
	// ShouldSnapshot returns true if a snapshot should be taken.
	ShouldSnapshot(aggregate Aggregate, eventsSinceSnapshot int) bool
}

// EveryNEventsStrategy takes a snapshot every N events.
type EveryNEventsStrategy struct {
	n int
}

// NewEveryNEventsStrategy creates a new every-N-events strategy.
func NewEveryNEventsStrategy(n int) *EveryNEventsStrategy {
	return &EveryNEventsStrategy{n: n}
}

// ShouldSnapshot returns true if N events have occurred since last snapshot.
func (s *EveryNEventsStrategy) ShouldSnapshot(aggregate Aggregate, eventsSinceSnapshot int) bool {
	return eventsSinceSnapshot >= s.n
}
