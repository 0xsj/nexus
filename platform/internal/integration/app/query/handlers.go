package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Integration context.
type Handlers struct {
	lookup *postgres.IntegrationLookup
	logger log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.IntegrationLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup: lookup,
		logger: logger,
	}
}

// ============================================================================
// GetIntegration Handler
// ============================================================================

// HandleGetIntegration handles the GetIntegration query.
func (h *Handlers) HandleGetIntegration(ctx context.Context, q GetIntegration) (*IntegrationView, error) {
	const op = "Handlers.HandleGetIntegration"

	proj, err := h.lookup.GetByID(ctx, q.IntegrationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListIntegrationsByUser Handler
// ============================================================================

// HandleListIntegrationsByUser handles the ListIntegrationsByUser query.
func (h *Handlers) HandleListIntegrationsByUser(ctx context.Context, q ListIntegrationsByUser) (*IntegrationListView, error) {
	const op = "Handlers.HandleListIntegrationsByUser"

	integrations, err := h.lookup.ListByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]IntegrationSummaryView, len(integrations))
	for i, integ := range integrations {
		summaries[i] = mapProjectionToSummary(&integ)
	}

	return &IntegrationListView{
		Integrations: summaries,
		TotalCount:   totalCount,
	}, nil
}

// ============================================================================
// GetIntegrationByUserAndProvider Handler
// ============================================================================

// HandleGetIntegrationByUserAndProvider handles the GetIntegrationByUserAndProvider query.
func (h *Handlers) HandleGetIntegrationByUserAndProvider(ctx context.Context, q GetIntegrationByUserAndProvider) (*IntegrationView, error) {
	const op = "Handlers.HandleGetIntegrationByUserAndProvider"

	proj, err := h.lookup.GetByUserAndProvider(ctx, q.UserID, q.ProviderType)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps an IntegrationProjection to an IntegrationView.
func mapProjectionToView(proj *postgres.IntegrationProjection) *IntegrationView {
	return &IntegrationView{
		IntegrationID:    proj.ID,
		UserID:           proj.UserID,
		ProviderType:     proj.ProviderType,
		Status:           proj.Status,
		ProviderUserID:   proj.ProviderUserID,
		ProviderUsername: proj.ProviderUsername,
		Scopes:           proj.Scopes,
		LastFetchAt:      proj.LastFetchAt,
		FetchCount:       proj.FetchCount,
		Metadata:         proj.Metadata,
		ConnectedAt:      proj.ConnectedAt,
		DisconnectedAt:   proj.DisconnectedAt,
		SuspendedAt:      proj.SuspendedAt,
		SuspensionReason: proj.SuspensionReason,
		CreatedAt:        proj.CreatedAt,
		UpdatedAt:        proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps an IntegrationProjection to an IntegrationSummaryView.
func mapProjectionToSummary(proj *postgres.IntegrationProjection) IntegrationSummaryView {
	return IntegrationSummaryView{
		IntegrationID:    proj.ID,
		UserID:           proj.UserID,
		ProviderType:     proj.ProviderType,
		Status:           proj.Status,
		ProviderUsername: proj.ProviderUsername,
		ConnectedAt:      proj.ConnectedAt,
	}
}
