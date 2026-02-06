package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
)

// ProfileLookup provides read-side queries on the profiles projection table.
type ProfileLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewProfileLookup creates a new ProfileLookup.
func NewProfileLookup(pool *pgxpool.Pool) *ProfileLookup {
	return &ProfileLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a profile projection by ID.
func (l *ProfileLookup) GetByID(ctx context.Context, id string) (*ProfileProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetProfileByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := ProfileRowToProjection(row)
	return &proj, nil
}

// ExistsByID checks if a profile exists by ID.
func (l *ProfileLookup) ExistsByID(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	return l.queries.ProfileExistsByID(ctx, uid)
}

// GetByUserID returns a profile projection by user ID.
func (l *ProfileLookup) GetByUserID(ctx context.Context, userID string) (*ProfileProjection, error) {
	row, err := l.queries.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	proj := ProfileRowToProjection(row)
	return &proj, nil
}

// GetByVanitySlug returns a profile projection by vanity slug.
func (l *ProfileLookup) GetByVanitySlug(ctx context.Context, slug string) (*ProfileProjection, error) {
	row, err := l.queries.GetProfileByVanitySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	proj := ProfileRowToProjection(row)
	return &proj, nil
}

// ExistsByVanitySlug checks if a profile exists with the given vanity slug.
func (l *ProfileLookup) ExistsByVanitySlug(ctx context.Context, slug string) (bool, error) {
	return l.queries.ProfileExistsByVanitySlug(ctx, slug)
}

// ListProfiles returns profiles with pagination.
func (l *ProfileLookup) ListProfiles(ctx context.Context, limit, offset int) ([]ProfileProjection, error) {
	rows, err := l.queries.ListProfiles(ctx, generated.ListProfilesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]ProfileProjection, len(rows))
	for i, row := range rows {
		projections[i] = ProfileRowToProjection(row)
	}
	return projections, nil
}

// CountProfiles returns the total count of profiles.
func (l *ProfileLookup) CountProfiles(ctx context.Context) (int, error) {
	count, err := l.queries.CountProfiles(ctx)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
