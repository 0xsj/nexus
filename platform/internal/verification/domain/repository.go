package domain

import (
	"context"
	"time"
)

// ============================================================================
// Verification Repository
// ============================================================================

// VerificationRepository manages Verification aggregate persistence.
type VerificationRepository interface {
	// Save persists a verification.
	Save(ctx context.Context, verification *Verification) error

	// FindByID retrieves a verification by ID.
	FindByID(ctx context.Context, id string) (*Verification, error)

	// FindByOAuthState retrieves a verification by OAuth state.
	FindByOAuthState(ctx context.Context, state string) (*Verification, error)

	// FindByUserAndProvider retrieves active verifications for a user and provider.
	FindByUserAndProvider(ctx context.Context, userID string, provider Provider) ([]*Verification, error)

	// FindActiveByUser retrieves all active verifications for a user.
	FindActiveByUser(ctx context.Context, userID string) ([]*Verification, error)

	// FindPendingExpired retrieves verifications that have expired but not yet marked.
	FindPendingExpired(ctx context.Context, before time.Time) ([]*Verification, error)

	// ExistsByID checks if a verification exists.
	ExistsByID(ctx context.Context, id string) (bool, error)

	// Delete removes a verification.
	Delete(ctx context.Context, id string) error
}

// ============================================================================
// OAuth State Repository
// ============================================================================

// OAuthStateRepository manages OAuth state persistence for CSRF protection.
type OAuthStateRepository interface {
	// Save persists an OAuth state.
	Save(ctx context.Context, state *OAuthState) error

	// FindByValue retrieves an OAuth state by its value.
	FindByValue(ctx context.Context, value string) (*OAuthState, error)

	// Delete removes an OAuth state.
	Delete(ctx context.Context, value string) error

	// DeleteExpired removes all expired OAuth states.
	DeleteExpired(ctx context.Context) (int64, error)

	// ExistsByValue checks if an OAuth state exists.
	ExistsByValue(ctx context.Context, value string) (bool, error)
}

// ============================================================================
// Provider Token Repository
// ============================================================================

// ProviderTokenRepository manages OAuth tokens for providers.
// Tokens are stored encrypted and associated with a user and provider.
type ProviderTokenRepository interface {
	// Save persists provider tokens.
	Save(ctx context.Context, userID string, provider Provider, tokens *OAuthTokens) error

	// FindByUserAndProvider retrieves tokens for a user and provider.
	FindByUserAndProvider(ctx context.Context, userID string, provider Provider) (*OAuthTokens, error)

	// Delete removes tokens for a user and provider.
	Delete(ctx context.Context, userID string, provider Provider) error

	// DeleteByUser removes all tokens for a user.
	DeleteByUser(ctx context.Context, userID string) error

	// ExistsByUserAndProvider checks if tokens exist for a user and provider.
	ExistsByUserAndProvider(ctx context.Context, userID string, provider Provider) (bool, error)
}

// ============================================================================
// Read Models / Views
// ============================================================================

// VerificationView is a read model for verification queries.
type VerificationView struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Provider       string     `json:"provider"`
	ProviderName   string     `json:"provider_name"`
	CredentialType string     `json:"credential_type"`
	Status         string     `json:"status"`
	ProviderUserID string     `json:"provider_user_id,omitempty"`
	Username       string     `json:"username,omitempty"`
	CredentialID   string     `json:"credential_id,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
	FailureCode    string     `json:"failure_code,omitempty"`
	InitiatedAt    time.Time  `json:"initiated_at"`
	AuthorizedAt   *time.Time `json:"authorized_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	FailedAt       *time.Time `json:"failed_at,omitempty"`
	ExpiresAt      time.Time  `json:"expires_at"`
}

// VerificationSummary is a lightweight read model for lists.
type VerificationSummary struct {
	ID           string     `json:"id"`
	Provider     string     `json:"provider"`
	ProviderName string     `json:"provider_name"`
	Status       string     `json:"status"`
	InitiatedAt  time.Time  `json:"initiated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// ProviderConnectionView shows a user's connection status with a provider.
type ProviderConnectionView struct {
	Provider       string     `json:"provider"`
	ProviderName   string     `json:"provider_name"`
	Connected      bool       `json:"connected"`
	ProviderUserID string     `json:"provider_user_id,omitempty"`
	Username       string     `json:"username,omitempty"`
	CredentialID   string     `json:"credential_id,omitempty"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
}

// ============================================================================
// Query Filters
// ============================================================================

// VerificationFilter defines filters for querying verifications.
type VerificationFilter struct {
	UserID         string
	Provider       Provider
	Status         VerificationStatus
	CredentialType CredentialType
	FromDate       *time.Time
	ToDate         *time.Time
}

// VerificationListOptions defines pagination and sorting options.
type VerificationListOptions struct {
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

// DefaultListOptions returns default list options.
func DefaultListOptions() VerificationListOptions {
	return VerificationListOptions{
		Limit:    20,
		Offset:   0,
		SortBy:   "initiated_at",
		SortDesc: true,
	}
}

// ============================================================================
// Read Repository
// ============================================================================

// VerificationReadRepository provides read-only access to verification data.
type VerificationReadRepository interface {
	// GetByID retrieves a verification view by ID.
	GetByID(ctx context.Context, id string) (*VerificationView, error)

	// GetByUser retrieves all verifications for a user.
	GetByUser(ctx context.Context, userID string, opts VerificationListOptions) ([]*VerificationSummary, int, error)

	// GetByUserAndProvider retrieves verifications for a user and provider.
	GetByUserAndProvider(ctx context.Context, userID string, provider Provider) ([]*VerificationView, error)

	// GetLatestByUserAndProvider retrieves the most recent verification.
	GetLatestByUserAndProvider(ctx context.Context, userID string, provider Provider) (*VerificationView, error)

	// GetProviderConnections retrieves all provider connection statuses for a user.
	GetProviderConnections(ctx context.Context, userID string) ([]*ProviderConnectionView, error)

	// List retrieves verifications with filters and pagination.
	List(ctx context.Context, filter VerificationFilter, opts VerificationListOptions) ([]*VerificationView, int, error)

	// CountByUser returns the count of verifications for a user.
	CountByUser(ctx context.Context, userID string) (int, error)

	// CountByStatus returns the count of verifications by status.
	CountByStatus(ctx context.Context, status VerificationStatus) (int, error)
}
