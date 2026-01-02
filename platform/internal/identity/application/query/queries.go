package query

import (
	"github.com/0xsj/nexus/platform/internal/identity/domain"
)

// ============================================================================
// User Queries
// ============================================================================

// GetUser retrieves a user by ID.
type GetUser struct {
	UserID string
}

// GetUserByDID retrieves a user by their primary DID.
type GetUserByDID struct {
	DID string
}

// GetUserByWallet retrieves a user by a linked wallet address.
type GetUserByWallet struct {
	Address string
	Chain   domain.Chain
}

// GetUserByEmail retrieves a user by a linked email.
type GetUserByEmail struct {
	Email string
}

// ListUsers retrieves a paginated list of users.
type ListUsers struct {
	Limit     int
	Offset    int
	Status    *domain.UserStatus
	SortBy    string // "created_at", "updated_at", "last_login_at"
	SortOrder string // "asc", "desc"
}

// GetUserStats retrieves statistics for a user.
type GetUserStats struct {
	UserID string
}

// GetPublicProfile retrieves the public profile for a user.
type GetPublicProfile struct {
	DID string
}

// ============================================================================
// Session Queries
// ============================================================================

// GetSession retrieves a session by ID.
type GetSession struct {
	UserID    string
	SessionID string
}

// ListUserSessions retrieves all sessions for a user.
type ListUserSessions struct {
	UserID         string
	CurrentSession string // To mark which is current
	ActiveOnly     bool
	Limit          int
	Offset         int
}

// ValidateSession validates a session is active and not expired.
type ValidateSession struct {
	SessionID string
	TokenHash string
}

// CountUserSessions counts sessions for a user.
type CountUserSessions struct {
	UserID     string
	ActiveOnly bool
}

// ============================================================================
// Connection Queries
// ============================================================================

// GetConnection retrieves a connection by ID.
type GetConnection struct {
	UserID       string
	ConnectionID string
}

// GetConnectionByProvider retrieves a connection by provider.
type GetConnectionByProvider struct {
	UserID   string
	Provider domain.OAuthProvider
}

// ListUserConnections retrieves all connections for a user.
type ListUserConnections struct {
	UserID     string
	ActiveOnly bool
	Limit      int
	Offset     int
}

// CheckConnectionExists checks if a user has a connection to a provider.
type CheckConnectionExists struct {
	UserID   string
	Provider domain.OAuthProvider
}

// ============================================================================
// API Key Queries
// ============================================================================

// GetAPIKey retrieves an API key by ID.
type GetAPIKey struct {
	UserID string
	KeyID  string
}

// ListUserAPIKeys retrieves all API keys for a user.
type ListUserAPIKeys struct {
	UserID     string
	ActiveOnly bool
	Limit      int
	Offset     int
}

// ValidateAPIKey validates an API key.
type ValidateAPIKey struct {
	RawKey string
}

// CountUserAPIKeys counts API keys for a user.
type CountUserAPIKeys struct {
	UserID     string
	ActiveOnly bool
}

// ============================================================================
// Token Queries
// ============================================================================

// ValidateAccessToken validates an access token.
type ValidateAccessToken struct {
	Token string
}

// ValidateRefreshToken validates a refresh token.
type ValidateRefreshToken struct {
	Token string
}

// ============================================================================
// Search Queries
// ============================================================================

// SearchUsers searches for users by various criteria.
type SearchUsers struct {
	Query     string // Search in DID, email, wallet address
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ============================================================================
// Existence Queries
// ============================================================================

// CheckUserExists checks if a user exists by various identifiers.
type CheckUserExists struct {
	UserID *string
	DID    *string
	Email  *string
	Wallet *WalletIdentifier
}

// WalletIdentifier identifies a wallet.
type WalletIdentifier struct {
	Address string
	Chain   domain.Chain
}

// ============================================================================
// Query Options
// ============================================================================

// ListOptions contains common list query options.
type ListOptions struct {
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// DefaultListOptions returns default list options.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithLimit sets the limit.
func (o ListOptions) WithLimit(limit int) ListOptions {
	o.Limit = limit
	return o
}

// WithOffset sets the offset.
func (o ListOptions) WithOffset(offset int) ListOptions {
	o.Offset = offset
	return o
}

// WithSort sets the sort options.
func (o ListOptions) WithSort(sortBy, sortOrder string) ListOptions {
	o.SortBy = sortBy
	o.SortOrder = sortOrder
	return o
}

// Validate validates and normalizes list options.
func (o *ListOptions) Validate() {
	if o.Limit <= 0 {
		o.Limit = 20
	}
	if o.Limit > 100 {
		o.Limit = 100
	}
	if o.Offset < 0 {
		o.Offset = 0
	}
	if o.SortOrder != "asc" && o.SortOrder != "desc" {
		o.SortOrder = "desc"
	}
}
