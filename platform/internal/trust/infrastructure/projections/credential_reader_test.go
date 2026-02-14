package projections

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockCredentialReaderQueries struct {
	existsResult bool
	existsErr    error
	existsArg    string

	countResult int32
	countErr    error
	countArg    string
}

func (m *mockCredentialReaderQueries) CredentialProjectionExists(ctx context.Context, credentialID string) (bool, error) {
	m.existsArg = credentialID
	return m.existsResult, m.existsErr
}

func (m *mockCredentialReaderQueries) CountCredentialProjectionsByUserID(ctx context.Context, userID string) (int32, error) {
	m.countArg = userID
	return m.countResult, m.countErr
}

func TestCredentialReader_GetCredential(t *testing.T) {
	t.Run("credential exists", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{existsResult: true}
		reader := &CredentialReader{queries: mock}

		err := reader.GetCredential(context.Background(), "cred-001")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if mock.existsArg != "cred-001" {
			t.Errorf("expected credentialID %q, got %q", "cred-001", mock.existsArg)
		}
	})

	t.Run("credential not found", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{existsResult: false}
		reader := &CredentialReader{queries: mock}

		err := reader.GetCredential(context.Background(), "cred-missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "credential not found") {
			t.Errorf("expected error containing %q, got %q", "credential not found", err.Error())
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{existsErr: errors.New("db connection failed")}
		reader := &CredentialReader{queries: mock}

		err := reader.GetCredential(context.Background(), "cred-err")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "db connection failed") {
			t.Errorf("expected wrapped error containing %q, got %q", "db connection failed", err.Error())
		}
	})
}

func TestCredentialReader_CountUserCredentials(t *testing.T) {
	t.Run("returns count", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{countResult: 3}
		reader := &CredentialReader{queries: mock}

		count, err := reader.CountUserCredentials(context.Background(), "user-001")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if count != 3 {
			t.Errorf("expected count 3, got %d", count)
		}
		if mock.countArg != "user-001" {
			t.Errorf("expected userID %q, got %q", "user-001", mock.countArg)
		}
	})

	t.Run("returns zero", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{countResult: 0}
		reader := &CredentialReader{queries: mock}

		count, err := reader.CountUserCredentials(context.Background(), "user-empty")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock := &mockCredentialReaderQueries{countErr: errors.New("db timeout")}
		reader := &CredentialReader{queries: mock}

		_, err := reader.CountUserCredentials(context.Background(), "user-err")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "db timeout") {
			t.Errorf("expected error containing %q, got %q", "db timeout", err.Error())
		}
	})
}
