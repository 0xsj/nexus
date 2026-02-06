package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Verification context.
type Handlers struct {
	lookup *postgres.VerificationLookup
	logger log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.VerificationLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup: lookup,
		logger: logger,
	}
}

// ============================================================================
// GetVerification Handler
// ============================================================================

// HandleGetVerification handles the GetVerification query.
func (h *Handlers) HandleGetVerification(ctx context.Context, q GetVerification) (*VerificationView, error) {
	const op = "Handlers.HandleGetVerification"

	proj, err := h.lookup.GetByID(ctx, q.VerificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListVerificationsByUser Handler
// ============================================================================

// HandleListVerificationsByUser handles the ListVerificationsByUser query.
func (h *Handlers) HandleListVerificationsByUser(ctx context.Context, q ListVerificationsByUser) (*VerificationListView, error) {
	const op = "Handlers.HandleListVerificationsByUser"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	verifications, err := h.lookup.ListByUserID(ctx, q.UserID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]VerificationSummaryView, len(verifications))
	for i, v := range verifications {
		summaries[i] = mapProjectionToSummary(&v)
	}

	return &VerificationListView{
		Verifications: summaries,
		TotalCount:    totalCount,
		Limit:         limit,
		Offset:        q.Offset,
		HasMore:       q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps a VerificationProjection to a VerificationView.
func mapProjectionToView(proj *postgres.VerificationProjection) *VerificationView {
	return &VerificationView{
		VerificationID: proj.ID,
		UserID:         proj.UserID,
		ProviderType:   proj.ProviderType,
		Status:         proj.Status,
		OAuthState:     proj.OAuthState,
		ErrorMessage:   proj.ErrorMessage,
		ErrorCode:      proj.ErrorCode,
		CredentialID:   proj.CredentialID,
		StartedAt:      proj.StartedAt,
		CompletedAt:    proj.CompletedAt,
		CreatedAt:      proj.CreatedAt,
		UpdatedAt:      proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps a VerificationProjection to a VerificationSummaryView.
func mapProjectionToSummary(proj *postgres.VerificationProjection) VerificationSummaryView {
	return VerificationSummaryView{
		VerificationID: proj.ID,
		UserID:         proj.UserID,
		ProviderType:   proj.ProviderType,
		Status:         proj.Status,
		StartedAt:      proj.StartedAt,
		CompletedAt:    proj.CompletedAt,
	}
}
