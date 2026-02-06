package v1

import "time"

// ============================================================================
// Presentation Requests
// ============================================================================

// CreatePresentationRequest represents a request to create a new presentation.
type CreatePresentationRequest struct {
	HolderDID            string   `json:"holder_did" validate:"required"`
	CredentialIDs        []string `json:"credential_ids" validate:"required"`
	DisclosurePolicyType string   `json:"disclosure_policy_type" validate:"required"`
	AllowedClaims        []string `json:"allowed_claims,omitempty"`
	BlockedClaims        []string `json:"blocked_claims,omitempty"`
	Purpose              string   `json:"purpose,omitempty"`
}

// RevokePresentationRequest represents a request to revoke a presentation.
type RevokePresentationRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}

// ============================================================================
// ShareLink Requests
// ============================================================================

// CreateShareLinkRequest represents a request to create a new share link.
type CreateShareLinkRequest struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxViews  int        `json:"max_views,omitempty"`
	Pin       string     `json:"pin,omitempty"`
	Audience  string     `json:"audience,omitempty"`
}

// AccessShareLinkRequest represents a request to access a share link.
type AccessShareLinkRequest struct {
	VerifierDID string `json:"verifier_did,omitempty"`
}
