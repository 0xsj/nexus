package projections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
	schemadomain "github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

type mockSchemaProjectorQueries struct {
	upsertCalled       bool
	upsertParams       generated.UpsertSchemaProjectionParams
	updateStatusCalled bool
	updateStatusParams generated.UpdateSchemaProjectionStatusParams
	updateClaimsCalled bool
	updateClaimsParams generated.UpdateSchemaProjectionClaimsParams
	err                error
}

func (m *mockSchemaProjectorQueries) UpsertSchemaProjection(_ context.Context, arg generated.UpsertSchemaProjectionParams) error {
	m.upsertCalled = true
	m.upsertParams = arg
	return m.err
}

func (m *mockSchemaProjectorQueries) UpdateSchemaProjectionStatus(_ context.Context, arg generated.UpdateSchemaProjectionStatusParams) error {
	m.updateStatusCalled = true
	m.updateStatusParams = arg
	return m.err
}

func (m *mockSchemaProjectorQueries) UpdateSchemaProjectionClaims(_ context.Context, arg generated.UpdateSchemaProjectionClaimsParams) error {
	m.updateClaimsCalled = true
	m.updateClaimsParams = arg
	return m.err
}

func TestSchemaProjector_Handle(t *testing.T) {
	claims := []schemadomain.ClaimDefinitionData{
		{
			Key:         "name",
			DataType:    "string",
			Required:    true,
			DisplayName: "Full Name",
			Description: "The user's full name",
			Disclosable: true,
			Order:       1,
			Constraints: nil,
		},
	}

	t.Run("Schema.Registered upserts projection", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{}
		projector := &SchemaProjector{queries: mock}

		eventData, _ := json.Marshal(schemadomain.SchemaRegisteredEvent{
			SchemaID:   "schema-123",
			SchemaType: "VerifiedDeveloper",
			Claims:     claims,
		})

		env := eventsourcing.EventEnvelope{
			Type:        schemadomain.EventTypeSchemaRegistered,
			AggregateID: "schema-123",
			Data:        json.RawMessage(eventData),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.upsertCalled {
			t.Fatal("expected UpsertSchemaProjection to be called")
		}
		if mock.upsertParams.SchemaID != "schema-123" {
			t.Errorf("expected SchemaID %q, got %q", "schema-123", mock.upsertParams.SchemaID)
		}
		if mock.upsertParams.SchemaType != "VerifiedDeveloper" {
			t.Errorf("expected SchemaType %q, got %q", "VerifiedDeveloper", mock.upsertParams.SchemaType)
		}
		if mock.upsertParams.Status != "active" {
			t.Errorf("expected Status %q, got %q", "active", mock.upsertParams.Status)
		}

		var gotClaims []schemadomain.ClaimDefinitionData
		if err := json.Unmarshal(mock.upsertParams.Claims, &gotClaims); err != nil {
			t.Fatalf("failed to unmarshal claims: %v", err)
		}
		if len(gotClaims) != 1 || gotClaims[0].Key != "name" {
			t.Errorf("unexpected claims: %+v", gotClaims)
		}
	})

	t.Run("Schema.VersionAdded updates claims", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{}
		projector := &SchemaProjector{queries: mock}

		eventData, _ := json.Marshal(schemadomain.SchemaVersionAddedEvent{
			SchemaID: "schema-456",
			Claims:   claims,
		})

		env := eventsourcing.EventEnvelope{
			Type:        schemadomain.EventTypeSchemaVersionAdded,
			AggregateID: "schema-456",
			Data:        json.RawMessage(eventData),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateClaimsCalled {
			t.Fatal("expected UpdateSchemaProjectionClaims to be called")
		}
		if mock.updateClaimsParams.SchemaID != "schema-456" {
			t.Errorf("expected SchemaID %q, got %q", "schema-456", mock.updateClaimsParams.SchemaID)
		}

		var gotClaims []schemadomain.ClaimDefinitionData
		if err := json.Unmarshal(mock.updateClaimsParams.Claims, &gotClaims); err != nil {
			t.Fatalf("failed to unmarshal claims: %v", err)
		}
		if len(gotClaims) != 1 || gotClaims[0].Key != "name" {
			t.Errorf("unexpected claims: %+v", gotClaims)
		}
	})

	t.Run("Schema.Deprecated updates status to deprecated", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{}
		projector := &SchemaProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        schemadomain.EventTypeSchemaDeprecated,
			AggregateID: "schema-789",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateSchemaProjectionStatus to be called")
		}
		if mock.updateStatusParams.SchemaID != "schema-789" {
			t.Errorf("expected SchemaID %q, got %q", "schema-789", mock.updateStatusParams.SchemaID)
		}
		if mock.updateStatusParams.Status != "deprecated" {
			t.Errorf("expected Status %q, got %q", "deprecated", mock.updateStatusParams.Status)
		}
	})

	t.Run("Schema.Activated updates status to active", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{}
		projector := &SchemaProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        schemadomain.EventTypeSchemaActivated,
			AggregateID: "schema-789",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateStatusCalled {
			t.Fatal("expected UpdateSchemaProjectionStatus to be called")
		}
		if mock.updateStatusParams.SchemaID != "schema-789" {
			t.Errorf("expected SchemaID %q, got %q", "schema-789", mock.updateStatusParams.SchemaID)
		}
		if mock.updateStatusParams.Status != "active" {
			t.Errorf("expected Status %q, got %q", "active", mock.updateStatusParams.Status)
		}
	})

	t.Run("unknown event returns nil", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{}
		projector := &SchemaProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        "SomeOther.Event",
			AggregateID: "agg-1",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &env)
		if err != nil {
			t.Fatalf("expected nil for unknown event, got %v", err)
		}
		if mock.upsertCalled || mock.updateStatusCalled || mock.updateClaimsCalled {
			t.Fatal("expected no query calls for unknown event")
		}
	})

	t.Run("malformed JSON returns error", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{}
		projector := &SchemaProjector{queries: mock}

		env := eventsourcing.EventEnvelope{
			Type:        schemadomain.EventTypeSchemaRegistered,
			AggregateID: "schema-bad",
			Data:        json.RawMessage(`{not valid json`),
		}

		err := projector.Handle(context.Background(), &env)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})

	t.Run("query error propagates", func(t *testing.T) {
		mock := &mockSchemaProjectorQueries{err: errors.New("db connection lost")}
		projector := &SchemaProjector{queries: mock}

		eventData, _ := json.Marshal(schemadomain.SchemaRegisteredEvent{
			SchemaID:   "schema-err",
			SchemaType: "VerifiedDeveloper",
			Claims:     claims,
		})

		env := eventsourcing.EventEnvelope{
			Type:        schemadomain.EventTypeSchemaRegistered,
			AggregateID: "schema-err",
			Data:        json.RawMessage(eventData),
		}

		err := projector.Handle(context.Background(), &env)
		if err == nil {
			t.Fatal("expected error to propagate from query, got nil")
		}
		if !errors.Is(err, mock.err) {
			t.Errorf("expected wrapped db error, got %v", err)
		}
	})
}
