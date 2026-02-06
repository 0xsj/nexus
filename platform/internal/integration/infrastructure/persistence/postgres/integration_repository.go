package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure IntegrationRepository implements domain.IntegrationRepository.
var _ domain.IntegrationRepository = (*IntegrationRepository)(nil)

// IntegrationRepository is a PostgreSQL implementation of domain.IntegrationRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type IntegrationRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewIntegrationRepository creates a new PostgreSQL integration repository.
func NewIntegrationRepository(pool *pgxpool.Pool) *IntegrationRepository {
	return &IntegrationRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists an integration aggregate by appending new events and updating the projection.
func (r *IntegrationRepository) Save(ctx context.Context, integ *domain.Integration) error {
	if !integ.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("IntegrationRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestIntegrationEventVersion(ctx, uuidFromIntegrationID(integ.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("IntegrationRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := integ.Version() - len(integ.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"IntegrationRepository.Save",
			integ.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range integ.Changes() {
		version := expectedVersion + i + 1
		params, err := IntegrationEventToInsertParams(integ.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("IntegrationRepository.Save", err)
		}

		if err := qtx.InsertIntegrationEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("IntegrationRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := IntegrationToUpsertParams(integ)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("IntegrationRepository.Save", "integration", err)
	}

	if err := qtx.UpsertIntegration(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("IntegrationRepository.Save", "integration", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("IntegrationRepository.Save", err)
	}

	integ.ClearChanges()
	return nil
}

// Get retrieves an integration by ID, replaying events to rebuild state.
func (r *IntegrationRepository) Get(ctx context.Context, id domain.IntegrationID) (*domain.Integration, error) {
	rows, err := r.queries.GetIntegrationEvents(ctx, uuidFromIntegrationID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("IntegrationRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("IntegrationRepository.Get", domain.AggregateTypeIntegration, id.String())
	}

	events, err := IntegrationEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("IntegrationRepository.Get", id.String(), err)
	}

	integ := domain.NewIntegrationFromEvents(id.String())
	eventsourcing.Hydrate(integ, events)

	return integ, nil
}

// Exists checks if an integration with the given ID exists.
func (r *IntegrationRepository) Exists(ctx context.Context, id domain.IntegrationID) (bool, error) {
	exists, err := r.queries.IntegrationAggregateExists(ctx, uuidFromIntegrationID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("IntegrationRepository.Exists", id.String(), err)
	}
	return exists, nil
}
