package v1

import (
	"time"
)

// ============================================================================
// Auth Responses
// ============================================================================

// AuthResponse is the response after successful authentication.
type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	ExpiresAt    time.Time     `json:"expires_at"`
	User         *UserResponse `json:"user,omitempty"`
}

// MagicLinkSentResponse is the response after sending a magic link.
type MagicLinkSentResponse struct {
	Message   string `json:"message"`
	Email     string `json:"email"`
	ExpiresIn int    `json:"expires_in"`
}

// ============================================================================
// OAuth Responses
// ============================================================================

// OAuthAuthorizeResponse is the response with OAuth authorization URL.
type OAuthAuthorizeResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
	ExpiresIn        int    `json:"expires_in"`
}

// OAuthInitiatedResponse is the response after initiating OAuth flow.
type OAuthInitiatedResponse struct {
	AuthURL   string `json:"auth_url"`
	State     string `json:"state"`
	ExpiresIn int    `json:"expires_in"`
}

// ============================================================================
// Wallet Auth Responses
// ============================================================================

// WalletChallengeResponse is the response with a SIWE challenge.
type WalletChallengeResponse struct {
	Message   string    `json:"message"`
	Nonce     string    `json:"nonce"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ============================================================================
// Session Responses
// ============================================================================

// SessionResponse represents a session in API responses.
type SessionResponse struct {
	ID         string    `json:"id"`
	AuthMethod string    `json:"auth_method"`
	Status     string    `json:"status"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	IsCurrent  bool      `json:"is_current,omitempty"`
}

// SessionListResponse is a list of sessions.
type SessionListResponse struct {
	Sessions   []SessionResponse `json:"sessions"`
	TotalCount int               `json:"total_count"`
}

// SessionRevokedResponse is the response after revoking a session.
type SessionRevokedResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// ============================================================================
// User Responses
// ============================================================================

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	Status      string    `json:"status,omitempty"`
	PrimaryDID  string    `json:"primary_did,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// UserProfileResponse represents a detailed user profile.
type UserProfileResponse struct {
	User          UserResponse           `json:"user"`
	DIDs          []string               `json:"dids"`
	AuthMethods   []string               `json:"auth_methods"`
	OAuthAccounts []OAuthAccountResponse `json:"oauth_accounts"`
}

// PublicProfileResponse represents a public user profile.
type PublicProfileResponse struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	PrimaryDID  string    `json:"primary_did"`
	CreatedAt   time.Time `json:"created_at"`
}

// ============================================================================
// DID Responses
// ============================================================================

// DIDResponse represents a DID in API responses.
type DIDResponse struct {
	DID       string    `json:"did"`
	Method    string    `json:"method,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	AddedAt   time.Time `json:"added_at"`
}

// DIDListResponse is the list of DIDs for a user.
type DIDListResponse struct {
	UserID     string        `json:"user_id"`
	PrimaryDID string        `json:"primary_did"`
	DIDs       []DIDResponse `json:"dids"`
}

// DIDAddedResponse is the response after adding a DID.
type DIDAddedResponse struct {
	Message string `json:"message"`
	DID     string `json:"did"`
}

// DIDRemovedResponse is the response after removing a DID.
type DIDRemovedResponse struct {
	Message string `json:"message"`
	DID     string `json:"did"`
}

// DIDResolutionResponse is the response for DID resolution.
type DIDResolutionResponse struct {
	DID         string `json:"did"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	IsPrimary   bool   `json:"is_primary"`
}

// ============================================================================
// OAuth Account Responses
// ============================================================================

// OAuthAccountResponse represents an OAuth account in API responses.
type OAuthAccountResponse struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id"`
	Email      string `json:"email,omitempty"`
}

// OAuthAccountsResponse is the list of OAuth accounts for a user.
type OAuthAccountsResponse struct {
	UserID   string                 `json:"user_id"`
	Accounts []OAuthAccountResponse `json:"accounts"`
}

// ============================================================================
// OAuth Link Responses
// ============================================================================

// OAuthLinkResponse represents an OAuth link in API responses (with linked_at).
type OAuthLinkResponse struct {
	Provider   string    `json:"provider"`
	ExternalID string    `json:"external_id"`
	Email      string    `json:"email,omitempty"`
	LinkedAt   time.Time `json:"linked_at"`
}

// OAuthLinkedResponse is the response after linking an OAuth account.
type OAuthLinkedResponse struct {
	Message  string `json:"message"`
	Provider string `json:"provider"`
}

// OAuthUnlinkedResponse is the response after unlinking an OAuth account.
type OAuthUnlinkedResponse struct {
	Message  string `json:"message"`
	Provider string `json:"provider"`
}

// ============================================================================
// Profile Update Responses
// ============================================================================

// DisplayNameUpdatedResponse is the response after updating display name.
type DisplayNameUpdatedResponse struct {
	Message     string `json:"message"`
	DisplayName string `json:"display_name"`
}

// EmailUpdatedResponse is the response after updating email.
type EmailUpdatedResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// ============================================================================
// Account Responses
// ============================================================================

// AccountActivatedResponse is the response after activating an account.
type AccountActivatedResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
}

// AccountDeletedResponse is the response after deleting an account.
type AccountDeletedResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// Health/Status Responses
// ============================================================================

// HealthResponse represents the service health status.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// Generic Responses
// ============================================================================

// MessageResponse is a generic message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// SuccessResponse is a generic success response.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
