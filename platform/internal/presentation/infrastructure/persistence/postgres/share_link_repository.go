package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/presentation/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure ShareLinkRepository implements domain.ShareLinkRepository.
var _ domain.ShareLinkRepository = (*ShareLinkRepository)(nil)

// ShareLinkRepository is a PostgreSQL implementation of domain.ShareLinkRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type ShareLinkRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewShareLinkRepository creates a new PostgreSQL share link repository.
func NewShareLinkRepository(pool *pgxpool.Pool) *ShareLinkRepository {
	return &ShareLinkRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a share link aggregate by appending new events and updating the projection.
func (r *ShareLinkRepository) Save(ctx context.Context, sl *domain.ShareLink) error {
	if !sl.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("ShareLinkRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestShareLinkEventVersion(ctx, uuidFromShareLinkID(sl.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("ShareLinkRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := sl.Version() - len(sl.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"ShareLinkRepository.Save",
			sl.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range sl.Changes() {
		version := expectedVersion + i + 1
		params, err := ShareLinkEventToInsertParams(sl.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("ShareLinkRepository.Save", err)
		}

		if err := qtx.InsertShareLinkEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("ShareLinkRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := ShareLinkToUpsertParams(sl)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("ShareLinkRepository.Save", "share_link", err)
	}

	if err := qtx.UpsertShareLink(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("ShareLinkRepository.Save", "share_link", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("ShareLinkRepository.Save", err)
	}

	sl.ClearChanges()
	return nil
}

// Get retrieves a share link by ID, replaying events to rebuild state.
func (r *ShareLinkRepository) Get(ctx context.Context, id domain.ShareLinkID) (*domain.ShareLink, error) {
	rows, err := r.queries.GetShareLinkEvents(ctx, uuidFromShareLinkID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("ShareLinkRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("ShareLinkRepository.Get", domain.AggregateTypeShareLink, id.String())
	}

	events, err := ShareLinkEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("ShareLinkRepository.Get", id.String(), err)
	}

	sl := domain.NewShareLinkFromEvents(id.String())
	eventsourcing.Hydrate(sl, events)

	return sl, nil
}

// Exists checks if a share link with the given ID exists.
func (r *ShareLinkRepository) Exists(ctx context.Context, id domain.ShareLinkID) (bool, error) {
	exists, err := r.queries.ShareLinkAggregateExists(ctx, uuidFromShareLinkID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("ShareLinkRepository.Exists", id.String(), err)
	}
	return exists, nil
}
