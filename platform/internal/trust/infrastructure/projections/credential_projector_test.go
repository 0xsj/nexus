package projections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

type mockCredentialProjectorQueries struct {
	upsertCalled bool
	upsertArg    generated.UpsertCredentialProjectionParams
	upsertErr    error

	updateStatusCalled bool
	updateStatusArg    generated.UpdateCredentialProjectionStatusParams
	updateStatusErr    error

	getUserIDByDIDResult string
	getUserIDByDIDErr    error
	getUserIDByDIDArg    string
}

func (m *mockCredentialProjectorQueries) UpsertCredentialProjection(ctx context.Context, arg generated.UpsertCredentialProjectionParams) error {
	m.upsertCalled = true
	m.upsertArg = arg
	return m.upsertErr
}

func (m *mockCredentialProjectorQueries) UpdateCredentialProjectionStatus(ctx context.Context, arg generated.UpdateCredentialProjectionStatusParams) error {
	m.updateStatusCalled = true
	m.updateStatusArg = arg
	return m.updateStatusErr
}

func (m *mockCredentialProjectorQueries) GetUserIDByDID(ctx context.Context, subjectDid string) (string, error) {
	m.getUserIDByDIDArg = subjectDid
	return m.getUserIDByDIDResult, m.getUserIDByDIDErr
}

func TestCredentialProjector_Handle(t *testing.T) {
	t.Run("Credential.Issued with DID lookup success", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{
			getUserIDByDIDResult: "user-123",
		}
		projector := &CredentialProjector{queries: mock}

		data, _ := json.Marshal(map[string]any{
			"credential_id":   "cred-001",
			"credential_type": "VerifiedEmail",
			"subject_did":     "did:key:z6Mk123",
		})

		env := eventsourcing.EventEnvelope{
			Type:        "Credential.Issued",
			AggregateID: "cred-001",
			Data:        json.RawMessage(data),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !mock.upsertCalled {
			t.Fatal("expected UpsertCredentialProjection to be called")
		}
		if mock.getUserIDByDIDArg != "did:key:z6Mk123" {
			t.Errorf("expected GetUserIDByDID called with %q, got %q", "did:key:z6Mk123", mock.getUserIDByDIDArg)
		}
		if mock.upsertArg.CredentialID != "cred-001" {
			t.Errorf("expected CredentialID %q, got %q", "cred-001", mock.upsertArg.CredentialID)
		}
		if mock.upsertArg.CredentialType != "VerifiedEmail" {
			t.Errorf("expected CredentialType %q, got %q", "VerifiedEmail", mock.upsertArg.CredentialType)
		}
		if mock.upsertArg.SubjectDid != "did:key:z6Mk123" {
			t.Errorf("expected SubjectDid %q, got %q", "did:key:z6Mk123", mock.upsertArg.SubjectDid)
		}
		if mock.upsertArg.UserID != "user-123" {
			t.Errorf("expected UserID %q, got %q", "user-123", mock.upsertArg.UserID)
		}
		if mock.upsertArg.Status != "active" {
			t.Errorf("expected Status %q, got %q", "active", mock.upsertArg.Status)
		}
	})

	t.Run("Credential.Issued with DID lookup failure", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{
			getUserIDByDIDResult: "",
			getUserIDByDIDErr:    errors.New("not found"),
		}
		projector := &CredentialProjector{queries: mock}

		data, _ := json.Marshal(map[string]any{
			"credential_id":   "cred-002",
			"credential_type": "VerifiedGitHub",
			"subject_did":     "did:key:z6Mk456",
		})

		env := eventsourcing.EventEnvelope{
			Type:        "Credential.Issued",
			AggregateID: "cred-002",
			Data:        json.RawMessage(data),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !mock.upsertCalled {
			t.Fatal("expected UpsertCredentialProjection to be called")
		}
		if mock.upsertArg.UserID != "" {
			t.Errorf("expected empty UserID, got %q", mock.upsertArg.UserID)
		}
		if mock.upsertArg.Status != "active" {
			t.Errorf("expected Status %q, got %q", "active", mock.upsertArg.Status)
		}
	})

	t.Run("Credential.Revoked", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        "Credential.Revoked",
			AggregateID: "cred-003",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateCredentialProjectionStatus to be called")
		}
		if mock.updateStatusArg.CredentialID != "cred-003" {
			t.Errorf("expected CredentialID %q, got %q", "cred-003", mock.updateStatusArg.CredentialID)
		}
		if mock.updateStatusArg.Status != "revoked" {
			t.Errorf("expected Status %q, got %q", "revoked", mock.updateStatusArg.Status)
		}
	})

	t.Run("Credential.Expired", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        "Credential.Expired",
			AggregateID: "cred-004",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateCredentialProjectionStatus to be called")
		}
		if mock.updateStatusArg.CredentialID != "cred-004" {
			t.Errorf("expected CredentialID %q, got %q", "cred-004", mock.updateStatusArg.CredentialID)
		}
		if mock.updateStatusArg.Status != "expired" {
			t.Errorf("expected Status %q, got %q", "expired", mock.updateStatusArg.Status)
		}
	})

	t.Run("unknown event returns nil", func(t *testing.T) {
		mock := &mockCredentialProjectorQueries{}
		projector := &CredentialProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        "SomeOther.Event",
			AggregateID: "agg-999",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
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
			Type:        "Credential.Issued",
			AggregateID: "cred-bad",
			Data:        json.RawMessage(`{not valid json`),
		}

		err := projector.Handle(context.Background(), &env)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})
}
