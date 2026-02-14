package projections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	credentialdomain "github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

type mockCredentialProjectorQueries struct {
	upsertCalled       bool
	upsertParams       generated.UpsertCredentialProjectionParams
	updateStatusCalled bool
	updateStatusParams generated.UpdateCredentialProjectionStatusParams
	upsertErr          error
	updateStatusErr    error
}

func (m *mockCredentialProjectorQueries) UpsertCredentialProjection(ctx context.Context, arg generated.UpsertCredentialProjectionParams) error {
	m.upsertCalled = true
	m.upsertParams = arg
	return m.upsertErr
}

func (m *mockCredentialProjectorQueries) UpdateCredentialProjectionStatus(ctx context.Context, arg generated.UpdateCredentialProjectionStatusParams) error {
	m.updateStatusCalled = true
	m.updateStatusParams = arg
	return m.updateStatusErr
}

func TestCredentialProjector_Handle(t *testing.T) {
	t.Run("credential issued upserts projection", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		data, _ := json.Marshal(credentialdomain.CredentialIssuedEvent{
			CredentialID:   "cred-123",
			CredentialType: "GitHubContributor",
		})

		env := eventsourcing.EventEnvelope{
			Type:        credentialdomain.EventTypeCredentialIssued,
			AggregateID: "agg-1",
			Data:        data,
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.upsertCalled {
			t.Fatal("expected UpsertCredentialProjection to be called")
		}
		if mock.upsertParams.CredentialID != "cred-123" {
			t.Errorf("expected CredentialID %q, got %q", "cred-123", mock.upsertParams.CredentialID)
		}
		if mock.upsertParams.CredentialType != "GitHubContributor" {
			t.Errorf("expected CredentialType %q, got %q", "GitHubContributor", mock.upsertParams.CredentialType)
		}
		if mock.upsertParams.Status != "active" {
			t.Errorf("expected Status %q, got %q", "active", mock.upsertParams.Status)
		}
	})

	t.Run("credential revoked updates status to revoked", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        credentialdomain.EventTypeCredentialRevoked,
			AggregateID: "cred-456",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateCredentialProjectionStatus to be called")
		}
		if mock.updateStatusParams.CredentialID != "cred-456" {
			t.Errorf("expected CredentialID %q, got %q", "cred-456", mock.updateStatusParams.CredentialID)
		}
		if mock.updateStatusParams.Status != "revoked" {
			t.Errorf("expected Status %q, got %q", "revoked", mock.updateStatusParams.Status)
		}
	})

	t.Run("credential expired updates status to expired", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        credentialdomain.EventTypeCredentialExpired,
			AggregateID: "cred-789",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateCredentialProjectionStatus to be called")
		}
		if mock.updateStatusParams.CredentialID != "cred-789" {
			t.Errorf("expected CredentialID %q, got %q", "cred-789", mock.updateStatusParams.CredentialID)
		}
		if mock.updateStatusParams.Status != "expired" {
			t.Errorf("expected Status %q, got %q", "expired", mock.updateStatusParams.Status)
		}
	})

	t.Run("unknown event returns nil without calling queries", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        "SomeUnknown.Event",
			AggregateID: "agg-999",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if mock.upsertCalled {
			t.Error("expected UpsertCredentialProjection not to be called")
		}
		if mock.updateStatusCalled {
			t.Error("expected UpdateCredentialProjectionStatus not to be called")
		}
	})

	t.Run("malformed JSON returns error", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        credentialdomain.EventTypeCredentialIssued,
			AggregateID: "agg-bad",
			Data:        json.RawMessage(`{not valid json`),
		}

		err := projector.Handle(context.Background(), &env)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})

	t.Run("upsert query error propagates", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{
			upsertErr: errors.New("db connection lost"),
		}
		projector := &CredentialProjector{queries: mock}

		data, _ := json.Marshal(credentialdomain.CredentialIssuedEvent{
			CredentialID:   "cred-err",
			CredentialType: "LinkedInProfile",
		})

		env := eventsourcing.EventEnvelope{
			Type:        credentialdomain.EventTypeCredentialIssued,
			AggregateID: "agg-err",
			Data:        data,
		}

		err := projector.Handle(context.Background(), &env)
		if err == nil {
			t.Fatal("expected error from upsert failure, got nil")
		}
	})
}
