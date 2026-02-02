//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
)

// ============================================================================
// Test Database Configuration
// ============================================================================

const (
	defaultTestDSN = "postgres://postgres:postgres@localhost:5432/nexus_test?sslmode=disable"
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
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Run schema migration
	if err := runMigrations(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
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

// runMigrations applies the schema to the test database.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
		CREATE TABLE IF NOT EXISTS ledger_entries (
			id UUID PRIMARY KEY,
			occurred_at TIMESTAMPTZ NOT NULL,
			recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			event_type VARCHAR(255) NOT NULL,
			actor_id VARCHAR(255) NOT NULL,
			actor_type VARCHAR(50) NOT NULL,
			subject_id VARCHAR(255) NOT NULL,
			subject_type VARCHAR(50) NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}',
			context_id VARCHAR(255)
		);

		CREATE INDEX IF NOT EXISTS idx_ledger_entries_occurred_at 
		ON ledger_entries (occurred_at DESC);

		CREATE INDEX IF NOT EXISTS idx_ledger_entries_actor 
		ON ledger_entries (actor_id, actor_type, occurred_at DESC);

		CREATE INDEX IF NOT EXISTS idx_ledger_entries_subject 
		ON ledger_entries (subject_id, subject_type, occurred_at DESC);

		CREATE INDEX IF NOT EXISTS idx_ledger_entries_event_type 
		ON ledger_entries (event_type, occurred_at DESC);

		CREATE INDEX IF NOT EXISTS idx_ledger_entries_context_id 
		ON ledger_entries (context_id) 
		WHERE context_id IS NOT NULL;

		CREATE INDEX IF NOT EXISTS idx_ledger_entries_subject_event_type 
		ON ledger_entries (subject_id, subject_type, event_type, occurred_at DESC);
	`

	_, err := pool.Exec(ctx, schema)
	return err
}

// ============================================================================
// Test Data Builders
// ============================================================================

// entryBuilder helps construct test audit entries.
type entryBuilder struct {
	id          string
	occurredAt  time.Time
	eventType   string
	actorID     string
	actorType   string
	subjectID   string
	subjectType string
	metadata    map[string]interface{}
	contextID   string
}

func newEntryBuilder() *entryBuilder {
	return &entryBuilder{
		occurredAt:  time.Now().UTC(),
		eventType:   "test.event",
		actorID:     "user-123",
		actorType:   "user",
		subjectID:   "subject-456",
		subjectType: "credential",
		metadata:    make(map[string]interface{}),
	}
}

func (b *entryBuilder) withID(id string) *entryBuilder {
	b.id = id
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
	b.metadata[key] = value
	return b
}

func (b *entryBuilder) withContextID(contextID string) *entryBuilder {
	b.contextID = contextID
	return b
}

func (b *entryBuilder) build(t *testing.T) *generated.InsertEntryParams {
	t.Helper()

	id := b.id
	if id == "" {
		id = fmt.Sprintf("00000000-0000-0000-0000-%012d", time.Now().UnixNano()%1000000000000)
	}

	uid, err := parseUUID(id)
	if err != nil {
		t.Fatalf("invalid UUID: %v", err)
	}

	metadata, err := marshalMetadata(b.metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	params := &generated.InsertEntryParams{
		ID:          uid,
		OccurredAt:  b.occurredAt,
		EventType:   b.eventType,
		ActorID:     b.actorID,
		ActorType:   b.actorType,
		SubjectID:   b.subjectID,
		SubjectType: b.subjectType,
		Metadata:    metadata,
	}

	if b.contextID != "" {
		params.ContextID = &b.contextID
	}

	return params
}

// insertTestEntry is a helper to insert a test entry directly.
func (db *testDB) insertTestEntry(t *testing.T, params *generated.InsertEntryParams) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.queries.InsertEntry(ctx, *params)
	if err != nil {
		t.Fatalf("failed to insert test entry: %v", err)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func parseUUID(s string) ([16]byte, error) {
	var uid [16]byte
	parsed, err := parseUUIDString(s)
	if err != nil {
		return uid, err
	}
	copy(uid[:], parsed)
	return uid, nil
}

func parseUUIDString(s string) ([]byte, error) {
	// Simple UUID parser - expects format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(s) != 36 {
		return nil, fmt.Errorf("invalid UUID length: %d", len(s))
	}

	hexStr := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	if len(hexStr) != 32 {
		return nil, fmt.Errorf("invalid UUID format")
	}

	result := make([]byte, 16)
	for i := 0; i < 16; i++ {
		var val byte
		for j := 0; j < 2; j++ {
			c := hexStr[i*2+j]
			switch {
			case c >= '0' && c <= '9':
				val = val*16 + (c - '0')
			case c >= 'a' && c <= 'f':
				val = val*16 + (c - 'a' + 10)
			case c >= 'A' && c <= 'F':
				val = val*16 + (c - 'A' + 10)
			default:
				return nil, fmt.Errorf("invalid hex character: %c", c)
			}
		}
		result[i] = val
	}

	return result, nil
}

func marshalMetadata(m map[string]interface{}) ([]byte, error) {
	if m == nil || len(m) == 0 {
		return []byte("{}"), nil
	}

	// Simple JSON marshaling for test purposes
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		first = false

		result += fmt.Sprintf(`"%s":`, k)
		switch val := v.(type) {
		case string:
			result += fmt.Sprintf(`"%s"`, val)
		case int:
			result += fmt.Sprintf("%d", val)
		case bool:
			result += fmt.Sprintf("%t", val)
		default:
			result += fmt.Sprintf(`"%v"`, val)
		}
	}
	result += "}"

	return []byte(result), nil
}
