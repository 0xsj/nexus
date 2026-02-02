package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/0xsj/nexus/platform/internal/ledger/app/query"
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Query Handler Interface
// ============================================================================

// QueryHandler defines the interface for ledger query operations.
type QueryHandler interface {
	HandleGetEntry(ctx context.Context, q query.GetEntry) (*query.EntryView, error)
	HandleGetActivity(ctx context.Context, q query.GetActivity) (*query.ActivityView, error)
	HandleGetUserActivity(ctx context.Context, q query.GetUserActivity) (*query.UserActivityView, error)
	HandleGetCredentialHistory(ctx context.Context, q query.GetCredentialHistory) (*query.CredentialHistoryView, error)
	HandleGetSubjectHistory(ctx context.Context, q query.GetSubjectHistory) (*query.SubjectHistoryView, error)
	HandleGetVerificationLog(ctx context.Context, q query.GetVerificationLog) (*query.VerificationLogView, error)
	HandleGetActivityStats(ctx context.Context, q query.GetActivityStats) (*query.ActivityStatsView, error)
}

// Compile-time check that *query.Handlers implements QueryHandler
var _ QueryHandler = (*query.Handlers)(nil)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Ledger context.
type Handler struct {
	queries QueryHandler
}

// NewHandler creates a new Handler.
func NewHandler(queries QueryHandler) *Handler {
	return &Handler{
		queries: queries,
	}
}

// ============================================================================
// Get Entry
// ============================================================================

// GetEntry handles GET /entries/{id}
func (h *Handler) GetEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("entry id is required"))
		return
	}

	result, err := h.queries.HandleGetEntry(r.Context(), query.GetEntry{
		ID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toEntryResponse(result))
}

// ============================================================================
// Get Activity
// ============================================================================

// GetActivity handles GET /activity
func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	req := h.parseActivityRequest(r)

	result, err := h.queries.HandleGetActivity(r.Context(), query.GetActivity{
		EventTypes:  req.EventTypes,
		ActorID:     req.ActorID,
		ActorType:   req.ActorType,
		SubjectID:   req.SubjectID,
		SubjectType: req.SubjectType,
		ContextID:   req.ContextID,
		FromTime:    req.FromTime,
		ToTime:      req.ToTime,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toActivityResponse(result))
}

// ============================================================================
// Get User Activity
// ============================================================================

// GetUserActivity handles GET /users/{userId}/activity
func (h *Handler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user id is required"))
		return
	}

	req := h.parseUserActivityRequest(r)

	result, err := h.queries.HandleGetUserActivity(r.Context(), query.GetUserActivity{
		UserID:     userID,
		EventTypes: req.EventTypes,
		FromTime:   req.FromTime,
		ToTime:     req.ToTime,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toUserActivityResponse(result))
}

// ============================================================================
// Get Credential History
// ============================================================================

// GetCredentialHistory handles GET /credentials/{credentialId}/history
func (h *Handler) GetCredentialHistory(w http.ResponseWriter, r *http.Request) {
	credentialID := chi.URLParam(r, "credentialId")
	if credentialID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("credential id is required"))
		return
	}

	result, err := h.queries.HandleGetCredentialHistory(r.Context(), query.GetCredentialHistory{
		CredentialID: credentialID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toCredentialHistoryResponse(result))
}

// ============================================================================
// Get Subject History
// ============================================================================

// GetSubjectHistory handles GET /subjects/{subjectType}/{subjectId}/history
func (h *Handler) GetSubjectHistory(w http.ResponseWriter, r *http.Request) {
	subjectType := chi.URLParam(r, "subjectType")
	subjectID := chi.URLParam(r, "subjectId")

	if subjectType == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("subject type is required"))
		return
	}
	if subjectID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("subject id is required"))
		return
	}

	page, pageSize := h.parsePagination(r)

	result, err := h.queries.HandleGetSubjectHistory(r.Context(), query.GetSubjectHistory{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toSubjectHistoryResponse(result))
}

// ============================================================================
// Get Verification Log
// ============================================================================

// GetVerificationLog handles GET /users/{userId}/verifications
func (h *Handler) GetVerificationLog(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user id is required"))
		return
	}

	fromTime, toTime := h.parseTimeRange(r)
	page, pageSize := h.parsePagination(r)

	result, err := h.queries.HandleGetVerificationLog(r.Context(), query.GetVerificationLog{
		UserID:   userID,
		FromTime: fromTime,
		ToTime:   toTime,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toVerificationLogResponse(result))
}

// ============================================================================
// Get Activity Stats
// ============================================================================

// GetActivityStats handles GET /stats
func (h *Handler) GetActivityStats(w http.ResponseWriter, r *http.Request) {
	req := h.parseStatsRequest(r)

	result, err := h.queries.HandleGetActivityStats(r.Context(), query.GetActivityStats{
		ActorID:     req.ActorID,
		SubjectID:   req.SubjectID,
		SubjectType: req.SubjectType,
		FromTime:    req.FromTime,
		ToTime:      req.ToTime,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toActivityStatsResponse(result))
}

// ============================================================================
// Request Parsing Helpers
// ============================================================================

func (h *Handler) parseActivityRequest(r *http.Request) GetActivityRequest {
	fromTime, toTime := h.parseTimeRange(r)
	page, pageSize := h.parsePagination(r)

	return GetActivityRequest{
		EventTypes:  r.URL.Query()["event_types"],
		ActorID:     r.URL.Query().Get("actor_id"),
		ActorType:   r.URL.Query().Get("actor_type"),
		SubjectID:   r.URL.Query().Get("subject_id"),
		SubjectType: r.URL.Query().Get("subject_type"),
		ContextID:   r.URL.Query().Get("context_id"),
		FromTime:    fromTime,
		ToTime:      toTime,
		Page:        page,
		PageSize:    pageSize,
	}
}

func (h *Handler) parseUserActivityRequest(r *http.Request) GetUserActivityRequest {
	fromTime, toTime := h.parseTimeRange(r)
	page, pageSize := h.parsePagination(r)

	return GetUserActivityRequest{
		EventTypes: r.URL.Query()["event_types"],
		FromTime:   fromTime,
		ToTime:     toTime,
		Page:       page,
		PageSize:   pageSize,
	}
}

func (h *Handler) parseStatsRequest(r *http.Request) GetActivityStatsRequest {
	fromTime, toTime := h.parseTimeRange(r)

	return GetActivityStatsRequest{
		ActorID:     r.URL.Query().Get("actor_id"),
		SubjectID:   r.URL.Query().Get("subject_id"),
		SubjectType: r.URL.Query().Get("subject_type"),
		FromTime:    fromTime,
		ToTime:      toTime,
	}
}

func (h *Handler) parseTimeRange(r *http.Request) (time.Time, time.Time) {
	var fromTime, toTime time.Time

	if from := r.URL.Query().Get("from_time"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			fromTime = t
		}
	}

	if to := r.URL.Query().Get("to_time"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			toTime = t
		}
	}

	return fromTime, toTime
}

func (h *Handler) parsePagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed := parseInt(p); parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed := parseInt(ps); parsed > 0 {
			pageSize = parsed
		}
	}

	return page, pageSize
}

func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return n
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func toEntryResponse(v *query.EntryView) EntryResponse {
	return EntryResponse{
		ID:          v.ID,
		OccurredAt:  v.OccurredAt,
		RecordedAt:  v.RecordedAt,
		EventType:   v.EventType,
		ActorID:     v.ActorID,
		ActorType:   v.ActorType,
		SubjectID:   v.SubjectID,
		SubjectType: v.SubjectType,
		Metadata:    v.Metadata,
		ContextID:   v.ContextID,
	}
}

func toEntryResponses(views []query.EntryView) []EntryResponse {
	responses := make([]EntryResponse, len(views))
	for i, v := range views {
		responses[i] = toEntryResponse(&v)
	}
	return responses
}

func toActivityResponse(v *query.ActivityView) ActivityResponse {
	return ActivityResponse{
		Entries:    toEntryResponses(v.Entries),
		TotalCount: v.TotalCount,
		Page:       v.Page,
		PageSize:   v.PageSize,
		HasMore:    v.HasMore,
	}
}

func toUserActivityResponse(v *query.UserActivityView) UserActivityResponse {
	return UserActivityResponse{
		UserID:   v.UserID,
		Activity: toActivityResponse(&v.Activity),
	}
}

func toCredentialHistoryResponse(v *query.CredentialHistoryView) CredentialHistoryResponse {
	return CredentialHistoryResponse{
		CredentialID: v.CredentialID,
		Entries:      toEntryResponses(v.Entries),
		TotalCount:   v.TotalCount,
	}
}

func toSubjectHistoryResponse(v *query.SubjectHistoryView) SubjectHistoryResponse {
	return SubjectHistoryResponse{
		SubjectID:   v.SubjectID,
		SubjectType: v.SubjectType,
		Entries:     toEntryResponses(v.Entries),
		TotalCount:  v.TotalCount,
	}
}

func toVerificationLogResponse(v *query.VerificationLogView) VerificationLogResponse {
	entries := make([]VerificationLogEntryResponse, len(v.Entries))
	for i, e := range v.Entries {
		entries[i] = VerificationLogEntryResponse{
			ID:           e.ID,
			OccurredAt:   e.OccurredAt,
			VerifierID:   e.VerifierID,
			VerifierType: e.VerifierType,
			CredentialID: e.CredentialID,
			Outcome:      e.Outcome,
			Metadata:     e.Metadata,
		}
	}

	return VerificationLogResponse{
		UserID:     v.UserID,
		Entries:    entries,
		TotalCount: v.TotalCount,
		Page:       v.Page,
		PageSize:   v.PageSize,
		HasMore:    v.HasMore,
	}
}

func toActivityStatsResponse(v *query.ActivityStatsView) ActivityStatsResponse {
	counts := make([]EventTypeCountResponse, len(v.EventTypeCounts))
	for i, c := range v.EventTypeCounts {
		counts[i] = EventTypeCountResponse{
			EventType: c.EventType,
			Count:     c.Count,
		}
	}

	return ActivityStatsResponse{
		TotalEntries:    v.TotalEntries,
		EventTypeCounts: counts,
		FromTime:        v.FromTime,
		ToTime:          v.ToTime,
	}
}

// ============================================================================
// Response Writing Helpers
// ============================================================================

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errResp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}
