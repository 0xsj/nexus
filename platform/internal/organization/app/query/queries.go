package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetOrganization       = "organization.GetOrganization"
	QueryGetOrganizationBySlug = "organization.GetOrganizationBySlug"
	QueryListOrganizations     = "organization.ListOrganizations"
)

// ============================================================================
// GetOrganization
// ============================================================================

// GetOrganization retrieves a single organization by ID.
type GetOrganization struct {
	OrganizationID types.ID `json:"organization_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetOrganization) QueryName() string {
	return QueryGetOrganization
}

// Validate implements cqrs.Validatable.
func (q GetOrganization) Validate() error {
	if q.OrganizationID.IsZero() {
		return cqrs.ErrQueryValidation("GetOrganization.Validate", "organization_id is required")
	}
	return nil
}

// ============================================================================
// GetOrganizationBySlug
// ============================================================================

// GetOrganizationBySlug retrieves an organization by its slug.
type GetOrganizationBySlug struct {
	Slug string `json:"slug" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetOrganizationBySlug) QueryName() string {
	return QueryGetOrganizationBySlug
}

// Validate implements cqrs.Validatable.
func (q GetOrganizationBySlug) Validate() error {
	if q.Slug == "" {
		return cqrs.ErrQueryValidation("GetOrganizationBySlug.Validate", "slug is required")
	}
	return nil
}

// ============================================================================
// ListOrganizations
// ============================================================================

// ListOrganizations lists organizations with pagination.
type ListOrganizations struct {
	Limit  int `json:"limit" validate:"omitempty"`
	Offset int `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListOrganizations) QueryName() string {
	return QueryListOrganizations
}

// Validate implements cqrs.Validatable.
func (q ListOrganizations) Validate() error {
	return nil
}
