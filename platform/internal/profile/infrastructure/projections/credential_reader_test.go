package projections

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
)

type mockCredentialReaderQueries struct {
	projection generated.ProfileCredentialProjection
	err        error
}

func (m *mockCredentialReaderQueries) GetCredentialProjection(ctx context.Context, credentialID string) (generated.ProfileCredentialProjection, error) {
	return m.projection, m.err
}

func TestCredentialReader_GetCredentialByID(t *testing.T) {
	t.Run("credential found returns true with type", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{
			projection: generated.ProfileCredentialProjection{
				CredentialID:   "cred-123",
				CredentialType: "GitHubContributor",
				Status:         "active",
				UpdatedAt:      time.Now(),
			},
		}
		reader := &CredentialReader{queries: mock}

		found, credType, err := reader.GetCredentialByID(context.Background(), "cred-123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !found {
			t.Error("expected found to be true")
		}
		if credType != "GitHubContributor" {
			t.Errorf("expected credential type %q, got %q", "GitHubContributor", credType)
		}
	})

	t.Run("credential not found returns false with empty type", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{
			err: errors.New("no rows in result set"),
		}
		reader := &CredentialReader{queries: mock}

		found, credType, err := reader.GetCredentialByID(context.Background(), "cred-nonexistent")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if found {
			t.Error("expected found to be false")
		}
		if credType != "" {
			t.Errorf("expected empty credential type, got %q", credType)
		}
	})
}
