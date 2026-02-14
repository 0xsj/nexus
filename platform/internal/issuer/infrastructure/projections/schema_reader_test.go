package projections

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockSchemaReaderQueries struct {
	exists bool
	err    error
}

func (m *mockSchemaReaderQueries) SchemaProjectionExistsByType(_ context.Context, _ string) (bool, error) {
	return m.exists, m.err
}

func TestSchemaReader_GetSchema(t *testing.T) {
	t.Run("schema exists returns nil", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{exists: true}
		reader := &SchemaReader{queries: mock}

		err := reader.GetSchema(context.Background(), "VerifiedDeveloper")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("schema not found returns error", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{exists: false}
		reader := &SchemaReader{queries: mock}

		err := reader.GetSchema(context.Background(), "NonExistent")
		if err == nil {
			t.Fatal("expected error for non-existent schema, got nil")
		}
		if !strings.Contains(err.Error(), "schema not found") {
			t.Errorf("expected error to contain %q, got %q", "schema not found", err.Error())
		}
		if !strings.Contains(err.Error(), "NonExistent") {
			t.Errorf("expected error to contain schema type %q, got %q", "NonExistent", err.Error())
		}
	})

	t.Run("query error is wrapped", func(t *testing.T) {
		dbErr := errors.New("connection refused")
		mock := &mockSchemaReaderQueries{err: dbErr}
		reader := &SchemaReader{queries: mock}

		err := reader.GetSchema(context.Background(), "VerifiedDeveloper")
		if err == nil {
			t.Fatal("expected error when query fails, got nil")
		}
		if !errors.Is(err, dbErr) {
			t.Errorf("expected wrapped db error, got %v", err)
		}
	})
}

func TestSchemaReader_SchemaExists(t *testing.T) {
	t.Run("schema exists returns true", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{exists: true}
		reader := &SchemaReader{queries: mock}

		exists, err := reader.SchemaExists(context.Background(), "VerifiedDeveloper")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !exists {
			t.Error("expected true, got false")
		}
	})

	t.Run("schema does not exist returns false", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{exists: false}
		reader := &SchemaReader{queries: mock}

		exists, err := reader.SchemaExists(context.Background(), "NonExistent")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Error("expected false, got true")
		}
	})

	t.Run("query error propagates", func(t *testing.T) {
		dbErr := errors.New("timeout")
		mock := &mockSchemaReaderQueries{err: dbErr}
		reader := &SchemaReader{queries: mock}

		_, err := reader.SchemaExists(context.Background(), "VerifiedDeveloper")
		if err == nil {
			t.Fatal("expected error when query fails, got nil")
		}
		if !errors.Is(err, dbErr) {
			t.Errorf("expected db error, got %v", err)
		}
	})
}
