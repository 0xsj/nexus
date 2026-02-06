// Package domain contains the core business logic for the Organization bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes
// ============================================================================

const (
	CodeOrganizationNotFound      pkgerrors.Code = "ORGANIZATION_NOT_FOUND"
	CodeOrganizationAlreadyExists pkgerrors.Code = "ORGANIZATION_ALREADY_EXISTS"
	CodeOrganizationInvalid       pkgerrors.Code = "ORGANIZATION_INVALID"
	CodeOrganizationNotVerified   pkgerrors.Code = "ORGANIZATION_NOT_VERIFIED"
	CodeMemberNotFound            pkgerrors.Code = "MEMBER_NOT_FOUND"
	CodeMemberAlreadyExists       pkgerrors.Code = "MEMBER_ALREADY_EXISTS"
	CodeInvitationNotFound        pkgerrors.Code = "INVITATION_NOT_FOUND"
	CodeInvitationExpired         pkgerrors.Code = "INVITATION_EXPIRED"
	CodeInsufficientPermissions   pkgerrors.Code = "INSUFFICIENT_PERMISSIONS"
	CodeSlugAlreadyTaken          pkgerrors.Code = "SLUG_ALREADY_TAKEN"
	CodeOwnerCannotLeave          pkgerrors.Code = "OWNER_CANNOT_LEAVE"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrOrganizationAlreadyExists = errors.New("organization already exists")
	ErrOrganizationInvalid       = errors.New("organization is invalid")
	ErrOrganizationNotVerified   = errors.New("organization is not verified")
	ErrMemberNotFound            = errors.New("member not found")
	ErrMemberAlreadyExists       = errors.New("member already exists")
	ErrInvitationNotFound        = errors.New("invitation not found")
	ErrInvitationExpired         = errors.New("invitation has expired")
	ErrInsufficientPermissions   = errors.New("insufficient permissions")
	ErrSlugAlreadyTaken          = errors.New("slug already taken")
	ErrOwnerCannotLeave          = errors.New("owner cannot leave organization")
)

// ============================================================================
// Error Constructors
// ============================================================================

// OrganizationNotFound creates an organization not found error.
func OrganizationNotFound(operation string, orgID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "organization").
		WithCode(CodeOrganizationNotFound).
		WithMeta("organization_id", orgID)
}

// OrganizationAlreadyExists creates an organization already exists error.
func OrganizationAlreadyExists(operation string, slug string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "organization").
		WithCode(CodeOrganizationAlreadyExists).
		WithMeta("slug", slug)
}

// OrganizationInvalid creates an organization invalid error.
func OrganizationInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "organization is invalid: "+reason).
		WithCode(CodeOrganizationInvalid)
}

// OrganizationNotVerified creates an organization not verified error.
func OrganizationNotVerified(operation string, orgID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "organization is not verified").
		WithCode(CodeOrganizationNotVerified).
		WithMeta("organization_id", orgID)
}

// MemberNotFound creates a member not found error.
func MemberNotFound(operation string, memberID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "member").
		WithCode(CodeMemberNotFound).
		WithMeta("member_id", memberID)
}

// MemberAlreadyExists creates a member already exists error.
func MemberAlreadyExists(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "member").
		WithCode(CodeMemberAlreadyExists).
		WithMeta("user_id", userID)
}

// InvitationNotFound creates an invitation not found error.
func InvitationNotFound(operation string, invitationID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "invitation").
		WithCode(CodeInvitationNotFound).
		WithMeta("invitation_id", invitationID)
}

// InvitationExpired creates an invitation expired error.
func InvitationExpired(operation string, invitationID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "invitation has expired").
		WithCode(CodeInvitationExpired).
		WithMeta("invitation_id", invitationID)
}

// InsufficientPermissions creates an insufficient permissions error.
func InsufficientPermissions(operation string, userID string, requiredRole string) *pkgerrors.Error {
	return pkgerrors.Forbidden(operation, "insufficient permissions").
		WithCode(CodeInsufficientPermissions).
		WithMeta("user_id", userID).
		WithMeta("required_role", requiredRole)
}

// SlugAlreadyTaken creates a slug already taken error.
func SlugAlreadyTaken(operation string, slug string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "slug").
		WithCode(CodeSlugAlreadyTaken).
		WithMeta("slug", slug)
}

// OwnerCannotLeave creates an owner cannot leave error.
func OwnerCannotLeave(operation string, orgID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "owner cannot leave organization without transferring ownership").
		WithCode(CodeOwnerCannotLeave).
		WithMeta("organization_id", orgID)
}
