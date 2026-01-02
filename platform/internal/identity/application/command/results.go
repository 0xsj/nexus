package command

import (
	"time"
)

// ============================================================================
// Authentication Results
// ============================================================================

// ChallengeResult is returned when requesting an auth challenge.
type ChallengeResult struct {
	Nonce     string    `json:"nonce"`
	Message   string    `json:"message"`
	Domain    string    `json:"domain"`
	URI       string    `json:"uri"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthResult is returned after successful authentication.
type AuthResult struct {
	UserID       string    `json:"user_id"`
	DID          string    `json:"did"`
	SessionID    string    `json:"session_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"` // Seconds until access token expires
	ExpiresAt    time.Time `json:"expires_at"`
	IsNewUser    bool      `json:"is_new_user,omitempty"`
}

// RefreshResult is returned after refreshing tokens.
type RefreshResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"` // Only if rotated
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// APIKeyAuthResult is returned after API key authentication.
type APIKeyAuthResult struct {
	UserID  string   `json:"user_id"`
	DID     string   `json:"did"`
	KeyID   string   `json:"key_id"`
	KeyName string   `json:"key_name"`
	Scopes  []string `json:"scopes"`
}

// ============================================================================
// Registration Results
// ============================================================================

// RegisterResult is returned after successful registration.
type RegisterResult struct {
	UserID       string    `json:"user_id"`
	DID          string    `json:"did"`
	SessionID    string    `json:"session_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ============================================================================
// Session Results
// ============================================================================

// RevokeSessionResult is returned after revoking a session.
type RevokeSessionResult struct {
	SessionID string `json:"session_id"`
	Revoked   bool   `json:"revoked"`
}

// RevokeAllSessionsResult is returned after revoking all sessions.
type RevokeAllSessionsResult struct {
	RevokedCount int      `json:"revoked_count"`
	SessionIDs   []string `json:"session_ids,omitempty"`
}

// ============================================================================
// Wallet Results
// ============================================================================

// LinkWalletResult is returned after linking a wallet.
type LinkWalletResult struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
	DID     string `json:"did"` // did:pkh derived from wallet
}

// UnlinkWalletResult is returned after unlinking a wallet.
type UnlinkWalletResult struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// ============================================================================
// Email Results
// ============================================================================

// LinkEmailResult is returned after linking an email.
type LinkEmailResult struct {
	UserID           string `json:"user_id"`
	Email            string `json:"email"`
	VerificationSent bool   `json:"verification_sent"`
}

// VerifyEmailResult is returned after verifying an email.
type VerifyEmailResult struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

// ============================================================================
// OAuth Results
// ============================================================================

// InitiateOAuthResult is returned when starting OAuth flow.
type InitiateOAuthResult struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// CompleteOAuthResult is returned after completing OAuth flow.
type CompleteOAuthResult struct {
	ConnectionID string `json:"connection_id"`
	Provider     string `json:"provider"`
	Username     string `json:"username,omitempty"`
	Email        string `json:"email,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	IsNew        bool   `json:"is_new"`
}

// UnlinkOAuthResult is returned after unlinking OAuth.
type UnlinkOAuthResult struct {
	UserID   string `json:"user_id"`
	Provider string `json:"provider"`
}

// SyncConnectionResult is returned after syncing a connection.
type SyncConnectionResult struct {
	ConnectionID   string    `json:"connection_id"`
	Provider       string    `json:"provider"`
	SyncedAt       time.Time `json:"synced_at"`
	ProfileUpdated bool      `json:"profile_updated"`
}

// ============================================================================
// API Key Results
// ============================================================================

// CreateAPIKeyResult is returned after creating an API key.
// RawKey is only returned ONCE at creation time.
type CreateAPIKeyResult struct {
	KeyID     string     `json:"key_id"`
	RawKey    string     `json:"raw_key"` // Only returned once!
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// RevokeAPIKeyResult is returned after revoking an API key.
type RevokeAPIKeyResult struct {
	KeyID     string    `json:"key_id"`
	Revoked   bool      `json:"revoked"`
	RevokedAt time.Time `json:"revoked_at"`
}

// UpdateAPIKeyResult is returned after updating an API key.
type UpdateAPIKeyResult struct {
	KeyID       string    `json:"key_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================================================
// User Management Results
// ============================================================================

// SuspendUserResult is returned after suspending a user.
type SuspendUserResult struct {
	UserID      string    `json:"user_id"`
	Suspended   bool      `json:"suspended"`
	SuspendedAt time.Time `json:"suspended_at"`
}

// ActivateUserResult is returned after activating a user.
type ActivateUserResult struct {
	UserID      string    `json:"user_id"`
	Activated   bool      `json:"activated"`
	ActivatedAt time.Time `json:"activated_at"`
}

// UpdatePrimaryDIDResult is returned after updating primary DID.
type UpdatePrimaryDIDResult struct {
	UserID string `json:"user_id"`
	OldDID string `json:"old_did"`
	NewDID string `json:"new_did"`
}

// DeleteUserResult is returned after deleting a user.
type DeleteUserResult struct {
	UserID    string    `json:"user_id"`
	Deleted   bool      `json:"deleted"`
	DeletedAt time.Time `json:"deleted_at"`
}
