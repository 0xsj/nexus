package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
)

// ReputationLookup provides read-side queries on the reputations projection table.
type ReputationLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewReputationLookup creates a new ReputationLookup.
func NewReputationLookup(pool *pgxpool.Pool) *ReputationLookup {
	return &ReputationLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByUserID returns a reputation projection by user ID.
func (l *ReputationLookup) GetByUserID(ctx context.Context, userID string) (*ReputationProjection, error) {
	row, err := l.queries.GetReputationByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	proj := ReputationRowToProjection(row)
	return &proj, nil
}
