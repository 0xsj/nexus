package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Nonce Repository
// ============================================================================

// NonceRepository implements domain.NonceRepository using PostgreSQL.
type NonceRepository struct {
	adapter *postgres.BaseAdapter
}

// NewNonceRepository creates a new PostgreSQL nonce repository.
func NewNonceRepository(adapter *postgres.BaseAdapter) *NonceRepository {
	return &NonceRepository{adapter: adapter}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save persists a nonce.
func (r *NonceRepository) Save(ctx context.Context, nonce *domain.Nonce) error {
	const op = "NonceRepository.Save"

	_, err := r.adapter.Exec(ctx, queryInsertNonce,
		nonce.Value,
		nonce.CreatedAt,
		nonce.ExpiresAt,
		nonce.Used,
		ptrToNullTime(nonce.UsedAt),
	)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	return nil
}

// MarkUsed marks a nonce as used.
func (r *NonceRepository) MarkUsed(ctx context.Context, value string) error {
	const op = "NonceRepository.MarkUsed"

	now := time.Now()
	result, err := r.adapter.Exec(ctx, queryUpdateNonce,
		true,  // used
		now,   // used_at
		value, // nonce
	)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrInvalidNonce(op, value)
	}

	return nil
}

// Delete removes a nonce.
func (r *NonceRepository) Delete(ctx context.Context, value string) error {
	const op = "NonceRepository.Delete"

	result, err := r.adapter.Exec(ctx, queryDeleteNonce, value)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrInvalidNonce(op, value)
	}

	return nil
}

// DeleteExpired removes all expired nonces.
func (r *NonceRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "NonceRepository.DeleteExpired"

	result, err := r.adapter.Exec(ctx, queryDeleteExpiredNonces, time.Now())
	if err != nil {
		return 0, pkgerrors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, pkgerrors.Wrap(err, op)
	}

	return rowsAffected, nil
}

// ============================================================================
// Read Operations
// ============================================================================

// FindByValue finds a nonce by its value.
func (r *NonceRepository) FindByValue(ctx context.Context, value string) (*domain.Nonce, error) {
	const op = "NonceRepository.FindByValue"

	row := r.adapter.Executor().QueryRow(ctx, querySelectNonce, value)

	nonceRow, err := scanNonceRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidNonce(op, value)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return nonceRow.toDomain(), nil
}

// IsUsed checks if a nonce has been used.
func (r *NonceRepository) IsUsed(ctx context.Context, value string) (bool, error) {
	const op = "NonceRepository.IsUsed"

	var used bool
	row := r.adapter.Executor().QueryRow(ctx, queryIsNonceUsed, value)
	if err := row.Scan(&used); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Nonce doesn't exist, treat as not used
			return false, nil
		}
		return false, pkgerrors.Wrap(err, op)
	}

	return used, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanNonceRow scans a single nonce row.
func scanNonceRow(row interface{ Scan(...interface{}) error }) (*nonceRow, error) {
	var r nonceRow
	err := row.Scan(
		&r.Nonce,
		&r.CreatedAt,
		&r.ExpiresAt,
		&r.Used,
		&r.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// toDomain converts a nonce row to a domain nonce.
func (r *nonceRow) toDomain() *domain.Nonce {
	return &domain.Nonce{
		Value:     r.Nonce,
		Used:      r.Used,
		UsedAt:    nullTimeToPtr(r.UsedAt),
		CreatedAt: r.CreatedAt,
		ExpiresAt: r.ExpiresAt,
	}
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.NonceRepository = (*NonceRepository)(nil)