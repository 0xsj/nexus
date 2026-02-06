package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/persistence/postgres/generated"
)

// WalletLookup provides read-side queries on the wallets projection table.
type WalletLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewWalletLookup creates a new WalletLookup.
func NewWalletLookup(pool *pgxpool.Pool) *WalletLookup {
	return &WalletLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a wallet projection by ID.
func (l *WalletLookup) GetByID(ctx context.Context, id string) (*WalletProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetWalletByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := WalletRowToProjection(row)
	return &proj, nil
}

// ExistsByID checks if a wallet exists by ID.
func (l *WalletLookup) ExistsByID(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	return l.queries.WalletExistsByID(ctx, uid)
}

// GetByAddress returns a wallet projection by address.
func (l *WalletLookup) GetByAddress(ctx context.Context, address string) (*WalletProjection, error) {
	row, err := l.queries.GetWalletByAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	proj := WalletRowToProjection(row)
	return &proj, nil
}

// ExistsByAddress checks if a wallet exists by address.
func (l *WalletLookup) ExistsByAddress(ctx context.Context, address string) (bool, error) {
	return l.queries.WalletExistsByAddress(ctx, address)
}

// ListByUserID returns wallets for a user ID with pagination.
func (l *WalletLookup) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]WalletProjection, error) {
	rows, err := l.queries.ListWalletsByUserID(ctx, generated.ListWalletsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]WalletProjection, len(rows))
	for i, row := range rows {
		projections[i] = WalletRowToProjection(row)
	}
	return projections, nil
}

// CountByUserID returns the count of wallets for a user ID.
func (l *WalletLookup) CountByUserID(ctx context.Context, userID string) (int, error) {
	count, err := l.queries.CountWalletsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetPrimaryByUserID returns the primary wallet for a user ID.
func (l *WalletLookup) GetPrimaryByUserID(ctx context.Context, userID string) (*WalletProjection, error) {
	row, err := l.queries.GetPrimaryWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	proj := WalletRowToProjection(row)
	return &proj, nil
}
