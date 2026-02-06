package v1

import "time"

// ============================================================================
// Credential Requests
// ============================================================================

// IssueCredentialRequest represents a request to issue a new credential.
type IssueCredentialRequest struct {
	CredentialType string         `json:"credential_type" validate:"required"`
	IssuerDID      string         `json:"issuer_did" validate:"required"`
	SubjectDID     string         `json:"subject_did" validate:"required"`
	Claims         map[string]any `json:"claims" validate:"required"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	VerificationID string         `json:"verification_id,omitempty"`
}

// RevokeCredentialRequest represents a request to revoke a credential.
type RevokeCredentialRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}
