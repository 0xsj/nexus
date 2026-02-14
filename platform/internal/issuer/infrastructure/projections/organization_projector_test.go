package projections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	orgdomain "github.com/0xsj/nexus/platform/internal/organization/domain"

	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

type mockOrgProjectorQueries struct {
	upsertCalled          bool
	upsertParams          generated.UpsertOrganizationProjectionParams
	updateVerifiedCalled  bool
	updateVerifiedParams  generated.UpdateOrganizationProjectionVerifiedParams
	updateActiveCalled    bool
	updateActiveParams    generated.UpdateOrganizationProjectionActiveParams
	err                   error
}

func (m *mockOrgProjectorQueries) UpsertOrganizationProjection(_ context.Context, arg generated.UpsertOrganizationProjectionParams) error {
	m.upsertCalled = true
	m.upsertParams = arg
	return m.err
}

func (m *mockOrgProjectorQueries) UpdateOrganizationProjectionVerified(_ context.Context, arg generated.UpdateOrganizationProjectionVerifiedParams) error {
	m.updateVerifiedCalled = true
	m.updateVerifiedParams = arg
	return m.err
}

func (m *mockOrgProjectorQueries) UpdateOrganizationProjectionActive(_ context.Context, arg generated.UpdateOrganizationProjectionActiveParams) error {
	m.updateActiveCalled = true
	m.updateActiveParams = arg
	return m.err
}

func TestOrganizationProjector_Handle(t *testing.T) {
	t.Run("Organization.Created upserts projection", func(t *testing.T) {
		mock := &mockOrgProjectorQueries{}
		projector := &OrganizationProjector{queries: mock}

		orgID := "org-123"
		data, _ := json.Marshal(orgdomain.OrganizationCreatedEvent{
			OrganizationID: orgID,
			Name:           "Acme Corp",
			Slug:           "acme-corp",
			OrgType:        "company",
			OwnerUserID:    "user-1",
			OwnerMemberID:  "member-1",
			CreatedAt:      time.Now(),
		})

		envelope := eventsourcing.EventEnvelope{
			Type:        orgdomain.EventTypeOrganizationCreated,
			AggregateID: orgID,
			Data:        json.RawMessage(data),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.upsertCalled {
			t.Fatal("expected UpsertOrganizationProjection to be called")
		}
		if mock.upsertParams.OrganizationID != orgID {
			t.Errorf("expected OrganizationID %q, got %q", orgID, mock.upsertParams.OrganizationID)
		}
		if mock.upsertParams.VerificationStatus != "unverified" {
			t.Errorf("expected VerificationStatus %q, got %q", "unverified", mock.upsertParams.VerificationStatus)
		}
		if !mock.upsertParams.Active {
			t.Error("expected Active to be true")
		}
	})

	t.Run("Organization.Verified updates verification status", func(t *testing.T) {
		mock := &mockOrgProjectorQueries{}
		projector := &OrganizationProjector{queries: mock}

		aggregateID := "org-456"
		envelope := eventsourcing.EventEnvelope{
			Type:        orgdomain.EventTypeOrganizationVerified,
			AggregateID: aggregateID,
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateVerifiedCalled {
			t.Fatal("expected UpdateOrganizationProjectionVerified to be called")
		}
		if mock.updateVerifiedParams.OrganizationID != aggregateID {
			t.Errorf("expected OrganizationID %q, got %q", aggregateID, mock.updateVerifiedParams.OrganizationID)
		}
		if mock.updateVerifiedParams.VerificationStatus != "verified" {
			t.Errorf("expected VerificationStatus %q, got %q", "verified", mock.updateVerifiedParams.VerificationStatus)
		}
	})

	t.Run("Organization.Deleted sets active to false", func(t *testing.T) {
		mock := &mockOrgProjectorQueries{}
		projector := &OrganizationProjector{queries: mock}

		aggregateID := "org-789"
		envelope := eventsourcing.EventEnvelope{
			Type:        orgdomain.EventTypeOrganizationDeleted,
			AggregateID: aggregateID,
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateActiveCalled {
			t.Fatal("expected UpdateOrganizationProjectionActive to be called")
		}
		if mock.updateActiveParams.OrganizationID != aggregateID {
			t.Errorf("expected OrganizationID %q, got %q", aggregateID, mock.updateActiveParams.OrganizationID)
		}
		if mock.updateActiveParams.Active {
			t.Error("expected Active to be false")
		}
	})

	t.Run("unknown event returns nil", func(t *testing.T) {
		mock := &mockOrgProjectorQueries{}
		projector := &OrganizationProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        "SomeOther.Event",
			AggregateID: "org-000",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected nil for unknown event, got %v", err)
		}
		if mock.upsertCalled || mock.updateVerifiedCalled || mock.updateActiveCalled {
			t.Fatal("expected no queries to be called for unknown event")
		}
	})

	t.Run("malformed JSON returns error", func(t *testing.T) {
		mock := &mockOrgProjectorQueries{}
		projector := &OrganizationProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        orgdomain.EventTypeOrganizationCreated,
			AggregateID: "org-bad",
			Data:        json.RawMessage(`{not valid json`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})

	t.Run("query error propagates", func(t *testing.T) {
		queryErr := errors.New("database connection lost")
		mock := &mockOrgProjectorQueries{err: queryErr}
		projector := &OrganizationProjector{queries: mock}

		orgID := "org-err"
		data, _ := json.Marshal(orgdomain.OrganizationCreatedEvent{
			OrganizationID: orgID,
			Name:           "Fail Corp",
			Slug:           "fail-corp",
			OrgType:        "company",
			OwnerUserID:    "user-1",
			OwnerMemberID:  "member-1",
			CreatedAt:      time.Now(),
		})

		envelope := eventsourcing.EventEnvelope{
			Type:        orgdomain.EventTypeOrganizationCreated,
			AggregateID: orgID,
			Data:        json.RawMessage(data),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err == nil {
			t.Fatal("expected error to propagate from query, got nil")
		}
	})
}
