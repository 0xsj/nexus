package eventsourcing

import (
	"context"
)

// ============================================================================
// Aggregate Interface
// ============================================================================

// Aggregate represents an event-sourced aggregate root.
type Aggregate interface {
	// AggregateID returns the unique identifier of the aggregate.
	AggregateID() string

	// AggregateType returns the type name of the aggregate.
	AggregateType() string

	// Version returns the current version (number of events applied).
	Version() int

	// Changes returns uncommitted events.
	Changes() []Event

	// ClearChanges clears uncommitted events after saving.
	ClearChanges()

	// ApplyEvent applies an event to update aggregate state.
	// Called during hydration (replay) and when raising new events.
	ApplyEvent(event Event)
}

// ============================================================================
// Aggregate Root Base
// ============================================================================

// AggregateRoot provides base functionality for event-sourced aggregates.
// Embed this in your domain aggregates.
type AggregateRoot struct {
	id      string
	typ     string
	version int
	changes []Event
}

// InitAggregate initializes the aggregate root.
// Call this in your aggregate's constructor.
func (a *AggregateRoot) InitAggregate(aggregateType, aggregateID string) {
	a.id = aggregateID
	a.typ = aggregateType
	a.version = 0
	a.changes = make([]Event, 0)
}

// AggregateID returns the aggregate ID.
func (a *AggregateRoot) AggregateID() string {
	return a.id
}

// AggregateType returns the aggregate type.
func (a *AggregateRoot) AggregateType() string {
	return a.typ
}

// Version returns the current version.
func (a *AggregateRoot) Version() int {
	return a.version
}

// Changes returns uncommitted events.
func (a *AggregateRoot) Changes() []Event {
	return a.changes
}

// ClearChanges clears uncommitted events.
func (a *AggregateRoot) ClearChanges() {
	a.changes = make([]Event, 0)
}

// SetVersion sets the version (used during hydration).
func (a *AggregateRoot) SetVersion(version int) {
	a.version = version
}

// IncrementVersion increments the version.
func (a *AggregateRoot) IncrementVersion() {
	a.version++
}

// Raise records a new event and applies it.
// Use this in command methods to emit events.
func (a *AggregateRoot) Raise(aggregate Aggregate, event Event) {
	a.changes = append(a.changes, event)
	a.version++
	aggregate.ApplyEvent(event)
}

// HasChanges returns true if there are uncommitted events.
func (a *AggregateRoot) HasChanges() bool {
	return len(a.changes) > 0
}

// ============================================================================
// Aggregate Hydration
// ============================================================================

// Hydrate rebuilds an aggregate from events.
func Hydrate(aggregate Aggregate, events []Event) {
	for _, event := range events {
		aggregate.ApplyEvent(event)
		if root, ok := getAggregateRoot(aggregate); ok {
			root.IncrementVersion()
		}
	}
}

// HydrateFromEnvelopes rebuilds an aggregate from event envelopes.
func HydrateFromEnvelopes(aggregate Aggregate, envelopes []*EventEnvelope, registry *EventRegistry) error {
	for _, envelope := range envelopes {
		event, err := registry.Deserialize(envelope)
		if err != nil {
			return err
		}
		aggregate.ApplyEvent(event)
		if root, ok := getAggregateRoot(aggregate); ok {
			root.SetVersion(envelope.Version)
		}
	}
	return nil
}

// getAggregateRoot extracts the embedded AggregateRoot if present.
func getAggregateRoot(aggregate Aggregate) (*AggregateRoot, bool) {
	type hasRoot interface {
		GetAggregateRoot() *AggregateRoot
	}

	if ar, ok := aggregate.(hasRoot); ok {
		return ar.GetAggregateRoot(), true
	}
	return nil, false
}

// ============================================================================
// Aggregate Factory
// ============================================================================

// AggregateFactory creates new aggregate instances.
type AggregateFactory interface {
	// Create creates a new aggregate instance with the given ID.
	Create(aggregateID string) Aggregate
}

// AggregateFactoryFunc is a function adapter for AggregateFactory.
type AggregateFactoryFunc func(aggregateID string) Aggregate

// Create implements AggregateFactory.
func (f AggregateFactoryFunc) Create(aggregateID string) Aggregate {
	return f(aggregateID)
}

// ============================================================================
// Aggregate Repository
// ============================================================================

// AggregateRepository loads and saves aggregates.
type AggregateRepository[T Aggregate] interface {
	// Load loads an aggregate by ID.
	Load(ctx context.Context, aggregateID string) (T, error)

	// Save saves an aggregate's uncommitted events.
	Save(ctx context.Context, aggregate T) error

	// Exists checks if an aggregate exists.
	Exists(ctx context.Context, aggregateID string) (bool, error)
}

// ============================================================================
// Generic Repository Implementation
// ============================================================================

// GenericRepository is a generic aggregate repository implementation.
type GenericRepository[T Aggregate] struct {
	store    EventStore
	factory  AggregateFactory
	registry *EventRegistry
}

// NewGenericRepository creates a new generic repository.
func NewGenericRepository[T Aggregate](
	store EventStore,
	factory AggregateFactory,
	registry *EventRegistry,
) *GenericRepository[T] {
	if registry == nil {
		registry = DefaultRegistry
	}
	return &GenericRepository[T]{
		store:    store,
		factory:  factory,
		registry: registry,
	}
}

// Load loads an aggregate by ID.
func (r *GenericRepository[T]) Load(ctx context.Context, aggregateID string) (T, error) {
	var zero T

	aggregate := r.factory.Create(aggregateID)

	envelopes, err := r.store.Load(ctx, aggregateID)
	if err != nil {
		return zero, err
	}

	if len(envelopes) == 0 {
		return zero, ErrAggregateNotFound("Repository.Load", aggregate.AggregateType(), aggregateID)
	}

	if err := HydrateFromEnvelopes(aggregate, envelopes, r.registry); err != nil {
		return zero, err
	}

	typed, ok := aggregate.(T)
	if !ok {
		return zero, ErrAggregateValidation("Repository.Load", "type assertion failed")
	}

	return typed, nil
}

// Save saves an aggregate's uncommitted events.
func (r *GenericRepository[T]) Save(ctx context.Context, aggregate T) error {
	changes := aggregate.Changes()
	if len(changes) == 0 {
		return nil
	}

	// Calculate expected version (version before changes)
	expectedVersion := aggregate.Version() - len(changes)

	// Create envelopes
	envelopes := make([]*EventEnvelope, 0, len(changes))
	for i, event := range changes {
		envelope, err := NewEventEnvelope(event, expectedVersion+i+1)
		if err != nil {
			return err
		}
		envelopes = append(envelopes, envelope)
	}

	// Save to store
	if err := r.store.Save(ctx, aggregate.AggregateID(), envelopes, expectedVersion); err != nil {
		return err
	}

	// Clear changes after successful save
	aggregate.ClearChanges()

	return nil
}

// Exists checks if an aggregate exists.
func (r *GenericRepository[T]) Exists(ctx context.Context, aggregateID string) (bool, error) {
	envelopes, err := r.store.Load(ctx, aggregateID)
	if err != nil {
		if IsEventNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return len(envelopes) > 0, nil
}

// ============================================================================
// Aggregate Validator
// ============================================================================

// ValidatableAggregate is implemented by aggregates that can validate themselves.
type ValidatableAggregate interface {
	Aggregate
	Validate() error
}

// ValidateAggregate validates an aggregate if it implements ValidatableAggregate.
func ValidateAggregate(aggregate Aggregate) error {
	if v, ok := aggregate.(ValidatableAggregate); ok {
		return v.Validate()
	}
	return nil
}
