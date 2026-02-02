package domain

import (
	"testing"
	"time"
)

// ============================================================================
// NewAuditEntry Tests
// ============================================================================

func TestNewAuditEntry_Valid(t *testing.T) {
	id := NewEntryID()
	occurredAt := time.Now().UTC()
	eventType := MustNewEventType("credential.issued")
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata().Set("key", "value")

	entry, err := NewAuditEntry(id, occurredAt, eventType, actor, subject, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.ID().Equals(id) {
		t.Errorf("ID mismatch: got %v, want %v", entry.ID(), id)
	}
	if !entry.OccurredAt().Equal(occurredAt) {
		t.Errorf("OccurredAt mismatch: got %v, want %v", entry.OccurredAt(), occurredAt)
	}
	if entry.RecordedAt().IsZero() {
		t.Error("RecordedAt should be set")
	}
	if !entry.EventType().Equals(eventType) {
		t.Errorf("EventType mismatch: got %v, want %v", entry.EventType(), eventType)
	}
	if !entry.Actor().Equals(actor) {
		t.Errorf("Actor mismatch: got %v, want %v", entry.Actor(), actor)
	}
	if !entry.Subject().Equals(subject) {
		t.Errorf("Subject mismatch: got %v, want %v", entry.Subject(), subject)
	}
	if entry.Metadata().Get("key") != "value" {
		t.Errorf("Metadata mismatch: got %v, want %v", entry.Metadata().Get("key"), "value")
	}
}

func TestNewAuditEntry_WithCorrelationID(t *testing.T) {
	id := NewEntryID()
	occurredAt := time.Now().UTC()
	eventType := MustNewEventType("credential.issued")
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata().WithCorrelationID("corr-789")

	entry, err := NewAuditEntry(id, occurredAt, eventType, actor, subject, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ContextID() != "corr-789" {
		t.Errorf("ContextID mismatch: got %v, want %v", entry.ContextID(), "corr-789")
	}
}

func TestNewAuditEntry_ZeroID(t *testing.T) {
	var zeroID EntryID
	occurredAt := time.Now().UTC()
	eventType := MustNewEventType("credential.issued")
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata()

	_, err := NewAuditEntry(zeroID, occurredAt, eventType, actor, subject, metadata)

	if err == nil {
		t.Error("expected error for zero ID")
	}
}

func TestNewAuditEntry_ZeroOccurredAt(t *testing.T) {
	id := NewEntryID()
	var zeroTime time.Time
	eventType := MustNewEventType("credential.issued")
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata()

	_, err := NewAuditEntry(id, zeroTime, eventType, actor, subject, metadata)

	if err == nil {
		t.Error("expected error for zero occurred_at")
	}
}

func TestNewAuditEntry_ZeroEventType(t *testing.T) {
	id := NewEntryID()
	occurredAt := time.Now().UTC()
	var zeroEventType EventType
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata()

	_, err := NewAuditEntry(id, occurredAt, zeroEventType, actor, subject, metadata)

	if err == nil {
		t.Error("expected error for zero event_type")
	}
}

func TestNewAuditEntry_ZeroActor(t *testing.T) {
	id := NewEntryID()
	occurredAt := time.Now().UTC()
	eventType := MustNewEventType("credential.issued")
	var zeroActor Actor
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata()

	_, err := NewAuditEntry(id, occurredAt, eventType, zeroActor, subject, metadata)

	if err == nil {
		t.Error("expected error for zero actor")
	}
}

func TestNewAuditEntry_ZeroSubject(t *testing.T) {
	id := NewEntryID()
	occurredAt := time.Now().UTC()
	eventType := MustNewEventType("credential.issued")
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	var zeroSubject Subject
	metadata := NewMetadata()

	_, err := NewAuditEntry(id, occurredAt, eventType, actor, zeroSubject, metadata)

	if err == nil {
		t.Error("expected error for zero subject")
	}
}

// ============================================================================
// Reconstitute Tests
// ============================================================================

func TestReconstitute(t *testing.T) {
	id := NewEntryID()
	occurredAt := time.Now().Add(-time.Hour).UTC()
	recordedAt := time.Now().UTC()
	eventType := MustNewEventType("credential.issued")
	actor := MustNewActor(ActorTypeUser, MustNewActorID("user-123"))
	subject := MustNewSubject(SubjectTypeCredential, MustNewSubjectID("cred-456"))
	metadata := NewMetadata().Set("key", "value")
	contextID := "corr-123"

	entry := Reconstitute(id, occurredAt, recordedAt, eventType, actor, subject, metadata, contextID)

	if !entry.ID().Equals(id) {
		t.Errorf("ID mismatch")
	}
	if !entry.OccurredAt().Equal(occurredAt) {
		t.Errorf("OccurredAt mismatch")
	}
	if !entry.RecordedAt().Equal(recordedAt) {
		t.Errorf("RecordedAt mismatch")
	}
	if !entry.EventType().Equals(eventType) {
		t.Errorf("EventType mismatch")
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
	if entry.ContextID() != contextID {
		t.Errorf("ContextID mismatch: got %v, want %v", entry.ContextID(), contextID)
	}
}

// ============================================================================
// Builder Tests
// ============================================================================

func TestAuditEntryBuilder_Build_Valid(t *testing.T) {
	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID().IsZero() {
		t.Error("expected non-zero ID")
	}
	if entry.OccurredAt().IsZero() {
		t.Error("expected non-zero OccurredAt")
	}
	if entry.EventType().String() != "credential.issued" {
		t.Errorf("EventType mismatch: got %v, want %v", entry.EventType().String(), "credential.issued")
	}
	if entry.Actor().Type() != ActorTypeUser {
		t.Errorf("Actor type mismatch: got %v, want %v", entry.Actor().Type(), ActorTypeUser)
	}
	if entry.Actor().ID().String() != "user-123" {
		t.Errorf("Actor ID mismatch: got %v, want %v", entry.Actor().ID().String(), "user-123")
	}
	if entry.Subject().Type() != SubjectTypeCredential {
		t.Errorf("Subject type mismatch: got %v, want %v", entry.Subject().Type(), SubjectTypeCredential)
	}
	if entry.Subject().ID().String() != "cred-456" {
		t.Errorf("Subject ID mismatch: got %v, want %v", entry.Subject().ID().String(), "cred-456")
	}
}

func TestAuditEntryBuilder_WithID(t *testing.T) {
	id := NewEntryID()

	entry, err := NewAuditEntryBuilder().
		WithID(id).
		WithOccurredNow().
		WithEventTypeString("user.registered").
		WithSystemActor("registration-service").
		WithSubjectUser("user-789").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.ID().Equals(id) {
		t.Errorf("ID mismatch: got %v, want %v", entry.ID(), id)
	}
}

func TestAuditEntryBuilder_WithOccurredAt(t *testing.T) {
	specificTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredAt(specificTime).
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.OccurredAt().Equal(specificTime) {
		t.Errorf("OccurredAt mismatch: got %v, want %v", entry.OccurredAt(), specificTime)
	}
}

func TestAuditEntryBuilder_WithEventType(t *testing.T) {
	eventType := MustNewEventType("wallet.linked")

	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventType(eventType).
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.EventType().Equals(eventType) {
		t.Errorf("EventType mismatch: got %v, want %v", entry.EventType(), eventType)
	}
}

func TestAuditEntryBuilder_WithActor(t *testing.T) {
	actor := MustNewActor(ActorTypeAdmin, MustNewActorID("admin-001"))

	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.revoked").
		WithActor(actor).
		WithSubjectCredential("cred-456").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.Actor().Equals(actor) {
		t.Errorf("Actor mismatch: got %v, want %v", entry.Actor(), actor)
	}
}

func TestAuditEntryBuilder_WithSystemActor(t *testing.T) {
	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.expired").
		WithSystemActor("expiration-worker").
		WithSubjectCredential("cred-456").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Actor().Type() != ActorTypeSystem {
		t.Errorf("Actor type mismatch: got %v, want %v", entry.Actor().Type(), ActorTypeSystem)
	}
	if entry.Actor().ID().String() != "expiration-worker" {
		t.Errorf("Actor ID mismatch: got %v, want %v", entry.Actor().ID().String(), "expiration-worker")
	}
}

func TestAuditEntryBuilder_WithSubject(t *testing.T) {
	subject := MustNewSubject(SubjectTypeWallet, MustNewSubjectID("wallet-789"))

	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("wallet.linked").
		WithUserActor("user-123").
		WithSubject(subject).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !entry.Subject().Equals(subject) {
		t.Errorf("Subject mismatch: got %v, want %v", entry.Subject(), subject)
	}
}

func TestAuditEntryBuilder_WithSubjectUser(t *testing.T) {
	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("user.registered").
		WithSystemActor("registration").
		WithSubjectUser("user-new").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Subject().Type() != SubjectTypeUser {
		t.Errorf("Subject type mismatch: got %v, want %v", entry.Subject().Type(), SubjectTypeUser)
	}
	if entry.Subject().ID().String() != "user-new" {
		t.Errorf("Subject ID mismatch: got %v, want %v", entry.Subject().ID().String(), "user-new")
	}
}

func TestAuditEntryBuilder_WithMetadata(t *testing.T) {
	metadata := NewMetadata().
		Set("key1", "value1").
		Set("key2", 42)

	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		WithMetadata(metadata).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Metadata().Get("key1") != "value1" {
		t.Errorf("Metadata key1 mismatch")
	}
	if entry.Metadata().Get("key2") != 42 {
		t.Errorf("Metadata key2 mismatch")
	}
}

func TestAuditEntryBuilder_WithMeta(t *testing.T) {
	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		WithMeta("reason", "initial issuance").
		WithMeta("schema_id", "schema-001").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Metadata().Get("reason") != "initial issuance" {
		t.Errorf("Metadata reason mismatch")
	}
	if entry.Metadata().Get("schema_id") != "schema-001" {
		t.Errorf("Metadata schema_id mismatch")
	}
}

func TestAuditEntryBuilder_WithCorrelationID(t *testing.T) {
	entry, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		WithCorrelationID("corr-999").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ContextID() != "corr-999" {
		t.Errorf("ContextID mismatch: got %v, want %v", entry.ContextID(), "corr-999")
	}
	if entry.Metadata().CorrelationID() != "corr-999" {
		t.Errorf("Metadata CorrelationID mismatch")
	}
}

func TestAuditEntryBuilder_Build_MissingID(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err == nil {
		t.Error("expected error for missing ID")
	}
}

func TestAuditEntryBuilder_Build_MissingOccurredAt(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err == nil {
		t.Error("expected error for missing OccurredAt")
	}
}

func TestAuditEntryBuilder_Build_MissingEventType(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err == nil {
		t.Error("expected error for missing EventType")
	}
}

func TestAuditEntryBuilder_Build_MissingActor(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithSubjectCredential("cred-456").
		Build()

	if err == nil {
		t.Error("expected error for missing Actor")
	}
}

func TestAuditEntryBuilder_Build_MissingSubject(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		Build()

	if err == nil {
		t.Error("expected error for missing Subject")
	}
}

func TestAuditEntryBuilder_MustBuild_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	entry := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		MustBuild()

	if entry == nil {
		t.Error("expected non-nil entry")
	}
}

func TestAuditEntryBuilder_MustBuild_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid entry")
		}
	}()

	NewAuditEntryBuilder().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		MustBuild() // Missing ID
}

// ============================================================================
// Builder with Invalid Inputs Tests
// ============================================================================

func TestAuditEntryBuilder_WithEventTypeString_Invalid(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("invalid"). // No dot separator
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		Build()

	if err == nil {
		t.Error("expected error for invalid event type string")
	}
}

func TestAuditEntryBuilder_WithUserActor_EmptyID(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor(""). // Empty user ID
		WithSubjectCredential("cred-456").
		Build()

	if err == nil {
		t.Error("expected error for empty user actor ID")
	}
}

func TestAuditEntryBuilder_WithSubjectCredential_EmptyID(t *testing.T) {
	_, err := NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential(""). // Empty credential ID
		Build()

	if err == nil {
		t.Error("expected error for empty subject credential ID")
	}
}
