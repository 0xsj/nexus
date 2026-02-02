//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
)

// ============================================================================
// Test Database Configuration
// ============================================================================

const (
	// Default DSN matches docker-compose.yaml (port 5439 externally)
	defaultTestDSN = "postgres://nexus:nexus@localhost:5439/nexus?sslmode=disable"
)

func getTestDSN() string {
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return defaultTestDSN
}

// ============================================================================
// Test Database Setup
// ============================================================================

// testDB holds the test database connection and queries.
type testDB struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// setupTestDB creates a connection to the test database.
func setupTestDB(t *testing.T) *testDB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, getTestDSN())
	if err != nil {
		t.Fatalf("failed to connect to test database: %v\nDSN: %s\nMake sure docker is running: make docker-up && make migrate-ledger", err, getTestDSN())
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Verify ledger_entries table exists
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'ledger_entries'
		)
	`).Scan(&exists)
	if err != nil {
		pool.Close()
		t.Fatalf("failed to check for ledger_entries table: %v", err)
	}
	if !exists {
		pool.Close()
		t.Fatalf("ledger_entries table does not exist. Run: make migrate-ledger")
	}

	return &testDB{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// close cleans up the test database connection.
func (db *testDB) close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// cleanup removes all data from ledger_entries table.
func (db *testDB) cleanup(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.pool.Exec(ctx, "TRUNCATE TABLE ledger_entries")
	if err != nil {
		t.Fatalf("failed to cleanup test data: %v", err)
	}
}

// ============================================================================
// Test Data Builders
// ============================================================================

// entryBuilder helps construct test audit entries.
type entryBuilder struct {
	id          uuid.UUID
	occurredAt  time.Time
	recordedAt  time.Time
	eventType   string
	actorID     string
	actorType   string
	subjectID   string
	subjectType string
	metadata    []byte
	contextID   *string
}

func newEntryBuilder() *entryBuilder {
	return &entryBuilder{
		id:          uuid.New(),
		occurredAt:  time.Now().UTC(),
		recordedAt:  time.Now().UTC(),
		eventType:   "test.event",
		actorID:     "user-123",
		actorType:   "user",
		subjectID:   "subject-456",
		subjectType: "credential",
		metadata:    []byte("{}"),
	}
}

func (b *entryBuilder) withID(id string) *entryBuilder {
	parsed, err := uuid.Parse(id)
	if err == nil {
		b.id = parsed
	}
	return b
}

func (b *entryBuilder) withOccurredAt(t time.Time) *entryBuilder {
	b.occurredAt = t
	return b
}

func (b *entryBuilder) withEventType(eventType string) *entryBuilder {
	b.eventType = eventType
	return b
}

func (b *entryBuilder) withActor(actorID, actorType string) *entryBuilder {
	b.actorID = actorID
	b.actorType = actorType
	return b
}

func (b *entryBuilder) withSubject(subjectID, subjectType string) *entryBuilder {
	b.subjectID = subjectID
	b.subjectType = subjectType
	return b
}

func (b *entryBuilder) withMetadata(key string, value interface{}) *entryBuilder {
	// Simple JSON construction
	switch v := value.(type) {
	case string:
		b.metadata = []byte(`{"` + key + `":"` + v + `"}`)
	default:
		b.metadata = []byte("{}")
	}
	return b
}

func (b *entryBuilder) withContextID(contextID string) *entryBuilder {
	b.contextID = &contextID
	return b
}

func (b *entryBuilder) build(t *testing.T) generated.InsertEntryParams {
	t.Helper()

	return generated.InsertEntryParams{
		ID:          b.id,
		OccurredAt:  b.occurredAt,
		RecordedAt:  b.recordedAt,
		EventType:   b.eventType,
		ActorID:     b.actorID,
		ActorType:   b.actorType,
		SubjectID:   b.subjectID,
		SubjectType: b.subjectType,
		Metadata:    b.metadata,
		ContextID:   b.contextID,
	}
}

// insertTestEntry is a helper to insert a test entry directly.
func (db *testDB) insertTestEntry(t *testing.T, params generated.InsertEntryParams) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.queries.InsertEntry(ctx, params)
	if err != nil {
		t.Fatalf("failed to insert test entry: %v", err)
	}
}
