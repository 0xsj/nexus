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

// Ensure SessionRepository implements domain.SessionRepository.
var _ domain.SessionRepository = (*SessionRepository)(nil)

// SessionRepository is a PostgreSQL implementation of domain.SessionRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type SessionRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewSessionRepository creates a new PostgreSQL session repository.
func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a session aggregate by appending new events and updating the projection.
func (r *SessionRepository) Save(ctx context.Context, session *domain.Session) error {
	if !session.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("SessionRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestSessionEventVersion(ctx, uuidFromSessionID(session.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("SessionRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := session.Version() - len(session.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"SessionRepository.Save",
			session.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range session.Changes() {
		version := expectedVersion + i + 1
		params, err := SessionEventToInsertParams(session.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("SessionRepository.Save", err)
		}

		if err := qtx.InsertSessionEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("SessionRepository.Save", err)
		}
	}

	// Update projection
	if err := qtx.UpsertSession(ctx, SessionToUpsertParams(session)); err != nil {
		return eventsourcing.ErrProjectionFailed("SessionRepository.Save", "session", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("SessionRepository.Save", err)
	}

	session.ClearChanges()
	return nil
}

// Get retrieves a session by ID, replaying events to rebuild state.
func (r *SessionRepository) Get(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	rows, err := r.queries.GetSessionEvents(ctx, uuidFromSessionID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("SessionRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("SessionRepository.Get", domain.AggregateTypeSession, id.String())
	}

	events, err := SessionEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("SessionRepository.Get", id.String(), err)
	}

	session := domain.NewSessionFromEvents(id.String())
	eventsourcing.Hydrate(session, events)

	return session, nil
}

// Exists checks if a session with the given ID exists.
func (r *SessionRepository) Exists(ctx context.Context, id domain.SessionID) (bool, error) {
	exists, err := r.queries.SessionAggregateExists(ctx, uuidFromSessionID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("SessionRepository.Exists", id.String(), err)
	}
	return exists, nil
}
