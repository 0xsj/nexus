package projections

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
)

type mockOrgReaderQueries struct {
	exists     bool
	existsErr  error
	projection generated.IssuerOrganizationProjection
	getErr     error
}

func (m *mockOrgReaderQueries) OrganizationProjectionExists(_ context.Context, _ string) (bool, error) {
	return m.exists, m.existsErr
}

func (m *mockOrgReaderQueries) GetOrganizationProjection(_ context.Context, _ string) (generated.IssuerOrganizationProjection, error) {
	return m.projection, m.getErr
}

func TestOrganizationReader_GetOrganization(t *testing.T) {
	t.Run("organization exists returns nil", func(t *testing.T) {
		mock := &mockOrgReaderQueries{exists: true}
		reader := &OrganizationReader{queries: mock}

		err := reader.GetOrganization(context.Background(), "org-123")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("organization not found returns error", func(t *testing.T) {
		mock := &mockOrgReaderQueries{exists: false}
		reader := &OrganizationReader{queries: mock}

		orgID := "org-missing"
		err := reader.GetOrganization(context.Background(), orgID)
		if err == nil {
			t.Fatal("expected error for missing organization, got nil")
		}
		if !strings.Contains(err.Error(), "organization not found") {
			t.Errorf("expected error to contain %q, got %q", "organization not found", err.Error())
		}
		if !strings.Contains(err.Error(), orgID) {
			t.Errorf("expected error to contain org ID %q, got %q", orgID, err.Error())
		}
	})

	t.Run("query error is wrapped", func(t *testing.T) {
		queryErr := errors.New("connection refused")
		mock := &mockOrgReaderQueries{existsErr: queryErr}
		reader := &OrganizationReader{queries: mock}

		err := reader.GetOrganization(context.Background(), "org-err")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, queryErr) {
			t.Errorf("expected error to wrap %v, got %v", queryErr, err)
		}
	})
}

func TestOrganizationReader_IsVerified(t *testing.T) {
	t.Run("verified organization returns true", func(t *testing.T) {
		mock := &mockOrgReaderQueries{
			projection: generated.IssuerOrganizationProjection{
				OrganizationID:     "org-verified",
				VerificationStatus: "verified",
				Active:             true,
				UpdatedAt:          time.Now(),
			},
		}
		reader := &OrganizationReader{queries: mock}

		verified, err := reader.IsVerified(context.Background(), "org-verified")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !verified {
			t.Error("expected true for verified organization")
		}
	})

	t.Run("unverified organization returns false", func(t *testing.T) {
		mock := &mockOrgReaderQueries{
			projection: generated.IssuerOrganizationProjection{
				OrganizationID:     "org-unverified",
				VerificationStatus: "unverified",
				Active:             true,
				UpdatedAt:          time.Now(),
			},
		}
		reader := &OrganizationReader{queries: mock}

		verified, err := reader.IsVerified(context.Background(), "org-unverified")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if verified {
			t.Error("expected false for unverified organization")
		}
	})

	t.Run("query error returns false and error", func(t *testing.T) {
		queryErr := errors.New("not found")
		mock := &mockOrgReaderQueries{getErr: queryErr}
		reader := &OrganizationReader{queries: mock}

		verified, err := reader.IsVerified(context.Background(), "org-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if verified {
			t.Error("expected false when query fails")
		}
	})
}
