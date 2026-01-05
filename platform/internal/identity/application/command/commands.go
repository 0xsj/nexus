package command

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Command Types
// ============================================================================

const (
	TypeRequestChallenge       = "identity.request_challenge"
	TypeRegisterWithWallet     = "identity.register_with_wallet"
	TypeAuthenticateWithWallet = "identity.authenticate_with_wallet"
	TypeRefreshToken           = "identity.refresh_token"
	TypeRevokeSession          = "identity.revoke_session"
	TypeRevokeAllSessions      = "identity.revoke_all_sessions"
	TypeCreateAPIKey           = "identity.create_api_key"
	TypeRevokeAPIKey           = "identity.revoke_api_key"
	TypeLinkWallet             = "identity.link_wallet"
	TypeUnlinkWallet           = "identity.unlink_wallet"
	TypeRequestMagicLink       = "identity.request_magic_link"
	TypeVerifyMagicLink        = "identity.verify_magic_link"
)

// ============================================================================
// Request Challenge Command
// ============================================================================

// RequestChallenge generates a SIWE challenge for wallet authentication.
type RequestChallenge struct {
	Address string       `json:"address"`
	Chain   domain.Chain `json:"chain"`
	Domain  string       `json:"domain"`
	URI     string       `json:"uri"`
}

// CommandName returns the command name.
func (c *RequestChallenge) CommandName() string {
	return TypeRequestChallenge
}

// RequestChallengeResult is the result of a challenge request.
type RequestChallengeResult struct {
	Nonce     string    `json:"nonce"`
	Message   string    `json:"message"`
	Domain    string    `json:"domain"`
	URI       string    `json:"uri"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ============================================================================
// Register With Wallet Command
// ============================================================================

// RegisterWithWallet registers a new user with a wallet signature.
type RegisterWithWallet struct {
	Address   string       `json:"address"`
	Chain     domain.Chain `json:"chain"`
	Signature string       `json:"signature"`
	Message   string       `json:"message"`
	Nonce     string       `json:"nonce"`
	UserAgent string       `json:"user_agent"`
	IPAddress string       `json:"ip_address"`
}

// CommandName returns the command name.
func (c *RegisterWithWallet) CommandName() string {
	return TypeRegisterWithWallet
}

// AuthResult is the result of authentication commands.
type AuthResult struct {
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
// Authenticate With Wallet Command
// ============================================================================

// AuthenticateWithWallet authenticates an existing user with a wallet signature.
type AuthenticateWithWallet struct {
	Address   string       `json:"address"`
	Chain     domain.Chain `json:"chain"`
	Signature string       `json:"signature"`
	Message   string       `json:"message"`
	Nonce     string       `json:"nonce"`
	UserAgent string       `json:"user_agent"`
	IPAddress string       `json:"ip_address"`
}

// CommandName returns the command name.
func (c *AuthenticateWithWallet) CommandName() string {
	return TypeAuthenticateWithWallet
}

// ============================================================================
// Refresh Token Command
// ============================================================================

// RefreshToken refreshes an access token.
type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
	UserAgent    string `json:"user_agent"`
	IPAddress    string `json:"ip_address"`
}

// CommandName returns the command name.
func (c *RefreshToken) CommandName() string {
	return TypeRefreshToken
}

// RefreshTokenResult is the result of a token refresh.
type RefreshTokenResult struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ============================================================================
// Revoke Session Command
// ============================================================================

// RevokeSession revokes a specific session.
type RevokeSession struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
}

// CommandName returns the command name.
func (c *RevokeSession) CommandName() string {
	return TypeRevokeSession
}

// RevokeSessionResult is the result of session revocation.
type RevokeSessionResult struct {
	SessionID string `json:"session_id"`
	Revoked   bool   `json:"revoked"`
}

// ============================================================================
// Revoke All Sessions Command
// ============================================================================

// RevokeAllSessions revokes all sessions for a user.
type RevokeAllSessions struct {
	UserID         string `json:"user_id"`
	ExceptCurrent  bool   `json:"except_current"`
	CurrentSession string `json:"current_session"`
	Reason         string `json:"reason"`
}

// CommandName returns the command name.
func (c *RevokeAllSessions) CommandName() string {
	return TypeRevokeAllSessions
}

// RevokeAllSessionsResult is the result of revoking all sessions.
type RevokeAllSessionsResult struct {
	RevokedCount int      `json:"revoked_count"`
	SessionIDs   []string `json:"session_ids"`
}

// ============================================================================
// Create API Key Command
// ============================================================================

// CreateAPIKey creates a new API key for a user.
type CreateAPIKey struct {
	UserID      string              `json:"user_id"`
	Name        string              `json:"name"`
	Scopes      domain.APIKeyScopes `json:"scopes"`
	ExpiresIn   time.Duration       `json:"expires_in"`
	Description string              `json:"description"`
}

// CommandName returns the command name.
func (c *CreateAPIKey) CommandName() string {
	return TypeCreateAPIKey
}

// CreateAPIKeyResult is the result of API key creation.
type CreateAPIKeyResult struct {
	KeyID     string     `json:"key_id"`
	Name      string     `json:"name"`
	RawKey    string     `json:"raw_key"` // Only returned once!
	Prefix    string     `json:"prefix"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// ============================================================================
// Revoke API Key Command
// ============================================================================

// RevokeAPIKey revokes an API key.
type RevokeAPIKey struct {
	UserID string `json:"user_id"`
	KeyID  string `json:"key_id"`
	Reason string `json:"reason"`
}

// CommandName returns the command name.
func (c *RevokeAPIKey) CommandName() string {
	return TypeRevokeAPIKey
}

// RevokeAPIKeyResult is the result of API key revocation.
type RevokeAPIKeyResult struct {
	KeyID     string    `json:"key_id"`
	Revoked   bool      `json:"revoked"`
	RevokedAt time.Time `json:"revoked_at"`
}

// ============================================================================
// Link Wallet Command
// ============================================================================

// LinkWallet links a wallet to an existing user.
type LinkWallet struct {
	UserID    string       `json:"user_id"`
	Address   string       `json:"address"`
	Chain     domain.Chain `json:"chain"`
	Signature string       `json:"signature"`
	Message   string       `json:"message"`
	Nonce     string       `json:"nonce"`
}

// CommandName returns the command name.
func (c *LinkWallet) CommandName() string {
	return TypeLinkWallet
}

// LinkWalletResult is the result of linking a wallet.
type LinkWalletResult struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
	DID     string `json:"did"`
}

// ============================================================================
// Unlink Wallet Command
// ============================================================================

// UnlinkWallet unlinks a wallet from a user.
type UnlinkWallet struct {
	UserID  string       `json:"user_id"`
	Address string       `json:"address"`
	Chain   domain.Chain `json:"chain"`
}

// CommandName returns the command name.
func (c *UnlinkWallet) CommandName() string {
	return TypeUnlinkWallet
}

// UnlinkWalletResult is the result of unlinking a wallet.
type UnlinkWalletResult struct {
	UserID  string `json:"user_id"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// ============================================================================
// Request Magic Link Command
// ============================================================================

// RequestMagicLink requests a magic link for email authentication.
type RequestMagicLink struct {
	Email     string                  `json:"email"`
	Purpose   domain.MagicLinkPurpose `json:"purpose"`
	IPAddress string                  `json:"ip_address"`
	UserAgent string                  `json:"user_agent"`
}

// CommandName returns the command name.
func (c *RequestMagicLink) CommandName() string {
	return TypeRequestMagicLink
}

// RequestMagicLinkResult is the result of requesting a magic link.
type RequestMagicLinkResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================================
// Verify Magic Link Command
// ============================================================================

// VerifyMagicLink verifies a magic link token and authenticates the user.
type VerifyMagicLink struct {
	Token     string `json:"token"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// CommandName returns the command name.
func (c *VerifyMagicLink) CommandName() string {
	return TypeVerifyMagicLink
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Command = (*RequestChallenge)(nil)
	_ cqrs.Command = (*RegisterWithWallet)(nil)
	_ cqrs.Command = (*AuthenticateWithWallet)(nil)
	_ cqrs.Command = (*RefreshToken)(nil)
	_ cqrs.Command = (*RevokeSession)(nil)
	_ cqrs.Command = (*RevokeAllSessions)(nil)
	_ cqrs.Command = (*CreateAPIKey)(nil)
	_ cqrs.Command = (*RevokeAPIKey)(nil)
	_ cqrs.Command = (*LinkWallet)(nil)
	_ cqrs.Command = (*UnlinkWallet)(nil)
	_ cqrs.Command = (*RequestMagicLink)(nil)
	_ cqrs.Command = (*VerifyMagicLink)(nil)
)
