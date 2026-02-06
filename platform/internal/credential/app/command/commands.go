package command

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandIssueCredential  = "credential.IssueCredential"
	CommandRevokeCredential = "credential.RevokeCredential"
	CommandExpireCredential = "credential.ExpireCredential"
)

// ============================================================================
// IssueCredential
// ============================================================================

// IssueCredential creates a new verifiable credential.
type IssueCredential struct {
	CredentialType string         `json:"credential_type" validate:"required"`
	IssuerDID      string         `json:"issuer_did" validate:"required"`
	SubjectDID     string         `json:"subject_did" validate:"required"`
	Claims         map[string]any `json:"claims" validate:"required"`
	ExpiresAt      time.Time      `json:"expires_at" validate:"omitempty"`
	VerificationID string         `json:"verification_id" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c IssueCredential) CommandName() string {
	return CommandIssueCredential
}

// Validate implements cqrs.Validatable.
func (c IssueCredential) Validate() error {
	if c.CredentialType == "" {
		return cqrs.ErrCommandValidation("IssueCredential.Validate", "credential_type is required")
	}
	if c.IssuerDID == "" {
		return cqrs.ErrCommandValidation("IssueCredential.Validate", "issuer_did is required")
	}
	if c.SubjectDID == "" {
		return cqrs.ErrCommandValidation("IssueCredential.Validate", "subject_did is required")
	}
	if len(c.Claims) == 0 {
		return cqrs.ErrCommandValidation("IssueCredential.Validate", "claims are required")
	}
	if !c.ExpiresAt.IsZero() && c.ExpiresAt.Before(time.Now()) {
		return cqrs.ErrCommandValidation("IssueCredential.Validate", "expires_at must be in the future")
	}
	return nil
}

// IssueCredentialResult is the result data for IssueCredential.
type IssueCredentialResult struct {
	CredentialID   string `json:"credential_id"`
	CredentialType string `json:"credential_type"`
	Status         string `json:"status"`
}

// ============================================================================
// RevokeCredential
// ============================================================================

// RevokeCredential revokes an active credential.
type RevokeCredential struct {
	CredentialID types.ID `json:"credential_id" validate:"required"`
	Reason       string   `json:"reason" validate:"required,max=500"`
}

// CommandName implements cqrs.Command.
func (c RevokeCredential) CommandName() string {
	return CommandRevokeCredential
}

// Validate implements cqrs.Validatable.
func (c RevokeCredential) Validate() error {
	if c.CredentialID.IsZero() {
		return cqrs.ErrCommandValidation("RevokeCredential.Validate", "credential_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("RevokeCredential.Validate", "reason is required")
	}
	if len(c.Reason) > 500 {
		return cqrs.ErrCommandValidation("RevokeCredential.Validate", "reason must be 500 characters or less")
	}
	return nil
}

// RevokeCredentialResult is the result data for RevokeCredential.
type RevokeCredentialResult struct {
	CredentialID string `json:"credential_id"`
	Status       string `json:"status"`
}

// ============================================================================
// ExpireCredential
// ============================================================================

// ExpireCredential marks an active credential as expired.
type ExpireCredential struct {
	CredentialID types.ID `json:"credential_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ExpireCredential) CommandName() string {
	return CommandExpireCredential
}

// Validate implements cqrs.Validatable.
func (c ExpireCredential) Validate() error {
	if c.CredentialID.IsZero() {
		return cqrs.ErrCommandValidation("ExpireCredential.Validate", "credential_id is required")
	}
	return nil
}

// ExpireCredentialResult is the result data for ExpireCredential.
type ExpireCredentialResult struct {
	CredentialID string `json:"credential_id"`
	Status       string `json:"status"`
}
