package v1

import "time"

// ============================================================================
// Credential Responses
// ============================================================================

// CredentialResponse represents a full credential in API responses.
type CredentialResponse struct {
	CredentialID     string         `json:"credential_id"`
	CredentialType   string         `json:"credential_type"`
	IssuerDID        string         `json:"issuer_did"`
	SubjectDID       string         `json:"subject_did"`
	Claims           map[string]any `json:"claims"`
	IssuedAt         time.Time      `json:"issued_at"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	Status           string         `json:"status"`
	RevokedAt        *time.Time     `json:"revoked_at,omitempty"`
	RevocationReason string         `json:"revocation_reason,omitempty"`
	JWT              string         `json:"jwt,omitempty"`
	VerificationID   string         `json:"verification_id,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// CredentialSummaryResponse represents a lightweight credential in list responses.
type CredentialSummaryResponse struct {
	CredentialID   string     `json:"credential_id"`
	CredentialType string     `json:"credential_type"`
	IssuerDID      string     `json:"issuer_did"`
	SubjectDID     string     `json:"subject_did"`
	Status         string     `json:"status"`
	IssuedAt       time.Time  `json:"issued_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// CredentialListResponse represents a paginated list of credentials.
type CredentialListResponse struct {
	Credentials []CredentialSummaryResponse `json:"credentials"`
	TotalCount  int                         `json:"total_count"`
	Limit       int                         `json:"limit"`
	Offset      int                         `json:"offset"`
	HasMore     bool                        `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// CredentialIssuedResponse represents the result of issuing a credential.
type CredentialIssuedResponse struct {
	CredentialID   string    `json:"credential_id"`
	CredentialType string    `json:"credential_type"`
	Status         string    `json:"status"`
	IssuedAt       time.Time `json:"issued_at"`
}

// CredentialRevokedResponse represents the result of revoking a credential.
type CredentialRevokedResponse struct {
	CredentialID string    `json:"credential_id"`
	Status       string    `json:"status"`
	RevokedAt    time.Time `json:"revoked_at"`
}

// CredentialExpiredResponse represents the result of expiring a credential.
type CredentialExpiredResponse struct {
	CredentialID string    `json:"credential_id"`
	Status       string    `json:"status"`
	ExpiredAt    time.Time `json:"expired_at"`
}
