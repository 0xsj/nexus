package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
)

type mockSchemaReaderQueries struct {
	projection generated.CredentialSchemaProjection
	err        error
}

func (m *mockSchemaReaderQueries) GetSchemaProjectionByType(_ context.Context, _ string) (generated.CredentialSchemaProjection, error) {
	return m.projection, m.err
}

func mustClaims(t *testing.T, data map[string]any) domain.Claims {
	t.Helper()
	c, err := domain.NewClaims(data)
	if err != nil {
		t.Fatalf("NewClaims: %v", err)
	}
	return c
}

func TestSchemaReader_GetSchema(t *testing.T) {
	tests := []struct {
		name     string
		mock     mockSchemaReaderQueries
		credType domain.CredentialType
		wantID   string
		wantErr  bool
	}{
		{
			name: "found returns schema ID",
			mock: mockSchemaReaderQueries{
				projection: generated.CredentialSchemaProjection{
					SchemaID:   "schema-abc",
					SchemaType: "github",
					Status:     "active",
					UpdatedAt:  time.Now(),
				},
			},
			credType: domain.CredentialType("github"),
			wantID:   "schema-abc",
		},
		{
			name: "not found returns error",
			mock: mockSchemaReaderQueries{
				err: fmt.Errorf("no rows"),
			},
			credType: domain.CredentialType("nonexistent"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &SchemaReader{queries: &tt.mock}

			id, err := reader.GetSchema(context.Background(), tt.credType)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("schema ID = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestSchemaReader_ValidateClaims(t *testing.T) {
	activeClaims := []struct {
		Key      string `json:"key"`
		Required bool   `json:"required"`
	}{
		{Key: "username", Required: true},
		{Key: "repo_count", Required: true},
		{Key: "bio", Required: false},
	}
	activeClaimsJSON, _ := json.Marshal(activeClaims)

	allOptionalClaims := []struct {
		Key      string `json:"key"`
		Required bool   `json:"required"`
	}{
		{Key: "bio", Required: false},
		{Key: "avatar", Required: false},
	}
	allOptionalJSON, _ := json.Marshal(allOptionalClaims)

	t.Run("active schema with all required claims present", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{
			projection: generated.CredentialSchemaProjection{
				SchemaID:   "schema-1",
				SchemaType: "github",
				Status:     "active",
				Claims:     activeClaimsJSON,
				UpdatedAt:  time.Now(),
			},
		}
		reader := &SchemaReader{queries: mock}
		claims := mustClaims(t, map[string]any{
			"username":   "octocat",
			"repo_count": 42,
			"bio":        "Hello",
		})

		err := reader.ValidateClaims(context.Background(), domain.CredentialType("github"), claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("schema not found", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{err: fmt.Errorf("no rows")}
		reader := &SchemaReader{queries: mock}
		claims := mustClaims(t, map[string]any{"username": "octocat"})

		err := reader.ValidateClaims(context.Background(), domain.CredentialType("github"), claims)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("deprecated schema returns error", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{
			projection: generated.CredentialSchemaProjection{
				SchemaID:   "schema-2",
				SchemaType: "github",
				Status:     "deprecated",
				Claims:     activeClaimsJSON,
				UpdatedAt:  time.Now(),
			},
		}
		reader := &SchemaReader{queries: mock}
		claims := mustClaims(t, map[string]any{"username": "octocat", "repo_count": 10})

		err := reader.ValidateClaims(context.Background(), domain.CredentialType("github"), claims)
		if err == nil {
			t.Fatal("expected error for deprecated schema, got nil")
		}
	})

	t.Run("required claim missing returns error", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{
			projection: generated.CredentialSchemaProjection{
				SchemaID:   "schema-3",
				SchemaType: "github",
				Status:     "active",
				Claims:     activeClaimsJSON,
				UpdatedAt:  time.Now(),
			},
		}
		reader := &SchemaReader{queries: mock}
		claims := mustClaims(t, map[string]any{"username": "octocat"})

		err := reader.ValidateClaims(context.Background(), domain.CredentialType("github"), claims)
		if err == nil {
			t.Fatal("expected error for missing required claim, got nil")
		}
	})

	t.Run("no required claims succeeds with any input", func(t *testing.T) {
		mock := &mockSchemaReaderQueries{
			projection: generated.CredentialSchemaProjection{
				SchemaID:   "schema-4",
				SchemaType: "github",
				Status:     "active",
				Claims:     allOptionalJSON,
				UpdatedAt:  time.Now(),
			},
		}
		reader := &SchemaReader{queries: mock}
		claims := mustClaims(t, map[string]any{"anything": "value"})

		err := reader.ValidateClaims(context.Background(), domain.CredentialType("github"), claims)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
