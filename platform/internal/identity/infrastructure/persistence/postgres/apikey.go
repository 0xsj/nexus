package postgres

import (
	"context"
	"database/sql"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// API Key Repository
// ============================================================================

// APIKeyRepository implements domain.APIKeyRepository using PostgreSQL.
type APIKeyRepository struct {
	adapter *postgres.BaseAdapter
}

// Ensure APIKeyRepository implements domain.APIKeyRepository.
var _ domain.APIKeyRepository = (*APIKeyRepository)(nil)

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(adapter *postgres.BaseAdapter) *APIKeyRepository {
	return &APIKeyRepository{
		adapter: adapter,
	}
}

// ============================================================================
// Save
// ============================================================================

// Save persists an API key (create or update).
func (r *APIKeyRepository) Save(ctx context.Context, apiKey *domain.APIKey) error {
	const op = "postgres.APIKeyRepository.Save"

	exists, err := r.existsByID(ctx, apiKey.ID())
	if err != nil {
		return errors.Wrap(err, op)
	}

	row := ToAPIKeyRow(apiKey)

	if exists {
		return r.update(ctx, row)
	}

	return r.insert(ctx, row)
}

func (r *APIKeyRepository) insert(ctx context.Context, row *APIKeyRow) error {
	const op = "postgres.APIKeyRepository.insert"

	_, err := r.adapter.Exec(ctx, queryAPIKeyInsert,
		row.ID,
		row.UserID,
		row.Name,
		row.KeyHash,
		row.KeyPrefix,
		row.Scopes,
		row.RateLimit,
		row.Status,
		row.CreatedAt,
		row.ExpiresAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *APIKeyRepository) update(ctx context.Context, row *APIKeyRow) error {
	const op = "postgres.APIKeyRepository.update"

	_, err := r.adapter.Exec(ctx, queryAPIKeyUpdate,
		row.ID,
		row.Name,
		row.Scopes,
		row.RateLimit,
		row.Status,
		row.LastUsedAt,
		row.UsageCount,
		row.RevokedAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Find
// ============================================================================

// FindByID finds an API key by ID.
func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*domain.APIKey, error) {
	const op = "postgres.APIKeyRepository.FindByID"

	row, err := r.scanAPIKey(ctx, queryAPIKeyByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAPIKeyNotFound(op, id)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainAPIKey(), nil
}

// FindByKeyHash finds an API key by key hash.
func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	const op = "postgres.APIKeyRepository.FindByKeyHash"

	row, err := r.scanAPIKey(ctx, queryAPIKeyByKeyHash, keyHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAPIKeyNotFound(op, keyHash)
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainAPIKey(), nil
}

// FindByKeyPrefix finds API keys by prefix (for listing).
func (r *APIKeyRepository) FindByKeyPrefix(ctx context.Context, prefix string) ([]*domain.APIKey, error) {
	const op = "postgres.APIKeyRepository.FindByKeyPrefix"

	return r.findAPIKeys(ctx, op, queryAPIKeysByKeyPrefix, prefix)
}

// FindByUserID finds all API keys for a user.
func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	const op = "postgres.APIKeyRepository.FindByUserID"

	return r.findAPIKeys(ctx, op, queryAPIKeysByUserID, userID)
}

// FindActiveByUserID finds all active API keys for a user.
func (r *APIKeyRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	const op = "postgres.APIKeyRepository.FindActiveByUserID"

	return r.findAPIKeys(ctx, op, queryAPIKeysActiveByUserID, userID)
}

// ============================================================================
// Delete
// ============================================================================

// Delete removes an API key.
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	const op = "postgres.APIKeyRepository.Delete"

	result, err := r.adapter.Exec(ctx, queryAPIKeyDelete, id)
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrAPIKeyNotFound(op, id)
	}

	return nil
}

// DeleteByUserID removes all API keys for a user.
func (r *APIKeyRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const op = "postgres.APIKeyRepository.DeleteByUserID"

	_, err := r.adapter.Exec(ctx, queryAPIKeyDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// ============================================================================
// Count
// ============================================================================

// CountByUserID counts API keys for a user.
func (r *APIKeyRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	const op = "postgres.APIKeyRepository.CountByUserID"

	var count int
	row := r.adapter.Executor().QueryRow(ctx, queryAPIKeyCountByUserID, userID)
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

// CountActiveByUserID counts active API keys for a user.
func (r *APIKeyRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	const op = "postgres.APIKeyRepository.CountActiveByUserID"

	var count int
	row := r.adapter.Executor().QueryRow(ctx, queryAPIKeyCountActiveByUserID, userID)
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (r *APIKeyRepository) existsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM api_keys WHERE id = $1)", id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *APIKeyRepository) scanAPIKey(ctx context.Context, query string, args ...interface{}) (*APIKeyRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var apiKeyRow APIKeyRow
	err := row.Scan(
		&apiKeyRow.ID,
		&apiKeyRow.UserID,
		&apiKeyRow.Name,
		&apiKeyRow.KeyHash,
		&apiKeyRow.KeyPrefix,
		&apiKeyRow.Scopes,
		&apiKeyRow.RateLimit,
		&apiKeyRow.Status,
		&apiKeyRow.LastUsedAt,
		&apiKeyRow.UsageCount,
		&apiKeyRow.CreatedAt,
		&apiKeyRow.ExpiresAt,
		&apiKeyRow.RevokedAt,
	)
	if err != nil {
		return nil, err
	}

	return &apiKeyRow, nil
}

func (r *APIKeyRepository) findAPIKeys(ctx context.Context, op, query string, args ...interface{}) ([]*domain.APIKey, error) {
	rows, err := r.adapter.Select(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var apiKeys []*domain.APIKey
	for rows.Next() {
		var apiKeyRow APIKeyRow
		err := rows.Scan(
			&apiKeyRow.ID,
			&apiKeyRow.UserID,
			&apiKeyRow.Name,
			&apiKeyRow.KeyHash,
			&apiKeyRow.KeyPrefix,
			&apiKeyRow.Scopes,
			&apiKeyRow.RateLimit,
			&apiKeyRow.Status,
			&apiKeyRow.LastUsedAt,
			&apiKeyRow.UsageCount,
			&apiKeyRow.CreatedAt,
			&apiKeyRow.ExpiresAt,
			&apiKeyRow.RevokedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		apiKeys = append(apiKeys, apiKeyRow.ToDomainAPIKey())
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return apiKeys, nil
}
