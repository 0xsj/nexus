package v1

import "time"

// ============================================================================
// Organization Responses
// ============================================================================

// OrganizationResponse represents a full organization in API responses.
type OrganizationResponse struct {
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

// OrganizationSummaryResponse represents a lightweight organization in list responses.
type OrganizationSummaryResponse struct {
	OrganizationID     string    `json:"organization_id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	OrgType            string    `json:"org_type"`
	VerificationStatus string    `json:"verification_status"`
	MemberCount        int       `json:"member_count"`
	CreatedAt          time.Time `json:"created_at"`
}

// OrganizationListResponse represents a paginated list of organizations.
type OrganizationListResponse struct {
	Organizations []OrganizationSummaryResponse `json:"organizations"`
	TotalCount    int                           `json:"total_count"`
	Limit         int                           `json:"limit"`
	Offset        int                           `json:"offset"`
	HasMore       bool                          `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// OrganizationCreatedResponse represents the result of creating an organization.
type OrganizationCreatedResponse struct {
	OrganizationID string    `json:"organization_id"`
	Slug           string    `json:"slug"`
	OwnerMemberID  string    `json:"owner_member_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// OrganizationUpdatedResponse represents the result of updating an organization.
type OrganizationUpdatedResponse struct {
	OrganizationID string    `json:"organization_id"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MemberAddedResponse represents the result of adding a member.
type MemberAddedResponse struct {
	OrganizationID string    `json:"organization_id"`
	MemberID       string    `json:"member_id"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

// MemberRemovedResponse represents the result of removing a member.
type MemberRemovedResponse struct {
	OrganizationID string    `json:"organization_id"`
	MemberID       string    `json:"member_id"`
	RemovedAt      time.Time `json:"removed_at"`
}

// MemberRoleChangedResponse represents the result of changing a member's role.
type MemberRoleChangedResponse struct {
	OrganizationID string    `json:"organization_id"`
	MemberID       string    `json:"member_id"`
	NewRole        string    `json:"new_role"`
	ChangedAt      time.Time `json:"changed_at"`
}

// OwnershipTransferredResponse represents the result of transferring ownership.
type OwnershipTransferredResponse struct {
	OrganizationID string    `json:"organization_id"`
	NewOwnerID     string    `json:"new_owner_id"`
	TransferredAt  time.Time `json:"transferred_at"`
}

// VerificationCompletedResponse represents the result of completing verification.
type VerificationCompletedResponse struct {
	OrganizationID     string    `json:"organization_id"`
	DID                string    `json:"did"`
	VerificationStatus string    `json:"verification_status"`
	VerifiedAt         time.Time `json:"verified_at"`
}

// OrganizationDeletedResponse represents the result of deleting an organization.
type OrganizationDeletedResponse struct {
	OrganizationID string    `json:"organization_id"`
	DeletedAt      time.Time `json:"deleted_at"`
}
