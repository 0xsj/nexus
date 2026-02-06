package query

import "time"

// ============================================================================
// Presentation Views
// ============================================================================

// PresentationView is the full presentation read model.
type PresentationView struct {
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

// PresentationSummaryView is a lightweight presentation representation for lists.
type PresentationSummaryView struct {
	PresentationID string    `json:"presentation_id"`
	HolderDID      string    `json:"holder_did"`
	Purpose        string    `json:"purpose,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// PresentationListView is a paginated list of presentations.
type PresentationListView struct {
	Presentations []PresentationSummaryView `json:"presentations"`
	TotalCount    int                       `json:"total_count"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
	HasMore       bool                      `json:"has_more"`
}

// ============================================================================
// ShareLink Views
// ============================================================================

// ShareLinkView is the full share link read model.
type ShareLinkView struct {
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

// ShareLinkListView is a paginated list of share links.
type ShareLinkListView struct {
	ShareLinks []ShareLinkView `json:"share_links"`
	TotalCount int             `json:"total_count"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	HasMore    bool            `json:"has_more"`
}

// ============================================================================
// Access Grant Views
// ============================================================================

// AccessGrantView represents a single access grant in the log.
type AccessGrantView struct {
	ID              string         `json:"id"`
	ShareLinkID     string         `json:"share_link_id"`
	VerifierDID     string         `json:"verifier_did,omitempty"`
	AccessedAt      time.Time      `json:"accessed_at"`
	IPAddress       string         `json:"ip_address,omitempty"`
	DisclosedClaims map[string]any `json:"disclosed_claims,omitempty"`
}

// AccessLogView is a paginated list of access grants.
type AccessLogView struct {
	AccessGrants []AccessGrantView `json:"access_grants"`
	TotalCount   int               `json:"total_count"`
	Limit        int               `json:"limit"`
	Offset       int               `json:"offset"`
	HasMore      bool              `json:"has_more"`
}
