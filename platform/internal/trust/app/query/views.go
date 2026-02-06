package query

import "time"

// ============================================================================
// Vouch Views
// ============================================================================

// VouchView is the full vouch read model.
type VouchView struct {
	VouchID          string     `json:"vouch_id"`
	VoucherID        string     `json:"voucher_id"`
	VoucheeID        string     `json:"vouchee_id"`
	CredentialID     string     `json:"credential_id,omitempty"`
	ClaimKey         string     `json:"claim_key,omitempty"`
	Relationship     string     `json:"relationship"`
	Strength         int        `json:"strength"`
	Statement        string     `json:"statement,omitempty"`
	Context          string     `json:"context,omitempty"`
	Status           string     `json:"status"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevocationReason string     `json:"revocation_reason,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// VouchSummaryView is a lightweight vouch representation for lists.
type VouchSummaryView struct {
	VouchID      string     `json:"vouch_id"`
	VoucherID    string     `json:"voucher_id"`
	VoucheeID    string     `json:"vouchee_id"`
	Relationship string     `json:"relationship"`
	Strength     int        `json:"strength"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// VouchListView is a paginated list of vouches.
type VouchListView struct {
	Vouches    []VouchSummaryView `json:"vouches"`
	TotalCount int                `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
}

// ============================================================================
// Reputation Views
// ============================================================================

// ReputationView is the reputation read model.
type ReputationView struct {
	UserID           string    `json:"user_id"`
	OverallScore     int       `json:"overall_score"`
	CredentialScore  int       `json:"credential_score"`
	VouchScore       int       `json:"vouch_score"`
	NetworkScore     int       `json:"network_score"`
	VouchCount       int       `json:"vouch_count"`
	LastCalculatedAt time.Time `json:"last_calculated_at"`
}
