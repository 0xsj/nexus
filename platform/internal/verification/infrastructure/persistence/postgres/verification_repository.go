package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure VerificationRepository implements domain.VerificationRepository.
var _ domain.VerificationRepository = (*VerificationRepository)(nil)

// VerificationRepository is a PostgreSQL implementation of domain.VerificationRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type VerificationRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewVerificationRepository creates a new PostgreSQL verification repository.
func NewVerificationRepository(pool *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a verification aggregate by appending new events and updating the projection.
func (r *VerificationRepository) Save(ctx context.Context, v *domain.Verification) error {
	if !v.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("VerificationRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestVerificationEventVersion(ctx, uuidFromVerificationID(v.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("VerificationRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := v.Version() - len(v.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"VerificationRepository.Save",
			v.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range v.Changes() {
		version := expectedVersion + i + 1
		params, err := VerificationEventToInsertParams(v.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("VerificationRepository.Save", err)
		}

		if err := qtx.InsertVerificationEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("VerificationRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := VerificationToUpsertParams(v)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("VerificationRepository.Save", "verification", err)
	}

	if err := qtx.UpsertVerification(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("VerificationRepository.Save", "verification", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("VerificationRepository.Save", err)
	}

	v.ClearChanges()
	return nil
}

// Get retrieves a verification by ID, replaying events to rebuild state.
func (r *VerificationRepository) Get(ctx context.Context, id domain.VerificationID) (*domain.Verification, error) {
	rows, err := r.queries.GetVerificationEvents(ctx, uuidFromVerificationID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("VerificationRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("VerificationRepository.Get", domain.AggregateTypeVerification, id.String())
	}

	events, err := VerificationEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("VerificationRepository.Get", id.String(), err)
	}

	v := domain.NewVerificationFromEvents(id.String())
	eventsourcing.Hydrate(v, events)

	return v, nil
}

// Exists checks if a verification with the given ID exists.
func (r *VerificationRepository) Exists(ctx context.Context, id domain.VerificationID) (bool, error) {
	exists, err := r.queries.VerificationAggregateExists(ctx, uuidFromVerificationID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("VerificationRepository.Exists", id.String(), err)
	}
	return exists, nil
}
