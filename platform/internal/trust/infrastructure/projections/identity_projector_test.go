package projections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	identitydomain "github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

type mockIdentityProjectorQueries struct {
	upsertCalled bool
	upsertArg    generated.UpsertUserProjectionParams

	updateActiveCalled bool
	updateActiveArg    generated.UpdateUserProjectionActiveParams

	err error
}

func (m *mockIdentityProjectorQueries) UpsertUserProjection(_ context.Context, arg generated.UpsertUserProjectionParams) error {
	m.upsertCalled = true
	m.upsertArg = arg
	return m.err
}

func (m *mockIdentityProjectorQueries) UpdateUserProjectionActive(_ context.Context, arg generated.UpdateUserProjectionActiveParams) error {
	m.updateActiveCalled = true
	m.updateActiveArg = arg
	return m.err
}

func TestIdentityProjector_Handle(t *testing.T) {
	t.Run("User.Registered upserts projection with active true", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		evt := identitydomain.UserRegisteredEvent{
			UserID: "user-123",
		}
		data, _ := json.Marshal(evt)

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserRegistered,
			AggregateID: "agg-001",
			Data:        data,
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.upsertCalled {
			t.Fatal("expected UpsertUserProjection to be called")
		}
		if mock.upsertArg.UserID != "user-123" {
			t.Errorf("expected UserID %q, got %q", "user-123", mock.upsertArg.UserID)
		}
		if !mock.upsertArg.Active {
			t.Error("expected Active to be true")
		}
	})

	t.Run("User.Activated sets active true", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserActivated,
			AggregateID: "user-456",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateActiveCalled {
			t.Fatal("expected UpdateUserProjectionActive to be called")
		}
		if mock.updateActiveArg.UserID != "user-456" {
			t.Errorf("expected UserID %q, got %q", "user-456", mock.updateActiveArg.UserID)
		}
		if !mock.updateActiveArg.Active {
			t.Error("expected Active to be true")
		}
	})

	t.Run("User.Reactivated sets active true", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserReactivated,
			AggregateID: "user-789",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateActiveCalled {
			t.Fatal("expected UpdateUserProjectionActive to be called")
		}
		if mock.updateActiveArg.UserID != "user-789" {
			t.Errorf("expected UserID %q, got %q", "user-789", mock.updateActiveArg.UserID)
		}
		if !mock.updateActiveArg.Active {
			t.Error("expected Active to be true")
		}
	})

	t.Run("User.Suspended sets active false", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserSuspended,
			AggregateID: "user-111",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateActiveCalled {
			t.Fatal("expected UpdateUserProjectionActive to be called")
		}
		if mock.updateActiveArg.UserID != "user-111" {
			t.Errorf("expected UserID %q, got %q", "user-111", mock.updateActiveArg.UserID)
		}
		if mock.updateActiveArg.Active {
			t.Error("expected Active to be false")
		}
	})

	t.Run("User.Deleted sets active false", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserDeleted,
			AggregateID: "user-222",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !mock.updateActiveCalled {
			t.Fatal("expected UpdateUserProjectionActive to be called")
		}
		if mock.updateActiveArg.UserID != "user-222" {
			t.Errorf("expected UserID %q, got %q", "user-222", mock.updateActiveArg.UserID)
		}
		if mock.updateActiveArg.Active {
			t.Error("expected Active to be false")
		}
	})

	t.Run("unknown event returns nil without calling queries", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        "SomeOther.Event",
			AggregateID: "agg-999",
			Data:        json.RawMessage(`{}`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if mock.upsertCalled {
			t.Error("UpsertUserProjection should not have been called")
		}
		if mock.updateActiveCalled {
			t.Error("UpdateUserProjectionActive should not have been called")
		}
	})

	t.Run("malformed JSON for User.Registered returns error", func(t *testing.T) {
		mock := &mockIdentityProjectorQueries{}
		projector := IdentityProjector{queries: mock}

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserRegistered,
			AggregateID: "agg-bad",
			Data:        json.RawMessage(`{not valid json`),
		}

		err := projector.Handle(context.Background(), &envelope)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
		if mock.upsertCalled {
			t.Error("UpsertUserProjection should not have been called on unmarshal failure")
		}
	})

	t.Run("query error propagates from UpsertUserProjection", func(t *testing.T) {
		queryErr := errors.New("db connection failed")
		mock := &mockIdentityProjectorQueries{err: queryErr}
		projector := IdentityProjector{queries: mock}

		evt := identitydomain.UserRegisteredEvent{UserID: "user-err"}
		data, _ := json.Marshal(evt)

		envelope := eventsourcing.EventEnvelope{
			Type:        identitydomain.EventTypeUserRegistered,
			AggregateID: "agg-err",
			Data:        data,
		}

		err := projector.Handle(context.Background(), &envelope)
		if err == nil {
			t.Fatal("expected error to propagate, got nil")
		}
	})
}
