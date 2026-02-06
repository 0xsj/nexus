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

// Ensure PresentationRepository implements domain.PresentationRepository.
var _ domain.PresentationRepository = (*PresentationRepository)(nil)

// PresentationRepository is a PostgreSQL implementation of domain.PresentationRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type PresentationRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPresentationRepository creates a new PostgreSQL presentation repository.
func NewPresentationRepository(pool *pgxpool.Pool) *PresentationRepository {
	return &PresentationRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a presentation aggregate by appending new events and updating the projection.
func (r *PresentationRepository) Save(ctx context.Context, pres *domain.Presentation) error {
	if !pres.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("PresentationRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestPresentationEventVersion(ctx, uuidFromPresentationID(pres.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("PresentationRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := pres.Version() - len(pres.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"PresentationRepository.Save",
			pres.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range pres.Changes() {
		version := expectedVersion + i + 1
		params, err := PresentationEventToInsertParams(pres.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("PresentationRepository.Save", err)
		}

		if err := qtx.InsertPresentationEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("PresentationRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := PresentationToUpsertParams(pres)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("PresentationRepository.Save", "presentation", err)
	}

	if err := qtx.UpsertPresentation(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("PresentationRepository.Save", "presentation", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("PresentationRepository.Save", err)
	}

	pres.ClearChanges()
	return nil
}

// Get retrieves a presentation by ID, replaying events to rebuild state.
func (r *PresentationRepository) Get(ctx context.Context, id domain.PresentationID) (*domain.Presentation, error) {
	rows, err := r.queries.GetPresentationEvents(ctx, uuidFromPresentationID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("PresentationRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("PresentationRepository.Get", domain.AggregateTypePresentation, id.String())
	}

	events, err := PresentationEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("PresentationRepository.Get", id.String(), err)
	}

	pres := domain.NewPresentationFromEvents(id.String())
	eventsourcing.Hydrate(pres, events)

	return pres, nil
}

// Exists checks if a presentation with the given ID exists.
func (r *PresentationRepository) Exists(ctx context.Context, id domain.PresentationID) (bool, error) {
	exists, err := r.queries.PresentationAggregateExists(ctx, uuidFromPresentationID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("PresentationRepository.Exists", id.String(), err)
	}
	return exists, nil
}
