package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Organization context.
type Handlers struct {
	lookup *postgres.OrganizationLookup
	logger log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.OrganizationLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup: lookup,
		logger: logger,
	}
}

// ============================================================================
// GetOrganization Handler
// ============================================================================

// HandleGetOrganization handles the GetOrganization query.
func (h *Handlers) HandleGetOrganization(ctx context.Context, q GetOrganization) (*OrganizationView, error) {
	const op = "Handlers.HandleGetOrganization"

	proj, err := h.lookup.GetByID(ctx, q.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// GetOrganizationBySlug Handler
// ============================================================================

// HandleGetOrganizationBySlug handles the GetOrganizationBySlug query.
func (h *Handlers) HandleGetOrganizationBySlug(ctx context.Context, q GetOrganizationBySlug) (*OrganizationView, error) {
	const op = "Handlers.HandleGetOrganizationBySlug"

	proj, err := h.lookup.GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListOrganizations Handler
// ============================================================================

// HandleListOrganizations handles the ListOrganizations query.
func (h *Handlers) HandleListOrganizations(ctx context.Context, q ListOrganizations) (*OrganizationListView, error) {
	const op = "Handlers.HandleListOrganizations"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	organizations, err := h.lookup.ListOrganizations(ctx, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountOrganizations(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]OrganizationSummaryView, len(organizations))
	for i, org := range organizations {
		summaries[i] = mapProjectionToSummary(&org)
	}

	return &OrganizationListView{
		Organizations: summaries,
		TotalCount:    totalCount,
		Limit:         limit,
		Offset:        q.Offset,
		HasMore:       q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps an OrganizationProjection to an OrganizationView.
func mapProjectionToView(proj *postgres.OrganizationProjection) *OrganizationView {
	return &OrganizationView{
		OrganizationID:     proj.ID,
		Name:               proj.Name,
		Slug:               proj.Slug,
		OrgType:            proj.OrgType,
		Description:        proj.Description,
		VerificationStatus: proj.VerificationStatus,
		DID:                proj.DID,
		OwnerMemberID:      proj.OwnerMemberID,
		MemberCount:        proj.MemberCount,
		CreatedAt:          proj.CreatedAt,
		UpdatedAt:          proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps an OrganizationProjection to an OrganizationSummaryView.
func mapProjectionToSummary(proj *postgres.OrganizationProjection) OrganizationSummaryView {
	return OrganizationSummaryView{
		OrganizationID:     proj.ID,
		Name:               proj.Name,
		Slug:               proj.Slug,
		OrgType:            proj.OrgType,
		VerificationStatus: proj.VerificationStatus,
		MemberCount:        proj.MemberCount,
		CreatedAt:          proj.CreatedAt,
	}
}
