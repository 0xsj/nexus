package eventsourcing

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Event Interface
// ============================================================================

// Event represents a domain event — an immutable fact that happened.
type Event interface {
	// EventType returns the unique type name of the event.
	// Format: "AggregateType.EventName" e.g., "Credential.Issued"
	EventType() string

	// OccurredAt returns when the event occurred.
	OccurredAt() time.Time

	// AggregateID returns the ID of the aggregate this event belongs to.
	AggregateID() string

	// AggregateType returns the type of aggregate.
	AggregateType() string
}

// ============================================================================
// Event Envelope
// ============================================================================

// EventEnvelope wraps an event with metadata for storage and transport.
type EventEnvelope struct {
	// ID is the unique identifier of this event instance.
	ID string `json:"id"`

	// Type is the event type name.
	Type string `json:"type"`

	// AggregateID is the ID of the aggregate.
	AggregateID string `json:"aggregate_id"`

	// AggregateType is the type of aggregate.
	AggregateType string `json:"aggregate_type"`

	// Version is the aggregate version after this event.
	Version int `json:"version"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Data is the serialized event payload.
	Data json.RawMessage `json:"data"`

	// Metadata contains additional context.
	Metadata EventMetadata `json:"metadata,omitempty"`
}

// EventMetadata contains contextual information about an event.
type EventMetadata struct {
	// CorrelationID links related operations.
	CorrelationID string `json:"correlation_id,omitempty"`

	// CausationID is the ID of the event/command that caused this event.
	CausationID string `json:"causation_id,omitempty"`

	// UserID is the user who triggered this event.
	UserID string `json:"user_id,omitempty"`

	// TenantID is the tenant context.
	TenantID string `json:"tenant_id,omitempty"`

	// TraceID for distributed tracing.
	TraceID string `json:"trace_id,omitempty"`

	// SpanID for distributed tracing.
	SpanID string `json:"span_id,omitempty"`

	// Custom allows for additional metadata.
	Custom map[string]any `json:"custom,omitempty"`
}

// NewEventEnvelope creates a new event envelope.
func NewEventEnvelope(event Event, version int) (*EventEnvelope, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, ErrEventValidation("NewEventEnvelope", "failed to marshal event: "+err.Error())
	}

	return &EventEnvelope{
		ID:            generateEventID(),
		Type:          event.EventType(),
		AggregateID:   event.AggregateID(),
		AggregateType: event.AggregateType(),
		Version:       version,
		Timestamp:     event.OccurredAt(),
		Data:          data,
	}, nil
}

// WithMetadata sets the metadata on the envelope.
func (e *EventEnvelope) WithMetadata(metadata EventMetadata) *EventEnvelope {
	e.Metadata = metadata
	return e
}

// WithCorrelationID sets the correlation ID.
func (e *EventEnvelope) WithCorrelationID(id string) *EventEnvelope {
	e.Metadata.CorrelationID = id
	return e
}

// WithCausationID sets the causation ID.
func (e *EventEnvelope) WithCausationID(id string) *EventEnvelope {
	e.Metadata.CausationID = id
	return e
}

// WithUserID sets the user ID.
func (e *EventEnvelope) WithUserID(id string) *EventEnvelope {
	e.Metadata.UserID = id
	return e
}

// WithTenantID sets the tenant ID.
func (e *EventEnvelope) WithTenantID(id string) *EventEnvelope {
	e.Metadata.TenantID = id
	return e
}

// WithTracing sets tracing IDs.
func (e *EventEnvelope) WithTracing(traceID, spanID string) *EventEnvelope {
	e.Metadata.TraceID = traceID
	e.Metadata.SpanID = spanID
	return e
}

// ============================================================================
// Base Event
// ============================================================================

// BaseEvent provides common event functionality.
// Embed this in your domain events.
type BaseEvent struct {
	id            string
	aggregateID   string
	aggregateType string
	occurredAt    time.Time
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(aggregateType, aggregateID string) BaseEvent {
	return BaseEvent{
		id:            generateEventID(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		occurredAt:    time.Now().UTC(),
	}
}

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// AggregateID returns the aggregate ID.
func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

// AggregateType returns the aggregate type.
func (e BaseEvent) AggregateType() string {
	return e.aggregateType
}

// ============================================================================
// Event Registry
// ============================================================================

// EventRegistry maps event type names to their Go types.
// Used for deserializing events from storage.
type EventRegistry struct {
	types map[string]func() Event
}

// NewEventRegistry creates a new event registry.
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		types: make(map[string]func() Event),
	}
}

// Register registers an event type.
func (r *EventRegistry) Register(eventType string, factory func() Event) {
	r.types[eventType] = factory
}

// RegisterType registers an event using a sample instance.
func RegisterType[E Event](r *EventRegistry, sample E) {
	r.types[sample.EventType()] = func() Event {
		var e E
		return e
	}
}

// Create creates a new instance of the registered event type.
func (r *EventRegistry) Create(eventType string) (Event, error) {
	factory, exists := r.types[eventType]
	if !exists {
		return nil, ErrEventValidation("EventRegistry.Create", "unknown event type: "+eventType)
	}
	return factory(), nil
}

// Deserialize deserializes an event envelope into a typed event.
func (r *EventRegistry) Deserialize(envelope *EventEnvelope) (Event, error) {
	event, err := r.Create(envelope.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(envelope.Data, event); err != nil {
		return nil, ErrEventValidation("EventRegistry.Deserialize", "failed to unmarshal event: "+err.Error())
	}

	return event, nil
}

// DefaultRegistry is the global event registry.
var DefaultRegistry = NewEventRegistry()

// Register registers an event type in the default registry.
func Register(eventType string, factory func() Event) {
	DefaultRegistry.Register(eventType, factory)
}

// ============================================================================
// Event Stream
// ============================================================================

// EventStream represents a sequence of events for an aggregate.
type EventStream struct {
	AggregateID   string
	AggregateType string
	Version       int
	Events        []*EventEnvelope
}

// NewEventStream creates a new event stream.
func NewEventStream(aggregateType, aggregateID string) *EventStream {
	return &EventStream{
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       0,
		Events:        make([]*EventEnvelope, 0),
	}
}

// Append adds events to the stream.
func (s *EventStream) Append(events ...*EventEnvelope) {
	s.Events = append(s.Events, events...)
	if len(events) > 0 {
		s.Version = events[len(events)-1].Version
	}
}

// IsEmpty returns true if the stream has no events.
func (s *EventStream) IsEmpty() bool {
	return len(s.Events) == 0
}

// Len returns the number of events in the stream.
func (s *EventStream) Len() int {
	return len(s.Events)
}

// ============================================================================
// Helpers
// ============================================================================

// generateEventID generates a unique event ID.
func generateEventID() string {
	return generateUUID()
}

// generateUUID generates a UUID v4.
func generateUUID() string {
	// Simple implementation - in production use a proper UUID library
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (i * 8))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10

	return formatUUID(b)
}

func formatUUID(b []byte) string {
	return sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func sprintf(format string, args ...any) string {
	// Simple hex formatter
	result := ""
	argIdx := 0

	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			i++ // skip %
			width := 0
			for i < len(format) && format[i] >= '0' && format[i] <= '9' {
				width = width*10 + int(format[i]-'0')
				i++
			}
			if i < len(format) && format[i] == 'x' && argIdx < len(args) {
				if bytes, ok := args[argIdx].([]byte); ok {
					for _, b := range bytes {
						result += hexByte(b)
					}
				}
				argIdx++
			}
		} else {
			result += string(format[i])
		}
	}

	return result
}

func hexByte(b byte) string {
	const hex = "0123456789abcdef"
	return string([]byte{hex[b>>4], hex[b&0x0f]})
}
