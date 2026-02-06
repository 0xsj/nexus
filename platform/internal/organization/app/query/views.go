package query

import "time"

// ============================================================================
// Organization Views
// ============================================================================

// OrganizationView is the full organization read model.
type OrganizationView struct {
	OrganizationID     string    `json:"organization_id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	OrgType            string    `json:"org_type"`
	Description        string    `json:"description,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	DID                string    `json:"did,omitempty"`
	OwnerMemberID      string    `json:"owner_member_id"`
	MemberCount        int       `json:"member_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// OrganizationSummaryView is a lightweight organization representation for lists.
type OrganizationSummaryView struct {
	OrganizationID     string    `json:"organization_id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	OrgType            string    `json:"org_type"`
	VerificationStatus string    `json:"verification_status"`
	MemberCount        int       `json:"member_count"`
	CreatedAt          time.Time `json:"created_at"`
}

// OrganizationListView is a paginated list of organizations.
type OrganizationListView struct {
	Organizations []OrganizationSummaryView `json:"organizations"`
	TotalCount    int                       `json:"total_count"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
	HasMore       bool                      `json:"has_more"`
}

// MemberView is a lightweight member representation.
type MemberView struct {
	MemberID string    `json:"member_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
