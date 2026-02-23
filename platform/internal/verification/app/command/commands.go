package command

import (
	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// Command name constants
const (
	CommandStartVerification    = "verification.StartVerification"
	CommandReceiveOAuthCallback = "verification.ReceiveOAuthCallback"
	CommandCompleteVerification = "verification.CompleteVerification"
	CommandFailVerification     = "verification.FailVerification"
)

// ============================================================================
// StartVerification
// ============================================================================

// StartVerification initiates a new verification for a provider.
type StartVerification struct {
	UserID       string              `json:"user_id" validate:"required"`
	ProviderType domain.ProviderType `json:"provider_type" validate:"required"`
	RedirectURI  string              `json:"redirect_uri"`
}

// CommandName implements cqrs.Command.
func (c StartVerification) CommandName() string {
	return CommandStartVerification
}

// Validate implements cqrs.Validatable.
func (c StartVerification) Validate() error {
	if c.UserID == "" {
		return cqrs.ErrCommandValidation("StartVerification.Validate", "user_id is required")
	}
	if !c.ProviderType.IsValid() {
		return cqrs.ErrCommandValidation("StartVerification.Validate", "provider_type is invalid")
	}
	return nil
}

// StartVerificationResult is the result data for StartVerification.
type StartVerificationResult struct {
	VerificationID string `json:"verification_id"`
	ProviderType   string `json:"provider_type"`
	Status         string `json:"status"`
	OAuthState     string `json:"oauth_state"`
	AuthURL        string `json:"auth_url,omitempty"`
}

// ============================================================================
// ReceiveOAuthCallback
// ============================================================================

// ReceiveOAuthCallback records that an OAuth callback has been received.
type ReceiveOAuthCallback struct {
	VerificationID string `json:"verification_id" validate:"required"`
	Code           string `json:"code" validate:"required"`
	State          string `json:"state" validate:"required"`
	RedirectURI    string `json:"redirect_uri"`
}

// CommandName implements cqrs.Command.
func (c ReceiveOAuthCallback) CommandName() string {
	return CommandReceiveOAuthCallback
}

// Validate implements cqrs.Validatable.
func (c ReceiveOAuthCallback) Validate() error {
	if c.VerificationID == "" {
		return cqrs.ErrCommandValidation("ReceiveOAuthCallback.Validate", "verification_id is required")
	}
	if c.Code == "" {
		return cqrs.ErrCommandValidation("ReceiveOAuthCallback.Validate", "code is required")
	}
	if c.State == "" {
		return cqrs.ErrCommandValidation("ReceiveOAuthCallback.Validate", "state is required")
	}
	return nil
}

// ReceiveOAuthCallbackResult is the result data for ReceiveOAuthCallback.
type ReceiveOAuthCallbackResult struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
	CredentialID   string `json:"credential_id,omitempty"`
}

// ============================================================================
// CompleteVerification
// ============================================================================

// CompleteVerification marks a verification as completed with a credential.
type CompleteVerification struct {
	VerificationID string `json:"verification_id" validate:"required"`
	CredentialID   string `json:"credential_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c CompleteVerification) CommandName() string {
	return CommandCompleteVerification
}

// Validate implements cqrs.Validatable.
func (c CompleteVerification) Validate() error {
	if c.VerificationID == "" {
		return cqrs.ErrCommandValidation("CompleteVerification.Validate", "verification_id is required")
	}
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation("CompleteVerification.Validate", "credential_id is required")
	}
	return nil
}

// CompleteVerificationResult is the result data for CompleteVerification.
type CompleteVerificationResult struct {
	VerificationID string `json:"verification_id"`
	CredentialID   string `json:"credential_id"`
	Status         string `json:"status"`
}

// ============================================================================
// FailVerification
// ============================================================================

// FailVerification marks a verification as failed.
type FailVerification struct {
	VerificationID string `json:"verification_id" validate:"required"`
	ErrorMessage   string `json:"error_message" validate:"required"`
	ErrorCode      string `json:"error_code" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c FailVerification) CommandName() string {
	return CommandFailVerification
}

// Validate implements cqrs.Validatable.
func (c FailVerification) Validate() error {
	if c.VerificationID == "" {
		return cqrs.ErrCommandValidation("FailVerification.Validate", "verification_id is required")
	}
	if c.ErrorMessage == "" {
		return cqrs.ErrCommandValidation("FailVerification.Validate", "error_message is required")
	}
	if c.ErrorCode == "" {
		return cqrs.ErrCommandValidation("FailVerification.Validate", "error_code is required")
	}
	return nil
}

// FailVerificationResult is the result data for FailVerification.
type FailVerificationResult struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
}
