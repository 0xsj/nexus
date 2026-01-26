package command

import (
	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Command Types
// ============================================================================

const (
	TypeInitiateVerification = "verification.initiate"
	TypeHandleOAuthCallback  = "verification.oauth_callback"
	TypeCompleteVerification = "verification.complete"
	TypeFailVerification     = "verification.fail"
	TypeExpireVerification   = "verification.expire"
	TypeCancelVerification   = "verification.cancel"
)

// ============================================================================
// Initiate Verification Command
// ============================================================================

// InitiateVerification starts a new verification flow with a provider.
type InitiateVerification struct {
	VerificationID string                `json:"verification_id"`
	UserID         string                `json:"user_id"`
	Provider       domain.Provider       `json:"provider"`
	CredentialType domain.CredentialType `json:"credential_type"`
	RedirectURL    string                `json:"redirect_url,omitempty"`
}

// CommandName returns the command type.
func (c *InitiateVerification) CommandName() string {
	return TypeInitiateVerification
}

// NewInitiateVerification creates a new InitiateVerification command.
func NewInitiateVerification(
	verificationID string,
	userID string,
	provider domain.Provider,
	credentialType domain.CredentialType,
) *InitiateVerification {
	return &InitiateVerification{
		VerificationID: verificationID,
		UserID:         userID,
		Provider:       provider,
		CredentialType: credentialType,
	}
}

// WithRedirectURL sets the redirect URL.
func (c *InitiateVerification) WithRedirectURL(url string) *InitiateVerification {
	c.RedirectURL = url
	return c
}

// InitiateVerificationResult is the result of initiating a verification.
type InitiateVerificationResult struct {
	VerificationID   string `json:"verification_id"`
	AuthorizationURL string `json:"authorization_url"`
	OAuthState       string `json:"oauth_state"`
	ExpiresIn        int64  `json:"expires_in"`
}

// ============================================================================
// Handle OAuth Callback Command
// ============================================================================

// HandleOAuthCallback processes the OAuth callback from a provider.
type HandleOAuthCallback struct {
	State string `json:"state"`
	Code  string `json:"code"`
	Error string `json:"error,omitempty"`
}

// CommandName returns the command type.
func (c *HandleOAuthCallback) CommandName() string {
	return TypeHandleOAuthCallback
}

// NewHandleOAuthCallback creates a new HandleOAuthCallback command.
func NewHandleOAuthCallback(state, code string) *HandleOAuthCallback {
	return &HandleOAuthCallback{
		State: state,
		Code:  code,
	}
}

// WithError sets the OAuth error.
func (c *HandleOAuthCallback) WithError(err string) *HandleOAuthCallback {
	c.Error = err
	return c
}

// HandleOAuthCallbackResult is the result of handling an OAuth callback.
type HandleOAuthCallbackResult struct {
	VerificationID string `json:"verification_id"`
	UserID         string `json:"user_id"`
	Provider       string `json:"provider"`
	RedirectURL    string `json:"redirect_url,omitempty"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

// ============================================================================
// Complete Verification Command
// ============================================================================

// CompleteVerification completes a verification and issues a credential.
type CompleteVerification struct {
	VerificationID string         `json:"verification_id"`
	ProviderUserID string         `json:"provider_user_id"`
	Username       string         `json:"username"`
	Claims         map[string]any `json:"claims"`
}

// CommandName returns the command type.
func (c *CompleteVerification) CommandName() string {
	return TypeCompleteVerification
}

// NewCompleteVerification creates a new CompleteVerification command.
func NewCompleteVerification(
	verificationID string,
	providerUserID string,
	username string,
	claims map[string]any,
) *CompleteVerification {
	return &CompleteVerification{
		VerificationID: verificationID,
		ProviderUserID: providerUserID,
		Username:       username,
		Claims:         claims,
	}
}

// CompleteVerificationResult is the result of completing a verification.
type CompleteVerificationResult struct {
	VerificationID string `json:"verification_id"`
	CredentialID   string `json:"credential_id"`
	CredentialType string `json:"credential_type"`
	Provider       string `json:"provider"`
	SignedVC       string `json:"signed_vc,omitempty"`
}

// ============================================================================
// Fail Verification Command
// ============================================================================

// FailVerification marks a verification as failed.
type FailVerification struct {
	VerificationID string `json:"verification_id"`
	Reason         string `json:"reason"`
	ErrorCode      string `json:"error_code,omitempty"`
}

// CommandName returns the command type.
func (c *FailVerification) CommandName() string {
	return TypeFailVerification
}

// NewFailVerification creates a new FailVerification command.
func NewFailVerification(verificationID, reason string) *FailVerification {
	return &FailVerification{
		VerificationID: verificationID,
		Reason:         reason,
	}
}

// WithErrorCode sets the error code.
func (c *FailVerification) WithErrorCode(code string) *FailVerification {
	c.ErrorCode = code
	return c
}

// FailVerificationResult is the result of failing a verification.
type FailVerificationResult struct {
	VerificationID string `json:"verification_id"`
	Failed         bool   `json:"failed"`
}

// ============================================================================
// Expire Verification Command
// ============================================================================

// ExpireVerification marks a verification as expired.
type ExpireVerification struct {
	VerificationID string `json:"verification_id"`
}

// CommandName returns the command type.
func (c *ExpireVerification) CommandName() string {
	return TypeExpireVerification
}

// NewExpireVerification creates a new ExpireVerification command.
func NewExpireVerification(verificationID string) *ExpireVerification {
	return &ExpireVerification{
		VerificationID: verificationID,
	}
}

// ExpireVerificationResult is the result of expiring a verification.
type ExpireVerificationResult struct {
	VerificationID string `json:"verification_id"`
	Expired        bool   `json:"expired"`
}

// ============================================================================
// Cancel Verification Command
// ============================================================================

// CancelVerification cancels an in-progress verification.
type CancelVerification struct {
	VerificationID string `json:"verification_id"`
	UserID         string `json:"user_id"`
	Reason         string `json:"reason,omitempty"`
}

// CommandName returns the command type.
func (c *CancelVerification) CommandName() string {
	return TypeCancelVerification
}

// NewCancelVerification creates a new CancelVerification command.
func NewCancelVerification(verificationID, userID string) *CancelVerification {
	return &CancelVerification{
		VerificationID: verificationID,
		UserID:         userID,
	}
}

// WithReason sets the cancellation reason.
func (c *CancelVerification) WithReason(reason string) *CancelVerification {
	c.Reason = reason
	return c
}

// CancelVerificationResult is the result of cancelling a verification.
type CancelVerificationResult struct {
	VerificationID string `json:"verification_id"`
	Cancelled      bool   `json:"cancelled"`
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Command = (*InitiateVerification)(nil)
	_ cqrs.Command = (*HandleOAuthCallback)(nil)
	_ cqrs.Command = (*CompleteVerification)(nil)
	_ cqrs.Command = (*FailVerification)(nil)
	_ cqrs.Command = (*ExpireVerification)(nil)
	_ cqrs.Command = (*CancelVerification)(nil)
)
