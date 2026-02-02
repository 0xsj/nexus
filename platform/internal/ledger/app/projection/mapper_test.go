package projection

import (
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/internal/ledger/domain"
)

// ============================================================================
// Mock Domain Event
// ============================================================================

type mockDomainEvent struct {
	eventType     string
	occurredAt    int64
	aggregateID   string
	aggregateType string
}

func (e *mockDomainEvent) EventType() string     { return e.eventType }
func (e *mockDomainEvent) OccurredAt() int64     { return e.occurredAt }
func (e *mockDomainEvent) AggregateID() string   { return e.aggregateID }
func (e *mockDomainEvent) AggregateType() string { return e.aggregateType }

func newMockEvent(eventType, aggregateType, aggregateID string) *mockDomainEvent {
	return &mockDomainEvent{
		eventType:     eventType,
		occurredAt:    time.Now().UnixMilli(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
	}
}

// ============================================================================
// NewMapper Tests
// ============================================================================

func TestNewMapper(t *testing.T) {
	mapper := NewMapper()

	if mapper == nil {
		t.Fatal("expected non-nil Mapper")
	}
}

// ============================================================================
// ExtractActor Tests
// ============================================================================

func TestExtractActor_WithUserID(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{
		UserID: "user-123",
	}

	actor, err := mapper.ExtractActor(meta)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Type() != domain.ActorTypeUser {
		t.Errorf("Actor.Type() = %v, want %v", actor.Type(), domain.ActorTypeUser)
	}
	if actor.ID().String() != "user-123" {
		t.Errorf("Actor.ID() = %v, want %v", actor.ID().String(), "user-123")
	}
}

func TestExtractActor_WithoutUserID(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{
		UserID: "",
	}

	actor, err := mapper.ExtractActor(meta)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Type() != domain.ActorTypeSystem {
		t.Errorf("Actor.Type() = %v, want %v", actor.Type(), domain.ActorTypeSystem)
	}
	if actor.ID().String() != "event-processor" {
		t.Errorf("Actor.ID() = %v, want %v", actor.ID().String(), "event-processor")
	}
}

func TestExtractActor_EmptyUserID(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{}

	actor, err := mapper.ExtractActor(meta)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor.Type() != domain.ActorTypeSystem {
		t.Errorf("expected system actor for empty metadata")
	}
}

// ============================================================================
// ExtractSubject Tests
// ============================================================================

func TestExtractSubject_KnownAggregateTypes(t *testing.T) {
	mapper := NewMapper()

	tests := []struct {
		aggregateType string
		aggregateID   string
		wantType      domain.SubjectType
	}{
		{"User", "user-123", domain.SubjectTypeUser},
		{"Credential", "cred-456", domain.SubjectTypeCredential},
		{"Presentation", "pres-789", domain.SubjectTypePresentation},
		{"Verification", "ver-001", domain.SubjectTypeVerification},
		{"Wallet", "wallet-002", domain.SubjectTypeWallet},
		{"Session", "sess-003", domain.SubjectTypeSession},
		{"Organization", "org-004", domain.SubjectTypeOrganization},
		{"DID", "did:key:z123", domain.SubjectTypeDID},
		{"Vouch", "vouch-005", domain.SubjectTypeVouch},
	}

	for _, tt := range tests {
		t.Run(tt.aggregateType, func(t *testing.T) {
			event := newMockEvent("test.event", tt.aggregateType, tt.aggregateID)

			subject, err := mapper.ExtractSubject(event)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if subject.Type() != tt.wantType {
				t.Errorf("Subject.Type() = %v, want %v", subject.Type(), tt.wantType)
			}
			if subject.ID().String() != tt.aggregateID {
				t.Errorf("Subject.ID() = %v, want %v", subject.ID().String(), tt.aggregateID)
			}
		})
	}
}

func TestExtractSubject_UnknownAggregateType(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("test.event", "UnknownType", "unknown-123")

	subject, err := mapper.ExtractSubject(event)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject.Type() != domain.SubjectTypeUnknown {
		t.Errorf("Subject.Type() = %v, want %v", subject.Type(), domain.SubjectTypeUnknown)
	}
	if subject.ID().String() != "unknown-123" {
		t.Errorf("Subject.ID() = %v, want %v", subject.ID().String(), "unknown-123")
	}
}

func TestExtractSubject_EmptyAggregateID(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("test.event", "User", "")

	_, err := mapper.ExtractSubject(event)

	if err == nil {
		t.Error("expected error for empty aggregate ID")
	}
}

// ============================================================================
// ExtractMetadata Tests
// ============================================================================

func TestExtractMetadata_AllFields(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{
		CorrelationID: "corr-123",
		CausationID:   "cause-456",
		TraceID:       "trace-789",
		SpanID:        "span-abc",
		TenantID:      "tenant-xyz",
		Custom: map[string]any{
			"custom_key": "custom_value",
		},
	}

	metadata := mapper.ExtractMetadata(meta)

	if metadata.CorrelationID() != "corr-123" {
		t.Errorf("CorrelationID() = %v, want %v", metadata.CorrelationID(), "corr-123")
	}
	if metadata.CausationID() != "cause-456" {
		t.Errorf("CausationID() = %v, want %v", metadata.CausationID(), "cause-456")
	}
	if metadata.GetString("trace_id") != "trace-789" {
		t.Errorf("trace_id = %v, want %v", metadata.GetString("trace_id"), "trace-789")
	}
	if metadata.GetString("span_id") != "span-abc" {
		t.Errorf("span_id = %v, want %v", metadata.GetString("span_id"), "span-abc")
	}
	if metadata.GetString("tenant_id") != "tenant-xyz" {
		t.Errorf("tenant_id = %v, want %v", metadata.GetString("tenant_id"), "tenant-xyz")
	}
	if metadata.Get("custom_key") != "custom_value" {
		t.Errorf("custom_key = %v, want %v", metadata.Get("custom_key"), "custom_value")
	}
}

func TestExtractMetadata_PartialFields(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{
		CorrelationID: "corr-123",
	}

	metadata := mapper.ExtractMetadata(meta)

	if metadata.CorrelationID() != "corr-123" {
		t.Errorf("CorrelationID() = %v, want %v", metadata.CorrelationID(), "corr-123")
	}
	if metadata.CausationID() != "" {
		t.Errorf("CausationID() should be empty, got %v", metadata.CausationID())
	}
	if metadata.Has("trace_id") {
		t.Error("should not have trace_id")
	}
}

func TestExtractMetadata_EmptyFields(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{}

	metadata := mapper.ExtractMetadata(meta)

	if metadata.CorrelationID() != "" {
		t.Errorf("CorrelationID() should be empty")
	}
	if metadata.CausationID() != "" {
		t.Errorf("CausationID() should be empty")
	}
	if metadata.Len() != 0 {
		t.Errorf("expected empty metadata, got %d entries", metadata.Len())
	}
}

func TestExtractMetadata_CustomFieldsOnly(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{
		Custom: map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		},
	}

	metadata := mapper.ExtractMetadata(meta)

	if metadata.Get("key1") != "value1" {
		t.Errorf("key1 = %v, want %v", metadata.Get("key1"), "value1")
	}
	if metadata.Get("key2") != 42 {
		t.Errorf("key2 = %v, want %v", metadata.Get("key2"), 42)
	}
	if metadata.Get("key3") != true {
		t.Errorf("key3 = %v, want %v", metadata.Get("key3"), true)
	}
}

func TestExtractMetadata_NilCustom(t *testing.T) {
	mapper := NewMapper()
	meta := domain.EventMetadata{
		CorrelationID: "corr-123",
		Custom:        nil,
	}

	metadata := mapper.ExtractMetadata(meta)

	if metadata.CorrelationID() != "corr-123" {
		t.Errorf("CorrelationID() = %v, want %v", metadata.CorrelationID(), "corr-123")
	}
}

// ============================================================================
// MapToEntry Tests
// ============================================================================

func TestMapToEntry_Valid(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("credential.issued", "Credential", "cred-123")
	actor := domain.MustNewActor(domain.ActorTypeUser, domain.MustNewActorID("user-456"))
	subject := domain.MustNewSubject(domain.SubjectTypeCredential, domain.MustNewSubjectID("cred-123"))
	metadata := domain.NewMetadata().Set("key", "value")

	entry, err := mapper.MapToEntry(event, actor, subject, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.EventType().String() != "credential.issued" {
		t.Errorf("EventType() = %v, want %v", entry.EventType().String(), "credential.issued")
	}
	if !entry.Actor().Equals(actor) {
		t.Errorf("Actor mismatch")
	}
	if !entry.Subject().Equals(subject) {
		t.Errorf("Subject mismatch")
	}
	if entry.Metadata().Get("key") != "value" {
		t.Errorf("Metadata mismatch")
	}
	if entry.ID().IsZero() {
		t.Error("expected non-zero ID")
	}
	if entry.OccurredAt().IsZero() {
		t.Error("expected non-zero OccurredAt")
	}
}

func TestMapToEntry_InvalidEventType(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("invalid", "Credential", "cred-123") // No dot separator
	actor := domain.MustNewActor(domain.ActorTypeUser, domain.MustNewActorID("user-456"))
	subject := domain.MustNewSubject(domain.SubjectTypeCredential, domain.MustNewSubjectID("cred-123"))
	metadata := domain.NewMetadata()

	_, err := mapper.MapToEntry(event, actor, subject, metadata)

	if err == nil {
		t.Error("expected error for invalid event type")
	}
}

func TestMapToEntry_PreservesTimestamp(t *testing.T) {
	mapper := NewMapper()
	specificTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	event := &mockDomainEvent{
		eventType:     "user.registered",
		occurredAt:    specificTime.UnixMilli(),
		aggregateID:   "user-123",
		aggregateType: "User",
	}
	actor := domain.MustNewActor(domain.ActorTypeSystem, domain.MustNewActorID("registration"))
	subject := domain.MustNewSubject(domain.SubjectTypeUser, domain.MustNewSubjectID("user-123"))
	metadata := domain.NewMetadata()

	entry, err := mapper.MapToEntry(event, actor, subject, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.OccurredAt().Equal(specificTime) {
		t.Errorf("OccurredAt() = %v, want %v", entry.OccurredAt(), specificTime)
	}
}

// ============================================================================
// MapEvent Tests
// ============================================================================

func TestMapEvent_Valid(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("credential.issued", "Credential", "cred-123")
	meta := domain.EventMetadata{
		UserID:        "user-456",
		CorrelationID: "corr-789",
	}

	entry, err := mapper.MapEvent(event, meta)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.EventType().String() != "credential.issued" {
		t.Errorf("EventType() = %v, want %v", entry.EventType().String(), "credential.issued")
	}
	if entry.Actor().Type() != domain.ActorTypeUser {
		t.Errorf("Actor.Type() = %v, want %v", entry.Actor().Type(), domain.ActorTypeUser)
	}
	if entry.Actor().ID().String() != "user-456" {
		t.Errorf("Actor.ID() = %v, want %v", entry.Actor().ID().String(), "user-456")
	}
	if entry.Subject().Type() != domain.SubjectTypeCredential {
		t.Errorf("Subject.Type() = %v, want %v", entry.Subject().Type(), domain.SubjectTypeCredential)
	}
	if entry.Subject().ID().String() != "cred-123" {
		t.Errorf("Subject.ID() = %v, want %v", entry.Subject().ID().String(), "cred-123")
	}
	if entry.Metadata().CorrelationID() != "corr-789" {
		t.Errorf("CorrelationID() = %v, want %v", entry.Metadata().CorrelationID(), "corr-789")
	}
	if entry.ContextID() != "corr-789" {
		t.Errorf("ContextID() = %v, want %v", entry.ContextID(), "corr-789")
	}
}

func TestMapEvent_SystemActor(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("credential.expired", "Credential", "cred-123")
	meta := domain.EventMetadata{
		// No UserID - should default to system actor
	}

	entry, err := mapper.MapEvent(event, meta)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Actor().Type() != domain.ActorTypeSystem {
		t.Errorf("Actor.Type() = %v, want %v", entry.Actor().Type(), domain.ActorTypeSystem)
	}
}

func TestMapEvent_InvalidEventType(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("invalid", "Credential", "cred-123")
	meta := domain.EventMetadata{
		UserID: "user-456",
	}

	_, err := mapper.MapEvent(event, meta)

	if err == nil {
		t.Error("expected error for invalid event type")
	}
}

func TestMapEvent_EmptyAggregateID(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("credential.issued", "Credential", "")
	meta := domain.EventMetadata{
		UserID: "user-456",
	}

	_, err := mapper.MapEvent(event, meta)

	if err == nil {
		t.Error("expected error for empty aggregate ID")
	}
}

func TestMapEvent_FullMetadata(t *testing.T) {
	mapper := NewMapper()
	event := newMockEvent("user.registered", "User", "user-123")
	meta := domain.EventMetadata{
		UserID:        "admin-001",
		CorrelationID: "corr-123",
		CausationID:   "cause-456",
		TraceID:       "trace-789",
		SpanID:        "span-abc",
		TenantID:      "tenant-xyz",
		Custom: map[string]any{
			"source": "api",
		},
	}

	entry, err := mapper.MapEvent(event, meta)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Metadata().CorrelationID() != "corr-123" {
		t.Errorf("CorrelationID mismatch")
	}
	if entry.Metadata().CausationID() != "cause-456" {
		t.Errorf("CausationID mismatch")
	}
	if entry.Metadata().GetString("trace_id") != "trace-789" {
		t.Errorf("trace_id mismatch")
	}
	if entry.Metadata().GetString("span_id") != "span-abc" {
		t.Errorf("span_id mismatch")
	}
	if entry.Metadata().GetString("tenant_id") != "tenant-xyz" {
		t.Errorf("tenant_id mismatch")
	}
	if entry.Metadata().Get("source") != "api" {
		t.Errorf("custom source mismatch")
	}
}

// ============================================================================
// DefaultSubjectMappings Tests
// ============================================================================

func TestDefaultSubjectMappings_ContainsExpectedTypes(t *testing.T) {
	expectedMappings := map[string]domain.SubjectType{
		"User":         domain.SubjectTypeUser,
		"Credential":   domain.SubjectTypeCredential,
		"Presentation": domain.SubjectTypePresentation,
		"Verification": domain.SubjectTypeVerification,
		"Wallet":       domain.SubjectTypeWallet,
		"Session":      domain.SubjectTypeSession,
		"Organization": domain.SubjectTypeOrganization,
		"DID":          domain.SubjectTypeDID,
		"Vouch":        domain.SubjectTypeVouch,
	}

	for aggregateType, expectedSubjectType := range expectedMappings {
		t.Run(aggregateType, func(t *testing.T) {
			actual, ok := DefaultSubjectMappings[aggregateType]
			if !ok {
				t.Errorf("missing mapping for %s", aggregateType)
				return
			}
			if actual != expectedSubjectType {
				t.Errorf("mapping for %s = %v, want %v", aggregateType, actual, expectedSubjectType)
			}
		})
	}
}
