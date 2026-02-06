package command

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandCreateOrganization   = "organization.CreateOrganization"
	CommandUpdateOrganization   = "organization.UpdateOrganization"
	CommandAddMember            = "organization.AddMember"
	CommandRemoveMember         = "organization.RemoveMember"
	CommandChangeMemberRole     = "organization.ChangeMemberRole"
	CommandTransferOwnership    = "organization.TransferOwnership"
	CommandCompleteVerification = "organization.CompleteVerification"
	CommandDeleteOrganization   = "organization.DeleteOrganization"
)

// ============================================================================
// CreateOrganization
// ============================================================================

// CreateOrganization creates a new organization with an owner.
type CreateOrganization struct {
	Name        string `json:"name" validate:"required"`
	Slug        string `json:"slug" validate:"required"`
	OrgType     string `json:"org_type" validate:"required"`
	OwnerUserID string `json:"owner_user_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c CreateOrganization) CommandName() string {
	return CommandCreateOrganization
}

// Validate implements cqrs.Validatable.
func (c CreateOrganization) Validate() error {
	if c.Name == "" {
		return cqrs.ErrCommandValidation("CreateOrganization.Validate", "name is required")
	}
	if c.Slug == "" {
		return cqrs.ErrCommandValidation("CreateOrganization.Validate", "slug is required")
	}
	if c.OrgType == "" {
		return cqrs.ErrCommandValidation("CreateOrganization.Validate", "org_type is required")
	}
	if c.OwnerUserID == "" {
		return cqrs.ErrCommandValidation("CreateOrganization.Validate", "owner_user_id is required")
	}
	return nil
}

// CreateOrganizationResult is the result data for CreateOrganization.
type CreateOrganizationResult struct {
	OrganizationID string `json:"organization_id"`
	Slug           string `json:"slug"`
	OwnerMemberID  string `json:"owner_member_id"`
}

// ============================================================================
// UpdateOrganization
// ============================================================================

// UpdateOrganization updates an organization's metadata.
type UpdateOrganization struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
	Name           string   `json:"name" validate:"omitempty"`
	Description    string   `json:"description" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c UpdateOrganization) CommandName() string {
	return CommandUpdateOrganization
}

// Validate implements cqrs.Validatable.
func (c UpdateOrganization) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("UpdateOrganization.Validate", "organization_id is required")
	}
	if c.Name == "" && c.Description == "" {
		return cqrs.ErrCommandValidation("UpdateOrganization.Validate", "at least one of name or description is required")
	}
	return nil
}

// UpdateOrganizationResult is the result data for UpdateOrganization.
type UpdateOrganizationResult struct {
	OrganizationID string `json:"organization_id"`
}

// ============================================================================
// AddMember
// ============================================================================

// AddMember adds a new member to an organization.
type AddMember struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
	UserID         string   `json:"user_id" validate:"required"`
	Role           string   `json:"role" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c AddMember) CommandName() string {
	return CommandAddMember
}

// Validate implements cqrs.Validatable.
func (c AddMember) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("AddMember.Validate", "organization_id is required")
	}
	if c.UserID == "" {
		return cqrs.ErrCommandValidation("AddMember.Validate", "user_id is required")
	}
	if c.Role == "" {
		return cqrs.ErrCommandValidation("AddMember.Validate", "role is required")
	}
	return nil
}

// AddMemberResult is the result data for AddMember.
type AddMemberResult struct {
	OrganizationID string `json:"organization_id"`
	MemberID       string `json:"member_id"`
	Role           string `json:"role"`
}

// ============================================================================
// RemoveMember
// ============================================================================

// RemoveMember removes a member from an organization.
type RemoveMember struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
	MemberID       types.ID `json:"member_id" validate:"required"`
	Reason         string   `json:"reason" validate:"omitempty,max=500"`
}

// CommandName implements cqrs.Command.
func (c RemoveMember) CommandName() string {
	return CommandRemoveMember
}

// Validate implements cqrs.Validatable.
func (c RemoveMember) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("RemoveMember.Validate", "organization_id is required")
	}
	if c.MemberID.IsZero() {
		return cqrs.ErrCommandValidation("RemoveMember.Validate", "member_id is required")
	}
	return nil
}

// RemoveMemberResult is the result data for RemoveMember.
type RemoveMemberResult struct {
	OrganizationID string `json:"organization_id"`
	MemberID       string `json:"member_id"`
}

// ============================================================================
// ChangeMemberRole
// ============================================================================

// ChangeMemberRole changes a member's role within an organization.
type ChangeMemberRole struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
	MemberID       types.ID `json:"member_id" validate:"required"`
	NewRole        string   `json:"new_role" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ChangeMemberRole) CommandName() string {
	return CommandChangeMemberRole
}

// Validate implements cqrs.Validatable.
func (c ChangeMemberRole) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("ChangeMemberRole.Validate", "organization_id is required")
	}
	if c.MemberID.IsZero() {
		return cqrs.ErrCommandValidation("ChangeMemberRole.Validate", "member_id is required")
	}
	if c.NewRole == "" {
		return cqrs.ErrCommandValidation("ChangeMemberRole.Validate", "new_role is required")
	}
	return nil
}

// ChangeMemberRoleResult is the result data for ChangeMemberRole.
type ChangeMemberRoleResult struct {
	OrganizationID string `json:"organization_id"`
	MemberID       string `json:"member_id"`
	NewRole        string `json:"new_role"`
}

// ============================================================================
// TransferOwnership
// ============================================================================

// TransferOwnership transfers organization ownership to another member.
type TransferOwnership struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
	ToMemberID     types.ID `json:"to_member_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c TransferOwnership) CommandName() string {
	return CommandTransferOwnership
}

// Validate implements cqrs.Validatable.
func (c TransferOwnership) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("TransferOwnership.Validate", "organization_id is required")
	}
	if c.ToMemberID.IsZero() {
		return cqrs.ErrCommandValidation("TransferOwnership.Validate", "to_member_id is required")
	}
	return nil
}

// TransferOwnershipResult is the result data for TransferOwnership.
type TransferOwnershipResult struct {
	OrganizationID string `json:"organization_id"`
	NewOwnerID     string `json:"new_owner_id"`
}

// ============================================================================
// CompleteVerification
// ============================================================================

// CompleteVerification marks an organization as verified and assigns a DID.
type CompleteVerification struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
	DID            string   `json:"did" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c CompleteVerification) CommandName() string {
	return CommandCompleteVerification
}

// Validate implements cqrs.Validatable.
func (c CompleteVerification) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("CompleteVerification.Validate", "organization_id is required")
	}
	if c.DID == "" {
		return cqrs.ErrCommandValidation("CompleteVerification.Validate", "did is required")
	}
	return nil
}

// CompleteVerificationResult is the result data for CompleteVerification.
type CompleteVerificationResult struct {
	OrganizationID     string `json:"organization_id"`
	DID                string `json:"did"`
	VerificationStatus string `json:"verification_status"`
}

// ============================================================================
// DeleteOrganization
// ============================================================================

// DeleteOrganization soft-deletes an organization.
type DeleteOrganization struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c DeleteOrganization) CommandName() string {
	return CommandDeleteOrganization
}

// Validate implements cqrs.Validatable.
func (c DeleteOrganization) Validate() error {
	if c.OrganizationID.IsZero() {
		return cqrs.ErrCommandValidation("DeleteOrganization.Validate", "organization_id is required")
	}
	return nil
}

// DeleteOrganizationResult is the result data for DeleteOrganization.
type DeleteOrganizationResult struct {
	OrganizationID string `json:"organization_id"`
}
