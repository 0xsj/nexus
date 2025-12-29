package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing/memory"
)

// ============================================================================
// Test Helpers
// ============================================================================

func createTestEnvelope(aggregateType, aggregateID, eventType string, version int) *eventsourcing.EventEnvelope {
	return &eventsourcing.EventEnvelope{
		ID:            "event-" + eventType,
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       version,
		Timestamp:     time.Now().UTC(),
		Data:          []byte(`{}`),
	}
}

// ============================================================================
// Store Tests
// ============================================================================

func TestNewStore(t *testing.T) {
	store := memory.NewStore()
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestStore_Save(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}

	err := store.Save(ctx, "user-123", events, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.EventCount() != 1 {
		t.Errorf("expected 1 event, got %d", store.EventCount())
	}
}

func TestStore_Save_MultipleEvents(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
		createTestEnvelope("User", "user-123", "User.EmailChanged", 2),
		createTestEnvelope("User", "user-123", "User.Deleted", 3),
	}

	err := store.Save(ctx, "user-123", events, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.EventCount() != 3 {
		t.Errorf("expected 3 events, got %d", store.EventCount())
	}
}

func TestStore_Save_EmptyEvents(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	err := store.Save(ctx, "user-123", []*eventsourcing.EventEnvelope{}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.EventCount() != 0 {
		t.Errorf("expected 0 events, got %d", store.EventCount())
	}
}

func TestStore_Save_ConcurrencyConflict(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Save first event
	events1 := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}
	err := store.Save(ctx, "user-123", events1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to save with wrong expected version
	events2 := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.EmailChanged", 2),
	}
	err = store.Save(ctx, "user-123", events2, 0) // Expected 0, but actual is 1
	if err == nil {
		t.Fatal("expected concurrency conflict error")
	}

	if !eventsourcing.IsConcurrencyConflict(err) {
		t.Errorf("expected concurrency conflict error, got: %v", err)
	}
}

func TestStore_Save_CorrectVersion(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Save first event
	events1 := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}
	store.Save(ctx, "user-123", events1, 0)

	// Save second event with correct expected version
	events2 := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.EmailChanged", 2),
	}
	err := store.Save(ctx, "user-123", events2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.EventCount() != 2 {
		t.Errorf("expected 2 events, got %d", store.EventCount())
	}
}

func TestStore_Save_ContextCanceled(t *testing.T) {
	store := memory.NewStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}

	err := store.Save(ctx, "user-123", events, 0)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestStore_Load(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Save events
	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
		createTestEnvelope("User", "user-123", "User.EmailChanged", 2),
	}
	store.Save(ctx, "user-123", events, 0)

	// Load events
	loaded, err := store.Load(ctx, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(loaded))
	}

	if loaded[0].Type != "User.Created" {
		t.Errorf("expected first event type 'User.Created', got '%s'", loaded[0].Type)
	}

	if loaded[1].Type != "User.EmailChanged" {
		t.Errorf("expected second event type 'User.EmailChanged', got '%s'", loaded[1].Type)
	}
}

func TestStore_Load_NotFound(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	loaded, err := store.Load(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("expected empty slice, got %d events", len(loaded))
	}
}

func TestStore_Load_ReturnsCopy(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}
	store.Save(ctx, "user-123", events, 0)

	// Load twice
	loaded1, _ := store.Load(ctx, "user-123")
	loaded2, _ := store.Load(ctx, "user-123")

	// Modify first load
	loaded1[0].Type = "Modified"

	// Second load should be unaffected
	if loaded2[0].Type == "Modified" {
		t.Error("expected Load to return a copy, but modification affected other loads")
	}
}

func TestStore_LoadFrom(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
		createTestEnvelope("User", "user-123", "User.EmailChanged", 2),
		createTestEnvelope("User", "user-123", "User.Deleted", 3),
	}
	store.Save(ctx, "user-123", events, 0)

	// Load from version 2
	loaded, err := store.LoadFrom(ctx, "user-123", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(loaded))
	}

	if loaded[0].Version != 2 {
		t.Errorf("expected first event version 2, got %d", loaded[0].Version)
	}

	if loaded[1].Version != 3 {
		t.Errorf("expected second event version 3, got %d", loaded[1].Version)
	}
}

func TestStore_LoadFrom_BeyondVersion(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	events := []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}
	store.Save(ctx, "user-123", events, 0)

	loaded, err := store.LoadFrom(ctx, "user-123", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("expected 0 events, got %d", len(loaded))
	}
}

func TestStore_LoadByType(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Save events for different aggregates
	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
	}, 0)

	store.Save(ctx, "user-2", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-2", "User.Created", 1),
	}, 0)

	store.Save(ctx, "order-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("Order", "order-1", "Order.Created", 1),
	}, 0)

	// Load by type
	loaded, err := store.LoadByType(ctx, "User", eventsourcing.DefaultLoadOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 User events, got %d", len(loaded))
	}
}

func TestStore_LoadByType_WithLimit(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Save multiple events
	for i := 1; i <= 5; i++ {
		store.Save(ctx, "user-"+string(rune('0'+i)), []*eventsourcing.EventEnvelope{
			createTestEnvelope("User", "user-"+string(rune('0'+i)), "User.Created", 1),
		}, 0)
	}

	// Load with limit
	opts := eventsourcing.DefaultLoadOptions().WithLimit(2)
	loaded, err := store.LoadByType(ctx, "User", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 events with limit, got %d", len(loaded))
	}
}

func TestStore_LoadByType_WithEventTypes(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
		createTestEnvelope("User", "user-1", "User.EmailChanged", 2),
		createTestEnvelope("User", "user-1", "User.Deleted", 3),
	}, 0)

	// Load only specific event types
	opts := eventsourcing.DefaultLoadOptions().WithEventTypes("User.Created", "User.Deleted")
	loaded, err := store.LoadByType(ctx, "User", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 events, got %d", len(loaded))
	}

	for _, e := range loaded {
		if e.Type != "User.Created" && e.Type != "User.Deleted" {
			t.Errorf("unexpected event type: %s", e.Type)
		}
	}
}

func TestStore_LoadByType_WithVersionRange(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
		createTestEnvelope("User", "user-1", "User.EmailChanged", 2),
		createTestEnvelope("User", "user-1", "User.NameChanged", 3),
		createTestEnvelope("User", "user-1", "User.Deleted", 4),
	}, 0)

	// Load version range 2-3
	opts := eventsourcing.DefaultLoadOptions().WithFromVersion(2).WithToVersion(3)
	loaded, err := store.LoadByType(ctx, "User", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 events, got %d", len(loaded))
	}

	for _, e := range loaded {
		if e.Version < 2 || e.Version > 3 {
			t.Errorf("expected version 2-3, got %d", e.Version)
		}
	}
}

// ============================================================================
// Extended Operations Tests
// ============================================================================

func TestStore_LoadAll(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
	}, 0)

	store.Save(ctx, "order-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("Order", "order-1", "Order.Created", 1),
	}, 0)

	loaded, err := store.LoadAll(ctx, eventsourcing.DefaultLoadOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 events, got %d", len(loaded))
	}
}

func TestStore_Count(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-123", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
		createTestEnvelope("User", "user-123", "User.EmailChanged", 2),
	}, 0)

	count, err := store.Count(ctx, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestStore_CountByType(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
	}, 0)

	store.Save(ctx, "user-2", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-2", "User.Created", 1),
	}, 0)

	store.Save(ctx, "order-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("Order", "order-1", "Order.Created", 1),
	}, 0)

	count, err := store.CountByType(ctx, "User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestStore_Delete(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-123", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-123", "User.Created", 1),
	}, 0)

	if store.EventCount() != 1 {
		t.Errorf("expected 1 event before delete, got %d", store.EventCount())
	}

	err := store.Delete(ctx, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.EventCount() != 0 {
		t.Errorf("expected 0 events after delete, got %d", store.EventCount())
	}
}

// ============================================================================
// Testing Helpers Tests
// ============================================================================

func TestStore_Clear(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
	}, 0)

	store.Save(ctx, "user-2", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-2", "User.Created", 1),
	}, 0)

	if store.EventCount() != 2 {
		t.Errorf("expected 2 events before clear, got %d", store.EventCount())
	}

	store.Clear()

	if store.EventCount() != 0 {
		t.Errorf("expected 0 events after clear, got %d", store.EventCount())
	}
}

func TestStore_AggregateIDs(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	store.Save(ctx, "user-1", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-1", "User.Created", 1),
	}, 0)

	store.Save(ctx, "user-2", []*eventsourcing.EventEnvelope{
		createTestEnvelope("User", "user-2", "User.Created", 1),
	}, 0)

	ids := store.AggregateIDs()

	if len(ids) != 2 {
		t.Errorf("expected 2 aggregate IDs, got %d", len(ids))
	}

	// Check both IDs exist (order may vary)
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	if !idMap["user-1"] || !idMap["user-2"] {
		t.Errorf("expected user-1 and user-2 in IDs, got %v", ids)
	}
}

// ============================================================================
// Snapshot Store Tests
// ============================================================================

func TestNewSnapshotStore(t *testing.T) {
	store := memory.NewSnapshotStore()
	if store == nil {
		t.Fatal("expected non-nil snapshot store")
	}
}

func TestSnapshotStore_SaveAndLoad(t *testing.T) {
	store := memory.NewSnapshotStore()
	ctx := context.Background()

	snapshot := &eventsourcing.Snapshot{
		AggregateID:   "user-123",
		AggregateType: "User",
		Version:       10,
		Data:          []byte(`{"email":"test@example.com","name":"Test"}`),
		Timestamp:     time.Now().UTC(),
	}

	err := store.Save(ctx, snapshot)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	loaded, err := store.Load(ctx, "user-123")
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	if loaded.AggregateID != "user-123" {
		t.Errorf("expected aggregate ID 'user-123', got '%s'", loaded.AggregateID)
	}

	if loaded.Version != 10 {
		t.Errorf("expected version 10, got %d", loaded.Version)
	}

	if string(loaded.Data) != `{"email":"test@example.com","name":"Test"}` {
		t.Errorf("unexpected data: %s", string(loaded.Data))
	}
}

func TestSnapshotStore_Load_NotFound(t *testing.T) {
	store := memory.NewSnapshotStore()
	ctx := context.Background()

	_, err := store.Load(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}

	if !eventsourcing.IsSnapshotNotFound(err) {
		t.Errorf("expected snapshot not found error, got: %v", err)
	}
}

func TestSnapshotStore_Load_ReturnsCopy(t *testing.T) {
	store := memory.NewSnapshotStore()
	ctx := context.Background()

	snapshot := &eventsourcing.Snapshot{
		AggregateID:   "user-123",
		AggregateType: "User",
		Version:       10,
		Data:          []byte(`original`),
		Timestamp:     time.Now().UTC(),
	}
	store.Save(ctx, snapshot)

	// Load and modify
	loaded1, _ := store.Load(ctx, "user-123")
	loaded1.Data = []byte(`modified`)

	// Load again
	loaded2, _ := store.Load(ctx, "user-123")

	if string(loaded2.Data) != "original" {
		t.Error("expected Load to return a copy")
	}
}

func TestSnapshotStore_Delete(t *testing.T) {
	store := memory.NewSnapshotStore()
	ctx := context.Background()

	snapshot := &eventsourcing.Snapshot{
		AggregateID:   "user-123",
		AggregateType: "User",
		Version:       10,
		Data:          []byte(`data`),
		Timestamp:     time.Now().UTC(),
	}
	store.Save(ctx, snapshot)

	err := store.Delete(ctx, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.Load(ctx, "user-123")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestSnapshotStore_Clear(t *testing.T) {
	store := memory.NewSnapshotStore()
	ctx := context.Background()

	store.Save(ctx, &eventsourcing.Snapshot{
		AggregateID: "user-1",
		Version:     1,
		Data:        []byte(`data`),
	})

	store.Save(ctx, &eventsourcing.Snapshot{
		AggregateID: "user-2",
		Version:     1,
		Data:        []byte(`data`),
	})

	store.Clear()

	_, err := store.Load(ctx, "user-1")
	if err == nil {
		t.Error("expected error after clear for user-1")
	}

	_, err = store.Load(ctx, "user-2")
	if err == nil {
		t.Error("expected error after clear for user-2")
	}
}

func TestSnapshotStore_Overwrite(t *testing.T) {
	store := memory.NewSnapshotStore()
	ctx := context.Background()

	// Save initial snapshot
	store.Save(ctx, &eventsourcing.Snapshot{
		AggregateID: "user-123",
		Version:     5,
		Data:        []byte(`v5`),
	})

	// Save newer snapshot
	store.Save(ctx, &eventsourcing.Snapshot{
		AggregateID: "user-123",
		Version:     10,
		Data:        []byte(`v10`),
	})

	loaded, _ := store.Load(ctx, "user-123")

	if loaded.Version != 10 {
		t.Errorf("expected version 10, got %d", loaded.Version)
	}

	if string(loaded.Data) != "v10" {
		t.Errorf("expected data 'v10', got '%s'", string(loaded.Data))
	}
}

// ============================================================================
// Interface Compliance Tests
// ============================================================================

func TestStore_ImplementsEventStore(t *testing.T) {
	var _ eventsourcing.EventStore = (*memory.Store)(nil)
}

func TestStore_ImplementsExtendedEventStore(t *testing.T) {
	var _ eventsourcing.ExtendedEventStore = (*memory.Store)(nil)
}

func TestSnapshotStore_ImplementsSnapshotStore(t *testing.T) {
	var _ eventsourcing.SnapshotStore = (*memory.SnapshotStore)(nil)
}
