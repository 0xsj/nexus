package v1

import (
	"time"
)

// ============================================================================
// Registration Requests
// ============================================================================

// RegisterWithEmailRequest is the request to register a new user with email.
type RegisterWithEmailRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
}

// Validate validates the request.
func (r *RegisterWithEmailRequest) Validate() error {
	if r.Email == "" {
		return NewValidationError("email is required")
	}
	return nil
}

// RegisterWithWalletRequest is the request to register a new user with a wallet.
type RegisterWithWalletRequest struct {
	Address     string `json:"address"`
	ChainID     string `json:"chain_id,omitempty"`
	Message     string `json:"message"`
	Signature   string `json:"signature"`
	DisplayName string `json:"display_name,omitempty"`
}

// Validate validates the request.
func (r *RegisterWithWalletRequest) Validate() error {
	if r.Address == "" {
		return NewValidationError("address is required")
	}
	if r.Message == "" {
		return NewValidationError("message is required")
	}
	if r.Signature == "" {
		return NewValidationError("signature is required")
	}
	return nil
}

// ============================================================================
// Magic Link Requests
// ============================================================================

// SendMagicLinkRequest is the request to send a magic link.
type SendMagicLinkRequest struct {
	Email       string `json:"email"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// Validate validates the request.
func (r *SendMagicLinkRequest) Validate() error {
	if r.Email == "" {
		return NewValidationError("email is required")
	}
	return nil
}

// VerifyMagicLinkRequest is the request to verify a magic link.
type VerifyMagicLinkRequest struct {
	Token string `json:"token"`
}

// Validate validates the request.
func (r *VerifyMagicLinkRequest) Validate() error {
	if r.Token == "" {
		return NewValidationError("token is required")
	}
	return nil
}

// ============================================================================
// OAuth Requests
// ============================================================================

// OAuthAuthorizeRequest is the request to start OAuth flow.
type OAuthAuthorizeRequest struct {
	Provider    string `json:"provider"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// Validate validates the request.
func (r *OAuthAuthorizeRequest) Validate() error {
	if r.Provider == "" {
		return NewValidationError("provider is required")
	}
	return nil
}

// OAuthCallbackRequest is the request to complete OAuth flow.
type OAuthCallbackRequest struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

// Validate validates the request.
func (r *OAuthCallbackRequest) Validate() error {
	if r.Provider == "" {
		return NewValidationError("provider is required")
	}
	if r.Code == "" {
		return NewValidationError("code is required")
	}
	if r.State == "" {
		return NewValidationError("state is required")
	}
	return nil
}

// ============================================================================
// Wallet Auth Requests
// ============================================================================

// WalletChallengeRequest is the request to get a SIWE challenge.
type WalletChallengeRequest struct {
	Address string `json:"address"`
	ChainID string `json:"chain_id,omitempty"`
}

// Validate validates the request.
func (r *WalletChallengeRequest) Validate() error {
	if r.Address == "" {
		return NewValidationError("address is required")
	}
	return nil
}

// WalletVerifyRequest is the request to verify a SIWE signature.
type WalletVerifyRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// Validate validates the request.
func (r *WalletVerifyRequest) Validate() error {
	if r.Message == "" {
		return NewValidationError("message is required")
	}
	if r.Signature == "" {
		return NewValidationError("signature is required")
	}
	return nil
}

// ============================================================================
// Session Requests
// ============================================================================

// RefreshSessionRequest is the request to refresh a session.
type RefreshSessionRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Validate validates the request.
func (r *RefreshSessionRequest) Validate() error {
	if r.RefreshToken == "" {
		return NewValidationError("refresh_token is required")
	}
	return nil
}

// RevokeSessionRequest is the request to revoke a session.
type RevokeSessionRequest struct {
	SessionID string `json:"session_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// Validate validates the request.
func (r *RevokeSessionRequest) Validate() error {
	// SessionID is optional - if not provided, revoke current session
	return nil
}

// ============================================================================
// User Profile Requests
// ============================================================================

// UpdateDisplayNameRequest is the request to update display name.
type UpdateDisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

// Validate validates the request.
func (r *UpdateDisplayNameRequest) Validate() error {
	if r.DisplayName == "" {
		return NewValidationError("display_name is required")
	}
	return nil
}

// UpdateEmailRequest is the request to update email.
type UpdateEmailRequest struct {
	Email string `json:"email"`
}

// Validate validates the request.
func (r *UpdateEmailRequest) Validate() error {
	if r.Email == "" {
		return NewValidationError("email is required")
	}
	return nil
}

// ============================================================================
// DID Requests
// ============================================================================

// AddDIDRequest is the request to add a DID.
type AddDIDRequest struct {
	DID string `json:"did"`
}

// Validate validates the request.
func (r *AddDIDRequest) Validate() error {
	if r.DID == "" {
		return NewValidationError("did is required")
	}
	return nil
}

// RemoveDIDRequest is the request to remove a DID.
type RemoveDIDRequest struct {
	DID string `json:"did"`
}

// Validate validates the request.
func (r *RemoveDIDRequest) Validate() error {
	if r.DID == "" {
		return NewValidationError("did is required")
	}
	return nil
}

// ============================================================================
// OAuth Link Requests
// ============================================================================

// LinkOAuthRequest is the request to link an OAuth account.
type LinkOAuthRequest struct {
	Provider    string `json:"provider"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// Validate validates the request.
func (r *LinkOAuthRequest) Validate() error {
	if r.Provider == "" {
		return NewValidationError("provider is required")
	}
	return nil
}

// UnlinkOAuthRequest is the request to unlink an OAuth account.
type UnlinkOAuthRequest struct {
	Provider string `json:"provider"`
}

// Validate validates the request.
func (r *UnlinkOAuthRequest) Validate() error {
	if r.Provider == "" {
		return NewValidationError("provider is required")
	}
	return nil
}

// ============================================================================
// Validation Interface
// ============================================================================

// Validatable is implemented by requests that can validate themselves.
type Validatable interface {
	Validate() error
}

// ============================================================================
// Request Metadata
// ============================================================================

// RequestMetadata contains metadata extracted from the HTTP request.
type RequestMetadata struct {
	IPAddress   string
	UserAgent   string
	RequestID   string
	RequestedAt time.Time
}

// NewRequestMetadata creates request metadata with current time.
func NewRequestMetadata(ipAddress, userAgent, requestID string) RequestMetadata {
	return RequestMetadata{
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		RequestID:   requestID,
		RequestedAt: time.Now().UTC(),
	}
}
