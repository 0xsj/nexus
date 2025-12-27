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
// Request Credential Command
// ============================================================================

// RequestCredential requests a new credential from an issuer.
type RequestCredential struct {
	CredentialID   string         `json:"credential_id"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	Claims         map[string]any `json:"claims,omitempty"`
}

// CommandName returns the command type.
func (c *RequestCredential) CommandName() string {
	return TypeRequestCredential
}

// NewRequestCredential creates a new RequestCredential command.
func NewRequestCredential(
	credentialID string,
	holderDID string,
	issuerDID string,
	credentialType string,
	claims map[string]any,
) *RequestCredential {
	return &RequestCredential{
		CredentialID:   credentialID,
		HolderDID:      holderDID,
		IssuerDID:      issuerDID,
		CredentialType: credentialType,
		Claims:         claims,
	}
}

// ============================================================================
// Issue Credential Command
// ============================================================================

// IssueCredential issues a credential to a holder.
type IssueCredential struct {
	CredentialID   string         `json:"credential_id"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	SchemaID       string         `json:"schema_id,omitempty"`
	Claims         map[string]any `json:"claims"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
}

// CommandName returns the command type.
func (c *IssueCredential) CommandName() string {
	return TypeIssueCredential
}

// NewIssueCredential creates a new IssueCredential command.
func NewIssueCredential(
	credentialID string,
	holderDID string,
	issuerDID string,
	credentialType string,
	claims map[string]any,
) *IssueCredential {
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
// Revoke Credential Command
// ============================================================================

// RevokeCredential permanently revokes a credential.
type RevokeCredential struct {
	CredentialID string `json:"credential_id"`
	RevokedBy    string `json:"revoked_by"`
	Reason       string `json:"reason"`
}

// CommandName returns the command type.
func (c *RevokeCredential) CommandName() string {
	return TypeRevokeCredential
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
// Suspend Credential Command
// ============================================================================

// SuspendCredential temporarily suspends a credential.
type SuspendCredential struct {
	CredentialID string     `json:"credential_id"`
	SuspendedBy  string     `json:"suspended_by"`
	Reason       string     `json:"reason"`
	Until        *time.Time `json:"until,omitempty"`
}

// CommandName returns the command type.
func (c *SuspendCredential) CommandName() string {
	return TypeSuspendCredential
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
// Reinstate Credential Command
// ============================================================================

// ReinstateCredential reinstates a suspended credential.
type ReinstateCredential struct {
	CredentialID string `json:"credential_id"`
	ReinstatedBy string `json:"reinstated_by"`
	Reason       string `json:"reason,omitempty"`
}

// CommandName returns the command type.
func (c *ReinstateCredential) CommandName() string {
	return TypeReinstateCredential
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
	_ cqrs.Command = (*RequestCredential)(nil)
	_ cqrs.Command = (*IssueCredential)(nil)
	_ cqrs.Command = (*RevokeCredential)(nil)
	_ cqrs.Command = (*SuspendCredential)(nil)
	_ cqrs.Command = (*ReinstateCredential)(nil)
)
