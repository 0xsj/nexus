package v1

import "time"

// ============================================================================
// Vouch Responses
// ============================================================================

// VouchResponse represents a full vouch in API responses.
type VouchResponse struct {
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

// VouchSummaryResponse represents a lightweight vouch in list responses.
type VouchSummaryResponse struct {
	VouchID      string     `json:"vouch_id"`
	VoucherID    string     `json:"voucher_id"`
	VoucheeID    string     `json:"vouchee_id"`
	Relationship string     `json:"relationship"`
	Strength     int        `json:"strength"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// VouchListResponse represents a paginated list of vouches.
type VouchListResponse struct {
	Vouches    []VouchSummaryResponse `json:"vouches"`
	TotalCount int                    `json:"total_count"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	HasMore    bool                   `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// VouchGivenResponse represents the result of giving a vouch.
type VouchGivenResponse struct {
	VouchID   string    `json:"vouch_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// VouchAcceptedResponse represents the result of accepting a vouch.
type VouchAcceptedResponse struct {
	VouchID    string    `json:"vouch_id"`
	Status     string    `json:"status"`
	AcceptedAt time.Time `json:"accepted_at"`
}

// VouchRevokedResponse represents the result of revoking a vouch.
type VouchRevokedResponse struct {
	VouchID   string    `json:"vouch_id"`
	Status    string    `json:"status"`
	RevokedAt time.Time `json:"revoked_at"`
}

// ============================================================================
// Reputation Responses
// ============================================================================

// ReputationResponse represents a reputation score in API responses.
type ReputationResponse struct {
	UserID           string    `json:"user_id"`
	OverallScore     int       `json:"overall_score"`
	CredentialScore  int       `json:"credential_score"`
	VouchScore       int       `json:"vouch_score"`
	NetworkScore     int       `json:"network_score"`
	VouchCount       int       `json:"vouch_count"`
	LastCalculatedAt time.Time `json:"last_calculated_at"`
}
