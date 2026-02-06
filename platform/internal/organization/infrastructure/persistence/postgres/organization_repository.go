package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/organization/domain"
	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure OrganizationRepository implements domain.OrganizationRepository.
var _ domain.OrganizationRepository = (*OrganizationRepository)(nil)

// OrganizationRepository is a PostgreSQL implementation of domain.OrganizationRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type OrganizationRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewOrganizationRepository creates a new PostgreSQL organization repository.
func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists an organization aggregate by appending new events and updating the projection.
func (r *OrganizationRepository) Save(ctx context.Context, org *domain.Organization) error {
	if !org.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("OrganizationRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestOrganizationEventVersion(ctx, uuidFromOrganizationID(org.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("OrganizationRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := org.Version() - len(org.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"OrganizationRepository.Save",
			org.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range org.Changes() {
		version := expectedVersion + i + 1
		params, err := OrganizationEventToInsertParams(org.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("OrganizationRepository.Save", err)
		}

		if err := qtx.InsertOrganizationEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("OrganizationRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := OrganizationToUpsertParams(org)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("OrganizationRepository.Save", "organization", err)
	}

	if err := qtx.UpsertOrganization(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("OrganizationRepository.Save", "organization", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("OrganizationRepository.Save", err)
	}

	org.ClearChanges()
	return nil
}

// Get retrieves an organization by ID, replaying events to rebuild state.
func (r *OrganizationRepository) Get(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	rows, err := r.queries.GetOrganizationEvents(ctx, uuidFromOrganizationID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("OrganizationRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("OrganizationRepository.Get", domain.AggregateTypeOrganization, id.String())
	}

	events, err := OrganizationEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("OrganizationRepository.Get", id.String(), err)
	}

	org := domain.NewOrganizationFromEvents(id.String())
	eventsourcing.Hydrate(org, events)

	return org, nil
}

// Exists checks if an organization with the given ID exists.
func (r *OrganizationRepository) Exists(ctx context.Context, id domain.OrganizationID) (bool, error) {
	exists, err := r.queries.OrganizationAggregateExists(ctx, uuidFromOrganizationID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("OrganizationRepository.Exists", id.String(), err)
	}
	return exists, nil
}
