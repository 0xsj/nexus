package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
)

// PresentationLookup provides read-side queries on the presentations projection table.
type PresentationLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPresentationLookup creates a new PresentationLookup.
func NewPresentationLookup(pool *pgxpool.Pool) *PresentationLookup {
	return &PresentationLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a presentation projection by ID.
func (l *PresentationLookup) GetByID(ctx context.Context, id string) (*PresentationProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetPresentationByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := PresentationRowToProjection(row)
	return &proj, nil
}

// ListByHolderDID returns presentations for a holder DID with pagination.
func (l *PresentationLookup) ListByHolderDID(ctx context.Context, holderDID string, limit, offset int) ([]PresentationProjection, error) {
	rows, err := l.queries.ListPresentationsByHolderDID(ctx, generated.ListPresentationsByHolderDIDParams{
		HolderDid: holderDID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]PresentationProjection, len(rows))
	for i, row := range rows {
		projections[i] = PresentationRowToProjection(row)
	}
	return projections, nil
}

// CountByHolderDID returns the count of presentations for a holder DID.
func (l *PresentationLookup) CountByHolderDID(ctx context.Context, holderDID string) (int, error) {
	count, err := l.queries.CountPresentationsByHolderDID(ctx, holderDID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
