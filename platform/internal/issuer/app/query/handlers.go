package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Issuer context.
type Handlers struct {
	issuerLookup   *postgres.IssuerLookup
	templateLookup *postgres.TemplateLookup
	logger         log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	issuerLookup *postgres.IssuerLookup,
	templateLookup *postgres.TemplateLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		issuerLookup:   issuerLookup,
		templateLookup: templateLookup,
		logger:         logger,
	}
}

// ============================================================================
// GetIssuer Handler
// ============================================================================

// HandleGetIssuer handles the GetIssuer query.
func (h *Handlers) HandleGetIssuer(ctx context.Context, q GetIssuer) (*IssuerView, error) {
	const op = "Handlers.HandleGetIssuer"

	proj, err := h.issuerLookup.GetByID(ctx, q.IssuerID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapIssuerProjectionToView(proj), nil
}

// ============================================================================
// GetIssuerByOrganization Handler
// ============================================================================

// HandleGetIssuerByOrganization handles the GetIssuerByOrganization query.
func (h *Handlers) HandleGetIssuerByOrganization(ctx context.Context, q GetIssuerByOrganization) (*IssuerView, error) {
	const op = "Handlers.HandleGetIssuerByOrganization"

	proj, err := h.issuerLookup.GetByOrganizationID(ctx, q.OrganizationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapIssuerProjectionToView(proj), nil
}

// ============================================================================
// ListIssuers Handler
// ============================================================================

// HandleListIssuers handles the ListIssuers query.
func (h *Handlers) HandleListIssuers(ctx context.Context, q ListIssuers) (*IssuerListView, error) {
	const op = "Handlers.HandleListIssuers"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	issuers, err := h.issuerLookup.ListIssuers(ctx, q.Status, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.issuerLookup.CountIssuers(ctx, q.Status)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]IssuerSummaryView, len(issuers))
	for i, iss := range issuers {
		summaries[i] = mapIssuerProjectionToSummary(&iss)
	}

	return &IssuerListView{
		Issuers:    summaries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// GetTemplate Handler
// ============================================================================

// HandleGetTemplate handles the GetTemplate query.
func (h *Handlers) HandleGetTemplate(ctx context.Context, q GetTemplate) (*TemplateView, error) {
	const op = "Handlers.HandleGetTemplate"

	proj, err := h.templateLookup.GetByID(ctx, q.TemplateID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapTemplateProjectionToView(proj), nil
}

// ============================================================================
// ListTemplatesByIssuer Handler
// ============================================================================

// HandleListTemplatesByIssuer handles the ListTemplatesByIssuer query.
func (h *Handlers) HandleListTemplatesByIssuer(ctx context.Context, q ListTemplatesByIssuer) (*TemplateListView, error) {
	const op = "Handlers.HandleListTemplatesByIssuer"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	templates, err := h.templateLookup.ListByIssuerID(ctx, q.IssuerID.String(), limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.templateLookup.CountByIssuerID(ctx, q.IssuerID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]TemplateSummaryView, len(templates))
	for i, t := range templates {
		summaries[i] = mapTemplateProjectionToSummary(&t)
	}

	return &TemplateListView{
		Templates:  summaries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapIssuerProjectionToView maps an IssuerProjection to an IssuerView.
func mapIssuerProjectionToView(proj *postgres.IssuerProjection) *IssuerView {
	return &IssuerView{
		IssuerID:       proj.ID,
		OrganizationID: proj.OrganizationID,
		Name:           proj.Name,
		Description:    proj.Description,
		DID:            proj.DID,
		WebhookURL:     proj.WebhookURL,
		Status:         proj.Status,
		Branding:       proj.Branding,
		CreatedAt:      proj.CreatedAt,
		UpdatedAt:      proj.UpdatedAt,
	}
}

// mapIssuerProjectionToSummary maps an IssuerProjection to an IssuerSummaryView.
func mapIssuerProjectionToSummary(proj *postgres.IssuerProjection) IssuerSummaryView {
	return IssuerSummaryView{
		IssuerID:       proj.ID,
		OrganizationID: proj.OrganizationID,
		Name:           proj.Name,
		Status:         proj.Status,
		CreatedAt:      proj.CreatedAt,
	}
}

// mapTemplateProjectionToView maps a TemplateProjection to a TemplateView.
func mapTemplateProjectionToView(proj *postgres.TemplateProjection) *TemplateView {
	return &TemplateView{
		TemplateID:     proj.ID,
		IssuerID:       proj.IssuerID,
		Name:           proj.Name,
		Description:    proj.Description,
		SchemaType:     proj.SchemaType,
		ClaimMappings:  proj.ClaimMappings,
		DefaultValues:  proj.DefaultValues,
		ExpirationDays: proj.ExpirationDays,
		AutoApprove:    proj.AutoApprove,
		Status:         proj.Status,
		Version:        proj.Version,
		CreatedAt:      proj.CreatedAt,
		UpdatedAt:      proj.UpdatedAt,
	}
}

// mapTemplateProjectionToSummary maps a TemplateProjection to a TemplateSummaryView.
func mapTemplateProjectionToSummary(proj *postgres.TemplateProjection) TemplateSummaryView {
	return TemplateSummaryView{
		TemplateID: proj.ID,
		IssuerID:   proj.IssuerID,
		Name:       proj.Name,
		SchemaType: proj.SchemaType,
		Status:     proj.Status,
		Version:    proj.Version,
		CreatedAt:  proj.CreatedAt,
	}
}
