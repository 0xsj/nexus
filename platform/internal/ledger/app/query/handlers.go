package query

import (
	"context"
)

// ============================================================================
// Query Handlers
// ============================================================================

// Handlers contains all query handlers for the Ledger context.
type Handlers struct {
	reader Reader
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(reader Reader) *Handlers {
	return &Handlers{
		reader: reader,
	}
}

// ============================================================================
// Handler Methods
// ============================================================================

// HandleGetEntry handles the GetEntry query.
func (h *Handlers) HandleGetEntry(ctx context.Context, q GetEntry) (*EntryView, error) {
	return h.reader.GetByID(ctx, q.ID)
}

// HandleGetActivity handles the GetActivity query.
func (h *Handlers) HandleGetActivity(ctx context.Context, q GetActivity) (*ActivityView, error) {
	page, pageSize := NormalizePagination(q.Page, q.PageSize)

	return h.reader.List(ctx, ListFilter{
		EventTypes:  q.EventTypes,
		ActorID:     q.ActorID,
		ActorType:   q.ActorType,
		SubjectID:   q.SubjectID,
		SubjectType: q.SubjectType,
		ContextID:   q.ContextID,
		FromTime:    q.FromTime,
		ToTime:      q.ToTime,
		Page:        page,
		PageSize:    pageSize,
	})
}

// HandleGetUserActivity handles the GetUserActivity query.
func (h *Handlers) HandleGetUserActivity(ctx context.Context, q GetUserActivity) (*UserActivityView, error) {
	page, pageSize := NormalizePagination(q.Page, q.PageSize)

	activity, err := h.reader.GetByActor(ctx, ActorFilter{
		ActorID:    q.UserID,
		ActorType:  "user",
		EventTypes: q.EventTypes,
		FromTime:   q.FromTime,
		ToTime:     q.ToTime,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, err
	}

	return &UserActivityView{
		UserID:   q.UserID,
		Activity: *activity,
	}, nil
}

// HandleGetCredentialHistory handles the GetCredentialHistory query.
func (h *Handlers) HandleGetCredentialHistory(ctx context.Context, q GetCredentialHistory) (*CredentialHistoryView, error) {
	history, err := h.reader.GetBySubject(ctx, SubjectFilter{
		SubjectID:   q.CredentialID,
		SubjectType: "credential",
	})
	if err != nil {
		return nil, err
	}

	return &CredentialHistoryView{
		CredentialID: q.CredentialID,
		Entries:      history.Entries,
		TotalCount:   history.TotalCount,
	}, nil
}

// HandleGetSubjectHistory handles the GetSubjectHistory query.
func (h *Handlers) HandleGetSubjectHistory(ctx context.Context, q GetSubjectHistory) (*SubjectHistoryView, error) {
	page, pageSize := NormalizePagination(q.Page, q.PageSize)

	return h.reader.GetBySubject(ctx, SubjectFilter{
		SubjectID:   q.SubjectID,
		SubjectType: q.SubjectType,
		Page:        page,
		PageSize:    pageSize,
	})
}

// HandleGetVerificationLog handles the GetVerificationLog query.
func (h *Handlers) HandleGetVerificationLog(ctx context.Context, q GetVerificationLog) (*VerificationLogView, error) {
	page, pageSize := NormalizePagination(q.Page, q.PageSize)

	return h.reader.GetVerificationLog(ctx, VerificationLogFilter{
		UserID:   q.UserID,
		FromTime: q.FromTime,
		ToTime:   q.ToTime,
		Page:     page,
		PageSize: pageSize,
	})
}

// HandleGetActivityStats handles the GetActivityStats query.
func (h *Handlers) HandleGetActivityStats(ctx context.Context, q GetActivityStats) (*ActivityStatsView, error) {
	return h.reader.GetStats(ctx, StatsFilter{
		ActorID:     q.ActorID,
		SubjectID:   q.SubjectID,
		SubjectType: q.SubjectType,
		FromTime:    q.FromTime,
		ToTime:      q.ToTime,
	})
}
