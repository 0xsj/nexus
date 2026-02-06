package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure VouchRepository implements domain.VouchRepository.
var _ domain.VouchRepository = (*VouchRepository)(nil)

// VouchRepository is a PostgreSQL implementation of domain.VouchRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type VouchRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewVouchRepository creates a new PostgreSQL vouch repository.
func NewVouchRepository(pool *pgxpool.Pool) *VouchRepository {
	return &VouchRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a vouch aggregate by appending new events and updating the projection.
func (r *VouchRepository) Save(ctx context.Context, vouch *domain.Vouch) error {
	if !vouch.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("VouchRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestVouchEventVersion(ctx, uuidFromVouchID(vouch.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("VouchRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := vouch.Version() - len(vouch.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"VouchRepository.Save",
			vouch.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range vouch.Changes() {
		version := expectedVersion + i + 1
		params, err := VouchEventToInsertParams(vouch.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("VouchRepository.Save", err)
		}

		if err := qtx.InsertVouchEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("VouchRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := VouchToUpsertParams(vouch)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("VouchRepository.Save", "vouch", err)
	}

	if err := qtx.UpsertVouch(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("VouchRepository.Save", "vouch", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("VouchRepository.Save", err)
	}

	vouch.ClearChanges()
	return nil
}

// Get retrieves a vouch by ID, replaying events to rebuild state.
func (r *VouchRepository) Get(ctx context.Context, id domain.VouchID) (*domain.Vouch, error) {
	rows, err := r.queries.GetVouchEvents(ctx, uuidFromVouchID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("VouchRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("VouchRepository.Get", domain.AggregateTypeVouch, id.String())
	}

	events, err := VouchEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("VouchRepository.Get", id.String(), err)
	}

	vouch := domain.NewVouchFromEvents(id.String())
	eventsourcing.Hydrate(vouch, events)

	return vouch, nil
}

// Exists checks if a vouch with the given ID exists.
func (r *VouchRepository) Exists(ctx context.Context, id domain.VouchID) (bool, error) {
	exists, err := r.queries.VouchAggregateExists(ctx, uuidFromVouchID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("VouchRepository.Exists", id.String(), err)
	}
	return exists, nil
}
