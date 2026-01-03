package postgres

import (
	"context"
	"database/sql"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Token Repository
// ============================================================================

// TokenRepository implements domain.TokenRepository using PostgreSQL.
type TokenRepository struct {
	adapter *postgres.BaseAdapter
}

// Ensure TokenRepository implements domain.TokenRepository.
var _ domain.TokenRepository = (*TokenRepository)(nil)

// NewTokenRepository creates a new token repository.
func NewTokenRepository(adapter *postgres.BaseAdapter) *TokenRepository {
	return &TokenRepository{
		adapter: adapter,
	}
}

// ============================================================================
// Save
// ============================================================================

// Save stores a refresh token.
func (r *TokenRepository) Save(ctx context.Context, token *domain.RefreshToken) error {
	const op = "postgres.TokenRepository.Save"

	exists, err := r.existsByID(ctx, token.ID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	row := ToRefreshTokenRow(token)

	if exists {
		return r.update(ctx, row)
	}

	return r.insert(ctx, row)
}

func (r *TokenRepository) insert(ctx context.Context, row *RefreshTokenRow) error {
	const op = "postgres.TokenRepository.insert"

	_, err := r.adapter.Exec(ctx, queryRefreshTokenInsert,
		row.ID,
		row.UserID,
		row.SessionID,
		row.TokenHash,
		row.Revoked,
		row.CreatedAt,
		row.ExpiresAt,
	)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *TokenRepository) update(ctx context.Context, row *RefreshTokenRow) error {
	const op = "postgres.TokenRepository.update"

	_, err := r.adapter.Exec(ctx, queryRefreshTokenUpdate,
		row.ID,
		row.Revoked,
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

// FindByTokenHash finds a refresh token by hash.
func (r *TokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	const op = "postgres.TokenRepository.FindByTokenHash"

	row, err := r.scanToken(ctx, queryRefreshTokenByTokenHash, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenInvalid(op, "token not found")
		}
		return nil, errors.Wrap(err, op)
	}

	return row.ToDomainRefreshToken(), nil
}

// FindByUserID finds all refresh tokens for a user.
func (r *TokenRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	const op = "postgres.TokenRepository.FindByUserID"

	return r.findTokens(ctx, op, queryRefreshTokensByUserID, userID)
}

// ============================================================================
// Delete
// ============================================================================

// Delete removes a refresh token.
func (r *TokenRepository) Delete(ctx context.Context, id string) error {
	const op = "postgres.TokenRepository.Delete"

	result, err := r.adapter.Exec(ctx, queryRefreshTokenDelete, id)
	if err != nil {
		return errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrTokenInvalid(op, "token not found")
	}

	return nil
}

// DeleteByTokenHash removes a refresh token by hash.
func (r *TokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	const op = "postgres.TokenRepository.DeleteByTokenHash"

	_, err := r.adapter.Exec(ctx, queryRefreshTokenDeleteByTokenHash, tokenHash)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// DeleteByUserID removes all refresh tokens for a user.
func (r *TokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const op = "postgres.TokenRepository.DeleteByUserID"

	_, err := r.adapter.Exec(ctx, queryRefreshTokenDeleteByUserID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

// DeleteExpired removes all expired refresh tokens.
func (r *TokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "postgres.TokenRepository.DeleteExpired"

	result, err := r.adapter.Exec(ctx, queryRefreshTokenDeleteExpired)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	return rowsAffected, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (r *TokenRepository) existsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE id = $1)", id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *TokenRepository) scanToken(ctx context.Context, query string, args ...interface{}) (*RefreshTokenRow, error) {
	row := r.adapter.Executor().QueryRow(ctx, query, args...)

	var tokenRow RefreshTokenRow
	err := row.Scan(
		&tokenRow.ID,
		&tokenRow.UserID,
		&tokenRow.SessionID,
		&tokenRow.TokenHash,
		&tokenRow.Revoked,
		&tokenRow.CreatedAt,
		&tokenRow.ExpiresAt,
		&tokenRow.RevokedAt,
	)
	if err != nil {
		return nil, err
	}

	return &tokenRow, nil
}

func (r *TokenRepository) findTokens(ctx context.Context, op, query string, args ...interface{}) ([]*domain.RefreshToken, error) {
	rows, err := r.adapter.Select(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		var tokenRow RefreshTokenRow
		err := rows.Scan(
			&tokenRow.ID,
			&tokenRow.UserID,
			&tokenRow.SessionID,
			&tokenRow.TokenHash,
			&tokenRow.Revoked,
			&tokenRow.CreatedAt,
			&tokenRow.ExpiresAt,
			&tokenRow.RevokedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		tokens = append(tokens, tokenRow.ToDomainRefreshToken())
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return tokens, nil
}
