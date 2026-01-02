package query

import (
	"time"
)

// ============================================================================
// User Views
// ============================================================================

// UserView is the read model for a user.
type UserView struct {
	ID              string           `json:"id"`
	PrimaryDID      string           `json:"primary_did"`
	Status          string           `json:"status"`
	Wallets         []WalletView     `json:"wallets,omitempty"`
	LinkedEmail     *EmailView       `json:"linked_email,omitempty"`
	Connections     []ConnectionView `json:"connections,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	LastLoginAt     *time.Time       `json:"last_login_at,omitempty"`
	LastLoginMethod *string          `json:"last_login_method,omitempty"`
}

// UserSummaryView is a minimal user view for lists.
type UserSummaryView struct {
	ID          string     `json:"id"`
	PrimaryDID  string     `json:"primary_did"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// ============================================================================
// Wallet Views
// ============================================================================

// WalletView is the read model for a linked wallet.
type WalletView struct {
	Address  string    `json:"address"`
	Chain    string    `json:"chain"`
	DID      string    `json:"did"`
	LinkedAt time.Time `json:"linked_at"`
}

// ============================================================================
// Email Views
// ============================================================================

// EmailView is the read model for a linked email.
type EmailView struct {
	Email      string     `json:"email"`
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	LinkedAt   time.Time  `json:"linked_at"`
}

// ============================================================================
// Connection Views
// ============================================================================

// ConnectionView is the read model for an OAuth connection.
type ConnectionView struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	Provider     string     `json:"provider"`
	ProviderName string     `json:"provider_name"`
	Username     string     `json:"username,omitempty"`
	DisplayName  string     `json:"display_name,omitempty"`
	Email        string     `json:"email,omitempty"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	ProfileURL   string     `json:"profile_url,omitempty"`
	Status       string     `json:"status"`
	ConnectedAt  time.Time  `json:"connected_at"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

// ConnectionSummaryView is a minimal connection view.
type ConnectionSummaryView struct {
	ID          string    `json:"id"`
	Provider    string    `json:"provider"`
	Username    string    `json:"username,omitempty"`
	Status      string    `json:"status"`
	ConnectedAt time.Time `json:"connected_at"`
}

// ============================================================================
// Session Views
// ============================================================================

// SessionView is the read model for a session.
type SessionView struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Method     string    `json:"method"`
	Status     string    `json:"status"`
	UserAgent  string    `json:"user_agent,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	Device     string    `json:"device,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	IsCurrent  bool      `json:"is_current"`
}

// SessionSummaryView is a minimal session view for lists.
type SessionSummaryView struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	Status     string    `json:"status"`
	Device     string    `json:"device,omitempty"`
	LastSeenAt time.Time `json:"last_seen_at"`
	IsCurrent  bool      `json:"is_current"`
}

// ============================================================================
// API Key Views
// ============================================================================

// APIKeyView is the read model for an API key.
type APIKeyView struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Prefix      string     `json:"prefix"`
	Scopes      []string   `json:"scopes"`
	Status      string     `json:"status"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UsageCount  int64      `json:"usage_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

// APIKeySummaryView is a minimal API key view for lists.
type APIKeySummaryView struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// ============================================================================
// List Views
// ============================================================================

// UserListView is a paginated list of users.
type UserListView struct {
	Users   []UserSummaryView `json:"users"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasMore bool              `json:"has_more"`
}

// SessionListView is a paginated list of sessions.
type SessionListView struct {
	Sessions []SessionSummaryView `json:"sessions"`
	Total    int                  `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
	HasMore  bool                 `json:"has_more"`
}

// ConnectionListView is a paginated list of connections.
type ConnectionListView struct {
	Connections []ConnectionSummaryView `json:"connections"`
	Total       int                     `json:"total"`
	Limit       int                     `json:"limit"`
	Offset      int                     `json:"offset"`
	HasMore     bool                    `json:"has_more"`
}

// APIKeyListView is a paginated list of API keys.
type APIKeyListView struct {
	APIKeys []APIKeySummaryView `json:"api_keys"`
	Total   int                 `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasMore bool                `json:"has_more"`
}

// ============================================================================
// Validation Views
// ============================================================================

// TokenValidationView is returned when validating a token.
type TokenValidationView struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	DID       string    `json:"did,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	Scopes    []string  `json:"scopes,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// APIKeyValidationView is returned when validating an API key.
type APIKeyValidationView struct {
	Valid   bool     `json:"valid"`
	UserID  string   `json:"user_id,omitempty"`
	DID     string   `json:"did,omitempty"`
	KeyID   string   `json:"key_id,omitempty"`
	KeyName string   `json:"key_name,omitempty"`
	Scopes  []string `json:"scopes,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// ============================================================================
// Stats Views
// ============================================================================

// UserStatsView contains statistics about a user's identity.
type UserStatsView struct {
	UserID           string `json:"user_id"`
	WalletCount      int    `json:"wallet_count"`
	ConnectionCount  int    `json:"connection_count"`
	ActiveSessions   int    `json:"active_sessions"`
	ActiveAPIKeys    int    `json:"active_api_keys"`
	TotalAPIKeyUsage int64  `json:"total_api_key_usage"`
}

// ============================================================================
// Profile View (Public)
// ============================================================================

// PublicProfileView is the public-facing profile view.
// Only contains information the user has chosen to share.
type PublicProfileView struct {
	DID         string             `json:"did"`
	DisplayName string             `json:"display_name,omitempty"`
	AvatarURL   string             `json:"avatar_url,omitempty"`
	Wallets     []PublicWallet     `json:"wallets,omitempty"`
	Connections []PublicConnection `json:"connections,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// PublicWallet is a public wallet view.
type PublicWallet struct {
	Chain string `json:"chain"`
	DID   string `json:"did"`
	// Address intentionally omitted for privacy
}

// PublicConnection is a public connection view.
type PublicConnection struct {
	Provider    string    `json:"provider"`
	Username    string    `json:"username,omitempty"`
	ProfileURL  string    `json:"profile_url,omitempty"`
	ConnectedAt time.Time `json:"connected_at"`
}
