package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres/generated"
)

// OrganizationLookup provides read-side queries on the organizations projection table.
type OrganizationLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewOrganizationLookup creates a new OrganizationLookup.
func NewOrganizationLookup(pool *pgxpool.Pool) *OrganizationLookup {
	return &OrganizationLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns an organization projection by ID.
func (l *OrganizationLookup) GetByID(ctx context.Context, id string) (*OrganizationProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetOrganizationByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := OrganizationRowToProjection(row)
	return &proj, nil
}

// ExistsByID checks if an organization exists by ID.
func (l *OrganizationLookup) ExistsByID(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	return l.queries.OrganizationExistsByID(ctx, uid)
}

// GetBySlug returns an organization projection by slug.
func (l *OrganizationLookup) GetBySlug(ctx context.Context, slug string) (*OrganizationProjection, error) {
	row, err := l.queries.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	proj := OrganizationRowToProjection(row)
	return &proj, nil
}

// ExistsBySlug checks if an organization exists by slug.
func (l *OrganizationLookup) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return l.queries.OrganizationExistsBySlug(ctx, slug)
}

// ListOrganizations returns organizations with pagination.
func (l *OrganizationLookup) ListOrganizations(ctx context.Context, limit, offset int) ([]OrganizationProjection, error) {
	rows, err := l.queries.ListOrganizations(ctx, generated.ListOrganizationsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]OrganizationProjection, len(rows))
	for i, row := range rows {
		projections[i] = OrganizationRowToProjection(row)
	}
	return projections, nil
}

// CountOrganizations returns the total count of non-deleted organizations.
func (l *OrganizationLookup) CountOrganizations(ctx context.Context) (int, error) {
	count, err := l.queries.CountOrganizations(ctx)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// SlugExists checks if the given slug is already taken.
// Implements domain.SlugLookup interface.
func (l *OrganizationLookup) SlugExists(ctx context.Context, slug string) (bool, error) {
	return l.queries.OrganizationExistsBySlug(ctx, slug)
}
