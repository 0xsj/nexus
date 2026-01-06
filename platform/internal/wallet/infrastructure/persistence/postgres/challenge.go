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
// Challenge Repository
// ============================================================================

// ChallengeRepository implements domain.ChallengeRepository using PostgreSQL.
type ChallengeRepository struct {
	adapter *postgres.BaseAdapter
}

// NewChallengeRepository creates a new PostgreSQL challenge repository.
func NewChallengeRepository(adapter *postgres.BaseAdapter) *ChallengeRepository {
	return &ChallengeRepository{adapter: adapter}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save persists a challenge.
func (r *ChallengeRepository) Save(ctx context.Context, challenge *domain.Challenge) error {
	const op = "ChallengeRepository.Save"

	row := challengeToRow(challenge)

	_, err := r.adapter.Exec(ctx, queryInsertChallenge,
		row.Nonce,
		row.Message,
		row.Address,
		row.AddressNormalized,
		row.ChainID,
		row.Domain,
		row.URI,
		row.IssuedAt,
		row.ExpiresAt,
		row.Used,
		row.UsedAt,
	)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	return nil
}

// MarkUsed marks a challenge as used.
func (r *ChallengeRepository) MarkUsed(ctx context.Context, nonce string) error {
	const op = "ChallengeRepository.MarkUsed"

	now := time.Now()
	result, err := r.adapter.Exec(ctx, queryUpdateChallenge,
		true,  // used
		now,   // used_at
		nonce, // nonce
	)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrInvalidNonce(op, nonce)
	}

	return nil
}

// Delete removes a challenge.
func (r *ChallengeRepository) Delete(ctx context.Context, nonce string) error {
	const op = "ChallengeRepository.Delete"

	result, err := r.adapter.Exec(ctx, queryDeleteChallenge, nonce)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if rowsAffected == 0 {
		return domain.ErrInvalidNonce(op, nonce)
	}

	return nil
}

// DeleteExpired removes all expired challenges.
func (r *ChallengeRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "ChallengeRepository.DeleteExpired"

	result, err := r.adapter.Exec(ctx, queryDeleteExpiredChallenges, time.Now())
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

// FindByNonce finds a challenge by nonce.
func (r *ChallengeRepository) FindByNonce(ctx context.Context, nonce string) (*domain.Challenge, error) {
	const op = "ChallengeRepository.FindByNonce"

	row := r.adapter.Executor().QueryRow(ctx, querySelectChallengeByNonce, nonce)

	challengeRow, err := scanChallengeRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidNonce(op, nonce)
		}
		return nil, pkgerrors.Wrap(err, op)
	}

	return challengeRow.toDomain(), nil
}

// ExistsByNonce checks if a challenge with the nonce exists.
func (r *ChallengeRepository) ExistsByNonce(ctx context.Context, nonce string) (bool, error) {
	const op = "ChallengeRepository.ExistsByNonce"

	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExistsChallengeByNonce, nonce)
	if err := row.Scan(&exists); err != nil {
		return false, pkgerrors.Wrap(err, op)
	}

	return exists, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanChallengeRow scans a single challenge row.
func scanChallengeRow(row interface{ Scan(...interface{}) error }) (*challengeRow, error) {
	var r challengeRow
	err := row.Scan(
		&r.Nonce,
		&r.Message,
		&r.Address,
		&r.AddressNormalized,
		&r.ChainID,
		&r.Domain,
		&r.URI,
		&r.IssuedAt,
		&r.ExpiresAt,
		&r.Used,
		&r.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.ChallengeRepository = (*ChallengeRepository)(nil)
