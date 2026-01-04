package query

import (
	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Query Types
// ============================================================================

const (
	TypeGetUser          = "identity.get_user"
	TypeGetUserByDID     = "identity.get_user_by_did"
	TypeGetUserByWallet  = "identity.get_user_by_wallet"
	TypeGetUserByEmail   = "identity.get_user_by_email"
	TypeListUsers        = "identity.list_users"
	TypeGetUserStats     = "identity.get_user_stats"
	TypeGetPublicProfile = "identity.get_public_profile"
	TypeCheckUserExists  = "identity.check_user_exists"
	TypeGetSession       = "identity.get_session"
	TypeListUserSessions = "identity.list_user_sessions"
	TypeValidateSession  = "identity.validate_session"
	TypeGetAPIKey        = "identity.get_api_key"
	TypeListUserAPIKeys  = "identity.list_user_api_keys"
	TypeValidateAPIKey   = "identity.validate_api_key"
	TypeValidateToken    = "identity.validate_token"
)

// ============================================================================
// User Queries
// ============================================================================

// GetUser retrieves a user by ID.
type GetUser struct {
	UserID string `json:"user_id"`
}

// QueryName returns the query name.
func (q *GetUser) QueryName() string {
	return TypeGetUser
}

// GetUserByDID retrieves a user by DID.
type GetUserByDID struct {
	DID string `json:"did"`
}

// QueryName returns the query name.
func (q *GetUserByDID) QueryName() string {
	return TypeGetUserByDID
}

// GetUserByWallet retrieves a user by wallet address.
type GetUserByWallet struct {
	Address string       `json:"address"`
	Chain   domain.Chain `json:"chain"`
}

// QueryName returns the query name.
func (q *GetUserByWallet) QueryName() string {
	return TypeGetUserByWallet
}

// GetUserByEmail retrieves a user by email.
type GetUserByEmail struct {
	Email string `json:"email"`
}

// QueryName returns the query name.
func (q *GetUserByEmail) QueryName() string {
	return TypeGetUserByEmail
}

// ListUsers retrieves a paginated list of users.
type ListUsers struct {
	Status    *domain.UserStatus `json:"status,omitempty"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
	SortBy    string             `json:"sort_by,omitempty"`
	SortOrder string             `json:"sort_order,omitempty"`
}

// QueryName returns the query name.
func (q *ListUsers) QueryName() string {
	return TypeListUsers
}

// NewListUsers creates a new ListUsers query with defaults.
func NewListUsers() *ListUsers {
	return &ListUsers{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// WithStatus filters by status.
func (q *ListUsers) WithStatus(status domain.UserStatus) *ListUsers {
	q.Status = &status
	return q
}

// WithLimit sets the limit.
func (q *ListUsers) WithLimit(limit int) *ListUsers {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListUsers) WithOffset(offset int) *ListUsers {
	q.Offset = offset
	return q
}

// WithSort sets the sort field and order.
func (q *ListUsers) WithSort(sortBy, sortOrder string) *ListUsers {
	q.SortBy = sortBy
	q.SortOrder = sortOrder
	return q
}

// GetUserStats retrieves statistics for a user.
type GetUserStats struct {
	UserID string `json:"user_id"`
}

// QueryName returns the query name.
func (q *GetUserStats) QueryName() string {
	return TypeGetUserStats
}

// GetPublicProfile retrieves the public profile for a DID.
type GetPublicProfile struct {
	DID string `json:"did"`
}

// QueryName returns the query name.
func (q *GetPublicProfile) QueryName() string {
	return TypeGetPublicProfile
}

// CheckUserExists checks if a user exists by various identifiers.
type CheckUserExists struct {
	Email  *string           `json:"email,omitempty"`
	DID    *string           `json:"did,omitempty"`
	Wallet *WalletIdentifier `json:"wallet,omitempty"`
}

// QueryName returns the query name.
func (q *CheckUserExists) QueryName() string {
	return TypeCheckUserExists
}

// WalletIdentifier identifies a wallet.
type WalletIdentifier struct {
	Address string       `json:"address"`
	Chain   domain.Chain `json:"chain"`
}

// ============================================================================
// Session Queries
// ============================================================================

// GetSession retrieves a session by ID.
type GetSession struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// QueryName returns the query name.
func (q *GetSession) QueryName() string {
	return TypeGetSession
}

// ListUserSessions retrieves sessions for a user.
type ListUserSessions struct {
	UserID         string `json:"user_id"`
	CurrentSession string `json:"current_session,omitempty"`
	ActiveOnly     bool   `json:"active_only"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
}

// QueryName returns the query name.
func (q *ListUserSessions) QueryName() string {
	return TypeListUserSessions
}

// NewListUserSessions creates a new ListUserSessions query with defaults.
func NewListUserSessions(userID string) *ListUserSessions {
	return &ListUserSessions{
		UserID:     userID,
		ActiveOnly: false,
		Limit:      20,
		Offset:     0,
	}
}

// WithCurrentSession sets the current session for comparison.
func (q *ListUserSessions) WithCurrentSession(sessionID string) *ListUserSessions {
	q.CurrentSession = sessionID
	return q
}

// WithActiveOnly filters to active sessions only.
func (q *ListUserSessions) WithActiveOnly(activeOnly bool) *ListUserSessions {
	q.ActiveOnly = activeOnly
	return q
}

// WithLimit sets the limit.
func (q *ListUserSessions) WithLimit(limit int) *ListUserSessions {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListUserSessions) WithOffset(offset int) *ListUserSessions {
	q.Offset = offset
	return q
}

// ValidateSession validates a session.
type ValidateSession struct {
	SessionID string `json:"session_id"`
	TokenHash string `json:"token_hash"`
}

// QueryName returns the query name.
func (q *ValidateSession) QueryName() string {
	return TypeValidateSession
}

// ============================================================================
// API Key Queries
// ============================================================================

// GetAPIKey retrieves an API key by ID.
type GetAPIKey struct {
	UserID string `json:"user_id"`
	KeyID  string `json:"key_id"`
}

// QueryName returns the query name.
func (q *GetAPIKey) QueryName() string {
	return TypeGetAPIKey
}

// ListUserAPIKeys retrieves API keys for a user.
type ListUserAPIKeys struct {
	UserID     string `json:"user_id"`
	ActiveOnly bool   `json:"active_only"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// QueryName returns the query name.
func (q *ListUserAPIKeys) QueryName() string {
	return TypeListUserAPIKeys
}

// NewListUserAPIKeys creates a new ListUserAPIKeys query with defaults.
func NewListUserAPIKeys(userID string) *ListUserAPIKeys {
	return &ListUserAPIKeys{
		UserID:     userID,
		ActiveOnly: false,
		Limit:      20,
		Offset:     0,
	}
}

// WithActiveOnly filters to active keys only.
func (q *ListUserAPIKeys) WithActiveOnly(activeOnly bool) *ListUserAPIKeys {
	q.ActiveOnly = activeOnly
	return q
}

// WithLimit sets the limit.
func (q *ListUserAPIKeys) WithLimit(limit int) *ListUserAPIKeys {
	q.Limit = limit
	return q
}

// WithOffset sets the offset.
func (q *ListUserAPIKeys) WithOffset(offset int) *ListUserAPIKeys {
	q.Offset = offset
	return q
}

// ValidateAPIKey validates an API key.
type ValidateAPIKey struct {
	RawKey string `json:"raw_key"`
}

// QueryName returns the query name.
func (q *ValidateAPIKey) QueryName() string {
	return TypeValidateAPIKey
}

// ============================================================================
// Token Queries
// ============================================================================

// ValidateToken validates an access or refresh token.
type ValidateToken struct {
	Token     string           `json:"token"`
	TokenType domain.TokenType `json:"token_type"`
}

// QueryName returns the query name.
func (q *ValidateToken) QueryName() string {
	return TypeValidateToken
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Query = (*GetUser)(nil)
	_ cqrs.Query = (*GetUserByDID)(nil)
	_ cqrs.Query = (*GetUserByWallet)(nil)
	_ cqrs.Query = (*GetUserByEmail)(nil)
	_ cqrs.Query = (*ListUsers)(nil)
	_ cqrs.Query = (*GetUserStats)(nil)
	_ cqrs.Query = (*GetPublicProfile)(nil)
	_ cqrs.Query = (*CheckUserExists)(nil)
	_ cqrs.Query = (*GetSession)(nil)
	_ cqrs.Query = (*ListUserSessions)(nil)
	_ cqrs.Query = (*ValidateSession)(nil)
	_ cqrs.Query = (*GetAPIKey)(nil)
	_ cqrs.Query = (*ListUserAPIKeys)(nil)
	_ cqrs.Query = (*ValidateAPIKey)(nil)
	_ cqrs.Query = (*ValidateToken)(nil)
)
