package query

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ============================================================================
// Mock Reader
// ============================================================================

type mockReader struct {
	getByIDFunc            func(ctx context.Context, id string) (*EntryView, error)
	listFunc               func(ctx context.Context, filter ListFilter) (*ActivityView, error)
	getBySubjectFunc       func(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error)
	getByActorFunc         func(ctx context.Context, filter ActorFilter) (*ActivityView, error)
	getVerificationLogFunc func(ctx context.Context, filter VerificationLogFilter) (*VerificationLogView, error)
	getStatsFunc           func(ctx context.Context, filter StatsFilter) (*ActivityStatsView, error)
}

func (m *mockReader) GetByID(ctx context.Context, id string) (*EntryView, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReader) List(ctx context.Context, filter ListFilter) (*ActivityView, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReader) GetBySubject(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error) {
	if m.getBySubjectFunc != nil {
		return m.getBySubjectFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReader) GetByActor(ctx context.Context, filter ActorFilter) (*ActivityView, error) {
	if m.getByActorFunc != nil {
		return m.getByActorFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReader) GetVerificationLog(ctx context.Context, filter VerificationLogFilter) (*VerificationLogView, error) {
	if m.getVerificationLogFunc != nil {
		return m.getVerificationLogFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReader) GetStats(ctx context.Context, filter StatsFilter) (*ActivityStatsView, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

// ============================================================================
// NewHandlers Tests
// ============================================================================

func TestNewHandlers(t *testing.T) {
	reader := &mockReader{}
	handlers := NewHandlers(reader)

	if handlers == nil {
		t.Fatal("expected non-nil Handlers")
	}
}

// ============================================================================
// HandleGetEntry Tests
// ============================================================================

func TestHandleGetEntry_Success(t *testing.T) {
	expectedEntry := &EntryView{
		ID:          "entry-123",
		EventType:   "credential.issued",
		ActorID:     "user-456",
		ActorType:   "user",
		SubjectID:   "cred-789",
		SubjectType: "credential",
	}

	reader := &mockReader{
		getByIDFunc: func(ctx context.Context, id string) (*EntryView, error) {
			if id != "entry-123" {
				t.Errorf("unexpected ID: %v", id)
			}
			return expectedEntry, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetEntry(context.Background(), GetEntry{ID: "entry-123"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != expectedEntry.ID {
		t.Errorf("ID mismatch: got %v, want %v", result.ID, expectedEntry.ID)
	}
	if result.EventType != expectedEntry.EventType {
		t.Errorf("EventType mismatch: got %v, want %v", result.EventType, expectedEntry.EventType)
	}
}

func TestHandleGetEntry_NotFound(t *testing.T) {
	reader := &mockReader{
		getByIDFunc: func(ctx context.Context, id string) (*EntryView, error) {
			return nil, errors.New("entry not found")
		},
	}

	handlers := NewHandlers(reader)
	_, err := handlers.HandleGetEntry(context.Background(), GetEntry{ID: "nonexistent"})

	if err == nil {
		t.Error("expected error for nonexistent entry")
	}
}

// ============================================================================
// HandleGetActivity Tests
// ============================================================================

func TestHandleGetActivity_Success(t *testing.T) {
	expectedActivity := &ActivityView{
		Entries: []EntryView{
			{ID: "entry-1", EventType: "credential.issued"},
			{ID: "entry-2", EventType: "credential.revoked"},
		},
		TotalCount: 2,
		Page:       1,
		PageSize:   20,
		HasMore:    false,
	}

	reader := &mockReader{
		listFunc: func(ctx context.Context, filter ListFilter) (*ActivityView, error) {
			return expectedActivity, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetActivity(context.Background(), GetActivity{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v, want %v", result.TotalCount, 2)
	}
}

func TestHandleGetActivity_WithFilters(t *testing.T) {
	var capturedFilter ListFilter

	reader := &mockReader{
		listFunc: func(ctx context.Context, filter ListFilter) (*ActivityView, error) {
			capturedFilter = filter
			return &ActivityView{}, nil
		},
	}

	handlers := NewHandlers(reader)
	fromTime := time.Now().Add(-24 * time.Hour)
	toTime := time.Now()

	_, err := handlers.HandleGetActivity(context.Background(), GetActivity{
		EventTypes:  []string{"credential.issued", "credential.revoked"},
		ActorID:     "user-123",
		ActorType:   "user",
		SubjectID:   "cred-456",
		SubjectType: "credential",
		ContextID:   "corr-789",
		FromTime:    fromTime,
		ToTime:      toTime,
		Page:        2,
		PageSize:    10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedFilter.EventTypes) != 2 {
		t.Errorf("EventTypes mismatch: got %v", capturedFilter.EventTypes)
	}
	if capturedFilter.ActorID != "user-123" {
		t.Errorf("ActorID mismatch: got %v", capturedFilter.ActorID)
	}
	if capturedFilter.ActorType != "user" {
		t.Errorf("ActorType mismatch: got %v", capturedFilter.ActorType)
	}
	if capturedFilter.SubjectID != "cred-456" {
		t.Errorf("SubjectID mismatch: got %v", capturedFilter.SubjectID)
	}
	if capturedFilter.SubjectType != "credential" {
		t.Errorf("SubjectType mismatch: got %v", capturedFilter.SubjectType)
	}
	if capturedFilter.ContextID != "corr-789" {
		t.Errorf("ContextID mismatch: got %v", capturedFilter.ContextID)
	}
	if capturedFilter.Page != 2 {
		t.Errorf("Page mismatch: got %v, want %v", capturedFilter.Page, 2)
	}
	if capturedFilter.PageSize != 10 {
		t.Errorf("PageSize mismatch: got %v, want %v", capturedFilter.PageSize, 10)
	}
}

func TestHandleGetActivity_NormalizesPagination(t *testing.T) {
	var capturedFilter ListFilter

	reader := &mockReader{
		listFunc: func(ctx context.Context, filter ListFilter) (*ActivityView, error) {
			capturedFilter = filter
			return &ActivityView{}, nil
		},
	}

	handlers := NewHandlers(reader)

	// Zero values should be normalized
	_, err := handlers.HandleGetActivity(context.Background(), GetActivity{
		Page:     0,
		PageSize: 0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedFilter.Page != 1 {
		t.Errorf("Page should be normalized to 1, got %v", capturedFilter.Page)
	}
	if capturedFilter.PageSize != DefaultPageSize {
		t.Errorf("PageSize should be normalized to %v, got %v", DefaultPageSize, capturedFilter.PageSize)
	}
}

func TestHandleGetActivity_Error(t *testing.T) {
	reader := &mockReader{
		listFunc: func(ctx context.Context, filter ListFilter) (*ActivityView, error) {
			return nil, errors.New("database error")
		},
	}

	handlers := NewHandlers(reader)
	_, err := handlers.HandleGetActivity(context.Background(), GetActivity{})

	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================================
// HandleGetUserActivity Tests
// ============================================================================

func TestHandleGetUserActivity_Success(t *testing.T) {
	reader := &mockReader{
		getByActorFunc: func(ctx context.Context, filter ActorFilter) (*ActivityView, error) {
			if filter.ActorID != "user-123" {
				t.Errorf("ActorID mismatch: got %v", filter.ActorID)
			}
			if filter.ActorType != "user" {
				t.Errorf("ActorType mismatch: got %v", filter.ActorType)
			}
			return &ActivityView{
				Entries:    []EntryView{{ID: "entry-1"}},
				TotalCount: 1,
			}, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetUserActivity(context.Background(), GetUserActivity{
		UserID: "user-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != "user-123" {
		t.Errorf("UserID mismatch: got %v", result.UserID)
	}
	if len(result.Activity.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Activity.Entries))
	}
}

func TestHandleGetUserActivity_WithFilters(t *testing.T) {
	var capturedFilter ActorFilter

	reader := &mockReader{
		getByActorFunc: func(ctx context.Context, filter ActorFilter) (*ActivityView, error) {
			capturedFilter = filter
			return &ActivityView{}, nil
		},
	}

	handlers := NewHandlers(reader)
	fromTime := time.Now().Add(-24 * time.Hour)
	toTime := time.Now()

	_, err := handlers.HandleGetUserActivity(context.Background(), GetUserActivity{
		UserID:     "user-123",
		EventTypes: []string{"credential.issued"},
		FromTime:   fromTime,
		ToTime:     toTime,
		Page:       3,
		PageSize:   15,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedFilter.EventTypes) != 1 {
		t.Errorf("EventTypes mismatch")
	}
	if capturedFilter.Page != 3 {
		t.Errorf("Page mismatch: got %v", capturedFilter.Page)
	}
	if capturedFilter.PageSize != 15 {
		t.Errorf("PageSize mismatch: got %v", capturedFilter.PageSize)
	}
}

func TestHandleGetUserActivity_Error(t *testing.T) {
	reader := &mockReader{
		getByActorFunc: func(ctx context.Context, filter ActorFilter) (*ActivityView, error) {
			return nil, errors.New("database error")
		},
	}

	handlers := NewHandlers(reader)
	_, err := handlers.HandleGetUserActivity(context.Background(), GetUserActivity{
		UserID: "user-123",
	})

	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================================
// HandleGetCredentialHistory Tests
// ============================================================================

func TestHandleGetCredentialHistory_Success(t *testing.T) {
	reader := &mockReader{
		getBySubjectFunc: func(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error) {
			if filter.SubjectID != "cred-123" {
				t.Errorf("SubjectID mismatch: got %v", filter.SubjectID)
			}
			if filter.SubjectType != "credential" {
				t.Errorf("SubjectType mismatch: got %v", filter.SubjectType)
			}
			return &SubjectHistoryView{
				SubjectID:   "cred-123",
				SubjectType: "credential",
				Entries: []EntryView{
					{ID: "entry-1", EventType: "credential.issued"},
					{ID: "entry-2", EventType: "credential.updated"},
				},
				TotalCount: 2,
			}, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetCredentialHistory(context.Background(), GetCredentialHistory{
		CredentialID: "cred-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CredentialID != "cred-123" {
		t.Errorf("CredentialID mismatch: got %v", result.CredentialID)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v", result.TotalCount)
	}
}

func TestHandleGetCredentialHistory_Error(t *testing.T) {
	reader := &mockReader{
		getBySubjectFunc: func(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error) {
			return nil, errors.New("database error")
		},
	}

	handlers := NewHandlers(reader)
	_, err := handlers.HandleGetCredentialHistory(context.Background(), GetCredentialHistory{
		CredentialID: "cred-123",
	})

	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================================
// HandleGetSubjectHistory Tests
// ============================================================================

func TestHandleGetSubjectHistory_Success(t *testing.T) {
	reader := &mockReader{
		getBySubjectFunc: func(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error) {
			if filter.SubjectID != "wallet-123" {
				t.Errorf("SubjectID mismatch: got %v", filter.SubjectID)
			}
			if filter.SubjectType != "wallet" {
				t.Errorf("SubjectType mismatch: got %v", filter.SubjectType)
			}
			return &SubjectHistoryView{
				SubjectID:   "wallet-123",
				SubjectType: "wallet",
				Entries:     []EntryView{{ID: "entry-1"}},
				TotalCount:  1,
			}, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetSubjectHistory(context.Background(), GetSubjectHistory{
		SubjectID:   "wallet-123",
		SubjectType: "wallet",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SubjectID != "wallet-123" {
		t.Errorf("SubjectID mismatch: got %v", result.SubjectID)
	}
	if result.SubjectType != "wallet" {
		t.Errorf("SubjectType mismatch: got %v", result.SubjectType)
	}
}

func TestHandleGetSubjectHistory_WithPagination(t *testing.T) {
	var capturedFilter SubjectFilter

	reader := &mockReader{
		getBySubjectFunc: func(ctx context.Context, filter SubjectFilter) (*SubjectHistoryView, error) {
			capturedFilter = filter
			return &SubjectHistoryView{}, nil
		},
	}

	handlers := NewHandlers(reader)
	_, err := handlers.HandleGetSubjectHistory(context.Background(), GetSubjectHistory{
		SubjectID:   "user-123",
		SubjectType: "user",
		Page:        2,
		PageSize:    25,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedFilter.Page != 2 {
		t.Errorf("Page mismatch: got %v", capturedFilter.Page)
	}
	if capturedFilter.PageSize != 25 {
		t.Errorf("PageSize mismatch: got %v", capturedFilter.PageSize)
	}
}

// ============================================================================
// HandleGetVerificationLog Tests
// ============================================================================

func TestHandleGetVerificationLog_Success(t *testing.T) {
	reader := &mockReader{
		getVerificationLogFunc: func(ctx context.Context, filter VerificationLogFilter) (*VerificationLogView, error) {
			if filter.UserID != "user-123" {
				t.Errorf("UserID mismatch: got %v", filter.UserID)
			}
			return &VerificationLogView{
				UserID: "user-123",
				Entries: []VerificationLogEntry{
					{ID: "ver-1", CredentialID: "cred-1", Outcome: "success"},
					{ID: "ver-2", CredentialID: "cred-2", Outcome: "failed"},
				},
				TotalCount: 2,
				Page:       1,
				PageSize:   20,
			}, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetVerificationLog(context.Background(), GetVerificationLog{
		UserID: "user-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != "user-123" {
		t.Errorf("UserID mismatch: got %v", result.UserID)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestHandleGetVerificationLog_WithFilters(t *testing.T) {
	var capturedFilter VerificationLogFilter

	reader := &mockReader{
		getVerificationLogFunc: func(ctx context.Context, filter VerificationLogFilter) (*VerificationLogView, error) {
			capturedFilter = filter
			return &VerificationLogView{}, nil
		},
	}

	handlers := NewHandlers(reader)
	fromTime := time.Now().Add(-7 * 24 * time.Hour)
	toTime := time.Now()

	_, err := handlers.HandleGetVerificationLog(context.Background(), GetVerificationLog{
		UserID:   "user-123",
		FromTime: fromTime,
		ToTime:   toTime,
		Page:     2,
		PageSize: 50,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedFilter.UserID != "user-123" {
		t.Errorf("UserID mismatch")
	}
	if capturedFilter.Page != 2 {
		t.Errorf("Page mismatch: got %v", capturedFilter.Page)
	}
	if capturedFilter.PageSize != 50 {
		t.Errorf("PageSize mismatch: got %v", capturedFilter.PageSize)
	}
}

// ============================================================================
// HandleGetActivityStats Tests
// ============================================================================

func TestHandleGetActivityStats_Success(t *testing.T) {
	reader := &mockReader{
		getStatsFunc: func(ctx context.Context, filter StatsFilter) (*ActivityStatsView, error) {
			return &ActivityStatsView{
				TotalEntries: 100,
				EventTypeCounts: []EventTypeCount{
					{EventType: "credential.issued", Count: 50},
					{EventType: "credential.revoked", Count: 30},
					{EventType: "user.registered", Count: 20},
				},
			}, nil
		},
	}

	handlers := NewHandlers(reader)
	result, err := handlers.HandleGetActivityStats(context.Background(), GetActivityStats{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalEntries != 100 {
		t.Errorf("TotalEntries mismatch: got %v", result.TotalEntries)
	}
	if len(result.EventTypeCounts) != 3 {
		t.Errorf("expected 3 event type counts, got %d", len(result.EventTypeCounts))
	}
}

func TestHandleGetActivityStats_WithFilters(t *testing.T) {
	var capturedFilter StatsFilter

	reader := &mockReader{
		getStatsFunc: func(ctx context.Context, filter StatsFilter) (*ActivityStatsView, error) {
			capturedFilter = filter
			return &ActivityStatsView{}, nil
		},
	}

	handlers := NewHandlers(reader)
	fromTime := time.Now().Add(-30 * 24 * time.Hour)
	toTime := time.Now()

	_, err := handlers.HandleGetActivityStats(context.Background(), GetActivityStats{
		ActorID:     "user-123",
		SubjectID:   "cred-456",
		SubjectType: "credential",
		FromTime:    fromTime,
		ToTime:      toTime,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedFilter.ActorID != "user-123" {
		t.Errorf("ActorID mismatch: got %v", capturedFilter.ActorID)
	}
	if capturedFilter.SubjectID != "cred-456" {
		t.Errorf("SubjectID mismatch: got %v", capturedFilter.SubjectID)
	}
	if capturedFilter.SubjectType != "credential" {
		t.Errorf("SubjectType mismatch: got %v", capturedFilter.SubjectType)
	}
}

func TestHandleGetActivityStats_Error(t *testing.T) {
	reader := &mockReader{
		getStatsFunc: func(ctx context.Context, filter StatsFilter) (*ActivityStatsView, error) {
			return nil, errors.New("database error")
		},
	}

	handlers := NewHandlers(reader)
	_, err := handlers.HandleGetActivityStats(context.Background(), GetActivityStats{})

	if err == nil {
		t.Error("expected error")
	}
}
