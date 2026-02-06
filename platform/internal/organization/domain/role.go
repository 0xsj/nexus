package domain

import "fmt"

// Role represents a member's role within an organization.
type Role int

const (
	RoleViewer Role = 1
	RoleMember Role = 2
	RoleAdmin  Role = 3
	RoleOwner  Role = 4
)

// String returns the string representation of the role.
func (r Role) String() string {
	switch r {
	case RoleViewer:
		return "viewer"
	case RoleMember:
		return "member"
	case RoleAdmin:
		return "admin"
	case RoleOwner:
		return "owner"
	default:
		return "unknown"
	}
}

// ParseRole parses a string into a Role.
func ParseRole(s string) (Role, error) {
	switch s {
	case "viewer":
		return RoleViewer, nil
	case "member":
		return RoleMember, nil
	case "admin":
		return RoleAdmin, nil
	case "owner":
		return RoleOwner, nil
	default:
		return 0, fmt.Errorf("invalid role: %s", s)
	}
}

// IsValid returns true if the role is a known, valid role.
func (r Role) IsValid() bool {
	switch r {
	case RoleViewer, RoleMember, RoleAdmin, RoleOwner:
		return true
	default:
		return false
	}
}

// CanManageMembers returns true if the role can invite/remove members.
func (r Role) CanManageMembers() bool {
	return r >= RoleAdmin
}

// CanManageSettings returns true if the role can update org settings.
func (r Role) CanManageSettings() bool {
	return r >= RoleAdmin
}

// CanDeleteOrganization returns true if the role can delete the org.
func (r Role) CanDeleteOrganization() bool {
	return r == RoleOwner
}

// CanTransferOwnership returns true if the role can transfer ownership.
func (r Role) CanTransferOwnership() bool {
	return r == RoleOwner
}
