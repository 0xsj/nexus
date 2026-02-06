package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Profile context.
type Handlers struct {
	lookup *postgres.ProfileLookup
	logger log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.ProfileLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup: lookup,
		logger: logger,
	}
}

// ============================================================================
// GetProfile Handler
// ============================================================================

// HandleGetProfile handles the GetProfile query.
func (h *Handlers) HandleGetProfile(ctx context.Context, q GetProfile) (*ProfileView, error) {
	const op = "Handlers.HandleGetProfile"

	proj, err := h.lookup.GetByID(ctx, q.ProfileID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// GetProfileByUser Handler
// ============================================================================

// HandleGetProfileByUser handles the GetProfileByUser query.
func (h *Handlers) HandleGetProfileByUser(ctx context.Context, q GetProfileByUser) (*ProfileView, error) {
	const op = "Handlers.HandleGetProfileByUser"

	proj, err := h.lookup.GetByUserID(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// GetProfileByVanitySlug Handler
// ============================================================================

// HandleGetProfileByVanitySlug handles the GetProfileByVanitySlug query.
func (h *Handlers) HandleGetProfileByVanitySlug(ctx context.Context, q GetProfileByVanitySlug) (*ProfileView, error) {
	const op = "Handlers.HandleGetProfileByVanitySlug"

	proj, err := h.lookup.GetByVanitySlug(ctx, q.VanitySlug)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListProfiles Handler
// ============================================================================

// HandleListProfiles handles the ListProfiles query.
func (h *Handlers) HandleListProfiles(ctx context.Context, q ListProfiles) (*ProfileListView, error) {
	const op = "Handlers.HandleListProfiles"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	profiles, err := h.lookup.ListProfiles(ctx, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountProfiles(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]ProfileSummaryView, len(profiles))
	for i, prof := range profiles {
		summaries[i] = mapProjectionToSummary(&prof)
	}

	return &ProfileListView{
		Profiles:   summaries,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     q.Offset,
		HasMore:    q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps a ProfileProjection to a ProfileView.
func mapProjectionToView(proj *postgres.ProfileProjection) *ProfileView {
	return &ProfileView{
		ProfileID:   proj.ID,
		UserID:      proj.UserID,
		DisplayName: proj.DisplayName,
		Headline:    proj.Headline,
		Bio:         proj.Bio,
		VanitySlug:  proj.VanitySlug,
		Badges:      []BadgeView{}, // Badges are loaded separately or via event replay
		BadgeCount:  proj.BadgeCount,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps a ProfileProjection to a ProfileSummaryView.
func mapProjectionToSummary(proj *postgres.ProfileProjection) ProfileSummaryView {
	return ProfileSummaryView{
		ProfileID:   proj.ID,
		UserID:      proj.UserID,
		DisplayName: proj.DisplayName,
		Headline:    proj.Headline,
		VanitySlug:  proj.VanitySlug,
		BadgeCount:  proj.BadgeCount,
		CreatedAt:   proj.CreatedAt,
	}
}
