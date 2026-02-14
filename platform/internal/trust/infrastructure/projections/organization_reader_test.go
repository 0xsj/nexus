package projections

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
)

type mockOrgReaderQueries struct {
	projection generated.TrustOrganizationProjection
	err        error
}

func (m *mockOrgReaderQueries) GetOrganizationProjection(_ context.Context, _ string) (generated.TrustOrganizationProjection, error) {
	return m.projection, m.err
}

func TestOrganizationReader_IsTrustAnchor(t *testing.T) {
	t.Run("verified and active returns true", func(t *testing.T) {
		mock := &mockOrgReaderQueries{
			projection: generated.TrustOrganizationProjection{
				OrganizationID:     "org-1",
				VerificationStatus: "verified",
				Active:             true,
				UpdatedAt:          time.Now(),
			},
		}
		reader := &OrganizationReader{queries: mock}

		result, err := reader.IsTrustAnchor(context.Background(), "org-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result {
			t.Error("expected true for verified and active organization")
		}
	})

	t.Run("verified and inactive returns false", func(t *testing.T) {
		mock := &mockOrgReaderQueries{
			projection: generated.TrustOrganizationProjection{
				OrganizationID:     "org-2",
				VerificationStatus: "verified",
				Active:             false,
				UpdatedAt:          time.Now(),
			},
		}
		reader := &OrganizationReader{queries: mock}

		result, err := reader.IsTrustAnchor(context.Background(), "org-2")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result {
			t.Error("expected false for verified but inactive organization")
		}
	})

	t.Run("unverified and active returns false", func(t *testing.T) {
		mock := &mockOrgReaderQueries{
			projection: generated.TrustOrganizationProjection{
				OrganizationID:     "org-3",
				VerificationStatus: "unverified",
				Active:             true,
				UpdatedAt:          time.Now(),
			},
		}
		reader := &OrganizationReader{queries: mock}

		result, err := reader.IsTrustAnchor(context.Background(), "org-3")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result {
			t.Error("expected false for unverified organization")
		}
	})

	t.Run("query error returns false and error", func(t *testing.T) {
		mock := &mockOrgReaderQueries{
			err: errors.New("not found"),
		}
		reader := &OrganizationReader{queries: mock}

		result, err := reader.IsTrustAnchor(context.Background(), "org-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result {
			t.Error("expected false when query returns error")
		}
	})
}
