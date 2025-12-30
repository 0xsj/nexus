package v1

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/http/request"
)

// ============================================================================
// Issue Credential Request
// ============================================================================

// IssueCredentialRequest is the request body for issuing a credential.
type IssueCredentialRequest struct {
	CredentialID   string         `json:"credential_id,omitempty"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	Claims         map[string]any `json:"claims"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	SchemaID       string         `json:"schema_id,omitempty"`
}

// Validate validates the request.
func (r *IssueCredentialRequest) Validate() error {
	v := request.NewValidator()

	v.Required("holder_did", r.HolderDID)
	v.DID("holder_did", r.HolderDID)

	v.Required("issuer_did", r.IssuerDID)
	v.DID("issuer_did", r.IssuerDID)

	v.Required("credential_type", r.CredentialType)
	v.MinLength("credential_type", r.CredentialType, 1)
	v.MaxLength("credential_type", r.CredentialType, 255)

	v.MapNotEmpty("claims", r.Claims)

	if r.ExpiresAt != nil {
		v.NotInPast("expires_at", *r.ExpiresAt)
	}

	return v.Error()
}

// ============================================================================
// Request Credential Request
// ============================================================================

// RequestCredentialRequest is the request body for requesting a credential.
type RequestCredentialRequest struct {
	CredentialID   string         `json:"credential_id,omitempty"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	Claims         map[string]any `json:"claims,omitempty"`
}

// Validate validates the request.
func (r *RequestCredentialRequest) Validate() error {
	v := request.NewValidator()

	v.Required("holder_did", r.HolderDID)
	v.DID("holder_did", r.HolderDID)

	v.Required("issuer_did", r.IssuerDID)
	v.DID("issuer_did", r.IssuerDID)

	v.Required("credential_type", r.CredentialType)
	v.MinLength("credential_type", r.CredentialType, 1)
	v.MaxLength("credential_type", r.CredentialType, 255)

	return v.Error()
}

// ============================================================================
// Revoke Credential Request
// ============================================================================

// RevokeCredentialRequest is the request body for revoking a credential.
type RevokeCredentialRequest struct {
	RevokedBy string `json:"revoked_by"`
	Reason    string `json:"reason"`
}

// Validate validates the request.
func (r *RevokeCredentialRequest) Validate() error {
	v := request.NewValidator()

	v.Required("revoked_by", r.RevokedBy)
	v.DID("revoked_by", r.RevokedBy)

	v.Required("reason", r.Reason)
	v.MinLength("reason", r.Reason, 1)
	v.MaxLength("reason", r.Reason, 1000)

	return v.Error()
}

// ============================================================================
// Suspend Credential Request
// ============================================================================

// SuspendCredentialRequest is the request body for suspending a credential.
type SuspendCredentialRequest struct {
	SuspendedBy string     `json:"suspended_by"`
	Reason      string     `json:"reason"`
	Until       *time.Time `json:"until,omitempty"`
}

// Validate validates the request.
func (r *SuspendCredentialRequest) Validate() error {
	v := request.NewValidator()

	v.Required("suspended_by", r.SuspendedBy)
	v.DID("suspended_by", r.SuspendedBy)

	v.Required("reason", r.Reason)
	v.MinLength("reason", r.Reason, 1)
	v.MaxLength("reason", r.Reason, 1000)

	if r.Until != nil {
		v.NotInPast("until", *r.Until)
	}

	return v.Error()
}

// ============================================================================
// Reinstate Credential Request
// ============================================================================

// ReinstateCredentialRequest is the request body for reinstating a credential.
type ReinstateCredentialRequest struct {
	ReinstatedBy string `json:"reinstated_by"`
	Reason       string `json:"reason,omitempty"`
}

// Validate validates the request.
func (r *ReinstateCredentialRequest) Validate() error {
	v := request.NewValidator()

	v.Required("reinstated_by", r.ReinstatedBy)
	v.DID("reinstated_by", r.ReinstatedBy)

	if r.Reason != "" {
		v.MaxLength("reason", r.Reason, 1000)
	}

	return v.Error()
}

// ============================================================================
// List Credentials Query Params
// ============================================================================

// ListCredentialsParams holds query parameters for listing credentials.
type ListCredentialsParams struct {
	Status         string
	CredentialType string
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string
}

// DefaultListCredentialsParams returns default list parameters.
func DefaultListCredentialsParams() ListCredentialsParams {
	return ListCredentialsParams{
		Limit:     20,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
}

// Validate validates the parameters.
func (p *ListCredentialsParams) Validate() error {
	v := request.NewValidator()

	if p.Status != "" {
		v.OneOf("status", p.Status, []string{"pending", "active", "suspended", "revoked", "expired"})
	}

	v.Between("limit", p.Limit, 1, 100)
	v.NonNegative("offset", p.Offset)

	if p.SortBy != "" {
		v.OneOf("sort_by", p.SortBy, []string{"created_at", "updated_at", "issued_at", "credential_type", "status"})
	}

	if p.SortOrder != "" {
		v.OneOf("sort_order", p.SortOrder, []string{"asc", "desc"})
	}

	return v.Error()
}

// ============================================================================
// Verify Credential Request
// ============================================================================

// VerifyCredentialRequest is the request body for verifying a credential.
type VerifyCredentialRequest struct {
	JWT string `json:"jwt"`
}

// Validate validates the request.
func (r *VerifyCredentialRequest) Validate() error {
	v := request.NewValidator()

	v.Required("jwt", r.JWT)
	v.MinLength("jwt", r.JWT, 10)

	return v.Error()
}
