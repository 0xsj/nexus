package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Wallet context.
type Handlers struct {
	lookup *postgres.WalletLookup
	logger log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.WalletLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup: lookup,
		logger: logger,
	}
}

// ============================================================================
// GetWallet Handler
// ============================================================================

// HandleGetWallet handles the GetWallet query.
func (h *Handlers) HandleGetWallet(ctx context.Context, q GetWallet) (*WalletView, error) {
	const op = "Handlers.HandleGetWallet"

	proj, err := h.lookup.GetByID(ctx, q.WalletID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListWalletsByUser Handler
// ============================================================================

// HandleListWalletsByUser handles the ListWalletsByUser query.
func (h *Handlers) HandleListWalletsByUser(ctx context.Context, q ListWalletsByUser) (*WalletListView, error) {
	const op = "Handlers.HandleListWalletsByUser"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	wallets, err := h.lookup.ListByUserID(ctx, q.UserID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]WalletSummaryView, len(wallets))
	for i, w := range wallets {
		summaries[i] = mapProjectionToSummary(&w)
	}

	return &WalletListView{
		Wallets:    summaries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// GetPrimaryWallet Handler
// ============================================================================

// HandleGetPrimaryWallet handles the GetPrimaryWallet query.
func (h *Handlers) HandleGetPrimaryWallet(ctx context.Context, q GetPrimaryWallet) (*WalletView, error) {
	const op = "Handlers.HandleGetPrimaryWallet"

	proj, err := h.lookup.GetPrimaryByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps a WalletProjection to a WalletView.
func mapProjectionToView(proj *postgres.WalletProjection) *WalletView {
	return &WalletView{
		WalletID:   proj.ID,
		UserID:     proj.UserID,
		Address:    proj.Address,
		Chain:      proj.Chain,
		Label:      proj.Label,
		Status:     proj.Status,
		IsPrimary:  proj.IsPrimary,
		DID:        proj.DID,
		LinkedAt:   proj.LinkedAt,
		VerifiedAt: proj.VerifiedAt,
		CreatedAt:  proj.CreatedAt,
		UpdatedAt:  proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps a WalletProjection to a WalletSummaryView.
func mapProjectionToSummary(proj *postgres.WalletProjection) WalletSummaryView {
	return WalletSummaryView{
		WalletID:  proj.ID,
		Address:   proj.Address,
		Chain:     proj.Chain,
		Label:     proj.Label,
		Status:    proj.Status,
		IsPrimary: proj.IsPrimary,
	}
}
