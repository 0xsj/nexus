package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Trust context.
type Handlers struct {
	vouchLookup      *postgres.VouchLookup
	reputationLookup *postgres.ReputationLookup
	logger           log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	vouchLookup *postgres.VouchLookup,
	reputationLookup *postgres.ReputationLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		vouchLookup:      vouchLookup,
		reputationLookup: reputationLookup,
		logger:           logger,
	}
}

// ============================================================================
// GetVouch Handler
// ============================================================================

// HandleGetVouch handles the GetVouch query.
func (h *Handlers) HandleGetVouch(ctx context.Context, q GetVouch) (*VouchView, error) {
	const op = "Handlers.HandleGetVouch"

	proj, err := h.vouchLookup.GetByID(ctx, q.VouchID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListVouchesByVouchee Handler
// ============================================================================

// HandleListVouchesByVouchee handles the ListVouchesByVouchee query.
func (h *Handlers) HandleListVouchesByVouchee(ctx context.Context, q ListVouchesByVouchee) (*VouchListView, error) {
	const op = "Handlers.HandleListVouchesByVouchee"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	vouches, err := h.vouchLookup.ListByVoucheeID(ctx, q.VoucheeID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.vouchLookup.CountByVoucheeID(ctx, q.VoucheeID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]VouchSummaryView, len(vouches))
	for i, v := range vouches {
		summaries[i] = mapProjectionToSummary(&v)
	}

	return &VouchListView{
		Vouches:    summaries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// ListVouchesByVoucher Handler
// ============================================================================

// HandleListVouchesByVoucher handles the ListVouchesByVoucher query.
func (h *Handlers) HandleListVouchesByVoucher(ctx context.Context, q ListVouchesByVoucher) (*VouchListView, error) {
	const op = "Handlers.HandleListVouchesByVoucher"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	vouches, err := h.vouchLookup.ListByVoucherID(ctx, q.VoucherID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.vouchLookup.CountByVoucherID(ctx, q.VoucherID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]VouchSummaryView, len(vouches))
	for i, v := range vouches {
		summaries[i] = mapProjectionToSummary(&v)
	}

	return &VouchListView{
		Vouches:    summaries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// GetReputation Handler
// ============================================================================

// HandleGetReputation handles the GetReputation query.
func (h *Handlers) HandleGetReputation(ctx context.Context, q GetReputation) (*ReputationView, error) {
	const op = "Handlers.HandleGetReputation"

	proj, err := h.reputationLookup.GetByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapReputationProjectionToView(proj), nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps a VouchProjection to a VouchView.
func mapProjectionToView(proj *postgres.VouchProjection) *VouchView {
	return &VouchView{
		VouchID:          proj.ID,
		VoucherID:        proj.VoucherID,
		VoucheeID:        proj.VoucheeID,
		CredentialID:     proj.CredentialID,
		ClaimKey:         proj.ClaimKey,
		Relationship:     proj.Relationship,
		Strength:         proj.Strength,
		Statement:        proj.Statement,
		Context:          proj.Context,
		Status:           proj.Status,
		ExpiresAt:        proj.ExpiresAt,
		AcceptedAt:       proj.AcceptedAt,
		RevokedAt:        proj.RevokedAt,
		RevocationReason: proj.RevocationReason,
		CreatedAt:        proj.CreatedAt,
		UpdatedAt:        proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps a VouchProjection to a VouchSummaryView.
func mapProjectionToSummary(proj *postgres.VouchProjection) VouchSummaryView {
	return VouchSummaryView{
		VouchID:      proj.ID,
		VoucherID:    proj.VoucherID,
		VoucheeID:    proj.VoucheeID,
		Relationship: proj.Relationship,
		Strength:     proj.Strength,
		Status:       proj.Status,
		CreatedAt:    proj.CreatedAt,
		ExpiresAt:    proj.ExpiresAt,
	}
}

// mapReputationProjectionToView maps a ReputationProjection to a ReputationView.
func mapReputationProjectionToView(proj *postgres.ReputationProjection) *ReputationView {
	return &ReputationView{
		UserID:           proj.UserID,
		OverallScore:     proj.OverallScore,
		CredentialScore:  proj.CredentialScore,
		VouchScore:       proj.VouchScore,
		NetworkScore:     proj.NetworkScore,
		VouchCount:       proj.VouchCount,
		LastCalculatedAt: proj.LastCalculatedAt,
	}
}
