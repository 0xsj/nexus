package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/persistence/postgres/generated"
)

// VerificationLookup provides read-side queries on the verifications projection table.
type VerificationLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewVerificationLookup creates a new VerificationLookup.
func NewVerificationLookup(pool *pgxpool.Pool) *VerificationLookup {
	return &VerificationLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a verification projection by ID.
func (l *VerificationLookup) GetByID(ctx context.Context, id string) (*VerificationProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetVerificationByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := VerificationRowToProjection(row)
	return &proj, nil
}

// ExistsByID checks if a verification exists by ID.
func (l *VerificationLookup) ExistsByID(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	return l.queries.VerificationExistsByID(ctx, uid)
}

// ListByUserID returns verifications for a user with pagination.
func (l *VerificationLookup) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]VerificationProjection, error) {
	rows, err := l.queries.ListVerificationsByUserID(ctx, generated.ListVerificationsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]VerificationProjection, len(rows))
	for i, row := range rows {
		projections[i] = VerificationRowToProjection(row)
	}
	return projections, nil
}

// CountByUserID returns the count of verifications for a user.
func (l *VerificationLookup) CountByUserID(ctx context.Context, userID string) (int, error) {
	count, err := l.queries.CountVerificationsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetByUserAndProvider returns the most recent verification for a user and provider type.
func (l *VerificationLookup) GetByUserAndProvider(ctx context.Context, userID string, providerType string) (*VerificationProjection, error) {
	row, err := l.queries.GetVerificationByUserAndProvider(ctx, generated.GetVerificationByUserAndProviderParams{
		UserID:       userID,
		ProviderType: providerType,
	})
	if err != nil {
		return nil, err
	}

	proj := VerificationRowToProjection(row)
	return &proj, nil
}
