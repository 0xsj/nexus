package command

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Command Types
// ============================================================================

const (
	TypeRequestCredential   = "credential.request"
	TypeIssueCredential     = "credential.issue"
	TypeRevokeCredential    = "credential.revoke"
	TypeSuspendCredential   = "credential.suspend"
	TypeReinstateCredential = "credential.reinstate"
)

// ============================================================================
// RequestCredential Command
// ============================================================================

// RequestCredential initiates a credential request from a holder.
type RequestCredential struct {
	cqrs.BaseCommand

	CredentialID   string         `json:"credential_id"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	Claims         map[string]any `json:"claims,omitempty"`
}

// CommandType returns the command type.
func (c *RequestCredential) CommandType() string {
	return TypeRequestCredential
}

// Validate validates the command.
func (c *RequestCredential) Validate() error {
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation(TypeRequestCredential, "credential_id is required")
	}
	if c.HolderDID == "" {
		return cqrs.ErrCommandValidation(TypeRequestCredential, "holder_did is required")
	}
	if c.IssuerDID == "" {
		return cqrs.ErrCommandValidation(TypeRequestCredential, "issuer_did is required")
	}
	if c.CredentialType == "" {
		return cqrs.ErrCommandValidation(TypeRequestCredential, "credential_type is required")
	}
	return nil
}

// NewRequestCredential creates a new RequestCredential command.
func NewRequestCredential(credentialID, holderDID, issuerDID, credentialType string, claims map[string]any) *RequestCredential {
	return &RequestCredential{
		CredentialID:   credentialID,
		HolderDID:      holderDID,
		IssuerDID:      issuerDID,
		CredentialType: credentialType,
		Claims:         claims,
	}
}

// ============================================================================
// IssueCredential Command
// ============================================================================

// IssueCredential issues a credential to a holder.
type IssueCredential struct {
	cqrs.BaseCommand

	CredentialID   string         `json:"credential_id"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	Claims         map[string]any `json:"claims"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	SchemaID       string         `json:"schema_id,omitempty"`
}

// CommandType returns the command type.
func (c *IssueCredential) CommandType() string {
	return TypeIssueCredential
}

// Validate validates the command.
func (c *IssueCredential) Validate() error {
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation(TypeIssueCredential, "credential_id is required")
	}
	if c.HolderDID == "" {
		return cqrs.ErrCommandValidation(TypeIssueCredential, "holder_did is required")
	}
	if c.IssuerDID == "" {
		return cqrs.ErrCommandValidation(TypeIssueCredential, "issuer_did is required")
	}
	if c.CredentialType == "" {
		return cqrs.ErrCommandValidation(TypeIssueCredential, "credential_type is required")
	}
	if len(c.Claims) == 0 {
		return cqrs.ErrCommandValidation(TypeIssueCredential, "claims are required")
	}
	return nil
}

// NewIssueCredential creates a new IssueCredential command.
func NewIssueCredential(credentialID, holderDID, issuerDID, credentialType string, claims map[string]any) *IssueCredential {
	return &IssueCredential{
		CredentialID:   credentialID,
		HolderDID:      holderDID,
		IssuerDID:      issuerDID,
		CredentialType: credentialType,
		Claims:         claims,
	}
}

// WithExpiration sets the expiration time.
func (c *IssueCredential) WithExpiration(expiresAt time.Time) *IssueCredential {
	c.ExpiresAt = &expiresAt
	return c
}

// WithSchemaID sets the schema ID.
func (c *IssueCredential) WithSchemaID(schemaID string) *IssueCredential {
	c.SchemaID = schemaID
	return c
}

// ============================================================================
// RevokeCredential Command
// ============================================================================

// RevokeCredential permanently revokes a credential.
type RevokeCredential struct {
	cqrs.BaseCommand

	CredentialID string `json:"credential_id"`
	RevokedBy    string `json:"revoked_by"`
	Reason       string `json:"reason"`
}

// CommandType returns the command type.
func (c *RevokeCredential) CommandType() string {
	return TypeRevokeCredential
}

// Validate validates the command.
func (c *RevokeCredential) Validate() error {
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation(TypeRevokeCredential, "credential_id is required")
	}
	if c.RevokedBy == "" {
		return cqrs.ErrCommandValidation(TypeRevokeCredential, "revoked_by is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation(TypeRevokeCredential, "reason is required")
	}
	return nil
}

// NewRevokeCredential creates a new RevokeCredential command.
func NewRevokeCredential(credentialID, revokedBy, reason string) *RevokeCredential {
	return &RevokeCredential{
		CredentialID: credentialID,
		RevokedBy:    revokedBy,
		Reason:       reason,
	}
}

// ============================================================================
// SuspendCredential Command
// ============================================================================

// SuspendCredential temporarily suspends a credential.
type SuspendCredential struct {
	cqrs.BaseCommand

	CredentialID string     `json:"credential_id"`
	SuspendedBy  string     `json:"suspended_by"`
	Reason       string     `json:"reason"`
	Until        *time.Time `json:"until,omitempty"`
}

// CommandType returns the command type.
func (c *SuspendCredential) CommandType() string {
	return TypeSuspendCredential
}

// Validate validates the command.
func (c *SuspendCredential) Validate() error {
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation(TypeSuspendCredential, "credential_id is required")
	}
	if c.SuspendedBy == "" {
		return cqrs.ErrCommandValidation(TypeSuspendCredential, "suspended_by is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation(TypeSuspendCredential, "reason is required")
	}
	return nil
}

// NewSuspendCredential creates a new SuspendCredential command.
func NewSuspendCredential(credentialID, suspendedBy, reason string) *SuspendCredential {
	return &SuspendCredential{
		CredentialID: credentialID,
		SuspendedBy:  suspendedBy,
		Reason:       reason,
	}
}

// WithUntil sets the suspension end time.
func (c *SuspendCredential) WithUntil(until time.Time) *SuspendCredential {
	c.Until = &until
	return c
}

// ============================================================================
// ReinstateCredential Command
// ============================================================================

// ReinstateCredential reinstates a suspended credential.
type ReinstateCredential struct {
	cqrs.BaseCommand

	CredentialID string `json:"credential_id"`
	ReinstatedBy string `json:"reinstated_by"`
	Reason       string `json:"reason,omitempty"`
}

// CommandType returns the command type.
func (c *ReinstateCredential) CommandType() string {
	return TypeReinstateCredential
}

// Validate validates the command.
func (c *ReinstateCredential) Validate() error {
	if c.CredentialID == "" {
		return cqrs.ErrCommandValidation(TypeReinstateCredential, "credential_id is required")
	}
	if c.ReinstatedBy == "" {
		return cqrs.ErrCommandValidation(TypeReinstateCredential, "reinstated_by is required")
	}
	return nil
}

// NewReinstateCredential creates a new ReinstateCredential command.
func NewReinstateCredential(credentialID, reinstatedBy, reason string) *ReinstateCredential {
	return &ReinstateCredential{
		CredentialID: credentialID,
		ReinstatedBy: reinstatedBy,
		Reason:       reason,
	}
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.Command     = (*RequestCredential)(nil)
	_ cqrs.Validatable = (*RequestCredential)(nil)
	_ cqrs.Command     = (*IssueCredential)(nil)
	_ cqrs.Validatable = (*IssueCredential)(nil)
	_ cqrs.Command     = (*RevokeCredential)(nil)
	_ cqrs.Validatable = (*RevokeCredential)(nil)
	_ cqrs.Command     = (*SuspendCredential)(nil)
	_ cqrs.Validatable = (*SuspendCredential)(nil)
	_ cqrs.Command     = (*ReinstateCredential)(nil)
	_ cqrs.Validatable = (*ReinstateCredential)(nil)
)
