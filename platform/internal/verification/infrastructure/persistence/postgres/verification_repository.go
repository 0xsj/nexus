package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/database"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Verification Repository
// ============================================================================

// VerificationRepository implements domain.VerificationRepository using PostgreSQL.
type VerificationRepository struct {
	adapter *pgadapter.BaseAdapter
}

// NewVerificationRepository creates a new VerificationRepository.
func NewVerificationRepository(db *pgadapter.DB) *VerificationRepository {
	return &VerificationRepository{
		adapter: pgadapter.NewBaseAdapter(db),
	}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save persists a verification aggregate.
func (r *VerificationRepository) Save(ctx context.Context, verification *domain.Verification) error {
	exists, err := r.ExistsByID(ctx, verification.AggregateID())
	if err != nil {
		return err
	}

	if exists {
		return r.update(ctx, verification)
	}
	return r.insert(ctx, verification)
}

// insert creates a new verification record.
func (r *VerificationRepository) insert(ctx context.Context, v *domain.Verification) error {
	row := fromVerification(v)

	_, err := r.adapter.Exec(ctx, queryInsertVerification,
		row.ID,
		row.UserID,
		row.Provider,
		row.CredentialType,
		row.Status,
		row.OAuthState,
		row.RedirectURL,
		row.ExpiresAt,
		row.InitiatedAt,
		row.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to insert verification: %w", err)
	}

	return nil
}

// update updates an existing verification record with optimistic locking.
func (r *VerificationRepository) update(ctx context.Context, v *domain.Verification) error {
	row := fromVerification(v)

	result, err := r.adapter.Exec(ctx, queryUpdateVerification,
		row.ID,
		row.Status,
		row.OAuthState,
		row.ProviderUserID,
		row.Username,
		row.Email,
		row.DisplayName,
		row.AvatarURL,
		row.ProfileURL,
		row.RawData,
		row.CredentialID,
		row.FailureReason,
		row.FailureCode,
		row.AuthorizedAt,
		row.CompletedAt,
		row.FailedAt,
		row.ExpiresAt,
		row.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to update verification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Fetch actual version for error message
		var actualVersion int
		dbRow := r.adapter.Executor().QueryRow(ctx, "SELECT version FROM verifications WHERE id = $1", v.AggregateID())
		_ = dbRow.Scan(&actualVersion)

		return eventsourcing.ErrConcurrencyConflict(
			"VerificationRepository.update",
			v.AggregateID(),
			v.Version(),
			actualVersion,
		)
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// FindByID retrieves a verification by ID.
func (r *VerificationRepository) FindByID(ctx context.Context, id string) (*domain.Verification, error) {
	row := &verificationRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, querySelectVerificationByID, id)
	err := dbRow.Scan(
		&row.ID,
		&row.UserID,
		&row.Provider,
		&row.CredentialType,
		&row.Status,
		&row.OAuthState,
		&row.ProviderUserID,
		&row.Username,
		&row.Email,
		&row.DisplayName,
		&row.AvatarURL,
		&row.ProfileURL,
		&row.RawData,
		&row.CredentialID,
		&row.FailureReason,
		&row.FailureCode,
		&row.RedirectURL,
		&row.InitiatedAt,
		&row.AuthorizedAt,
		&row.CompletedAt,
		&row.FailedAt,
		&row.ExpiresAt,
		&row.Version,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrVerificationNotFound("VerificationRepository.FindByID", id)
		}
		return nil, fmt.Errorf("failed to find verification: %w", err)
	}

	return row.toDomain()
}

// FindByOAuthState retrieves a verification by OAuth state.
func (r *VerificationRepository) FindByOAuthState(ctx context.Context, state string) (*domain.Verification, error) {
	row := &verificationRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, querySelectVerificationByOAuthState, state)
	err := dbRow.Scan(
		&row.ID,
		&row.UserID,
		&row.Provider,
		&row.CredentialType,
		&row.Status,
		&row.OAuthState,
		&row.ProviderUserID,
		&row.Username,
		&row.Email,
		&row.DisplayName,
		&row.AvatarURL,
		&row.ProfileURL,
		&row.RawData,
		&row.CredentialID,
		&row.FailureReason,
		&row.FailureCode,
		&row.RedirectURL,
		&row.InitiatedAt,
		&row.AuthorizedAt,
		&row.CompletedAt,
		&row.FailedAt,
		&row.ExpiresAt,
		&row.Version,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrVerificationNotFound("VerificationRepository.FindByOAuthState", "state:"+state)
		}
		return nil, fmt.Errorf("failed to find verification by state: %w", err)
	}

	return row.toDomain()
}

// FindByUserAndProvider retrieves verifications for a user and provider.
func (r *VerificationRepository) FindByUserAndProvider(ctx context.Context, userID string, provider domain.Provider) ([]*domain.Verification, error) {
	rows, err := r.adapter.Select(ctx, querySelectVerificationsByUserAndProvider, userID, string(provider))
	if err != nil {
		return nil, fmt.Errorf("failed to query verifications: %w", err)
	}
	defer rows.Close()

	return r.scanVerifications(rows)
}

// FindActiveByUser retrieves all active verifications for a user.
func (r *VerificationRepository) FindActiveByUser(ctx context.Context, userID string) ([]*domain.Verification, error) {
	rows, err := r.adapter.Select(ctx, querySelectActiveVerificationsByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active verifications: %w", err)
	}
	defer rows.Close()

	return r.scanVerifications(rows)
}

// FindPendingExpired retrieves verifications that have expired but not yet marked.
func (r *VerificationRepository) FindPendingExpired(ctx context.Context, before time.Time) ([]*domain.Verification, error) {
	rows, err := r.adapter.Select(ctx, querySelectPendingExpired, before)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired verifications: %w", err)
	}
	defer rows.Close()

	return r.scanVerifications(rows)
}

// ExistsByID checks if a verification exists.
func (r *VerificationRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	dbRow := r.adapter.Executor().QueryRow(ctx, queryVerificationExists, id)
	err := dbRow.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check verification exists: %w", err)
	}

	return exists, nil
}

// Delete removes a verification.
func (r *VerificationRepository) Delete(ctx context.Context, id string) error {
	result, err := r.adapter.Exec(ctx, queryDeleteVerification, id)
	if err != nil {
		return fmt.Errorf("failed to delete verification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVerificationNotFound("VerificationRepository.Delete", id)
	}

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanVerifications scans rows into domain verifications.
func (r *VerificationRepository) scanVerifications(rows database.Rows) ([]*domain.Verification, error) {
	var verifications []*domain.Verification

	for rows.Next() {
		row := &verificationRow{}
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.Provider,
			&row.CredentialType,
			&row.Status,
			&row.OAuthState,
			&row.ProviderUserID,
			&row.Username,
			&row.Email,
			&row.DisplayName,
			&row.AvatarURL,
			&row.ProfileURL,
			&row.RawData,
			&row.CredentialID,
			&row.FailureReason,
			&row.FailureCode,
			&row.RedirectURL,
			&row.InitiatedAt,
			&row.AuthorizedAt,
			&row.CompletedAt,
			&row.FailedAt,
			&row.ExpiresAt,
			&row.Version,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification row: %w", err)
		}

		v, err := row.toDomain()
		if err != nil {
			return nil, err
		}
		verifications = append(verifications, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating verification rows: %w", err)
	}

	return verifications, nil
}

// isNoRows checks if the error is a "no rows" error.
func isNoRows(err error) bool {
	return err != nil && err.Error() == "no rows in result set"
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.VerificationRepository = (*VerificationRepository)(nil)
