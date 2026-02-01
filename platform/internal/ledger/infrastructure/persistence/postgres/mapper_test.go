package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
)

// ============================================================================
// ToInsertParams Tests
// ============================================================================

func TestToInsertParams_Success(t *testing.T) {
	entry := createTestAuditEntry(t)

	params, err := ToInsertParams(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.ID != entry.ID().Value().UUID() {
		t.Errorf("ID mismatch: got %v, want %v", params.ID, entry.ID().Value().UUID())
	}
	if !params.OccurredAt.Equal(entry.OccurredAt()) {
		t.Errorf("OccurredAt mismatch: got %v, want %v", params.OccurredAt, entry.OccurredAt())
	}
	if !params.RecordedAt.Equal(entry.RecordedAt()) {
		t.Errorf("RecordedAt mismatch: got %v, want %v", params.RecordedAt, entry.RecordedAt())
	}
	if params.EventType != entry.EventType().String() {
		t.Errorf("EventType mismatch: got %v, want %v", params.EventType, entry.EventType().String())
	}
	if params.ActorID != entry.Actor().ID().String() {
		t.Errorf("ActorID mismatch: got %v, want %v", params.ActorID, entry.Actor().ID().String())
	}
	if params.ActorType != entry.Actor().Type().String() {
		t.Errorf("ActorType mismatch: got %v, want %v", params.ActorType, entry.Actor().Type().String())
	}
	if params.SubjectID != entry.Subject().ID().String() {
		t.Errorf("SubjectID mismatch: got %v, want %v", params.SubjectID, entry.Subject().ID().String())
	}
	if params.SubjectType != entry.Subject().Type().String() {
		t.Errorf("SubjectType mismatch: got %v, want %v", params.SubjectType, entry.Subject().Type().String())
	}
}

func TestToInsertParams_WithContextID(t *testing.T) {
	metadata := domain.NewMetadata().WithCorrelationID("corr-123")
	entry := createTestAuditEntryWithMetadata(t, metadata)

	params, err := ToInsertParams(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.ContextID == nil {
		t.Fatal("expected ContextID to be non-nil")
	}
	if *params.ContextID != "corr-123" {
		t.Errorf("ContextID mismatch: got %v, want %v", *params.ContextID, "corr-123")
	}
}

func TestToInsertParams_WithoutContextID(t *testing.T) {
	entry := createTestAuditEntry(t)

	params, err := ToInsertParams(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.ContextID != nil {
		t.Errorf("expected ContextID to be nil, got %v", *params.ContextID)
	}
}

// ============================================================================
// ToBatchParams Tests
// ============================================================================

func TestToBatchParams_Success(t *testing.T) {
	entry := createTestAuditEntry(t)

	params, err := ToBatchParams(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.ID != entry.ID().Value().UUID() {
		t.Errorf("ID mismatch: got %v, want %v", params.ID, entry.ID().Value().UUID())
	}
	if params.EventType != entry.EventType().String() {
		t.Errorf("EventType mismatch: got %v, want %v", params.EventType, entry.EventType().String())
	}
}

// ============================================================================
// ToDomainEntry Tests
// ============================================================================

func TestToDomainEntry_Success(t *testing.T) {
	row := createTestLedgerEntry()

	entry, err := ToDomainEntry(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.ID().String() != row.ID.String() {
		t.Errorf("ID mismatch: got %v, want %v", entry.ID().String(), row.ID.String())
	}
	if !entry.OccurredAt().Equal(row.OccurredAt) {
		t.Errorf("OccurredAt mismatch: got %v, want %v", entry.OccurredAt(), row.OccurredAt)
	}
	if !entry.RecordedAt().Equal(row.RecordedAt) {
		t.Errorf("RecordedAt mismatch: got %v, want %v", entry.RecordedAt(), row.RecordedAt)
	}
	if entry.EventType().String() != row.EventType {
		t.Errorf("EventType mismatch: got %v, want %v", entry.EventType().String(), row.EventType)
	}
	if entry.Actor().ID().String() != row.ActorID {
		t.Errorf("ActorID mismatch: got %v, want %v", entry.Actor().ID().String(), row.ActorID)
	}
	if entry.Actor().Type().String() != row.ActorType {
		t.Errorf("ActorType mismatch: got %v, want %v", entry.Actor().Type().String(), row.ActorType)
	}
	if entry.Subject().ID().String() != row.SubjectID {
		t.Errorf("SubjectID mismatch: got %v, want %v", entry.Subject().ID().String(), row.SubjectID)
	}
	if entry.Subject().Type().String() != row.SubjectType {
		t.Errorf("SubjectType mismatch: got %v, want %v", entry.Subject().Type().String(), row.SubjectType)
	}
}

func TestToDomainEntry_WithContextID(t *testing.T) {
	contextID := "corr-456"
	row := createTestLedgerEntry()
	row.ContextID = &contextID

	entry, err := ToDomainEntry(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.ContextID() != contextID {
		t.Errorf("ContextID mismatch: got %v, want %v", entry.ContextID(), contextID)
	}
}

func TestToDomainEntry_InvalidActorType(t *testing.T) {
	row := createTestLedgerEntry()
	row.ActorType = "invalid_actor_type"

	_, err := ToDomainEntry(row)
	if err == nil {
		t.Fatal("expected error for invalid actor type")
	}
}

func TestToDomainEntry_InvalidSubjectType(t *testing.T) {
	row := createTestLedgerEntry()
	row.SubjectType = "invalid_subject_type"

	_, err := ToDomainEntry(row)
	if err == nil {
		t.Fatal("expected error for invalid subject type")
	}
}

func TestToDomainEntry_InvalidEventType(t *testing.T) {
	row := createTestLedgerEntry()
	row.EventType = "invalid" // missing dot separator

	_, err := ToDomainEntry(row)
	if err == nil {
		t.Fatal("expected error for invalid event type")
	}
}

// ============================================================================
// ToEntryView Tests
// ============================================================================

func TestToEntryView_Success(t *testing.T) {
	row := createTestLedgerEntry()

	view := ToEntryView(row)

	if view.ID != row.ID.String() {
		t.Errorf("ID mismatch: got %v, want %v", view.ID, row.ID.String())
	}
	if !view.OccurredAt.Equal(row.OccurredAt) {
		t.Errorf("OccurredAt mismatch: got %v, want %v", view.OccurredAt, row.OccurredAt)
	}
	if !view.RecordedAt.Equal(row.RecordedAt) {
		t.Errorf("RecordedAt mismatch: got %v, want %v", view.RecordedAt, row.RecordedAt)
	}
	if view.EventType != row.EventType {
		t.Errorf("EventType mismatch: got %v, want %v", view.EventType, row.EventType)
	}
	if view.ActorID != row.ActorID {
		t.Errorf("ActorID mismatch: got %v, want %v", view.ActorID, row.ActorID)
	}
	if view.ActorType != row.ActorType {
		t.Errorf("ActorType mismatch: got %v, want %v", view.ActorType, row.ActorType)
	}
	if view.SubjectID != row.SubjectID {
		t.Errorf("SubjectID mismatch: got %v, want %v", view.SubjectID, row.SubjectID)
	}
	if view.SubjectType != row.SubjectType {
		t.Errorf("SubjectType mismatch: got %v, want %v", view.SubjectType, row.SubjectType)
	}
}

func TestToEntryView_WithContextID(t *testing.T) {
	contextID := "corr-789"
	row := createTestLedgerEntry()
	row.ContextID = &contextID

	view := ToEntryView(row)

	if view.ContextID != contextID {
		t.Errorf("ContextID mismatch: got %v, want %v", view.ContextID, contextID)
	}
}

func TestToEntryView_WithMetadata(t *testing.T) {
	row := createTestLedgerEntry()
	row.Metadata = json.RawMessage(`{"key": "value", "count": 42}`)

	view := ToEntryView(row)

	if view.Metadata == nil {
		t.Fatal("expected Metadata to be non-nil")
	}
	if view.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] mismatch: got %v, want %v", view.Metadata["key"], "value")
	}
	// JSON numbers are unmarshaled as float64
	if view.Metadata["count"] != float64(42) {
		t.Errorf("Metadata[count] mismatch: got %v, want %v", view.Metadata["count"], float64(42))
	}
}

// ============================================================================
// ToEntryViews Tests
// ============================================================================

func TestToEntryViews_Success(t *testing.T) {
	rows := []generated.LedgerEntry{
		createTestLedgerEntry(),
		createTestLedgerEntry(),
		createTestLedgerEntry(),
	}

	views := ToEntryViews(rows)

	if len(views) != len(rows) {
		t.Errorf("length mismatch: got %v, want %v", len(views), len(rows))
	}

	for i, view := range views {
		if view.ID != rows[i].ID.String() {
			t.Errorf("views[%d].ID mismatch: got %v, want %v", i, view.ID, rows[i].ID.String())
		}
	}
}

func TestToEntryViews_Empty(t *testing.T) {
	rows := []generated.LedgerEntry{}

	views := ToEntryViews(rows)

	if len(views) != 0 {
		t.Errorf("expected empty slice, got %v", len(views))
	}
}

// ============================================================================
// ToVerificationLogEntry Tests
// ============================================================================

func TestToVerificationLogEntry_Success(t *testing.T) {
	row := createTestLedgerEntry()
	row.ActorType = "external_verifier"
	row.SubjectType = "credential"
	row.Metadata = json.RawMessage(`{"outcome": "success"}`)

	entry := ToVerificationLogEntry(row)

	if entry.ID != row.ID.String() {
		t.Errorf("ID mismatch: got %v, want %v", entry.ID, row.ID.String())
	}
	if entry.VerifierID != row.ActorID {
		t.Errorf("VerifierID mismatch: got %v, want %v", entry.VerifierID, row.ActorID)
	}
	if entry.VerifierType != row.ActorType {
		t.Errorf("VerifierType mismatch: got %v, want %v", entry.VerifierType, row.ActorType)
	}
	if entry.CredentialID != row.SubjectID {
		t.Errorf("CredentialID mismatch: got %v, want %v", entry.CredentialID, row.SubjectID)
	}
	if entry.Outcome != "success" {
		t.Errorf("Outcome mismatch: got %v, want %v", entry.Outcome, "success")
	}
}

func TestToVerificationLogEntry_NoOutcome(t *testing.T) {
	row := createTestLedgerEntry()
	row.Metadata = json.RawMessage(`{"other": "data"}`)

	entry := ToVerificationLogEntry(row)

	if entry.Outcome != "" {
		t.Errorf("expected empty Outcome, got %v", entry.Outcome)
	}
}

// ============================================================================
// ToVerificationLogEntries Tests
// ============================================================================

func TestToVerificationLogEntries_Success(t *testing.T) {
	rows := []generated.LedgerEntry{
		createTestLedgerEntry(),
		createTestLedgerEntry(),
	}

	entries := ToVerificationLogEntries(rows)

	if len(entries) != len(rows) {
		t.Errorf("length mismatch: got %v, want %v", len(entries), len(rows))
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

func createTestAuditEntry(t *testing.T) *domain.AuditEntry {
	t.Helper()
	return createTestAuditEntryWithMetadata(t, domain.NewMetadata())
}

func createTestAuditEntryWithMetadata(t *testing.T, metadata domain.Metadata) *domain.AuditEntry {
	t.Helper()

	entry, err := domain.NewAuditEntryBuilder().
		WithNewID().
		WithOccurredNow().
		WithEventTypeString("credential.issued").
		WithUserActor("user-123").
		WithSubjectCredential("cred-456").
		WithMetadata(metadata).
		Build()

	if err != nil {
		t.Fatalf("failed to create test audit entry: %v", err)
	}

	return entry
}

func createTestLedgerEntry() generated.LedgerEntry {
	now := time.Now().UTC()
	return generated.LedgerEntry{
		ID:          uuid.New(),
		OccurredAt:  now,
		RecordedAt:  now,
		EventType:   "credential.issued",
		ActorID:     "user-123",
		ActorType:   "user",
		SubjectID:   "cred-456",
		SubjectType: "credential",
		Metadata:    json.RawMessage(`{}`),
		ContextID:   nil,
	}
}
