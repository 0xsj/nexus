package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
)

// VouchLookup provides read-side queries on the vouches projection table.
type VouchLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewVouchLookup creates a new VouchLookup.
func NewVouchLookup(pool *pgxpool.Pool) *VouchLookup {
	return &VouchLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a vouch projection by ID.
func (l *VouchLookup) GetByID(ctx context.Context, id string) (*VouchProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetVouchByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := VouchRowToProjection(row)
	return &proj, nil
}

// ListByVoucheeID returns vouches received by a user with pagination.
func (l *VouchLookup) ListByVoucheeID(ctx context.Context, voucheeID string, limit, offset int) ([]VouchProjection, error) {
	rows, err := l.queries.ListVouchesByVoucheeID(ctx, generated.ListVouchesByVoucheeIDParams{
		VoucheeID: voucheeID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]VouchProjection, len(rows))
	for i, row := range rows {
		projections[i] = VouchRowToProjection(row)
	}
	return projections, nil
}

// CountByVoucheeID returns the count of vouches received by a user.
func (l *VouchLookup) CountByVoucheeID(ctx context.Context, voucheeID string) (int, error) {
	count, err := l.queries.CountVouchesByVoucheeID(ctx, voucheeID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ListByVoucherID returns vouches given by a user with pagination.
func (l *VouchLookup) ListByVoucherID(ctx context.Context, voucherID string, limit, offset int) ([]VouchProjection, error) {
	rows, err := l.queries.ListVouchesByVoucherID(ctx, generated.ListVouchesByVoucherIDParams{
		VoucherID: voucherID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]VouchProjection, len(rows))
	for i, row := range rows {
		projections[i] = VouchRowToProjection(row)
	}
	return projections, nil
}

// CountByVoucherID returns the count of vouches given by a user.
func (l *VouchLookup) CountByVoucherID(ctx context.Context, voucherID string) (int, error) {
	count, err := l.queries.CountVouchesByVoucherID(ctx, voucherID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
