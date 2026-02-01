// Package query contains the query definitions and handlers for the Identity context.
package query

import (
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
