package v1

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/application/query"
)

// ============================================================================
// Auth Responses
// ============================================================================

// ChallengeResponse is the response containing a SIWE challenge.
type ChallengeResponse struct {
	Nonce     string    `json:"nonce"`
	Message   string    `json:"message"`
	Domain    string    `json:"domain"`
	URI       string    `json:"uri"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthResponse is the response after successful authentication.
type AuthResponse struct {
	UserID       string    `json:"user_id"`
	DID          string    `json:"did"`
	SessionID    string    `json:"session_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshResponse is the response after refreshing a token.
type RefreshResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// MagicLinkResponse is the response after requesting a magic link.
type MagicLinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================================
// User Responses
// ============================================================================

// UserResponse is the response containing user details.
type UserResponse struct {
	ID              string               `json:"id"`
	DID             string               `json:"did"`
	Status          string               `json:"status"`
	Wallets         []WalletResponse     `json:"wallets,omitempty"`
	Email           *EmailResponse       `json:"email,omitempty"`
	Connections     []ConnectionResponse `json:"connections,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	LastLoginAt     *time.Time           `json:"last_login_at,omitempty"`
	LastLoginMethod *string              `json:"last_login_method,omitempty"`
}

// FromUserView converts a query.UserView to UserResponse.
func FromUserView(v *query.UserView) *UserResponse {
	if v == nil {
		return nil
	}

	resp := &UserResponse{
		ID:          v.ID,
		DID:         v.PrimaryDID,
		Status:      v.Status,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		LastLoginAt: v.LastLoginAt,
	}

	for _, w := range v.Wallets {
		resp.Wallets = append(resp.Wallets, WalletResponse{
			Address:  w.Address,
			Chain:    w.Chain,
			DID:      w.DID,
			LinkedAt: w.LinkedAt,
		})
	}

	if v.LinkedEmail != nil {
		resp.Email = &EmailResponse{
			Email:      v.LinkedEmail.Email,
			Verified:   v.LinkedEmail.Verified,
			VerifiedAt: v.LinkedEmail.VerifiedAt,
			LinkedAt:   v.LinkedEmail.LinkedAt,
		}
	}

	for _, c := range v.Connections {
		resp.Connections = append(resp.Connections, ConnectionResponse{
			ID:           c.ID,
			Provider:     c.Provider,
			ProviderName: c.ProviderName,
			Username:     c.Username,
			Status:       c.Status,
			ConnectedAt:  c.ConnectedAt,
		})
	}

	return resp
}

// WalletResponse is a wallet in the response.
type WalletResponse struct {
	Address  string    `json:"address"`
	Chain    string    `json:"chain"`
	DID      string    `json:"did"`
	LinkedAt time.Time `json:"linked_at"`
}

// EmailResponse is an email in the response.
type EmailResponse struct {
	Email      string     `json:"email"`
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	LinkedAt   time.Time  `json:"linked_at"`
}

// ConnectionResponse is a connection in the response.
type ConnectionResponse struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	ProviderName string    `json:"provider_name"`
	Username     string    `json:"username,omitempty"`
	Status       string    `json:"status"`
	ConnectedAt  time.Time `json:"connected_at"`
}

// UserStatsResponse is the response containing user statistics.
type UserStatsResponse struct {
	UserID           string `json:"user_id"`
	WalletCount      int    `json:"wallet_count"`
	ConnectionCount  int    `json:"connection_count"`
	ActiveSessions   int    `json:"active_sessions"`
	ActiveAPIKeys    int    `json:"active_api_keys"`
	TotalAPIKeyUsage int64  `json:"total_api_key_usage"`
}

// FromUserStatsView converts a query.UserStatsView to UserStatsResponse.
func FromUserStatsView(v *query.UserStatsView) *UserStatsResponse {
	if v == nil {
		return nil
	}
	return &UserStatsResponse{
		UserID:           v.UserID,
		WalletCount:      v.WalletCount,
		ConnectionCount:  v.ConnectionCount,
		ActiveSessions:   v.ActiveSessions,
		ActiveAPIKeys:    v.ActiveAPIKeys,
		TotalAPIKeyUsage: v.TotalAPIKeyUsage,
	}
}

// ============================================================================
// Session Responses
// ============================================================================

// SessionResponse is the response containing session details.
type SessionResponse struct {
	ID         string    `json:"id"`
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

// FromSessionView converts a query.SessionView to SessionResponse.
func FromSessionView(v *query.SessionView) *SessionResponse {
	if v == nil {
		return nil
	}
	return &SessionResponse{
		ID:         v.ID,
		Method:     v.Method,
		Status:     v.Status,
		UserAgent:  v.UserAgent,
		IPAddress:  v.IPAddress,
		Device:     v.Device,
		CreatedAt:  v.CreatedAt,
		ExpiresAt:  v.ExpiresAt,
		LastSeenAt: v.LastSeenAt,
		IsCurrent:  v.IsCurrent,
	}
}

// SessionListResponse is the response containing a list of sessions.
type SessionListResponse struct {
	Sessions []SessionSummaryResponse `json:"sessions"`
	Total    int                      `json:"total"`
	Limit    int                      `json:"limit"`
	Offset   int                      `json:"offset"`
	HasMore  bool                     `json:"has_more"`
}

// SessionSummaryResponse is a summary of a session.
type SessionSummaryResponse struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	Status     string    `json:"status"`
	Device     string    `json:"device,omitempty"`
	LastSeenAt time.Time `json:"last_seen_at"`
	IsCurrent  bool      `json:"is_current"`
}

// FromSessionListView converts a query.SessionListView to SessionListResponse.
func FromSessionListView(v *query.SessionListView) *SessionListResponse {
	if v == nil {
		return nil
	}

	resp := &SessionListResponse{
		Total:   v.Total,
		Limit:   v.Limit,
		Offset:  v.Offset,
		HasMore: v.HasMore,
	}

	for _, s := range v.Sessions {
		resp.Sessions = append(resp.Sessions, SessionSummaryResponse{
			ID:         s.ID,
			Method:     s.Method,
			Status:     s.Status,
			Device:     s.Device,
			LastSeenAt: s.LastSeenAt,
			IsCurrent:  s.IsCurrent,
		})
	}

	return resp
}

// RevokeSessionResponse is the response after revoking a session.
type RevokeSessionResponse struct {
	SessionID string `json:"session_id"`
	Revoked   bool   `json:"revoked"`
}

// RevokeAllSessionsResponse is the response after revoking all sessions.
type RevokeAllSessionsResponse struct {
	RevokedCount int      `json:"revoked_count"`
	SessionIDs   []string `json:"session_ids"`
}

// ============================================================================
// API Key Responses
// ============================================================================

// APIKeyResponse is the response containing API key details.
type APIKeyResponse struct {
	ID          string     `json:"id"`
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

// FromAPIKeyView converts a query.APIKeyView to APIKeyResponse.
func FromAPIKeyView(v *query.APIKeyView) *APIKeyResponse {
	if v == nil {
		return nil
	}
	return &APIKeyResponse{
		ID:          v.ID,
		Name:        v.Name,
		Prefix:      v.Prefix,
		Scopes:      v.Scopes,
		Status:      v.Status,
		Description: v.Description,
		ExpiresAt:   v.ExpiresAt,
		LastUsedAt:  v.LastUsedAt,
		UsageCount:  v.UsageCount,
		CreatedAt:   v.CreatedAt,
	}
}

// APIKeyCreatedResponse is the response after creating an API key.
// Includes the raw key which is only shown once.
type APIKeyCreatedResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"` // Only shown once!
	Prefix    string     `json:"prefix"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// APIKeyListResponse is the response containing a list of API keys.
type APIKeyListResponse struct {
	APIKeys []APIKeySummaryResponse `json:"api_keys"`
	Total   int                     `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasMore bool                    `json:"has_more"`
}

// APIKeySummaryResponse is a summary of an API key.
type APIKeySummaryResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// FromAPIKeyListView converts a query.APIKeyListView to APIKeyListResponse.
func FromAPIKeyListView(v *query.APIKeyListView) *APIKeyListResponse {
	if v == nil {
		return nil
	}

	resp := &APIKeyListResponse{
		Total:   v.Total,
		Limit:   v.Limit,
		Offset:  v.Offset,
		HasMore: v.HasMore,
	}

	for _, k := range v.APIKeys {
		resp.APIKeys = append(resp.APIKeys, APIKeySummaryResponse{
			ID:         k.ID,
			Name:       k.Name,
			Prefix:     k.Prefix,
			Status:     k.Status,
			ExpiresAt:  k.ExpiresAt,
			LastUsedAt: k.LastUsedAt,
		})
	}

	return resp
}

// RevokeAPIKeyResponse is the response after revoking an API key.
type RevokeAPIKeyResponse struct {
	KeyID     string    `json:"key_id"`
	Revoked   bool      `json:"revoked"`
	RevokedAt time.Time `json:"revoked_at"`
}

// ============================================================================
// Link Responses
// ============================================================================

// LinkWalletResponse is the response after linking a wallet.
type LinkWalletResponse struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
	DID     string `json:"did"`
}

// UnlinkWalletResponse is the response after unlinking a wallet.
type UnlinkWalletResponse struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}
