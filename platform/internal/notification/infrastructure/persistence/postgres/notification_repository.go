package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Ensure NotificationRepository implements domain.NotificationRepository.
var _ domain.NotificationRepository = (*NotificationRepository)(nil)

// NotificationRepository is a PostgreSQL implementation of domain.NotificationRepository.
// It uses event sourcing for the write model and maintains a projection for reads.
type NotificationRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewNotificationRepository creates a new PostgreSQL notification repository.
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// Save persists a notification aggregate by appending new events and updating the projection.
func (r *NotificationRepository) Save(ctx context.Context, n *domain.Notification) error {
	if !n.HasChanges() {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eventsourcing.ErrEventStoreFailed("NotificationRepository.Save", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get current version for optimistic concurrency
	currentVersion, err := qtx.GetLatestNotificationEventVersion(ctx, uuidFromNotificationID(n.ID()))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return eventsourcing.ErrEventStoreFailed("NotificationRepository.Save", err)
	}

	// Calculate expected version (version before changes)
	expectedVersion := n.Version() - len(n.Changes())
	if int(currentVersion) != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(
			"NotificationRepository.Save",
			n.ID().String(),
			expectedVersion,
			int(currentVersion),
		)
	}

	// Append events
	for i, event := range n.Changes() {
		version := expectedVersion + i + 1
		params, err := NotificationEventToInsertParams(n.ID().String(), event, version)
		if err != nil {
			return eventsourcing.ErrEventStoreFailed("NotificationRepository.Save", err)
		}

		if err := qtx.InsertNotificationEvent(ctx, params); err != nil {
			return eventsourcing.ErrEventStoreFailed("NotificationRepository.Save", err)
		}
	}

	// Update projection
	upsertParams := NotificationToUpsertParams(n)
	if err := qtx.UpsertNotification(ctx, upsertParams); err != nil {
		return eventsourcing.ErrProjectionFailed("NotificationRepository.Save", "notification", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return eventsourcing.ErrEventStoreFailed("NotificationRepository.Save", err)
	}

	n.ClearChanges()
	return nil
}

// Get retrieves a notification by ID, replaying events to rebuild state.
func (r *NotificationRepository) Get(ctx context.Context, id domain.NotificationID) (*domain.Notification, error) {
	rows, err := r.queries.GetNotificationEvents(ctx, uuidFromNotificationID(id))
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("NotificationRepository.Get", id.String(), err)
	}

	if len(rows) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("NotificationRepository.Get", domain.AggregateTypeNotification, id.String())
	}

	events, err := NotificationEventRowsToEvents(rows)
	if err != nil {
		return nil, eventsourcing.ErrEventLoadFailed("NotificationRepository.Get", id.String(), err)
	}

	n := domain.NewNotificationFromEvents(id.String())
	eventsourcing.Hydrate(n, events)

	return n, nil
}

// Exists checks if a notification with the given ID exists.
func (r *NotificationRepository) Exists(ctx context.Context, id domain.NotificationID) (bool, error) {
	exists, err := r.queries.NotificationAggregateExists(ctx, uuidFromNotificationID(id))
	if err != nil {
		return false, eventsourcing.ErrEventLoadFailed("NotificationRepository.Exists", id.String(), err)
	}
	return exists, nil
}
