package v1

// ============================================================================
// Organization Requests
// ============================================================================

// CreateOrganizationRequest represents a request to create a new organization.
type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required"`
	Slug        string `json:"slug" validate:"required"`
	OrgType     string `json:"org_type" validate:"required"`
	OwnerUserID string `json:"owner_user_id" validate:"required"`
}

// UpdateOrganizationRequest represents a request to update an organization.
type UpdateOrganizationRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// AddMemberRequest represents a request to add a member to an organization.
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required"`
}

// RemoveMemberRequest represents a request to remove a member.
type RemoveMemberRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ChangeMemberRoleRequest represents a request to change a member's role.
type ChangeMemberRoleRequest struct {
	NewRole string `json:"new_role" validate:"required"`
}

// TransferOwnershipRequest represents a request to transfer ownership.
type TransferOwnershipRequest struct {
	ToMemberID string `json:"to_member_id" validate:"required"`
}

// CompleteVerificationRequest represents a request to complete verification.
type CompleteVerificationRequest struct {
	DID string `json:"did" validate:"required"`
}
