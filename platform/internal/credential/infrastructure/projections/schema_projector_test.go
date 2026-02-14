package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
	schemadomain "github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

type mockSchemaProjectorQueries struct {
	upsertCalled       bool
	updateStatusCalled bool
	updateClaimsCalled bool

	upsertArgs       generated.UpsertSchemaProjectionParams
	updateStatusArgs generated.UpdateSchemaProjectionStatusParams
	updateClaimsArgs generated.UpdateSchemaProjectionClaimsParams

	err error
}

func (m *mockSchemaProjectorQueries) UpsertSchemaProjection(_ context.Context, arg generated.UpsertSchemaProjectionParams) error {
	m.upsertCalled = true
	m.upsertArgs = arg
	return m.err
}

func (m *mockSchemaProjectorQueries) UpdateSchemaProjectionStatus(_ context.Context, arg generated.UpdateSchemaProjectionStatusParams) error {
	m.updateStatusCalled = true
	m.updateStatusArgs = arg
	return m.err
}

func (m *mockSchemaProjectorQueries) UpdateSchemaProjectionClaims(_ context.Context, arg generated.UpdateSchemaProjectionClaimsParams) error {
	m.updateClaimsCalled = true
	m.updateClaimsArgs = arg
	return m.err
}

func (m *mockSchemaProjectorQueries) anyCalled() bool {
	return m.upsertCalled || m.updateStatusCalled || m.updateClaimsCalled
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
		},
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	registeredData, err := json.Marshal(schemadomain.SchemaRegisteredEvent{
		SchemaID:   "schema-123",
		SchemaType: "github",
		Claims:     claims,
	})
	if err != nil {
		t.Fatalf("marshal registered event: %v", err)
	}

	versionAddedData, err := json.Marshal(schemadomain.SchemaVersionAddedEvent{
		SchemaID: "schema-456",
		Claims:   claims,
	})
	if err != nil {
		t.Fatalf("marshal version added event: %v", err)
	}

	tests := []struct {
		name    string
		event   eventsourcing.EventEnvelope
		verify  func(t *testing.T, m *mockSchemaProjectorQueries)
		wantErr bool
	}{
		{
			name: "Schema.Registered upserts projection",
			event: eventsourcing.EventEnvelope{
				Type:        schemadomain.EventTypeSchemaRegistered,
				AggregateID: "schema-123",
				Data:        registeredData,
			},
			verify: func(t *testing.T, m *mockSchemaProjectorQueries) {
				if !m.upsertCalled {
					t.Fatal("expected UpsertSchemaProjection to be called")
				}
				if m.upsertArgs.SchemaID != "schema-123" {
					t.Errorf("SchemaID = %q, want %q", m.upsertArgs.SchemaID, "schema-123")
				}
				if m.upsertArgs.SchemaType != "github" {
					t.Errorf("SchemaType = %q, want %q", m.upsertArgs.SchemaType, "github")
				}
				if m.upsertArgs.Status != "active" {
					t.Errorf("Status = %q, want %q", m.upsertArgs.Status, "active")
				}
				if string(m.upsertArgs.Claims) != string(claimsJSON) {
					t.Errorf("Claims = %s, want %s", m.upsertArgs.Claims, claimsJSON)
				}
			},
		},
		{
			name: "Schema.VersionAdded updates claims",
			event: eventsourcing.EventEnvelope{
				Type:        schemadomain.EventTypeSchemaVersionAdded,
				AggregateID: "schema-456",
				Data:        versionAddedData,
			},
			verify: func(t *testing.T, m *mockSchemaProjectorQueries) {
				if !m.updateClaimsCalled {
					t.Fatal("expected UpdateSchemaProjectionClaims to be called")
				}
				if m.updateClaimsArgs.SchemaID != "schema-456" {
					t.Errorf("SchemaID = %q, want %q", m.updateClaimsArgs.SchemaID, "schema-456")
				}
				if string(m.updateClaimsArgs.Claims) != string(claimsJSON) {
					t.Errorf("Claims = %s, want %s", m.updateClaimsArgs.Claims, claimsJSON)
				}
			},
		},
		{
			name: "Schema.Deprecated updates status to deprecated",
			event: eventsourcing.EventEnvelope{
				Type:        schemadomain.EventTypeSchemaDeprecated,
				AggregateID: "schema-789",
				Data:        json.RawMessage(`{}`),
			},
			verify: func(t *testing.T, m *mockSchemaProjectorQueries) {
				if !m.updateStatusCalled {
					t.Fatal("expected UpdateSchemaProjectionStatus to be called")
				}
				if m.updateStatusArgs.SchemaID != "schema-789" {
					t.Errorf("SchemaID = %q, want %q", m.updateStatusArgs.SchemaID, "schema-789")
				}
				if m.updateStatusArgs.Status != "deprecated" {
					t.Errorf("Status = %q, want %q", m.updateStatusArgs.Status, "deprecated")
				}
			},
		},
		{
			name: "Schema.Activated updates status to active",
			event: eventsourcing.EventEnvelope{
				Type:        schemadomain.EventTypeSchemaActivated,
				AggregateID: "schema-789",
				Data:        json.RawMessage(`{}`),
			},
			verify: func(t *testing.T, m *mockSchemaProjectorQueries) {
				if !m.updateStatusCalled {
					t.Fatal("expected UpdateSchemaProjectionStatus to be called")
				}
				if m.updateStatusArgs.SchemaID != "schema-789" {
					t.Errorf("SchemaID = %q, want %q", m.updateStatusArgs.SchemaID, "schema-789")
				}
				if m.updateStatusArgs.Status != "active" {
					t.Errorf("Status = %q, want %q", m.updateStatusArgs.Status, "active")
				}
			},
		},
		{
			name: "unknown event returns nil and calls nothing",
			event: eventsourcing.EventEnvelope{
				Type:        "SomeOther.Event",
				AggregateID: "agg-1",
				Data:        json.RawMessage(`{}`),
			},
			verify: func(t *testing.T, m *mockSchemaProjectorQueries) {
				if m.anyCalled() {
					t.Fatal("expected no queries to be called for unknown event")
				}
			},
		},
		{
			name: "malformed JSON returns error",
			event: eventsourcing.EventEnvelope{
				Type:        schemadomain.EventTypeSchemaRegistered,
				AggregateID: "schema-bad",
				Data:        json.RawMessage(`{not valid json`),
			},
			wantErr: true,
			verify: func(t *testing.T, m *mockSchemaProjectorQueries) {
				if m.anyCalled() {
					t.Fatal("expected no queries to be called on malformed JSON")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSchemaProjectorQueries{}
			projector := &SchemaProjector{queries: mock}

			err := projector.Handle(context.Background(), &tt.event)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.verify(t, mock)
		})
	}
}

func TestSchemaProjector_Handle_QueryError(t *testing.T) {
	data, _ := json.Marshal(schemadomain.SchemaRegisteredEvent{
		SchemaID:   "schema-err",
		SchemaType: "github",
		Claims:     []schemadomain.ClaimDefinitionData{},
	})

	mock := &mockSchemaProjectorQueries{err: fmt.Errorf("db connection lost")}
	projector := &SchemaProjector{queries: mock}

	err := projector.Handle(context.Background(), &eventsourcing.EventEnvelope{
		Type:        schemadomain.EventTypeSchemaRegistered,
		AggregateID: "schema-err",
		Data:        data,
	})
	if err == nil {
		t.Fatal("expected error from query failure, got nil")
	}
}
