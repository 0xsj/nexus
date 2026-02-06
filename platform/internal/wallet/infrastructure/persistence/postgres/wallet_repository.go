package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure WalletRepository implements domain.WalletRepository.
var _ domain.WalletRepository = (*WalletRepository)(nil)

// WalletRepository is a PostgreSQL implementation of domain.WalletRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type WalletRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewWalletRepository creates a new PostgreSQL wallet repository.
func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a wallet aggregate by appending new events and updating the projection.
func (r *WalletRepository) Save(ctx context.Context, w *domain.Wallet) error {
	if !w.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("WalletRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestWalletEventVersion(ctx, uuidFromWalletID(w.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("WalletRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := w.Version() - len(w.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"WalletRepository.Save",
			w.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range w.Changes() {
		version := expectedVersion + i + 1
		params, err := WalletEventToInsertParams(w.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("WalletRepository.Save", err)
		}

		if err := qtx.InsertWalletEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("WalletRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := WalletToUpsertParams(w)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("WalletRepository.Save", "wallet", err)
	}

	if err := qtx.UpsertWallet(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("WalletRepository.Save", "wallet", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("WalletRepository.Save", err)
	}

	w.ClearChanges()
	return nil
}

// Get retrieves a wallet by ID, replaying events to rebuild state.
func (r *WalletRepository) Get(ctx context.Context, id domain.WalletID) (*domain.Wallet, error) {
	rows, err := r.queries.GetWalletEvents(ctx, uuidFromWalletID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("WalletRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("WalletRepository.Get", domain.AggregateTypeWallet, id.String())
	}

	events, err := WalletEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("WalletRepository.Get", id.String(), err)
	}

	w := domain.NewWalletFromEvents(id.String())
	eventsourcing.Hydrate(w, events)

	return w, nil
}

// Exists checks if a wallet with the given ID exists.
func (r *WalletRepository) Exists(ctx context.Context, id domain.WalletID) (bool, error) {
	exists, err := r.queries.WalletAggregateExists(ctx, uuidFromWalletID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("WalletRepository.Exists", id.String(), err)
	}
	return exists, nil
}
