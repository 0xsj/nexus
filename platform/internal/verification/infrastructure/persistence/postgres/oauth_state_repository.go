package postgres

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
)

// ============================================================================
// OAuth State Repository
// ============================================================================

// OAuthStateRepository implements domain.OAuthStateRepository using PostgreSQL.
type OAuthStateRepository struct {
	adapter *pgadapter.BaseAdapter
}

// NewOAuthStateRepository creates a new OAuthStateRepository.
func NewOAuthStateRepository(db *pgadapter.DB) *OAuthStateRepository {
	return &OAuthStateRepository{
		adapter: pgadapter.NewBaseAdapter(db),
	}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save persists an OAuth state.
func (r *OAuthStateRepository) Save(ctx context.Context, state *domain.OAuthState) error {
	_, err := r.adapter.Exec(ctx, queryInsertOAuthState,
		state.Value,
		state.VerificationID, // Now uses the verification ID from the state
		state.UserID,
		string(state.Provider),
		nil, // code_verifier - for PKCE (optional)
		nil, // nonce (optional)
		toNullString(state.RedirectURL),
		state.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert OAuth state: %w", err)
	}

	return nil
}

// SaveWithVerification persists an OAuth state with a verification ID.
func (r *OAuthStateRepository) SaveWithVerification(ctx context.Context, verificationID string, state *domain.OAuthState, codeVerifier, nonce *string) error {
	var cv, n interface{}
	if codeVerifier != nil {
		cv = *codeVerifier
	}
	if nonce != nil {
		n = *nonce
	}

	_, err := r.adapter.Exec(ctx, queryInsertOAuthState,
		state.Value,
		verificationID,
		state.UserID,
		string(state.Provider),
		cv,
		n,
		toNullString(state.RedirectURL),
		state.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert OAuth state: %w", err)
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// FindByValue retrieves an OAuth state by its value.
func (r *OAuthStateRepository) FindByValue(ctx context.Context, value string) (*domain.OAuthState, error) {
	row := &oauthStateRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, querySelectOAuthStateByValue, value)
	err := dbRow.Scan(
		&row.State,
		&row.VerificationID,
		&row.UserID,
		&row.Provider,
		&row.CodeVerifier,
		&row.Nonce,
		&row.RedirectURL,
		&row.CreatedAt,
		&row.ExpiresAt,
		&row.UsedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrOAuthStateMismatch("OAuthStateRepository.FindByValue")
		}
		return nil, fmt.Errorf("failed to find OAuth state: %w", err)
	}

	// Check if already used
	if row.UsedAt.Valid {
		return nil, domain.ErrOAuthStateMismatch("OAuthStateRepository.FindByValue")
	}

	return row.toDomain(), nil
}

// MarkUsed marks an OAuth state as used.
func (r *OAuthStateRepository) MarkUsed(ctx context.Context, value string) error {
	result, err := r.adapter.Exec(ctx, queryMarkOAuthStateUsed, value)
	if err != nil {
		return fmt.Errorf("failed to mark OAuth state as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrOAuthStateMismatch("OAuthStateRepository.MarkUsed")
	}

	return nil
}

// Delete removes an OAuth state.
func (r *OAuthStateRepository) Delete(ctx context.Context, value string) error {
	_, err := r.adapter.Exec(ctx, queryDeleteOAuthState, value)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth state: %w", err)
	}

	return nil
}

// DeleteExpired removes all expired OAuth states.
func (r *OAuthStateRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.adapter.Exec(ctx, queryDeleteExpiredOAuthStates)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired OAuth states: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// ExistsByValue checks if an OAuth state exists and hasn't been used.
func (r *OAuthStateRepository) ExistsByValue(ctx context.Context, value string) (bool, error) {
	var exists bool
	dbRow := r.adapter.Executor().QueryRow(ctx, queryOAuthStateExists, value)
	err := dbRow.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check OAuth state exists: %w", err)
	}

	return exists, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.OAuthStateRepository = (*OAuthStateRepository)(nil)
