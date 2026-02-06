package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/profile/domain"
	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure ProfileRepository implements domain.ProfileRepository.
var _ domain.ProfileRepository = (*ProfileRepository)(nil)

// ProfileRepository is a PostgreSQL implementation of domain.ProfileRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type ProfileRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewProfileRepository creates a new PostgreSQL profile repository.
func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a profile aggregate by appending new events and updating the projection.
func (r *ProfileRepository) Save(ctx context.Context, p *domain.Profile) error {
	if !p.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("ProfileRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestProfileEventVersion(ctx, uuidFromProfileID(p.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("ProfileRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := p.Version() - len(p.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"ProfileRepository.Save",
			p.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range p.Changes() {
		version := expectedVersion + i + 1
		params, err := ProfileEventToInsertParams(p.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("ProfileRepository.Save", err)
		}

		if err := qtx.InsertProfileEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("ProfileRepository.Save", err)
		}
	}

	// Update projection
	upsertParams := ProfileToUpsertParams(p)

	if err := qtx.UpsertProfile(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("ProfileRepository.Save", "profile", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("ProfileRepository.Save", err)
	}

	p.ClearChanges()
	return nil
}

// Get retrieves a profile by ID, replaying events to rebuild state.
func (r *ProfileRepository) Get(ctx context.Context, id domain.ProfileID) (*domain.Profile, error) {
	rows, err := r.queries.GetProfileEvents(ctx, uuidFromProfileID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("ProfileRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("ProfileRepository.Get", domain.AggregateTypeProfile, id.String())
	}

	events, err := ProfileEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("ProfileRepository.Get", id.String(), err)
	}

	p := domain.NewProfileFromEvents(id.String())
	eventsourcing.Hydrate(p, events)

	return p, nil
}

// Exists checks if a profile with the given ID exists.
func (r *ProfileRepository) Exists(ctx context.Context, id domain.ProfileID) (bool, error) {
	exists, err := r.queries.ProfileAggregateExists(ctx, uuidFromProfileID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("ProfileRepository.Exists", id.String(), err)
	}
	return exists, nil
}
