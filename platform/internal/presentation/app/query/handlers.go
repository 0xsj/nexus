package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Presentation context.
type Handlers struct {
	presentationLookup *postgres.PresentationLookup
	shareLinkLookup    *postgres.ShareLinkLookup
	logger             log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	presentationLookup *postgres.PresentationLookup,
	shareLinkLookup *postgres.ShareLinkLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		presentationLookup: presentationLookup,
		shareLinkLookup:    shareLinkLookup,
		logger:             logger,
	}
}

// ============================================================================
// GetPresentation Handler
// ============================================================================

// HandleGetPresentation handles the GetPresentation query.
func (h *Handlers) HandleGetPresentation(ctx context.Context, q GetPresentation) (*PresentationView, error) {
	const op = "Handlers.HandleGetPresentation"

	proj, err := h.presentationLookup.GetByID(ctx, q.PresentationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapPresentationProjectionToView(proj), nil
}

// ============================================================================
// ListPresentationsByHolder Handler
// ============================================================================

// HandleListPresentationsByHolder handles the ListPresentationsByHolder query.
func (h *Handlers) HandleListPresentationsByHolder(ctx context.Context, q ListPresentationsByHolder) (*PresentationListView, error) {
	const op = "Handlers.HandleListPresentationsByHolder"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	presentations, err := h.presentationLookup.ListByHolderDID(ctx, q.HolderDID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.presentationLookup.CountByHolderDID(ctx, q.HolderDID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]PresentationSummaryView, len(presentations))
	for i, pres := range presentations {
		summaries[i] = mapPresentationProjectionToSummary(&pres)
	}

	return &PresentationListView{
		Presentations: summaries,
		TotalCount:    totalCount,
		Limit:         limit,
		Offset:        q.Offset,
		HasMore:       q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// GetShareLink Handler
// ============================================================================

// HandleGetShareLink handles the GetShareLink query.
func (h *Handlers) HandleGetShareLink(ctx context.Context, q GetShareLink) (*ShareLinkView, error) {
	const op = "Handlers.HandleGetShareLink"

	proj, err := h.shareLinkLookup.GetByID(ctx, q.ShareLinkID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapShareLinkProjectionToView(proj), nil
}

// ============================================================================
// GetShareLinkByToken Handler
// ============================================================================

// HandleGetShareLinkByToken handles the GetShareLinkByToken query.
func (h *Handlers) HandleGetShareLinkByToken(ctx context.Context, q GetShareLinkByToken) (*ShareLinkView, error) {
	const op = "Handlers.HandleGetShareLinkByToken"

	proj, err := h.shareLinkLookup.GetByToken(ctx, q.Token)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapShareLinkProjectionToView(proj), nil
}

// ============================================================================
// ListShareLinks Handler
// ============================================================================

// HandleListShareLinks handles the ListShareLinks query.
func (h *Handlers) HandleListShareLinks(ctx context.Context, q ListShareLinks) (*ShareLinkListView, error) {
	const op = "Handlers.HandleListShareLinks"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	shareLinks, err := h.shareLinkLookup.ListByPresentationID(ctx, q.PresentationID.String(), limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.shareLinkLookup.CountByPresentationID(ctx, q.PresentationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	views := make([]ShareLinkView, len(shareLinks))
	for i, sl := range shareLinks {
		views[i] = mapShareLinkProjectionToViewValue(&sl)
	}

	return &ShareLinkListView{
		ShareLinks: views,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// GetAccessLog Handler
// ============================================================================

// HandleGetAccessLog handles the GetAccessLog query.
func (h *Handlers) HandleGetAccessLog(ctx context.Context, q GetAccessLog) (*AccessLogView, error) {
	const op = "Handlers.HandleGetAccessLog"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	grants, err := h.shareLinkLookup.ListAccessGrants(ctx, q.ShareLinkID.String(), limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.shareLinkLookup.CountAccessGrants(ctx, q.ShareLinkID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	views := make([]AccessGrantView, len(grants))
	for i, grant := range grants {
		views[i] = mapAccessGrantProjectionToView(&grant)
	}

	return &AccessLogView{
		AccessGrants: views,
		TotalCount:   totalCount,
		Limit:        limit,
		Offset:       q.Offset,
		HasMore:      q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapPresentationProjectionToView maps a PresentationProjection to a PresentationView.
func mapPresentationProjectionToView(proj *postgres.PresentationProjection) *PresentationView {
	return &PresentationView{
		PresentationID:   proj.ID,
		HolderDID:        proj.HolderDID,
		CredentialIDs:    proj.CredentialIDs,
		DisclosurePolicy: proj.DisclosurePolicy,
		VPJWT:            proj.VPJWT,
		Purpose:          proj.Purpose,
		Status:           proj.Status,
		RevokedAt:        proj.RevokedAt,
		RevocationReason: proj.RevocationReason,
		CreatedAt:        proj.CreatedAt,
		UpdatedAt:        proj.UpdatedAt,
	}
}

// mapPresentationProjectionToSummary maps a PresentationProjection to a PresentationSummaryView.
func mapPresentationProjectionToSummary(proj *postgres.PresentationProjection) PresentationSummaryView {
	return PresentationSummaryView{
		PresentationID: proj.ID,
		HolderDID:      proj.HolderDID,
		Purpose:        proj.Purpose,
		Status:         proj.Status,
		CreatedAt:      proj.CreatedAt,
	}
}

// mapShareLinkProjectionToView maps a ShareLinkProjection to a ShareLinkView pointer.
func mapShareLinkProjectionToView(proj *postgres.ShareLinkProjection) *ShareLinkView {
	v := mapShareLinkProjectionToViewValue(proj)
	return &v
}

// mapShareLinkProjectionToViewValue maps a ShareLinkProjection to a ShareLinkView value.
func mapShareLinkProjectionToViewValue(proj *postgres.ShareLinkProjection) ShareLinkView {
	return ShareLinkView{
		ShareLinkID:    proj.ID,
		PresentationID: proj.PresentationID,
		Token:          proj.Token,
		ExpiresAt:      proj.ExpiresAt,
		MaxViews:       proj.MaxViews,
		CurrentViews:   proj.CurrentViews,
		PinProtected:   proj.PinHash != "",
		Audience:       proj.Audience,
		Status:         proj.Status,
		CreatedAt:      proj.CreatedAt,
		UpdatedAt:      proj.UpdatedAt,
	}
}

// mapAccessGrantProjectionToView maps an AccessGrantProjection to an AccessGrantView.
func mapAccessGrantProjectionToView(proj *postgres.AccessGrantProjection) AccessGrantView {
	return AccessGrantView{
		ID:              proj.ID,
		ShareLinkID:     proj.ShareLinkID,
		VerifierDID:     proj.VerifierDID,
		AccessedAt:      proj.AccessedAt,
		IPAddress:       proj.IPAddress,
		DisclosedClaims: proj.DisclosedClaims,
	}
}
