package command

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
)

// ============================================================================
// Registration Commands
// ============================================================================

// RegisterWithWallet creates a new user from a wallet signature.
type RegisterWithWallet struct {
	Address   string
	Chain     domain.Chain
	Signature string
	Nonce     string
	Message   string
	UserAgent string
	IPAddress string
}

// RegisterWithOAuth creates a new user from OAuth authentication.
type RegisterWithOAuth struct {
	Provider    domain.OAuthProvider
	Code        string
	State       string
	RedirectURI string
	UserAgent   string
	IPAddress   string
}

// ============================================================================
// Authentication Commands
// ============================================================================

// RequestChallenge requests a new authentication challenge.
type RequestChallenge struct {
	Address string
	Chain   domain.Chain
	Domain  string
	URI     string
}

// AuthenticateWithWallet authenticates a user via wallet signature.
type AuthenticateWithWallet struct {
	Address   string
	Chain     domain.Chain
	Signature string
	Nonce     string
	Message   string
	UserAgent string
	IPAddress string
}

// AuthenticateWithOAuth authenticates a user via OAuth.
type AuthenticateWithOAuth struct {
	Provider    domain.OAuthProvider
	Code        string
	State       string
	RedirectURI string
	UserAgent   string
	IPAddress   string
}

// AuthenticateWithAPIKey authenticates a request via API key.
type AuthenticateWithAPIKey struct {
	RawKey    string
	IPAddress string
	UserAgent string
}

// RefreshToken refreshes an access token using a refresh token.
type RefreshToken struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

// ============================================================================
// Session Commands
// ============================================================================

// RevokeSession revokes a single session.
type RevokeSession struct {
	UserID    string
	SessionID string
	Reason    string
}

// RevokeAllSessions revokes all sessions for a user.
type RevokeAllSessions struct {
	UserID         string
	ExceptCurrent  bool
	CurrentSession string
	Reason         string
}

// ============================================================================
// Wallet Commands
// ============================================================================

// LinkWallet links a new wallet to an existing user.
type LinkWallet struct {
	UserID    string
	Address   string
	Chain     domain.Chain
	Signature string
	Nonce     string
	Message   string
}

// UnlinkWallet removes a wallet from a user.
type UnlinkWallet struct {
	UserID  string
	Address string
	Chain   domain.Chain
}

// ============================================================================
// Email Commands
// ============================================================================

// LinkEmail links an email address to a user.
type LinkEmail struct {
	UserID string
	Email  string
}

// VerifyEmail verifies a linked email address.
type VerifyEmail struct {
	Token string
}

// ResendEmailVerification resends the email verification.
type ResendEmailVerification struct {
	UserID string
	Email  string
}

// ============================================================================
// OAuth Connection Commands
// ============================================================================

// InitiateOAuthLink starts the OAuth linking flow.
type InitiateOAuthLink struct {
	UserID   string
	Provider domain.OAuthProvider
	Scopes   []string
}

// CompleteOAuthLink completes the OAuth linking flow.
type CompleteOAuthLink struct {
	UserID      string
	Provider    domain.OAuthProvider
	Code        string
	State       string
	RedirectURI string
}

// UnlinkOAuth removes an OAuth connection.
type UnlinkOAuth struct {
	UserID   string
	Provider domain.OAuthProvider
}

// SyncConnection refreshes OAuth connection profile data.
type SyncConnection struct {
	UserID       string
	ConnectionID string
}

// RefreshOAuthTokens refreshes OAuth tokens for a connection.
type RefreshOAuthTokens struct {
	UserID       string
	ConnectionID string
}

// ============================================================================
// API Key Commands
// ============================================================================

// CreateAPIKey creates a new API key.
type CreateAPIKey struct {
	UserID      string
	Name        string
	Scopes      []domain.APIKeyScope
	ExpiresIn   time.Duration // 0 = no expiration
	Description string
}

// RevokeAPIKey revokes an API key.
type RevokeAPIKey struct {
	UserID string
	KeyID  string
	Reason string
}

// UpdateAPIKey updates an API key's metadata.
type UpdateAPIKey struct {
	UserID      string
	KeyID       string
	Name        *string
	Description *string
}

// ============================================================================
// User Management Commands
// ============================================================================

// SuspendUser suspends a user account.
type SuspendUser struct {
	UserID      string
	Reason      string
	SuspendedBy string
}

// ActivateUser reactivates a suspended user account.
type ActivateUser struct {
	UserID      string
	ActivatedBy string
}

// UpdatePrimaryDID updates the user's primary DID.
type UpdatePrimaryDID struct {
	UserID string
	NewDID string
}

// DeleteUser deletes a user and all associated data.
type DeleteUser struct {
	UserID    string
	DeletedBy string
	Reason    string
}
