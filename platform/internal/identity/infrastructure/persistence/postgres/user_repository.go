package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure UserRepository implements domain.UserRepository.
var _ domain.UserRepository = (*UserRepository)(nil)

// UserRepository is a PostgreSQL implementation of domain.UserRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type UserRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a user aggregate by appending new events and updating the projection.
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	if !user.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("UserRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestUserEventVersion(ctx, uuidFromUserID(user.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("UserRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := user.Version() - len(user.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"UserRepository.Save",
			user.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range user.Changes() {
		version := expectedVersion + i + 1
		params, err := UserEventToInsertParams(user.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("UserRepository.Save", err)
		}

		if err := qtx.InsertUserEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("UserRepository.Save", err)
		}
	}

	// Update projection
	if err := r.updateProjection(ctx, qtx, user); err != nil {
		return eventsourcing.ErrProjectionFailed("UserRepository.Save", "user", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("UserRepository.Save", err)
	}

	user.ClearChanges()
	return nil
}

// Get retrieves a user by ID, replaying events to rebuild state.
func (r *UserRepository) Get(ctx context.Context, id domain.UserID) (*domain.User, error) {
	rows, err := r.queries.GetUserEvents(ctx, uuidFromUserID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("UserRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("UserRepository.Get", domain.AggregateTypeUser, id.String())
	}

	events, err := UserEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("UserRepository.Get", id.String(), err)
	}

	user := domain.NewUserFromEvents(id.String())
	eventsourcing.Hydrate(user, events)

	return user, nil
}

// Exists checks if a user with the given ID exists.
func (r *UserRepository) Exists(ctx context.Context, id domain.UserID) (bool, error) {
	exists, err := r.queries.UserAggregateExists(ctx, uuidFromUserID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("UserRepository.Exists", id.String(), err)
	}
	return exists, nil
}

// updateProjection updates the read model tables for a user.
func (r *UserRepository) updateProjection(ctx context.Context, qtx *generated.Queries, user *domain.User) error {
	// Upsert main user record
	if err := qtx.UpsertUser(ctx, UserToUpsertParams(user)); err != nil {
		return err
	}

	// Sync DIDs - delete all and re-insert
	if err := qtx.DeleteAllUserDIDs(ctx, uuidFromUserID(user.ID())); err != nil {
		return err
	}

	for _, did := range user.DIDs() {
		isPrimary := did == user.PrimaryDID()
		params := UserDIDToInsertParams(user.ID(), did, isPrimary, user.CreatedAt())
		if err := qtx.InsertUserDID(ctx, params); err != nil {
			return err
		}
	}

	// Sync OAuth links - delete all and re-insert
	if err := qtx.DeleteAllUserOAuthLinks(ctx, uuidFromUserID(user.ID())); err != nil {
		return err
	}

	for _, link := range user.OAuthLinks() {
		params := UserOAuthLinkToInsertParams(user.ID(), link, "", user.CreatedAt())
		if err := qtx.InsertUserOAuthLink(ctx, params); err != nil {
			return err
		}
	}

	return nil
}
