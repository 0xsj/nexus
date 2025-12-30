package query

import (
	"time"
)

// ============================================================================
// Credential View
// ============================================================================

// CredentialView is the read model for a credential.
type CredentialView struct {
	ID             string         `json:"id"`
	CredentialType string         `json:"credential_type"`
	SchemaID       string         `json:"schema_id,omitempty"`
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	Status         string         `json:"status"`
	Claims         map[string]any `json:"claims,omitempty"`
	SignedVC       string         `json:"signed_vc,omitempty"`
	IssuedAt       *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	RevokedAt      *time.Time     `json:"revoked_at,omitempty"`
	SuspendedAt    *time.Time     `json:"suspended_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Version        int            `json:"version"`
}

// HasSignedVC returns true if the credential has a signed VC.
func (v *CredentialView) HasSignedVC() bool {
	return v.SignedVC != ""
}

// ============================================================================
// Credential List Result
// ============================================================================

// CredentialListResult is the result of a list query.
type CredentialListResult struct {
	Credentials []*CredentialView `json:"credentials"`
	Total       int               `json:"total"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
	HasMore     bool              `json:"has_more"`
}
