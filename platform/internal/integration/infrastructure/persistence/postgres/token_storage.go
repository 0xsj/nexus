package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/crypto"
	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================================
// Token Storage Implementation
// ============================================================================

// TokenStorage provides encrypted storage for OAuth tokens.
type TokenStorage struct {
	pool      *pgxpool.Pool
	queries   *generated.Queries
	encryptor *crypto.Encryptor
	version   int32 // Current encryption version
}

// NewTokenStorage creates a new token storage.
func NewTokenStorage(pool *pgxpool.Pool, encryptor *crypto.Encryptor) *TokenStorage {
	return &TokenStorage{
		pool:      pool,
		queries:   generated.New(pool),
		encryptor: encryptor,
		version:   1, // Start with version 1
	}
}

// Store stores encrypted OAuth tokens for an integration.
func (s *TokenStorage) Store(ctx context.Context, integrationID string, tokens *domain.OAuthTokens) error {
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	// Parse integration ID
	id, err := uuid.Parse(integrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	// Encrypt access token
	accessTokenEncrypted, err := s.encryptor.Encrypt(tokens.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// Encrypt refresh token (if present)
	var refreshTokenEncrypted []byte
	if tokens.RefreshToken != "" {
		refreshTokenEncrypted, err = s.encryptor.Encrypt(tokens.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	// Calculate expiration time
	var expiresAt pgtype.Timestamptz
	if tokens.ExpiresIn > 0 {
		expiresAt = pgtype.Timestamptz{
			Time:  time.Now().UTC().Add(time.Duration(tokens.ExpiresIn) * time.Second),
			Valid: true,
		}
	}

	// Store in database
	err = s.queries.StoreTokens(ctx, generated.StoreTokensParams{
		IntegrationID:          id,
		AccessTokenEncrypted:   accessTokenEncrypted,
		RefreshTokenEncrypted:  refreshTokenEncrypted,
		TokenType:              tokens.TokenType,
		ExpiresAt:              expiresAt,
		Scopes:                 tokens.Scopes,
		EncryptionVersion:      s.version,
	})
	if err != nil {
		return fmt.Errorf("failed to store tokens: %w", err)
	}

	return nil
}

// Get retrieves and decrypts OAuth tokens for an integration.
func (s *TokenStorage) Get(ctx context.Context, integrationID string) (*domain.OAuthTokens, error) {
	// Parse integration ID
	id, err := uuid.Parse(integrationID)
	if err != nil {
		return nil, fmt.Errorf("invalid integration ID: %w", err)
	}

	// Fetch from database
	row, err := s.queries.GetTokens(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tokens not found for integration %s", integrationID)
		}
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	// Decrypt access token
	accessToken, err := s.encryptor.Decrypt(row.AccessTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	// Decrypt refresh token (if present)
	var refreshToken string
	if len(row.RefreshTokenEncrypted) > 0 {
		refreshToken, err = s.encryptor.Decrypt(row.RefreshTokenEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
		}
	}

	// Calculate expires_in from expires_at
	var expiresIn int64
	if row.ExpiresAt.Valid {
		expiresIn = int64(time.Until(row.ExpiresAt.Time).Seconds())
		if expiresIn < 0 {
			expiresIn = 0 // Token already expired
		}
	}

	return &domain.OAuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    row.TokenType,
		ExpiresIn:    expiresIn,
		Scopes:       row.Scopes,
	}, nil
}

// Delete deletes OAuth tokens for an integration.
func (s *TokenStorage) Delete(ctx context.Context, integrationID string) error {
	// Parse integration ID
	id, err := uuid.Parse(integrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	// Delete from database
	err = s.queries.DeleteTokens(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	return nil
}

// Update updates OAuth tokens for an integration.
// This is typically used after refreshing tokens.
func (s *TokenStorage) Update(ctx context.Context, integrationID string, tokens *domain.OAuthTokens) error {
	// Update is the same as Store (upsert)
	return s.Store(ctx, integrationID, tokens)
}

// ============================================================================
// Additional Operations
// ============================================================================

// Exists checks if tokens exist for an integration.
func (s *TokenStorage) Exists(ctx context.Context, integrationID string) (bool, error) {
	id, err := uuid.Parse(integrationID)
	if err != nil {
		return false, fmt.Errorf("invalid integration ID: %w", err)
	}

	exists, err := s.queries.TokensExist(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check tokens existence: %w", err)
	}

	return exists, nil
}

// ListExpiredTokens returns a list of expired tokens for cleanup/refresh.
func (s *TokenStorage) ListExpiredTokens(ctx context.Context, limit int32) ([]string, error) {
	rows, err := s.queries.ListExpiredTokens(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired tokens: %w", err)
	}

	integrationIDs := make([]string, len(rows))
	for i, row := range rows {
		integrationIDs[i] = row.IntegrationID.String()
	}

	return integrationIDs, nil
}

// CountByEncryptionVersion returns the count of tokens using a specific encryption version.
// Useful for tracking progress during key rotation.
func (s *TokenStorage) CountByEncryptionVersion(ctx context.Context, version int32) (int, error) {
	count, err := s.queries.CountTokensByEncryptionVersion(ctx, version)
	if err != nil {
		return 0, fmt.Errorf("failed to count tokens by version: %w", err)
	}

	return int(count), nil
}
