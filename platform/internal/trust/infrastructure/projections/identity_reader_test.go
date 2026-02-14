package projections

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockIdentityReaderQueries struct {
	exists    bool
	existsErr error

	count    int32
	countErr error
}

func (m *mockIdentityReaderQueries) UserProjectionExists(_ context.Context, _ string) (bool, error) {
	return m.exists, m.existsErr
}

func (m *mockIdentityReaderQueries) CountCredentialProjectionsByUserID(_ context.Context, _ string) (int32, error) {
	return m.count, m.countErr
}

func TestIdentityReader_GetUser(t *testing.T) {
	t.Run("user exists returns nil", func(t *testing.T) {
		mock := &mockIdentityReaderQueries{exists: true}
		reader := IdentityReader{queries: mock}

		err := reader.GetUser(context.Background(), "user-123")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("user not found returns error", func(t *testing.T) {
		mock := &mockIdentityReaderQueries{exists: false}
		reader := IdentityReader{queries: mock}

		err := reader.GetUser(context.Background(), "user-missing")
		if err == nil {
			t.Fatal("expected error for missing user, got nil")
		}
		if !strings.Contains(err.Error(), "user not found") {
			t.Errorf("expected error containing %q, got %q", "user not found", err.Error())
		}
	})

	t.Run("query error is propagated", func(t *testing.T) {
		queryErr := errors.New("database timeout")
		mock := &mockIdentityReaderQueries{existsErr: queryErr}
		reader := IdentityReader{queries: mock}

		err := reader.GetUser(context.Background(), "user-err")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, queryErr) {
			if !strings.Contains(err.Error(), "database timeout") {
				t.Errorf("expected error to wrap or contain original, got %q", err.Error())
			}
		}
	})
}

func TestIdentityReader_HasVerifiedCredential(t *testing.T) {
	t.Run("count greater than zero returns true", func(t *testing.T) {
		mock := &mockIdentityReaderQueries{count: 3}
		reader := IdentityReader{queries: mock}

		has, err := reader.HasVerifiedCredential(context.Background(), "user-123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !has {
			t.Error("expected true when count > 0")
		}
	})

	t.Run("count zero returns false", func(t *testing.T) {
		mock := &mockIdentityReaderQueries{count: 0}
		reader := IdentityReader{queries: mock}

		has, err := reader.HasVerifiedCredential(context.Background(), "user-456")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if has {
			t.Error("expected false when count == 0")
		}
	})

	t.Run("query error is returned", func(t *testing.T) {
		queryErr := errors.New("connection refused")
		mock := &mockIdentityReaderQueries{countErr: queryErr}
		reader := IdentityReader{queries: mock}

		_, err := reader.HasVerifiedCredential(context.Background(), "user-err")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, queryErr) {
			if !strings.Contains(err.Error(), "connection refused") {
				t.Errorf("expected error to wrap or contain original, got %q", err.Error())
			}
		}
	})
}
