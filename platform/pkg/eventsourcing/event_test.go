package eventsourcing_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Test Events
// ============================================================================

type UserCreatedEvent struct {
	eventsourcing.BaseEvent
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func NewUserCreatedEvent(aggregateID, userID, email, name string) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("User", aggregateID),
		UserID:    userID,
		Email:     email,
		Name:      name,
	}
}

func (e *UserCreatedEvent) EventType() string {
	return "User.Created"
}

type UserEmailChangedEvent struct {
	eventsourcing.BaseEvent
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

func NewUserEmailChangedEvent(aggregateID, oldEmail, newEmail string) *UserEmailChangedEvent {
	return &UserEmailChangedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("User", aggregateID),
		OldEmail:  oldEmail,
		NewEmail:  newEmail,
	}
}

func (e *UserEmailChangedEvent) EventType() string {
	return "User.EmailChanged"
}

type UserDeletedEvent struct {
	eventsourcing.BaseEvent
	Reason string `json:"reason"`
}

func NewUserDeletedEvent(aggregateID, reason string) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("User", aggregateID),
		Reason:    reason,
	}
}

func (e *UserDeletedEvent) EventType() string {
	return "User.Deleted"
}

// ============================================================================
// BaseEvent Tests
// ============================================================================

func TestNewBaseEvent(t *testing.T) {
	base := eventsourcing.NewBaseEvent("User", "user-123")

	if base.AggregateType() != "User" {
		t.Errorf("expected aggregate type 'User', got '%s'", base.AggregateType())
	}

	if base.AggregateID() != "user-123" {
		t.Errorf("expected aggregate ID 'user-123', got '%s'", base.AggregateID())
	}

	if base.OccurredAt().IsZero() {
		t.Error("expected non-zero occurred at time")
	}

	// OccurredAt should be recent (within last second)
	if time.Since(base.OccurredAt()) > time.Second {
		t.Error("expected occurred at to be recent")
	}
}

func TestUserCreatedEvent(t *testing.T) {
	event := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")

	if event.EventType() != "User.Created" {
		t.Errorf("expected event type 'User.Created', got '%s'", event.EventType())
	}

	if event.AggregateType() != "User" {
		t.Errorf("expected aggregate type 'User', got '%s'", event.AggregateType())
	}

	if event.AggregateID() != "user-123" {
		t.Errorf("expected aggregate ID 'user-123', got '%s'", event.AggregateID())
	}

	if event.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", event.UserID)
	}

	if event.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", event.Email)
	}

	if event.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got '%s'", event.Name)
	}
}

// ============================================================================
// EventEnvelope Tests
// ============================================================================

func TestNewEventEnvelope(t *testing.T) {
	event := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")

	envelope, err := eventsourcing.NewEventEnvelope(event, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if envelope.ID == "" {
		t.Error("expected non-empty envelope ID")
	}

	if envelope.Type != "User.Created" {
		t.Errorf("expected type 'User.Created', got '%s'", envelope.Type)
	}

	if envelope.AggregateID != "user-123" {
		t.Errorf("expected aggregate ID 'user-123', got '%s'", envelope.AggregateID)
	}

	if envelope.AggregateType != "User" {
		t.Errorf("expected aggregate type 'User', got '%s'", envelope.AggregateType)
	}

	if envelope.Version != 1 {
		t.Errorf("expected version 1, got %d", envelope.Version)
	}

	if envelope.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	if len(envelope.Data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestEventEnvelope_WithMetadata(t *testing.T) {
	event := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")

	envelope, err := eventsourcing.NewEventEnvelope(event, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metadata := eventsourcing.EventMetadata{
		CorrelationID: "corr-123",
		CausationID:   "cause-456",
		UserID:        "actor-789",
		TenantID:      "tenant-abc",
		TraceID:       "trace-def",
		SpanID:        "span-ghi",
	}

	envelope.WithMetadata(metadata)

	if envelope.Metadata.CorrelationID != "corr-123" {
		t.Errorf("expected correlation ID 'corr-123', got '%s'", envelope.Metadata.CorrelationID)
	}

	if envelope.Metadata.CausationID != "cause-456" {
		t.Errorf("expected causation ID 'cause-456', got '%s'", envelope.Metadata.CausationID)
	}

	if envelope.Metadata.UserID != "actor-789" {
		t.Errorf("expected user ID 'actor-789', got '%s'", envelope.Metadata.UserID)
	}

	if envelope.Metadata.TenantID != "tenant-abc" {
		t.Errorf("expected tenant ID 'tenant-abc', got '%s'", envelope.Metadata.TenantID)
	}

	if envelope.Metadata.TraceID != "trace-def" {
		t.Errorf("expected trace ID 'trace-def', got '%s'", envelope.Metadata.TraceID)
	}

	if envelope.Metadata.SpanID != "span-ghi" {
		t.Errorf("expected span ID 'span-ghi', got '%s'", envelope.Metadata.SpanID)
	}
}

func TestEventEnvelope_BuilderMethods(t *testing.T) {
	event := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")

	envelope, err := eventsourcing.NewEventEnvelope(event, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope.
		WithCorrelationID("corr-123").
		WithCausationID("cause-456").
		WithUserID("actor-789").
		WithTenantID("tenant-abc").
		WithTracing("trace-def", "span-ghi")

	if envelope.Metadata.CorrelationID != "corr-123" {
		t.Errorf("expected correlation ID 'corr-123', got '%s'", envelope.Metadata.CorrelationID)
	}

	if envelope.Metadata.CausationID != "cause-456" {
		t.Errorf("expected causation ID 'cause-456', got '%s'", envelope.Metadata.CausationID)
	}

	if envelope.Metadata.UserID != "actor-789" {
		t.Errorf("expected user ID 'actor-789', got '%s'", envelope.Metadata.UserID)
	}

	if envelope.Metadata.TenantID != "tenant-abc" {
		t.Errorf("expected tenant ID 'tenant-abc', got '%s'", envelope.Metadata.TenantID)
	}

	if envelope.Metadata.TraceID != "trace-def" {
		t.Errorf("expected trace ID 'trace-def', got '%s'", envelope.Metadata.TraceID)
	}

	if envelope.Metadata.SpanID != "span-ghi" {
		t.Errorf("expected span ID 'span-ghi', got '%s'", envelope.Metadata.SpanID)
	}
}

func TestEventEnvelope_JSONRoundTrip(t *testing.T) {
	event := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")

	envelope, err := eventsourcing.NewEventEnvelope(event, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope.WithCorrelationID("corr-123").WithUserID("actor-789")

	// Marshal
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("failed to marshal envelope: %v", err)
	}

	// Unmarshal
	var decoded eventsourcing.EventEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	if decoded.ID != envelope.ID {
		t.Errorf("expected ID '%s', got '%s'", envelope.ID, decoded.ID)
	}

	if decoded.Type != envelope.Type {
		t.Errorf("expected type '%s', got '%s'", envelope.Type, decoded.Type)
	}

	if decoded.AggregateID != envelope.AggregateID {
		t.Errorf("expected aggregate ID '%s', got '%s'", envelope.AggregateID, decoded.AggregateID)
	}

	if decoded.Version != envelope.Version {
		t.Errorf("expected version %d, got %d", envelope.Version, decoded.Version)
	}

	if decoded.Metadata.CorrelationID != envelope.Metadata.CorrelationID {
		t.Errorf("expected correlation ID '%s', got '%s'", envelope.Metadata.CorrelationID, decoded.Metadata.CorrelationID)
	}
}

// ============================================================================
// EventRegistry Tests
// ============================================================================

func TestNewEventRegistry(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()

	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestEventRegistry_Register(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()

	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})

	event, err := registry.Create("User.Created")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event == nil {
		t.Fatal("expected non-nil event")
	}

	if event.EventType() != "" {
		// BaseEvent hasn't been initialized, so EventType comes from the struct
		_, ok := event.(*UserCreatedEvent)
		if !ok {
			t.Error("expected *UserCreatedEvent")
		}
	}
}

func TestEventRegistry_Create_UnknownType(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()

	_, err := registry.Create("Unknown.Event")
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}

func TestEventRegistry_Deserialize(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()

	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})

	// Create an envelope with event data
	originalEvent := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")
	envelope, err := eventsourcing.NewEventEnvelope(originalEvent, 1)
	if err != nil {
		t.Fatalf("unexpected error creating envelope: %v", err)
	}

	// Deserialize
	event, err := registry.Deserialize(envelope)
	if err != nil {
		t.Fatalf("unexpected error deserializing: %v", err)
	}

	userCreated, ok := event.(*UserCreatedEvent)
	if !ok {
		t.Fatalf("expected *UserCreatedEvent, got %T", event)
	}

	if userCreated.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", userCreated.UserID)
	}

	if userCreated.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", userCreated.Email)
	}

	if userCreated.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got '%s'", userCreated.Name)
	}
}

func TestEventRegistry_Deserialize_UnknownType(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()

	envelope := &eventsourcing.EventEnvelope{
		Type: "Unknown.Event",
		Data: []byte(`{}`),
	}

	_, err := registry.Deserialize(envelope)
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Register in default registry
	eventsourcing.Register("Test.Event", func() eventsourcing.Event {
		return &UserDeletedEvent{}
	})

	event, err := eventsourcing.DefaultRegistry.Create("Test.Event")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := event.(*UserDeletedEvent)
	if !ok {
		t.Errorf("expected *UserDeletedEvent, got %T", event)
	}
}

// ============================================================================
// EventStream Tests
// ============================================================================

func TestNewEventStream(t *testing.T) {
	stream := eventsourcing.NewEventStream("User", "user-123")

	if stream.AggregateID != "user-123" {
		t.Errorf("expected aggregate ID 'user-123', got '%s'", stream.AggregateID)
	}

	if stream.AggregateType != "User" {
		t.Errorf("expected aggregate type 'User', got '%s'", stream.AggregateType)
	}

	if stream.Version != 0 {
		t.Errorf("expected version 0, got %d", stream.Version)
	}

	if !stream.IsEmpty() {
		t.Error("expected empty stream")
	}

	if stream.Len() != 0 {
		t.Errorf("expected length 0, got %d", stream.Len())
	}
}

func TestEventStream_Append(t *testing.T) {
	stream := eventsourcing.NewEventStream("User", "user-123")

	event1 := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")
	envelope1, _ := eventsourcing.NewEventEnvelope(event1, 1)

	event2 := NewUserEmailChangedEvent("user-123", "test@example.com", "new@example.com")
	envelope2, _ := eventsourcing.NewEventEnvelope(event2, 2)

	stream.Append(envelope1, envelope2)

	if stream.IsEmpty() {
		t.Error("expected non-empty stream")
	}

	if stream.Len() != 2 {
		t.Errorf("expected length 2, got %d", stream.Len())
	}

	if stream.Version != 2 {
		t.Errorf("expected version 2, got %d", stream.Version)
	}
}

func TestEventStream_Append_Empty(t *testing.T) {
	stream := eventsourcing.NewEventStream("User", "user-123")

	stream.Append()

	if !stream.IsEmpty() {
		t.Error("expected empty stream after appending nothing")
	}

	if stream.Version != 0 {
		t.Errorf("expected version 0, got %d", stream.Version)
	}
}

// ============================================================================
// EventMetadata Tests
// ============================================================================

func TestEventMetadata_Custom(t *testing.T) {
	metadata := eventsourcing.EventMetadata{
		CorrelationID: "corr-123",
		Custom: map[string]any{
			"request_id": "req-456",
			"ip_address": "192.168.1.1",
		},
	}

	if metadata.Custom["request_id"] != "req-456" {
		t.Errorf("expected request_id 'req-456', got '%v'", metadata.Custom["request_id"])
	}

	if metadata.Custom["ip_address"] != "192.168.1.1" {
		t.Errorf("expected ip_address '192.168.1.1', got '%v'", metadata.Custom["ip_address"])
	}
}

func TestEventMetadata_JSONRoundTrip(t *testing.T) {
	metadata := eventsourcing.EventMetadata{
		CorrelationID: "corr-123",
		CausationID:   "cause-456",
		UserID:        "user-789",
		TenantID:      "tenant-abc",
		TraceID:       "trace-def",
		SpanID:        "span-ghi",
		Custom: map[string]any{
			"key": "value",
		},
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	var decoded eventsourcing.EventMetadata
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}

	if decoded.CorrelationID != metadata.CorrelationID {
		t.Errorf("expected correlation ID '%s', got '%s'", metadata.CorrelationID, decoded.CorrelationID)
	}

	if decoded.Custom["key"] != "value" {
		t.Errorf("expected custom key 'value', got '%v'", decoded.Custom["key"])
	}
}

// ============================================================================
// Multiple Event Types Tests
// ============================================================================

func TestMultipleEventTypes(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()

	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})
	registry.Register("User.EmailChanged", func() eventsourcing.Event {
		return &UserEmailChangedEvent{}
	})
	registry.Register("User.Deleted", func() eventsourcing.Event {
		return &UserDeletedEvent{}
	})

	// Test User.Created
	createdEvent := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test")
	createdEnvelope, _ := eventsourcing.NewEventEnvelope(createdEvent, 1)

	deserializedCreated, err := registry.Deserialize(createdEnvelope)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := deserializedCreated.(*UserCreatedEvent); !ok {
		t.Errorf("expected *UserCreatedEvent, got %T", deserializedCreated)
	}

	// Test User.EmailChanged
	emailChangedEvent := NewUserEmailChangedEvent("user-123", "old@example.com", "new@example.com")
	emailChangedEnvelope, _ := eventsourcing.NewEventEnvelope(emailChangedEvent, 2)

	deserializedEmailChanged, err := registry.Deserialize(emailChangedEnvelope)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := deserializedEmailChanged.(*UserEmailChangedEvent); !ok {
		t.Errorf("expected *UserEmailChangedEvent, got %T", deserializedEmailChanged)
	}

	// Test User.Deleted
	deletedEvent := NewUserDeletedEvent("user-123", "requested")
	deletedEnvelope, _ := eventsourcing.NewEventEnvelope(deletedEvent, 3)

	deserializedDeleted, err := registry.Deserialize(deletedEnvelope)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := deserializedDeleted.(*UserDeletedEvent); !ok {
		t.Errorf("expected *UserDeletedEvent, got %T", deserializedDeleted)
	}
}
