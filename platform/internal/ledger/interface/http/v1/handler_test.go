package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/ledger/app/query"
	"github.com/0xsj/nexus/platform/internal/ledger/domain"
)

// ============================================================================
// Mock Query Handler
// ============================================================================

type mockQueryHandler struct {
	handleGetEntryFunc             func(ctx context.Context, q query.GetEntry) (*query.EntryView, error)
	handleGetActivityFunc          func(ctx context.Context, q query.GetActivity) (*query.ActivityView, error)
	handleGetUserActivityFunc      func(ctx context.Context, q query.GetUserActivity) (*query.UserActivityView, error)
	handleGetCredentialHistoryFunc func(ctx context.Context, q query.GetCredentialHistory) (*query.CredentialHistoryView, error)
	handleGetSubjectHistoryFunc    func(ctx context.Context, q query.GetSubjectHistory) (*query.SubjectHistoryView, error)
	handleGetVerificationLogFunc   func(ctx context.Context, q query.GetVerificationLog) (*query.VerificationLogView, error)
	handleGetActivityStatsFunc     func(ctx context.Context, q query.GetActivityStats) (*query.ActivityStatsView, error)
}

func (m *mockQueryHandler) HandleGetEntry(ctx context.Context, q query.GetEntry) (*query.EntryView, error) {
	if m.handleGetEntryFunc != nil {
		return m.handleGetEntryFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueryHandler) HandleGetActivity(ctx context.Context, q query.GetActivity) (*query.ActivityView, error) {
	if m.handleGetActivityFunc != nil {
		return m.handleGetActivityFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueryHandler) HandleGetUserActivity(ctx context.Context, q query.GetUserActivity) (*query.UserActivityView, error) {
	if m.handleGetUserActivityFunc != nil {
		return m.handleGetUserActivityFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueryHandler) HandleGetCredentialHistory(ctx context.Context, q query.GetCredentialHistory) (*query.CredentialHistoryView, error) {
	if m.handleGetCredentialHistoryFunc != nil {
		return m.handleGetCredentialHistoryFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueryHandler) HandleGetSubjectHistory(ctx context.Context, q query.GetSubjectHistory) (*query.SubjectHistoryView, error) {
	if m.handleGetSubjectHistoryFunc != nil {
		return m.handleGetSubjectHistoryFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueryHandler) HandleGetVerificationLog(ctx context.Context, q query.GetVerificationLog) (*query.VerificationLogView, error) {
	if m.handleGetVerificationLogFunc != nil {
		return m.handleGetVerificationLogFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueryHandler) HandleGetActivityStats(ctx context.Context, q query.GetActivityStats) (*query.ActivityStatsView, error) {
	if m.handleGetActivityStatsFunc != nil {
		return m.handleGetActivityStatsFunc(ctx, q)
	}
	return nil, errors.New("not implemented")
}

// Compile-time check
var _ QueryHandler = (*mockQueryHandler)(nil)

// ============================================================================
// Test Helpers
// ============================================================================

func newTestHandler(mock *mockQueryHandler) (*Handler, *chi.Mux) {
	h := NewHandler(mock)
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return h, r
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ============================================================================
// GetEntry Tests
// ============================================================================

func TestGetEntry_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	mock := &mockQueryHandler{
		handleGetEntryFunc: func(ctx context.Context, q query.GetEntry) (*query.EntryView, error) {
			if q.ID != "entry-123" {
				t.Errorf("unexpected ID: %v", q.ID)
			}
			return &query.EntryView{
				ID:          "entry-123",
				OccurredAt:  now,
				RecordedAt:  now,
				EventType:   "credential.issued",
				ActorID:     "user-456",
				ActorType:   "user",
				SubjectID:   "cred-789",
				SubjectType: "credential",
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/entries/entry-123")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp EntryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "entry-123" {
		t.Errorf("ID mismatch: got %v, want %v", resp.ID, "entry-123")
	}
	if resp.EventType != "credential.issued" {
		t.Errorf("EventType mismatch: got %v, want %v", resp.EventType, "credential.issued")
	}
	if resp.ActorID != "user-456" {
		t.Errorf("ActorID mismatch: got %v, want %v", resp.ActorID, "user-456")
	}
	if resp.SubjectID != "cred-789" {
		t.Errorf("SubjectID mismatch: got %v, want %v", resp.SubjectID, "cred-789")
	}
}

func TestGetEntry_NotFound(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetEntryFunc: func(ctx context.Context, q query.GetEntry) (*query.EntryView, error) {
			return nil, domain.EntryNotFound("test", q.ID)
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/entries/nonexistent")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d: %s", http.StatusNotFound, w.Code, w.Body.String())
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != string(domain.CodeEntryNotFound) {
		t.Errorf("Code mismatch: got %v, want %v", resp.Code, domain.CodeEntryNotFound)
	}
}

func TestGetEntry_InvalidID(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetEntryFunc: func(ctx context.Context, q query.GetEntry) (*query.EntryView, error) {
			return nil, domain.InvalidEntryID("test", q.ID, "invalid format")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/entries/bad-id")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestGetEntry_InternalError(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetEntryFunc: func(ctx context.Context, q query.GetEntry) (*query.EntryView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/entries/entry-123")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// GetActivity Tests
// ============================================================================

func TestGetActivity_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	mock := &mockQueryHandler{
		handleGetActivityFunc: func(ctx context.Context, q query.GetActivity) (*query.ActivityView, error) {
			return &query.ActivityView{
				Entries: []query.EntryView{
					{ID: "entry-1", EventType: "credential.issued", OccurredAt: now, RecordedAt: now},
					{ID: "entry-2", EventType: "credential.revoked", OccurredAt: now, RecordedAt: now},
				},
				TotalCount: 2,
				Page:       1,
				PageSize:   20,
				HasMore:    false,
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/activity")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp ActivityResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Entries))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v, want %v", resp.TotalCount, 2)
	}
}

func TestGetActivity_WithQueryParams(t *testing.T) {
	var capturedQuery query.GetActivity

	mock := &mockQueryHandler{
		handleGetActivityFunc: func(ctx context.Context, q query.GetActivity) (*query.ActivityView, error) {
			capturedQuery = q
			return &query.ActivityView{}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/activity?actor_id=user-123&actor_type=user&subject_id=cred-456&subject_type=credential&context_id=corr-789&page=2&page_size=10")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	if capturedQuery.ActorID != "user-123" {
		t.Errorf("ActorID mismatch: got %v", capturedQuery.ActorID)
	}
	if capturedQuery.ActorType != "user" {
		t.Errorf("ActorType mismatch: got %v", capturedQuery.ActorType)
	}
	if capturedQuery.SubjectID != "cred-456" {
		t.Errorf("SubjectID mismatch: got %v", capturedQuery.SubjectID)
	}
	if capturedQuery.SubjectType != "credential" {
		t.Errorf("SubjectType mismatch: got %v", capturedQuery.SubjectType)
	}
	if capturedQuery.ContextID != "corr-789" {
		t.Errorf("ContextID mismatch: got %v", capturedQuery.ContextID)
	}
	if capturedQuery.Page != 2 {
		t.Errorf("Page mismatch: got %v", capturedQuery.Page)
	}
	if capturedQuery.PageSize != 10 {
		t.Errorf("PageSize mismatch: got %v", capturedQuery.PageSize)
	}
}

func TestGetActivity_WithTimeRange(t *testing.T) {
	var capturedQuery query.GetActivity

	mock := &mockQueryHandler{
		handleGetActivityFunc: func(ctx context.Context, q query.GetActivity) (*query.ActivityView, error) {
			capturedQuery = q
			return &query.ActivityView{}, nil
		},
	}

	_, r := newTestHandler(mock)

	fromTime := "2024-01-01T00:00:00Z"
	toTime := "2024-12-31T23:59:59Z"
	w := performRequest(r, "GET", "/api/v1/ledger/activity?from_time="+fromTime+"&to_time="+toTime)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	expectedFrom, _ := time.Parse(time.RFC3339, fromTime)
	expectedTo, _ := time.Parse(time.RFC3339, toTime)

	if !capturedQuery.FromTime.Equal(expectedFrom) {
		t.Errorf("FromTime mismatch: got %v, want %v", capturedQuery.FromTime, expectedFrom)
	}
	if !capturedQuery.ToTime.Equal(expectedTo) {
		t.Errorf("ToTime mismatch: got %v, want %v", capturedQuery.ToTime, expectedTo)
	}
}

func TestGetActivity_Error(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetActivityFunc: func(ctx context.Context, q query.GetActivity) (*query.ActivityView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/activity")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// GetUserActivity Tests
// ============================================================================

func TestGetUserActivity_Success(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetUserActivityFunc: func(ctx context.Context, q query.GetUserActivity) (*query.UserActivityView, error) {
			if q.UserID != "user-123" {
				t.Errorf("UserID mismatch: got %v", q.UserID)
			}
			return &query.UserActivityView{
				UserID: "user-123",
				Activity: query.ActivityView{
					Entries:    []query.EntryView{{ID: "entry-1"}},
					TotalCount: 1,
					Page:       1,
					PageSize:   20,
				},
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/users/user-123/activity")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp UserActivityResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.UserID != "user-123" {
		t.Errorf("UserID mismatch: got %v", resp.UserID)
	}
	if len(resp.Activity.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp.Activity.Entries))
	}
}

func TestGetUserActivity_WithFilters(t *testing.T) {
	var capturedQuery query.GetUserActivity

	mock := &mockQueryHandler{
		handleGetUserActivityFunc: func(ctx context.Context, q query.GetUserActivity) (*query.UserActivityView, error) {
			capturedQuery = q
			return &query.UserActivityView{UserID: q.UserID}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/users/user-123/activity?page=3&page_size=15")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	if capturedQuery.UserID != "user-123" {
		t.Errorf("UserID mismatch: got %v", capturedQuery.UserID)
	}
	if capturedQuery.Page != 3 {
		t.Errorf("Page mismatch: got %v", capturedQuery.Page)
	}
	if capturedQuery.PageSize != 15 {
		t.Errorf("PageSize mismatch: got %v", capturedQuery.PageSize)
	}
}

func TestGetUserActivity_Error(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetUserActivityFunc: func(ctx context.Context, q query.GetUserActivity) (*query.UserActivityView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/users/user-123/activity")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// GetCredentialHistory Tests
// ============================================================================

func TestGetCredentialHistory_Success(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetCredentialHistoryFunc: func(ctx context.Context, q query.GetCredentialHistory) (*query.CredentialHistoryView, error) {
			if q.CredentialID != "cred-123" {
				t.Errorf("CredentialID mismatch: got %v", q.CredentialID)
			}
			return &query.CredentialHistoryView{
				CredentialID: "cred-123",
				Entries: []query.EntryView{
					{ID: "entry-1", EventType: "credential.issued"},
					{ID: "entry-2", EventType: "credential.updated"},
				},
				TotalCount: 2,
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/credentials/cred-123/history")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp CredentialHistoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.CredentialID != "cred-123" {
		t.Errorf("CredentialID mismatch: got %v", resp.CredentialID)
	}
	if len(resp.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Entries))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount mismatch: got %v", resp.TotalCount)
	}
}

func TestGetCredentialHistory_Error(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetCredentialHistoryFunc: func(ctx context.Context, q query.GetCredentialHistory) (*query.CredentialHistoryView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/credentials/cred-123/history")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// GetSubjectHistory Tests
// ============================================================================

func TestGetSubjectHistory_Success(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetSubjectHistoryFunc: func(ctx context.Context, q query.GetSubjectHistory) (*query.SubjectHistoryView, error) {
			if q.SubjectType != "wallet" {
				t.Errorf("SubjectType mismatch: got %v", q.SubjectType)
			}
			if q.SubjectID != "wallet-123" {
				t.Errorf("SubjectID mismatch: got %v", q.SubjectID)
			}
			return &query.SubjectHistoryView{
				SubjectID:   "wallet-123",
				SubjectType: "wallet",
				Entries:     []query.EntryView{{ID: "entry-1"}},
				TotalCount:  1,
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/subjects/wallet/wallet-123/history")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp SubjectHistoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.SubjectID != "wallet-123" {
		t.Errorf("SubjectID mismatch: got %v", resp.SubjectID)
	}
	if resp.SubjectType != "wallet" {
		t.Errorf("SubjectType mismatch: got %v", resp.SubjectType)
	}
}

func TestGetSubjectHistory_WithPagination(t *testing.T) {
	var capturedQuery query.GetSubjectHistory

	mock := &mockQueryHandler{
		handleGetSubjectHistoryFunc: func(ctx context.Context, q query.GetSubjectHistory) (*query.SubjectHistoryView, error) {
			capturedQuery = q
			return &query.SubjectHistoryView{
				SubjectID:   q.SubjectID,
				SubjectType: q.SubjectType,
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/subjects/user/user-123/history?page=2&page_size=25")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	if capturedQuery.Page != 2 {
		t.Errorf("Page mismatch: got %v", capturedQuery.Page)
	}
	if capturedQuery.PageSize != 25 {
		t.Errorf("PageSize mismatch: got %v", capturedQuery.PageSize)
	}
}

func TestGetSubjectHistory_Error(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetSubjectHistoryFunc: func(ctx context.Context, q query.GetSubjectHistory) (*query.SubjectHistoryView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/subjects/wallet/wallet-123/history")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// GetVerificationLog Tests
// ============================================================================

func TestGetVerificationLog_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	mock := &mockQueryHandler{
		handleGetVerificationLogFunc: func(ctx context.Context, q query.GetVerificationLog) (*query.VerificationLogView, error) {
			if q.UserID != "user-123" {
				t.Errorf("UserID mismatch: got %v", q.UserID)
			}
			return &query.VerificationLogView{
				UserID: "user-123",
				Entries: []query.VerificationLogEntry{
					{ID: "ver-1", CredentialID: "cred-1", Outcome: "success", OccurredAt: now},
					{ID: "ver-2", CredentialID: "cred-2", Outcome: "failed", OccurredAt: now},
				},
				TotalCount: 2,
				Page:       1,
				PageSize:   20,
				HasMore:    false,
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/users/user-123/verifications")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp VerificationLogResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.UserID != "user-123" {
		t.Errorf("UserID mismatch: got %v", resp.UserID)
	}
	if len(resp.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Entries))
	}
}

func TestGetVerificationLog_WithFilters(t *testing.T) {
	var capturedQuery query.GetVerificationLog

	mock := &mockQueryHandler{
		handleGetVerificationLogFunc: func(ctx context.Context, q query.GetVerificationLog) (*query.VerificationLogView, error) {
			capturedQuery = q
			return &query.VerificationLogView{UserID: q.UserID}, nil
		},
	}

	_, r := newTestHandler(mock)

	fromTime := "2024-01-01T00:00:00Z"
	toTime := "2024-12-31T23:59:59Z"
	w := performRequest(r, "GET", "/api/v1/ledger/users/user-123/verifications?from_time="+fromTime+"&to_time="+toTime+"&page=2&page_size=50")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	if capturedQuery.Page != 2 {
		t.Errorf("Page mismatch: got %v", capturedQuery.Page)
	}
	if capturedQuery.PageSize != 50 {
		t.Errorf("PageSize mismatch: got %v", capturedQuery.PageSize)
	}

	expectedFrom, _ := time.Parse(time.RFC3339, fromTime)
	if !capturedQuery.FromTime.Equal(expectedFrom) {
		t.Errorf("FromTime mismatch: got %v", capturedQuery.FromTime)
	}
}

func TestGetVerificationLog_Error(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetVerificationLogFunc: func(ctx context.Context, q query.GetVerificationLog) (*query.VerificationLogView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/users/user-123/verifications")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// GetActivityStats Tests
// ============================================================================

func TestGetActivityStats_Success(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetActivityStatsFunc: func(ctx context.Context, q query.GetActivityStats) (*query.ActivityStatsView, error) {
			return &query.ActivityStatsView{
				TotalEntries: 100,
				EventTypeCounts: []query.EventTypeCount{
					{EventType: "credential.issued", Count: 50},
					{EventType: "credential.revoked", Count: 30},
					{EventType: "user.registered", Count: 20},
				},
			}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/stats")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp ActivityStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.TotalEntries != 100 {
		t.Errorf("TotalEntries mismatch: got %v", resp.TotalEntries)
	}
	if len(resp.EventTypeCounts) != 3 {
		t.Errorf("expected 3 event type counts, got %d", len(resp.EventTypeCounts))
	}
}

func TestGetActivityStats_WithFilters(t *testing.T) {
	var capturedQuery query.GetActivityStats

	mock := &mockQueryHandler{
		handleGetActivityStatsFunc: func(ctx context.Context, q query.GetActivityStats) (*query.ActivityStatsView, error) {
			capturedQuery = q
			return &query.ActivityStatsView{}, nil
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/stats?actor_id=user-123&subject_id=cred-456&subject_type=credential")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	if capturedQuery.ActorID != "user-123" {
		t.Errorf("ActorID mismatch: got %v", capturedQuery.ActorID)
	}
	if capturedQuery.SubjectID != "cred-456" {
		t.Errorf("SubjectID mismatch: got %v", capturedQuery.SubjectID)
	}
	if capturedQuery.SubjectType != "credential" {
		t.Errorf("SubjectType mismatch: got %v", capturedQuery.SubjectType)
	}
}

func TestGetActivityStats_Error(t *testing.T) {
	mock := &mockQueryHandler{
		handleGetActivityStatsFunc: func(ctx context.Context, q query.GetActivityStats) (*query.ActivityStatsView, error) {
			return nil, errors.New("database error")
		},
	}

	_, r := newTestHandler(mock)
	w := performRequest(r, "GET", "/api/v1/ledger/stats")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

// ============================================================================
// Error Mapping Tests
// ============================================================================

func TestMapError_EntryNotFound(t *testing.T) {
	err := domain.EntryNotFound("test", "entry-123")
	status, resp := MapError(err)

	if status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
	}
	if resp.Code != string(domain.CodeEntryNotFound) {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestMapError_InvalidEntryID(t *testing.T) {
	err := domain.InvalidEntryID("test", "bad-id", "invalid format")
	status, resp := MapError(err)

	if status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
	if resp.Code != string(domain.CodeInvalidEntryID) {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestMapError_InvalidActorType(t *testing.T) {
	err := domain.InvalidActorType("test", "bad-type")
	status, resp := MapError(err)

	if status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
	if resp.Code != string(domain.CodeInvalidActorType) {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestMapError_InvalidSubjectType(t *testing.T) {
	err := domain.InvalidSubjectType("test", "bad-type")
	status, resp := MapError(err)

	if status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
	if resp.Code != string(domain.CodeInvalidSubjectType) {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestMapError_InvalidEventType(t *testing.T) {
	err := domain.InvalidEventType("test", "bad-type", "missing dot")
	status, resp := MapError(err)

	if status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
	if resp.Code != string(domain.CodeInvalidEventType) {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestMapError_GenericError(t *testing.T) {
	err := errors.New("something went wrong")
	status, resp := MapError(err)

	if status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, status)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestMapError_Nil(t *testing.T) {
	status, resp := MapError(nil)

	if status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, status)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

// ============================================================================
// Helper Response Tests
// ============================================================================

func TestBadRequestResponse(t *testing.T) {
	resp := BadRequestResponse("invalid input")

	if resp.Code != "BAD_REQUEST" {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
	if resp.Message != "invalid input" {
		t.Errorf("Message mismatch: got %v", resp.Message)
	}
}

func TestNotFoundResponse(t *testing.T) {
	resp := NotFoundResponse("entry")

	if resp.Code != "NOT_FOUND" {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
	if resp.Message != "entry not found" {
		t.Errorf("Message mismatch: got %v", resp.Message)
	}
}

func TestInternalErrorResponse(t *testing.T) {
	resp := InternalErrorResponse()

	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
}

func TestValidationErrorResponse(t *testing.T) {
	resp := ValidationErrorResponse("email", "must be valid email")

	if resp.Code != "VALIDATION_ERROR" {
		t.Errorf("Code mismatch: got %v", resp.Code)
	}
	if resp.Details["field"] != "email" {
		t.Errorf("field mismatch: got %v", resp.Details["field"])
	}
}

// ============================================================================
// parseInt Tests
// ============================================================================

func TestParseInt_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"100", 100},
		{"999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseInt(tt.input)
			if got != tt.want {
				t.Errorf("parseInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInt_Invalid(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"12abc",
		"-1",
		"1.5",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got := parseInt(input)
			if got != 0 {
				t.Errorf("parseInt(%q) = %v, want 0", input, got)
			}
		})
	}
}
