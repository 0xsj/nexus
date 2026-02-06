package query

import "time"

// ============================================================================
// Credential Views
// ============================================================================

// CredentialView is the full credential read model.
type CredentialView struct {
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

// CredentialSummaryView is a lightweight credential representation for lists.
type CredentialSummaryView struct {
	CredentialID   string     `json:"credential_id"`
	CredentialType string     `json:"credential_type"`
	IssuerDID      string     `json:"issuer_did"`
	SubjectDID     string     `json:"subject_did"`
	Status         string     `json:"status"`
	IssuedAt       time.Time  `json:"issued_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// CredentialListView is a paginated list of credentials.
type CredentialListView struct {
	Credentials []CredentialSummaryView `json:"credentials"`
	TotalCount  int                     `json:"total_count"`
	Limit       int                     `json:"limit"`
	Offset      int                     `json:"offset"`
	HasMore     bool                    `json:"has_more"`
}
