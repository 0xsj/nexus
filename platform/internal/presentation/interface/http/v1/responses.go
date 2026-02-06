package v1

import "time"

// ============================================================================
// Presentation Responses
// ============================================================================

// PresentationResponse represents a full presentation in API responses.
type PresentationResponse struct {
	PresentationID   string         `json:"presentation_id"`
	HolderDID        string         `json:"holder_did"`
	CredentialIDs    []string       `json:"credential_ids"`
	DisclosurePolicy map[string]any `json:"disclosure_policy"`
	VPJWT            string         `json:"vp_jwt,omitempty"`
	Purpose          string         `json:"purpose,omitempty"`
	Status           string         `json:"status"`
	RevokedAt        *time.Time     `json:"revoked_at,omitempty"`
	RevocationReason string         `json:"revocation_reason,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// PresentationSummaryResponse represents a lightweight presentation in list responses.
type PresentationSummaryResponse struct {
	PresentationID string    `json:"presentation_id"`
	HolderDID      string    `json:"holder_did"`
	Purpose        string    `json:"purpose,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// PresentationListResponse represents a paginated list of presentations.
type PresentationListResponse struct {
	Presentations []PresentationSummaryResponse `json:"presentations"`
	TotalCount    int                           `json:"total_count"`
	Limit         int                           `json:"limit"`
	Offset        int                           `json:"offset"`
	HasMore       bool                          `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// PresentationCreatedResponse represents the result of creating a presentation.
type PresentationCreatedResponse struct {
	PresentationID string    `json:"presentation_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// PresentationRevokedResponse represents the result of revoking a presentation.
type PresentationRevokedResponse struct {
	PresentationID string    `json:"presentation_id"`
	Status         string    `json:"status"`
	RevokedAt      time.Time `json:"revoked_at"`
}

// ============================================================================
// ShareLink Responses
// ============================================================================

// ShareLinkResponse represents a full share link in API responses.
type ShareLinkResponse struct {
	ShareLinkID    string     `json:"share_link_id"`
	PresentationID string     `json:"presentation_id"`
	Token          string     `json:"token"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	MaxViews       int        `json:"max_views"`
	CurrentViews   int        `json:"current_views"`
	PinProtected   bool       `json:"pin_protected"`
	Audience       string     `json:"audience,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ShareLinkListResponse represents a paginated list of share links.
type ShareLinkListResponse struct {
	ShareLinks []ShareLinkResponse `json:"share_links"`
	TotalCount int                 `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	HasMore    bool                `json:"has_more"`
}

// ShareLinkCreatedResponse represents the result of creating a share link.
type ShareLinkCreatedResponse struct {
	ShareLinkID    string    `json:"share_link_id"`
	PresentationID string    `json:"presentation_id"`
	Token          string    `json:"token"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// ShareLinkAccessedResponse represents the result of accessing a share link.
type ShareLinkAccessedResponse struct {
	ShareLinkID  string `json:"share_link_id"`
	CurrentViews int    `json:"current_views"`
	Status       string `json:"status"`
}

// ShareLinkRevokedResponse represents the result of revoking a share link.
type ShareLinkRevokedResponse struct {
	ShareLinkID string    `json:"share_link_id"`
	Status      string    `json:"status"`
	RevokedAt   time.Time `json:"revoked_at"`
}

// ============================================================================
// Access Log Responses
// ============================================================================

// AccessGrantResponse represents a single access grant in API responses.
type AccessGrantResponse struct {
	ID              string         `json:"id"`
	ShareLinkID     string         `json:"share_link_id"`
	VerifierDID     string         `json:"verifier_did,omitempty"`
	AccessedAt      time.Time      `json:"accessed_at"`
	IPAddress       string         `json:"ip_address,omitempty"`
	DisclosedClaims map[string]any `json:"disclosed_claims,omitempty"`
}

// AccessLogResponse represents a paginated access log.
type AccessLogResponse struct {
	AccessGrants []AccessGrantResponse `json:"access_grants"`
	TotalCount   int                   `json:"total_count"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
	HasMore      bool                  `json:"has_more"`
}
