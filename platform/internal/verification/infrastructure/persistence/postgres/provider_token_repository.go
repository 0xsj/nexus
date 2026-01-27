package postgres

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
)

// ============================================================================
// Provider Token Repository
// ============================================================================

// ProviderTokenRepository implements domain.ProviderTokenRepository using PostgreSQL.
type ProviderTokenRepository struct {
	adapter *pgadapter.BaseAdapter
}

// NewProviderTokenRepository creates a new ProviderTokenRepository.
func NewProviderTokenRepository(db *pgadapter.DB) *ProviderTokenRepository {
	return &ProviderTokenRepository{
		adapter: pgadapter.NewBaseAdapter(db),
	}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save persists provider tokens (upsert).
func (r *ProviderTokenRepository) Save(ctx context.Context, userID string, provider domain.Provider, tokens *domain.OAuthTokens) error {
	return r.SaveWithProviderUserID(ctx, userID, provider, tokens, "")
}

// SaveWithProviderUserID persists provider tokens with optional provider user ID.
func (r *ProviderTokenRepository) SaveWithProviderUserID(ctx context.Context, userID string, provider domain.Provider, tokens *domain.OAuthTokens, providerUserID string) error {
	row := fromOAuthTokens(userID, provider, tokens, providerUserID)

	_, err := r.adapter.Exec(ctx, queryUpsertProviderTokens,
		row.UserID,
		row.Provider,
		row.AccessToken,
		row.RefreshToken,
		row.TokenType,
		row.Scopes,
		row.ExpiresAt,
		row.ProviderUserID,
	)
	if err != nil {
		return fmt.Errorf("failed to save provider tokens: %w", err)
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// FindByUserAndProvider retrieves tokens for a user and provider.
func (r *ProviderTokenRepository) FindByUserAndProvider(ctx context.Context, userID string, provider domain.Provider) (*domain.OAuthTokens, error) {
	row := &providerTokensRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, querySelectProviderTokens, userID, string(provider))
	err := dbRow.Scan(
		&row.UserID,
		&row.Provider,
		&row.AccessToken,
		&row.RefreshToken,
		&row.TokenType,
		&row.Scopes,
		&row.ExpiresAt,
		&row.ProviderUserID,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.LastUsedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrProviderTokenInvalid("ProviderTokenRepository.FindByUserAndProvider", string(provider))
		}
		return nil, fmt.Errorf("failed to find provider tokens: %w", err)
	}

	// Update last used timestamp
	_, _ = r.adapter.Exec(ctx, queryUpdateProviderTokensLastUsed, userID, string(provider))

	return row.toDomain(), nil
}

// ============================================================================
// Delete Operations
// ============================================================================

// Delete removes tokens for a user and provider.
func (r *ProviderTokenRepository) Delete(ctx context.Context, userID string, provider domain.Provider) error {
	result, err := r.adapter.Exec(ctx, queryDeleteProviderTokens, userID, string(provider))
	if err != nil {
		return fmt.Errorf("failed to delete provider tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrProviderTokenInvalid("ProviderTokenRepository.Delete", string(provider))
	}

	return nil
}

// DeleteByUser removes all tokens for a user.
func (r *ProviderTokenRepository) DeleteByUser(ctx context.Context, userID string) error {
	_, err := r.adapter.Exec(ctx, queryDeleteProviderTokensByUser, userID)
	if err != nil {
		return fmt.Errorf("failed to delete provider tokens by user: %w", err)
	}

	return nil
}

// ExistsByUserAndProvider checks if tokens exist for a user and provider.
func (r *ProviderTokenRepository) ExistsByUserAndProvider(ctx context.Context, userID string, provider domain.Provider) (bool, error) {
	var exists bool
	dbRow := r.adapter.Executor().QueryRow(ctx, queryProviderTokensExist, userID, string(provider))
	err := dbRow.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check provider tokens exist: %w", err)
	}

	return exists, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.ProviderTokenRepository = (*ProviderTokenRepository)(nil)
