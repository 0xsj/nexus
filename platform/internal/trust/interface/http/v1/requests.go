package v1

import "time"

// ============================================================================
// Trust Requests
// ============================================================================

// GiveVouchRequest represents a request to give a vouch.
type GiveVouchRequest struct {
	VoucherID    string     `json:"voucher_id" validate:"required"`
	VoucheeID    string     `json:"vouchee_id" validate:"required"`
	CredentialID string     `json:"credential_id,omitempty"`
	ClaimKey     string     `json:"claim_key,omitempty"`
	Relationship string     `json:"relationship" validate:"required"`
	Strength     int        `json:"strength" validate:"required,min=1,max=10"`
	Statement    string     `json:"statement,omitempty"`
	Context      string     `json:"context,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// AcceptVouchRequest represents a request to accept a vouch.
type AcceptVouchRequest struct {
	VoucheeID string `json:"vouchee_id" validate:"required"`
}

// RevokeVouchRequest represents a request to revoke a vouch.
type RevokeVouchRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}
