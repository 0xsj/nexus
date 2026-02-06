package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure IssuerRepository implements domain.IssuerRepository.
var _ domain.IssuerRepository = (*IssuerRepository)(nil)

// IssuerRepository is a PostgreSQL implementation of domain.IssuerRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type IssuerRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewIssuerRepository creates a new PostgreSQL issuer repository.
func NewIssuerRepository(pool *pgxpool.Pool) *IssuerRepository {
	return &IssuerRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists an issuer aggregate by appending new events and updating the projection.
func (r *IssuerRepository) Save(ctx context.Context, iss *domain.Issuer) error {
	if !iss.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("IssuerRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestIssuerEventVersion(ctx, uuidFromIssuerID(iss.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("IssuerRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := iss.Version() - len(iss.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"IssuerRepository.Save",
			iss.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range iss.Changes() {
		version := expectedVersion + i + 1
		params, err := IssuerEventToInsertParams(iss.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("IssuerRepository.Save", err)
		}

		if err := qtx.InsertIssuerEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("IssuerRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := IssuerToUpsertParams(iss)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("IssuerRepository.Save", "issuer", err)
	}

	if err := qtx.UpsertIssuer(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("IssuerRepository.Save", "issuer", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("IssuerRepository.Save", err)
	}

	iss.ClearChanges()
	return nil
}

// Get retrieves an issuer by ID, replaying events to rebuild state.
func (r *IssuerRepository) Get(ctx context.Context, id domain.IssuerID) (*domain.Issuer, error) {
	rows, err := r.queries.GetIssuerEvents(ctx, uuidFromIssuerID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("IssuerRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("IssuerRepository.Get", domain.AggregateTypeIssuer, id.String())
	}

	events, err := IssuerEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("IssuerRepository.Get", id.String(), err)
	}

	iss := domain.NewIssuerFromEvents(id.String())
	eventsourcing.Hydrate(iss, events)

	return iss, nil
}

// Exists checks if an issuer with the given ID exists.
func (r *IssuerRepository) Exists(ctx context.Context, id domain.IssuerID) (bool, error) {
	exists, err := r.queries.IssuerAggregateExists(ctx, uuidFromIssuerID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("IssuerRepository.Exists", id.String(), err)
	}
	return exists, nil
}
