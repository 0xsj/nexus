package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres/generated"
)

// NotificationLookup provides read-side queries on the notifications projection table.
type NotificationLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewNotificationLookup creates a new NotificationLookup.
func NewNotificationLookup(pool *pgxpool.Pool) *NotificationLookup {
	return &NotificationLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a notification projection by ID.
func (l *NotificationLookup) GetByID(ctx context.Context, id string) (*NotificationProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetNotificationByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := NotificationRowToProjection(row)
	return &proj, nil
}

// ExistsByID checks if a notification exists by ID.
func (l *NotificationLookup) ExistsByID(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	return l.queries.NotificationExistsByID(ctx, uid)
}

// ListByRecipient returns notifications for a recipient with pagination.
func (l *NotificationLookup) ListByRecipient(ctx context.Context, recipientID string, limit, offset int) ([]NotificationProjection, error) {
	rows, err := l.queries.ListNotificationsByRecipient(ctx, generated.ListNotificationsByRecipientParams{
		RecipientID: recipientID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]NotificationProjection, len(rows))
	for i, row := range rows {
		projections[i] = NotificationRowToProjection(row)
	}
	return projections, nil
}

// CountByRecipient returns the count of notifications for a recipient.
func (l *NotificationLookup) CountByRecipient(ctx context.Context, recipientID string) (int, error) {
	count, err := l.queries.CountNotificationsByRecipient(ctx, recipientID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// CountUnreadByRecipient returns the count of unread notifications for a recipient.
func (l *NotificationLookup) CountUnreadByRecipient(ctx context.Context, recipientID string) (int, error) {
	count, err := l.queries.CountUnreadByRecipient(ctx, recipientID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
