//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/internal/ledger/app/query"
	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	"github.com/google/uuid"
)

// ============================================================================
// Writer Integration Tests
// ============================================================================

func TestWriter_Append(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	writer := NewWriter(db.queries)
	reader := NewReader(db.queries)

	// Create a domain entry
	entryID := domain.NewEntryID()
	now := time.Now().UTC().Truncate(time.Microsecond)

	actor, err := domain.NewActor(domain.ActorTypeUser, domain.MustNewActorID("user-123"))
	if err != nil {
		t.Fatalf("failed to create actor: %v", err)
	}

	subject, err := domain.NewSubject(domain.SubjectTypeCredential, domain.MustNewSubjectID("cred-456"))
	if err != nil {
		t.Fatalf("failed to create subject: %v", err)
	}

	eventType, err := domain.NewEventType("credential.issued")
	if err != nil {
		t.Fatalf("failed to create event type: %v", err)
	}

	entry, err := domain.NewAuditEntry(
		entryID,
		now,
		eventType,
		actor,
		subject,
		domain.NewMetadata().Set("reason", "test"),
		"",
	)
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	// Append entry
	ctx := context.Background()
	if err := writer.Append(ctx, entry); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	// Verify it was persisted
	view, err := reader.GetByID(ctx, entryID.String())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if view.ID != entryID.String() {
		t.Errorf("ID mismatch: got %v, want %v", view.ID, entryID.String())
	}
	if view.EventType != "credential.issued" {
		t.Errorf("EventType mismatch: got %v, want %v", view.EventType, "credential.issued")
	}
	if view.ActorID != "user-123" {
		t.Errorf("ActorID mismatch: got %v, want %v", view.ActorID, "user-123")
	}
	if view.SubjectID != "cred-456" {
		t.Errorf("SubjectID mismatch: got %v, want %v", view.SubjectID, "cred-456")
	}
}

func TestWriter_AppendBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	writer := NewWriter(db.queries)
	reader := NewReader(db.queries)

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create multiple entries
	entries := make([]*domain.AuditEntry, 3)
	for i := 0; i < 3; i++ {
		entryID := domain.NewEntryID()
		actor, _ := domain.NewActor(domain.ActorTypeUser, domain.MustNewActorID("user-123"))
		subject, _ := domain.NewSubject(domain.SubjectTypeCredential, domain.MustNewSubjectID("cred-"+string(rune('A'+i))))
		eventType, _ := domain.NewEventType("credential.issued")

		entry, err := domain.NewAuditEntry(
			entryID,
			now.Add(time.Duration(i)*time.Second),
			eventType,
			actor,
			subject,
			domain.NewMetadata(),
			"",
		)
		if err != nil {
			t.Fatalf("failed to create entry %d: %v", i, err)
		}
		entries[i] = entry
	}

	// Append batch
	if err := writer.AppendBatch(ctx, entries); err != nil {
		t.Fatalf("AppendBatch() error = %v", err)
	}

	// Verify all were persisted
	activity, err := reader.List(ctx, query.ListFilter{PageSize: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 3 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 3)
	}
}

func TestWriter_AppendBatch_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	writer := NewWriter(db.queries)
	ctx := context.Background()

	// Empty batch should succeed
	if err := writer.AppendBatch(ctx, []*domain.AuditEntry{}); err != nil {
		t.Fatalf("AppendBatch() with empty slice should succeed, got error = %v", err)
	}
}

// ============================================================================
// Reader Integration Tests
// ============================================================================

func TestReader_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert test data
	entryID := uuid.New()
	params := newEntryBuilder().
		withID(entryID.String()).
		withEventType("credential.issued").
		withActor("user-123", "user").
		withSubject("cred-456", "credential").
		build(t)
	db.insertTestEntry(t, params)

	// Test retrieval
	view, err := reader.GetByID(ctx, entryID.String())
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if view.ID != entryID.String() {
		t.Errorf("ID mismatch: got %v, want %v", view.ID, entryID.String())
	}
	if view.EventType != "credential.issued" {
		t.Errorf("EventType mismatch: got %v", view.EventType)
	}
}

func TestReader_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	_, err := reader.GetByID(ctx, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for non-existent entry")
	}

	// Should be a not found error
	if !isNotFoundError(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestReader_GetByID_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	_, err := reader.GetByID(ctx, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestReader_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert test data
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withOccurredAt(now.Add(time.Duration(i)*time.Minute)).
			withEventType("credential.issued").
			withActor("user-123", "user").
			withSubject("cred-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Test list
	activity, err := reader.List(ctx, query.ListFilter{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 5 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 5)
	}
	if len(activity.Entries) != 5 {
		t.Errorf("Entries count mismatch: got %v, want %v", len(activity.Entries), 5)
	}
}

func TestReader_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert 10 entries
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withOccurredAt(now.Add(time.Duration(i) * time.Minute)).
			withEventType("credential.issued").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Page 1
	page1, err := reader.List(ctx, query.ListFilter{Page: 1, PageSize: 3})
	if err != nil {
		t.Fatalf("List() page 1 error = %v", err)
	}

	if len(page1.Entries) != 3 {
		t.Errorf("Page 1 entries: got %v, want %v", len(page1.Entries), 3)
	}
	if !page1.HasMore {
		t.Error("Page 1 should have more")
	}
	if page1.TotalCount != 10 {
		t.Errorf("TotalCount: got %v, want %v", page1.TotalCount, 10)
	}

	// Page 4 (last page with 1 entry)
	page4, err := reader.List(ctx, query.ListFilter{Page: 4, PageSize: 3})
	if err != nil {
		t.Fatalf("List() page 4 error = %v", err)
	}

	if len(page4.Entries) != 1 {
		t.Errorf("Page 4 entries: got %v, want %v", len(page4.Entries), 1)
	}
	if page4.HasMore {
		t.Error("Page 4 should not have more")
	}
}

func TestReader_List_FilterByEventType(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert mixed event types
	eventTypes := []string{"credential.issued", "credential.revoked", "user.registered"}
	for i, et := range eventTypes {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withEventType(et).
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Filter by credential.issued
	activity, err := reader.List(ctx, query.ListFilter{
		EventTypes: []string{"credential.issued"},
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 1 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 1)
	}
	if len(activity.Entries) > 0 && activity.Entries[0].EventType != "credential.issued" {
		t.Errorf("EventType mismatch: got %v", activity.Entries[0].EventType)
	}
}

func TestReader_List_FilterByActor(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries with different actors
	actors := []string{"user-1", "user-2", "user-1"}
	for i, actor := range actors {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withActor(actor, "user").
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Filter by user-1
	activity, err := reader.List(ctx, query.ListFilter{
		ActorID:  "user-1",
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 2)
	}
}

func TestReader_List_FilterBySubject(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries with different subjects
	for i := 0; i < 3; i++ {
		subjectID := "cred-A"
		if i == 1 {
			subjectID = "cred-B"
		}
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withSubject(subjectID, "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Filter by cred-A
	activity, err := reader.List(ctx, query.ListFilter{
		SubjectID:   "cred-A",
		SubjectType: "credential",
		PageSize:    10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 2)
	}
}

func TestReader_List_FilterByTimeRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries at different times
	baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withOccurredAt(baseTime.Add(time.Duration(i)*24*time.Hour)).
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Filter by time range (days 1-3)
	fromTime := baseTime.Add(24 * time.Hour)
	toTime := baseTime.Add(4 * 24 * time.Hour)

	activity, err := reader.List(ctx, query.ListFilter{
		FromTime: fromTime,
		ToTime:   toTime,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 3 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 3)
	}
}

func TestReader_GetBySubject(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries for different subjects
	for i := 0; i < 3; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withEventType("credential.issued").
			withSubject("cred-target", "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Insert entry for different subject
	params := newEntryBuilder().
		withID(uuid.New().String()).
		withSubject("cred-other", "credential").
		build(t)
	db.insertTestEntry(t, params)

	// Get history for target credential
	history, err := reader.GetBySubject(ctx, query.SubjectFilter{
		SubjectID:   "cred-target",
		SubjectType: "credential",
		PageSize:    10,
	})
	if err != nil {
		t.Fatalf("GetBySubject() error = %v", err)
	}

	if history.TotalCount != 3 {
		t.Errorf("TotalCount mismatch: got %v, want %v", history.TotalCount, 3)
	}
	if history.SubjectID != "cred-target" {
		t.Errorf("SubjectID mismatch: got %v", history.SubjectID)
	}
}

func TestReader_GetByActor(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries for target actor
	for i := 0; i < 3; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withActor("user-target", "user").
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Insert entry for different actor
	params := newEntryBuilder().
		withID(uuid.New().String()).
		withActor("user-other", "user").
		withSubject("subject-X", "credential").
		build(t)
	db.insertTestEntry(t, params)

	// Get activity for target user
	activity, err := reader.GetByActor(ctx, query.ActorFilter{
		ActorID:   "user-target",
		ActorType: "user",
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("GetByActor() error = %v", err)
	}

	if activity.TotalCount != 3 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 3)
	}
}

func TestReader_GetStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries with various event types
	eventTypes := []string{
		"credential.issued",
		"credential.issued",
		"credential.revoked",
		"user.registered",
	}
	for i, et := range eventTypes {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withEventType(et).
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Get stats
	stats, err := reader.GetStats(ctx, query.StatsFilter{})
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalEntries != 4 {
		t.Errorf("TotalEntries mismatch: got %v, want %v", stats.TotalEntries, 4)
	}

	// Verify event type counts
	countMap := make(map[string]int)
	for _, c := range stats.EventTypeCounts {
		countMap[c.EventType] = c.Count
	}

	if countMap["credential.issued"] != 2 {
		t.Errorf("credential.issued count: got %v, want %v", countMap["credential.issued"], 2)
	}
	if countMap["credential.revoked"] != 1 {
		t.Errorf("credential.revoked count: got %v, want %v", countMap["credential.revoked"], 1)
	}
	if countMap["user.registered"] != 1 {
		t.Errorf("user.registered count: got %v, want %v", countMap["user.registered"], 1)
	}
}

func TestReader_GetStats_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries for different actors
	for i := 0; i < 3; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withActor("user-target", "user").
			withEventType("credential.issued").
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	params := newEntryBuilder().
		withID(uuid.New().String()).
		withActor("user-other", "user").
		withEventType("credential.issued").
		withSubject("subject-X", "credential").
		build(t)
	db.insertTestEntry(t, params)

	// Get stats for target actor only
	stats, err := reader.GetStats(ctx, query.StatsFilter{
		ActorID: "user-target",
	})
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalEntries != 3 {
		t.Errorf("TotalEntries mismatch: got %v, want %v", stats.TotalEntries, 3)
	}
}

func TestReader_GetVerificationLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert verification events for target user
	for i := 0; i < 3; i++ {
		outcome := "success"
		if i == 1 {
			outcome = "failed"
		}
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withEventType("verification.completed").
			withActor("user-target", "user").
			withSubject("cred-"+string(rune('A'+i)), "credential").
			withMetadata("outcome", outcome).
			build(t)
		db.insertTestEntry(t, params)
	}

	// Insert for different user
	params := newEntryBuilder().
		withID(uuid.New().String()).
		withEventType("verification.completed").
		withActor("user-other", "user").
		withSubject("cred-X", "credential").
		build(t)
	db.insertTestEntry(t, params)

	// Get verification log
	log, err := reader.GetVerificationLog(ctx, query.VerificationLogFilter{
		UserID:   "user-target",
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetVerificationLog() error = %v", err)
	}

	if log.TotalCount != 3 {
		t.Errorf("TotalCount mismatch: got %v, want %v", log.TotalCount, 3)
	}
	if log.UserID != "user-target" {
		t.Errorf("UserID mismatch: got %v", log.UserID)
	}
}

func TestReader_List_WithContextID(t *testing.T) {
	db := setupTestDB(t)
	defer db.close()
	defer db.cleanup(t)

	reader := NewReader(db.queries)
	ctx := context.Background()

	// Insert entries with context ID
	for i := 0; i < 2; i++ {
		params := newEntryBuilder().
			withID(uuid.New().String()).
			withContextID("corr-123").
			withSubject("subject-"+string(rune('A'+i)), "credential").
			build(t)
		db.insertTestEntry(t, params)
	}

	// Insert entry without context ID
	params := newEntryBuilder().
		withID(uuid.New().String()).
		withSubject("subject-X", "credential").
		build(t)
	db.insertTestEntry(t, params)

	// Filter by context ID
	activity, err := reader.List(ctx, query.ListFilter{
		ContextID: "corr-123",
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if activity.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v, want %v", activity.TotalCount, 2)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func isNotFoundError(err error) bool {
	// Check if the error message contains "not found"
	return err != nil && (err.Error() == "entry not found" ||
		contains(err.Error(), "not found") ||
		contains(err.Error(), "LEDGER_ENTRY_NOT_FOUND"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
