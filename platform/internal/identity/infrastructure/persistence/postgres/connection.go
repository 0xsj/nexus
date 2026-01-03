package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Connection Repository
// ============================================================================

// ConnectionRepository implements domain.ConnectionRepository using PostgreSQL.
type ConnectionRepository struct {
	adapter *postgres.BaseAdapter
}

// Ensure ConnectionRepository implements domain.ConnectionRepository.
var _ domain.ConnectionRepository = (*ConnectionRepository)(nil)

// NewConnectionRepository creates a new connection repository.
func NewConnectionRepository(adapter *postgres.BaseAdapter) *ConnectionRepository {
	return &ConnectionRepository{
		adapter: adapter,
	}
}

// ============================================================================
// Save
// ============================================================================

// Save persists a connection (create or update).
func (r *ConnectionRepository) Save(ctx context.Context, connection *domain.Connection) error {
	const op = "postgres.ConnectionRepository.Save"

	exists, err := r.existsByID(ctx, connection.ID())
	if err != nil {
		return errors.Wrap(err, op)
	}

	row, err := ToConnectionRow(connection)
	if err != nil {
		return errors.Wrap(err, op)
	}

	if exists {
		return r.update(ctx, row)
	}

	return r.insert(ctx, row)
}

func (r *ConnectionRepository) insert(ctx context.Context, row *ConnectionRow) error {
	const op = "postgres.ConnectionRepository.insert"

	_, err := r.adapter.Exec(ctx, queryConnectionInsert,
		row.ID,
		row.UserID,
		row.Provider,
		row.ProviderUserID,
		row.ProviderUsername,
		row.ProviderEmail,
		row.ProviderAvatar,
		row.ProfileData,
		row.AccessToken,
		row.RefreshToken,
		row.TokenExpiresAt,
		row.Scopes,
		row.Status,
		row.LastSyncedAt,
		row.ConnectedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *ConnectionRepository) update(ctx context.Context, row *ConnectionRow) error {
	const op = "postgres.ConnectionRepository.update"

	_, err := r.adapter.Exec(ctx, queryConnectionUpdate,
		row.ID,
		row.ProviderUsername,
		row.ProviderEmail,
		row.ProviderAvatar,
		row.ProfileData,
		row.AccessToken,
		row.RefreshToken,
		row.TokenExpiresAt,
		row.Scopes,
		row.Status,
		row.LastSyncedAt,
		row.SyncError,
		row.UpdatedAt,
		row.DisconnectedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Find
// ============================================================================

// FindByID finds a connection by ID.
func (r *ConnectionRepository) FindByID(ctx context.Context, id string) (*domain.Connection, error) {
	const op = "postgres.ConnectionRepository.FindByID"

	row, err := r.scanConnection(ctx, queryConnectionByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound(op, "", id)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainConnection(), nil
}

// FindByUserAndProvider finds a connection by user and provider.
func (r *ConnectionRepository) FindByUserAndProvider(ctx context.Context, userID string, provider domain.OAuthProvider) (*domain.Connection, error) {
	const op = "postgres.ConnectionRepository.FindByUserAndProvider"

	row, err := r.scanConnection(ctx, queryConnectionByUserAndProvider, userID, provider.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound(op, provider.String(), userID)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainConnection(), nil
}

// FindByProviderUserID finds a connection by provider user ID.
func (r *ConnectionRepository) FindByProviderUserID(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.Connection, error) {
	const op = "postgres.ConnectionRepository.FindByProviderUserID"

	row, err := r.scanConnection(ctx, queryConnectionByProviderUserID, provider.String(), providerUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConnectionNotFound(op, provider.String(), providerUserID)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainConnection(), nil
}

// FindByUserID finds all connections for a user.
func (r *ConnectionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Connection, error) {
	const op = "postgres.ConnectionRepository.FindByUserID"

	return r.findConnections(ctx, op, queryConnectionsByUserID, userID)
}

// FindActiveByUserID finds all active connections for a user.
func (r *ConnectionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.Connection, error) {
	const op = "postgres.ConnectionRepository.FindActiveByUserID"

	return r.findConnections(ctx, op, queryConnectionsActiveByUserID, userID)
}

// FindStale finds connections that need syncing.
func (r *ConnectionRepository) FindStale(ctx context.Context, staleDuration time.Duration, limit int) ([]*domain.Connection, error) {
	const op = "postgres.ConnectionRepository.FindStale"

	staleTime := time.Now().Add(-staleDuration)
	return r.findConnections(ctx, op, queryConnectionsStale, staleTime, limit)
}

// ============================================================================
// Delete
// ============================================================================

// Delete removes a connection.
func (r *ConnectionRepository) Delete(ctx context.Context, id string) error {
	const op = "postgres.ConnectionRepository.Delete"

	result, err := r.adapter.Exec(ctx, queryConnectionDelete, id)
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrConnectionNotFound(op, "", id)
	}

	return nil
}

// DeleteByUserID removes all connections for a user.
func (r *ConnectionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const op = "postgres.ConnectionRepository.DeleteByUserID"

	_, err := r.adapter.Exec(ctx, queryConnectionDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Exists
// ============================================================================

// ExistsByUserAndProvider checks if a connection exists.
func (r *ConnectionRepository) ExistsByUserAndProvider(ctx context.Context, userID string, provider domain.OAuthProvider) (bool, error) {
	const op = "postgres.ConnectionRepository.ExistsByUserAndProvider"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryConnectionExistsByUserAndProvider, userID, provider.String())
	if err := row.Scan(&exists); err != nil {
		return false, errors.Wrap(err, op)
	}

	return exists, nil
}

func (r *ConnectionRepository) existsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM connections WHERE id = $1)", id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (r *ConnectionRepository) scanConnection(ctx context.Context, query string, args ...interface{}) (*ConnectionRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var connRow ConnectionRow
	err := row.Scan(
		&connRow.ID,
		&connRow.UserID,
		&connRow.Provider,
		&connRow.ProviderUserID,
		&connRow.ProviderUsername,
		&connRow.ProviderEmail,
		&connRow.ProviderAvatar,
		&connRow.ProfileData,
		&connRow.AccessToken,
		&connRow.RefreshToken,
		&connRow.TokenExpiresAt,
		&connRow.Scopes,
		&connRow.Status,
		&connRow.LastSyncedAt,
		&connRow.SyncError,
		&connRow.ConnectedAt,
		&connRow.UpdatedAt,
		&connRow.DisconnectedAt,
	)
	if err != nil {
		return nil, err
	}

	return &connRow, nil
}

func (r *ConnectionRepository) findConnections(ctx context.Context, op, query string, args ...interface{}) ([]*domain.Connection, error) {
	rows, err := r.adapter.Select(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var connections []*domain.Connection
	for rows.Next() {
		var connRow ConnectionRow
		err := rows.Scan(
			&connRow.ID,
			&connRow.UserID,
			&connRow.Provider,
			&connRow.ProviderUserID,
			&connRow.ProviderUsername,
			&connRow.ProviderEmail,
			&connRow.ProviderAvatar,
			&connRow.ProfileData,
			&connRow.AccessToken,
			&connRow.RefreshToken,
			&connRow.TokenExpiresAt,
			&connRow.Scopes,
			&connRow.Status,
			&connRow.LastSyncedAt,
			&connRow.SyncError,
			&connRow.ConnectedAt,
			&connRow.UpdatedAt,
			&connRow.DisconnectedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		connections = append(connections, connRow.ToDomainConnection())
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return connections, nil
}
