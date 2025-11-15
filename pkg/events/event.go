// Package events provides event bus abstractions for domain events.
package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event that has occurred.
// All domain events must implement this interface.
type Event interface {
	// EventID returns the unique identifier for this event instance.
	EventID() string

	// EventType returns the type/name of the event (e.g., "user.created").
	EventType() string

	// EventVersion returns the schema version of the event.
	EventVersion() int

	// OccurredAt returns when the event occurred.
	OccurredAt() time.Time

	// AggregateID returns the ID of the aggregate that produced this event.
	AggregateID() string

	// AggregateType returns the type of aggregate (e.g., "user", "page").
	AggregateType() string

	// Payload returns the event data.
	Payload() interface{}

	// Metadata returns additional event metadata.
	Metadata() map[string]string
}

// BaseEvent provides a default implementation of Event.
// Domain events can embed this to avoid implementing boilerplate.
type BaseEvent struct {
	ID             string            `json:"event_id"`
	Type           string            `json:"event_type"`
	Version        int               `json:"event_version"`
	Timestamp      time.Time         `json:"occurred_at"`
	AggregateID_   string            `json:"aggregate_id"`
	AggregateType_ string            `json:"aggregate_type"`
	Data           interface{}       `json:"payload"`
	Meta           map[string]string `json:"metadata,omitempty"`
}

// NewBaseEvent creates a new base event with generated ID and timestamp.
func NewBaseEvent(eventType, aggregateType, aggregateID string, payload interface{}) BaseEvent {
	return BaseEvent{
		ID:             uuid.New().String(),
		Type:           eventType,
		Version:        1,
		Timestamp:      time.Now().UTC(),
		AggregateID_:   aggregateID,
		AggregateType_: aggregateType,
		Data:           payload,
		Meta:           make(map[string]string),
	}
}

// EventID returns the unique identifier for this event instance.
func (e BaseEvent) EventID() string {
	return e.ID
}

// EventType returns the type/name of the event.
func (e BaseEvent) EventType() string {
	return e.Type
}

// EventVersion returns the schema version of the event.
func (e BaseEvent) EventVersion() int {
	return e.Version
}

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the ID of the aggregate that produced this event.
func (e BaseEvent) AggregateID() string {
	return e.AggregateID_
}

// AggregateType returns the type of aggregate.
func (e BaseEvent) AggregateType() string {
	return e.AggregateType_
}

// Payload returns the event data.
func (e BaseEvent) Payload() interface{} {
	return e.Data
}

// Metadata returns additional event metadata.
func (e BaseEvent) Metadata() map[string]string {
	if e.Meta == nil {
		e.Meta = make(map[string]string)
	}
	return e.Meta
}

// WithMetadata adds metadata to the event.
func (e *BaseEvent) WithMetadata(key, value string) *BaseEvent {
	if e.Meta == nil {
		e.Meta = make(map[string]string)
	}
	e.Meta[key] = value
	return e
}

// EventEnvelope wraps an event with additional context.
// Used internally by event bus implementations.
type EventEnvelope struct {
	Event      Event
	Context    context.Context
	ReceivedAt time.Time
	Attempts   int
}
