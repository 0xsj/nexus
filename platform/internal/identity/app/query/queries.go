// Package query contains the query definitions and handlers for the Identity context.
package query

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Query Name Constants
// ============================================================================

const (
	// User queries
	QueryGetUser        = "identity.GetUser"
	QueryGetUserByEmail = "identity.GetUserByEmail"
	QueryGetUserByDID   = "identity.GetUserByDID"
	QueryListUsers      = "identity.ListUsers"
	QuerySearchUsers    = "identity.SearchUsers"
	QueryGetUserProfile = "identity.GetUserProfile"

	// Session queries
	QueryGetSession            = "identity.GetSession"
	QueryListUserSessions      = "identity.ListUserSessions"
	QueryGetActiveSessionCount = "identity.GetActiveSessionCount"
	QueryValidateSession       = "identity.ValidateSession"

	// DID queries
	QueryGetUserDIDs = "identity.GetUserDIDs"
	QueryResolveDID  = "identity.ResolveDID"

	// OAuth queries
	QueryGetLinkedOAuthAccounts = "identity.GetLinkedOAuthAccounts"
	QueryGetOAuthState          = "identity.GetOAuthState"
)

// ============================================================================
// User Queries
// ============================================================================

// GetUser retrieves a user by ID.
type GetUser struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetUser) QueryName() string {
	return QueryGetUser
}

// Validate implements cqrs.Validatable.
func (q GetUser) Validate() error {
	if q.UserID.IsZero() {
		return cqrs.ErrQueryValidation("GetUser.Validate", "user_id is required")
	}
	return nil
}

// GetUserResult is the result for GetUser query.
type GetUserResult struct {
	UserID        string             `json:"user_id"`
	Email         string             `json:"email,omitempty"`
	DisplayName   string             `json:"display_name"`
	Status        string             `json:"status"`
	PrimaryDID    string             `json:"primary_did"`
	DIDs          []string           `json:"dids"`
	AuthMethods   []string           `json:"auth_methods"`
	OAuthAccounts []OAuthAccountInfo `json:"oauth_accounts,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// OAuthAccountInfo represents linked OAuth account information.
type OAuthAccountInfo struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id"`
	Email      string `json:"email,omitempty"`
}

// GetUserByEmail retrieves a user by email address.
type GetUserByEmail struct {
	Email types.Email `json:"email" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetUserByEmail) QueryName() string {
	return QueryGetUserByEmail
}

// Validate implements cqrs.Validatable.
func (q GetUserByEmail) Validate() error {
	if q.Email.IsEmpty() {
		return cqrs.ErrQueryValidation("GetUserByEmail.Validate", "email is required")
	}
	return nil
}

// GetUserByDID retrieves a user by DID.
type GetUserByDID struct {
	DID string `json:"did" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetUserByDID) QueryName() string {
	return QueryGetUserByDID
}

// Validate implements cqrs.Validatable.
func (q GetUserByDID) Validate() error {
	if q.DID == "" {
		return cqrs.ErrQueryValidation("GetUserByDID.Validate", "did is required")
	}
	return nil
}

// ListUsers lists users with pagination and filtering.
type ListUsers struct {
	Status   *string `json:"status" validate:"omitempty,oneof=pending active suspended deleted"`
	PageSize int     `json:"page_size" validate:"omitempty,min=1,max=100"`
	Cursor   *string `json:"cursor" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListUsers) QueryName() string {
	return QueryListUsers
}

// Validate implements cqrs.Validatable.
func (q ListUsers) Validate() error {
	if q.PageSize < 0 {
		return cqrs.ErrQueryValidation("ListUsers.Validate", "page_size must be non-negative")
	}
	if q.PageSize > 100 {
		return cqrs.ErrQueryValidation("ListUsers.Validate", "page_size must be 100 or less")
	}
	return nil
}

// ListUsersResult is the result for ListUsers query.
type ListUsersResult struct {
	Users      []UserSummary `json:"users"`
	NextCursor *string       `json:"next_cursor,omitempty"`
	TotalCount int           `json:"total_count"`
}

// UserSummary is a lightweight user representation for lists.
type UserSummary struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// SearchUsers searches users by display name or email.
type SearchUsers struct {
	Query    string  `json:"query" validate:"required,min=1,max=100"`
	PageSize int     `json:"page_size" validate:"omitempty,min=1,max=100"`
	Cursor   *string `json:"cursor" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q SearchUsers) QueryName() string {
	return QuerySearchUsers
}

// Validate implements cqrs.Validatable.
func (q SearchUsers) Validate() error {
	if q.Query == "" {
		return cqrs.ErrQueryValidation("SearchUsers.Validate", "query is required")
	}
	if len(q.Query) > 100 {
		return cqrs.ErrQueryValidation("SearchUsers.Validate", "query must be 100 characters or less")
	}
	if q.PageSize < 0 {
		return cqrs.ErrQueryValidation("SearchUsers.Validate", "page_size must be non-negative")
	}
	if q.PageSize > 100 {
		return cqrs.ErrQueryValidation("SearchUsers.Validate", "page_size must be 100 or less")
	}
	return nil
}

// GetUserProfile retrieves a user's public profile.
type GetUserProfile struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetUserProfile) QueryName() string {
	return QueryGetUserProfile
}

// Validate implements cqrs.Validatable.
func (q GetUserProfile) Validate() error {
	if q.UserID.IsZero() {
		return cqrs.ErrQueryValidation("GetUserProfile.Validate", "user_id is required")
	}
	return nil
}

// GetUserProfileResult is the result for GetUserProfile query.
type GetUserProfileResult struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	PrimaryDID  string    `json:"primary_did"`
	CreatedAt   time.Time `json:"created_at"`
	// Public profile excludes sensitive data like email
}

// ============================================================================
// Session Queries
// ============================================================================

// GetSession retrieves a session by ID.
type GetSession struct {
	SessionID types.ID `json:"session_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetSession) QueryName() string {
	return QueryGetSession
}

// Validate implements cqrs.Validatable.
func (q GetSession) Validate() error {
	if q.SessionID.IsZero() {
		return cqrs.ErrQueryValidation("GetSession.Validate", "session_id is required")
	}
	return nil
}

// GetSessionResult is the result for GetSession query.
type GetSessionResult struct {
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

// ListUserSessions lists all sessions for a user.
type ListUserSessions struct {
	UserID     types.ID `json:"user_id" validate:"required"`
	ActiveOnly bool     `json:"active_only"`
	PageSize   int      `json:"page_size" validate:"omitempty,min=1,max=100"`
	Cursor     *string  `json:"cursor" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListUserSessions) QueryName() string {
	return QueryListUserSessions
}

// Validate implements cqrs.Validatable.
func (q ListUserSessions) Validate() error {
	if q.UserID.IsZero() {
		return cqrs.ErrQueryValidation("ListUserSessions.Validate", "user_id is required")
	}
	if q.PageSize < 0 {
		return cqrs.ErrQueryValidation("ListUserSessions.Validate", "page_size must be non-negative")
	}
	if q.PageSize > 100 {
		return cqrs.ErrQueryValidation("ListUserSessions.Validate", "page_size must be 100 or less")
	}
	return nil
}

// ListUserSessionsResult is the result for ListUserSessions query.
type ListUserSessionsResult struct {
	Sessions   []SessionSummary `json:"sessions"`
	NextCursor *string          `json:"next_cursor,omitempty"`
	TotalCount int              `json:"total_count"`
}

// SessionSummary is a lightweight session representation for lists.
type SessionSummary struct {
	SessionID  string    `json:"session_id"`
	AuthMethod string    `json:"auth_method"`
	Status     string    `json:"status"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// GetActiveSessionCount returns the count of active sessions for a user.
type GetActiveSessionCount struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetActiveSessionCount) QueryName() string {
	return QueryGetActiveSessionCount
}

// Validate implements cqrs.Validatable.
func (q GetActiveSessionCount) Validate() error {
	if q.UserID.IsZero() {
		return cqrs.ErrQueryValidation("GetActiveSessionCount.Validate", "user_id is required")
	}
	return nil
}

// GetActiveSessionCountResult is the result for GetActiveSessionCount query.
type GetActiveSessionCountResult struct {
	UserID string `json:"user_id"`
	Count  int    `json:"count"`
}

// ValidateSession validates a session token and returns session info if valid.
type ValidateSession struct {
	SessionID types.ID `json:"session_id" validate:"required"`
	Token     string   `json:"token" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q ValidateSession) QueryName() string {
	return QueryValidateSession
}

// Validate implements cqrs.Validatable.
func (q ValidateSession) Validate() error {
	if q.SessionID.IsZero() {
		return cqrs.ErrQueryValidation("ValidateSession.Validate", "session_id is required")
	}
	if q.Token == "" {
		return cqrs.ErrQueryValidation("ValidateSession.Validate", "token is required")
	}
	return nil
}

// ValidateSessionResult is the result for ValidateSession query.
type ValidateSessionResult struct {
	Valid      bool      `json:"valid"`
	SessionID  string    `json:"session_id,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	AuthMethod string    `json:"auth_method,omitempty"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	Reason     string    `json:"reason,omitempty"` // If not valid, explains why
}

// ============================================================================
// DID Queries
// ============================================================================

// GetUserDIDs retrieves all DIDs associated with a user.
type GetUserDIDs struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetUserDIDs) QueryName() string {
	return QueryGetUserDIDs
}

// Validate implements cqrs.Validatable.
func (q GetUserDIDs) Validate() error {
	if q.UserID.IsZero() {
		return cqrs.ErrQueryValidation("GetUserDIDs.Validate", "user_id is required")
	}
	return nil
}

// GetUserDIDsResult is the result for GetUserDIDs query.
type GetUserDIDsResult struct {
	UserID     string    `json:"user_id"`
	PrimaryDID string    `json:"primary_did"`
	DIDs       []DIDInfo `json:"dids"`
}

// DIDInfo represents information about a DID.
type DIDInfo struct {
	DID       string    `json:"did"`
	Method    string    `json:"method"` // e.g., "key", "pkh", "web"
	IsPrimary bool      `json:"is_primary"`
	AddedAt   time.Time `json:"added_at"`
}

// ResolveDID resolves a DID to find the associated user.
type ResolveDID struct {
	DID string `json:"did" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q ResolveDID) QueryName() string {
	return QueryResolveDID
}

// Validate implements cqrs.Validatable.
func (q ResolveDID) Validate() error {
	if q.DID == "" {
		return cqrs.ErrQueryValidation("ResolveDID.Validate", "did is required")
	}
	return nil
}

// ResolveDIDResult is the result for ResolveDID query.
type ResolveDIDResult struct {
	DID         string `json:"did"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	IsPrimary   bool   `json:"is_primary"`
}

// ============================================================================
// OAuth Queries
// ============================================================================

// GetLinkedOAuthAccounts retrieves all OAuth accounts linked to a user.
type GetLinkedOAuthAccounts struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetLinkedOAuthAccounts) QueryName() string {
	return QueryGetLinkedOAuthAccounts
}

// Validate implements cqrs.Validatable.
func (q GetLinkedOAuthAccounts) Validate() error {
	if q.UserID.IsZero() {
		return cqrs.ErrQueryValidation("GetLinkedOAuthAccounts.Validate", "user_id is required")
	}
	return nil
}

// GetLinkedOAuthAccountsResult is the result for GetLinkedOAuthAccounts query.
type GetLinkedOAuthAccountsResult struct {
	UserID   string             `json:"user_id"`
	Accounts []OAuthAccountInfo `json:"accounts"`
}

// GetOAuthState retrieves an OAuth state record for validation.
type GetOAuthState struct {
	State string `json:"state" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetOAuthState) QueryName() string {
	return QueryGetOAuthState
}

// Validate implements cqrs.Validatable.
func (q GetOAuthState) Validate() error {
	if q.State == "" {
		return cqrs.ErrQueryValidation("GetOAuthState.Validate", "state is required")
	}
	return nil
}

// GetOAuthStateResult is the result for GetOAuthState query.
type GetOAuthStateResult struct {
	State       string    `json:"state"`
	Provider    string    `json:"provider"`
	RedirectURL string    `json:"redirect_url"`
	ExpiresAt   time.Time `json:"expires_at"`
	Valid       bool      `json:"valid"`
}
