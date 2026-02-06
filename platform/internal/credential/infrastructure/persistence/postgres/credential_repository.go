package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure CredentialRepository implements domain.CredentialRepository.
var _ domain.CredentialRepository = (*CredentialRepository)(nil)

// CredentialRepository is a PostgreSQL implementation of domain.CredentialRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type CredentialRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewCredentialRepository creates a new PostgreSQL credential repository.
func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a credential aggregate by appending new events and updating the projection.
func (r *CredentialRepository) Save(ctx context.Context, cred *domain.Credential) error {
	if !cred.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("CredentialRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestCredentialEventVersion(ctx, uuidFromCredentialID(cred.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("CredentialRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := cred.Version() - len(cred.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"CredentialRepository.Save",
			cred.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range cred.Changes() {
		version := expectedVersion + i + 1
		params, err := CredentialEventToInsertParams(cred.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("CredentialRepository.Save", err)
		}

		if err := qtx.InsertCredentialEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("CredentialRepository.Save", err)
		}
	}

	// Update projection
	upsertParams, err := CredentialToUpsertParams(cred)
	if err != nil {
		return eventsourcing.ErrProjectionFailed("CredentialRepository.Save", "credential", err)
	}

	if err := qtx.UpsertCredential(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("CredentialRepository.Save", "credential", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("CredentialRepository.Save", err)
	}

	cred.ClearChanges()
	return nil
}

// Get retrieves a credential by ID, replaying events to rebuild state.
func (r *CredentialRepository) Get(ctx context.Context, id domain.CredentialID) (*domain.Credential, error) {
	rows, err := r.queries.GetCredentialEvents(ctx, uuidFromCredentialID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("CredentialRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("CredentialRepository.Get", domain.AggregateTypeCredential, id.String())
	}

	events, err := CredentialEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("CredentialRepository.Get", id.String(), err)
	}

	cred := domain.NewCredentialFromEvents(id.String())
	eventsourcing.Hydrate(cred, events)

	return cred, nil
}

// Exists checks if a credential with the given ID exists.
func (r *CredentialRepository) Exists(ctx context.Context, id domain.CredentialID) (bool, error) {
	exists, err := r.queries.CredentialAggregateExists(ctx, uuidFromCredentialID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("CredentialRepository.Exists", id.String(), err)
	}
	return exists, nil
}
