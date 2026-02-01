package query

import "time"

// ============================================================================
// User Views
// ============================================================================

// UserView is the full user read model.
type UserView struct {
	UserID        string             `json:"user_id"`
	Email         string             `json:"email,omitempty"`
	DisplayName   string             `json:"display_name"`
	Status        string             `json:"status"`
	PrimaryDID    string             `json:"primary_did"`
	DIDs          []string           `json:"dids"`
	AuthMethods   []string           `json:"auth_methods"`
	OAuthAccounts []OAuthAccountView `json:"oauth_accounts,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// UserSummaryView is a lightweight user representation for lists.
type UserSummaryView struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserProfileView is a public user profile (excludes sensitive data).
type UserProfileView struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	PrimaryDID  string    `json:"primary_did"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserListView is a paginated list of users.
type UserListView struct {
	Users      []UserSummaryView `json:"users"`
	NextCursor *string           `json:"next_cursor,omitempty"`
	TotalCount int               `json:"total_count"`
}

// ============================================================================
// Session Views
// ============================================================================

// SessionView is the full session read model.
type SessionView struct {
	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
	AuthMethod string    `json:"auth_method"`
	Status     string    `json:"status"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// SessionSummaryView is a lightweight session representation for lists.
type SessionSummaryView struct {
	SessionID  string    `json:"session_id"`
	AuthMethod string    `json:"auth_method"`
	Status     string    `json:"status"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// SessionListView is a paginated list of sessions.
type SessionListView struct {
	Sessions   []SessionSummaryView `json:"sessions"`
	NextCursor *string              `json:"next_cursor,omitempty"`
	TotalCount int                  `json:"total_count"`
}

// SessionCountView is a session count result.
type SessionCountView struct {
	UserID string `json:"user_id"`
	Count  int    `json:"count"`
}

// SessionValidationView is the result of session validation.
type SessionValidationView struct {
	Valid      bool      `json:"valid"`
	SessionID  string    `json:"session_id,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	AuthMethod string    `json:"auth_method,omitempty"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	Reason     string    `json:"reason,omitempty"`
}

// ============================================================================
// DID Views
// ============================================================================

// DIDView represents information about a single DID.
type DIDView struct {
	DID       string    `json:"did"`
	Method    string    `json:"method"`
	IsPrimary bool      `json:"is_primary"`
	AddedAt   time.Time `json:"added_at"`
}

// UserDIDsView is the list of DIDs for a user.
type UserDIDsView struct {
	UserID     string    `json:"user_id"`
	PrimaryDID string    `json:"primary_did"`
	DIDs       []DIDView `json:"dids"`
}

// DIDResolutionView is the result of resolving a DID to a user.
type DIDResolutionView struct {
	DID         string `json:"did"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	IsPrimary   bool   `json:"is_primary"`
}

// ============================================================================
// OAuth Views
// ============================================================================

// OAuthAccountView represents a linked OAuth account.
type OAuthAccountView struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id"`
	Email      string `json:"email,omitempty"`
}

// LinkedOAuthAccountsView is the list of OAuth accounts for a user.
type LinkedOAuthAccountsView struct {
	UserID   string             `json:"user_id"`
	Accounts []OAuthAccountView `json:"accounts"`
}

// OAuthStateView is the OAuth state validation result.
type OAuthStateView struct {
	State       string    `json:"state"`
	Provider    string    `json:"provider"`
	RedirectURL string    `json:"redirect_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	Valid       bool      `json:"valid"`
}
