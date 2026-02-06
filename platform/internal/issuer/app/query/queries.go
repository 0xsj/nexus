package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetIssuer               = "issuer.GetIssuer"
	QueryGetIssuerByOrganization = "issuer.GetIssuerByOrganization"
	QueryListIssuers             = "issuer.ListIssuers"
	QueryGetTemplate             = "issuer.GetTemplate"
	QueryListTemplatesByIssuer   = "issuer.ListTemplatesByIssuer"
)

// ============================================================================
// GetIssuer
// ============================================================================

// GetIssuer retrieves a single issuer by ID.
type GetIssuer struct {
	IssuerID types.ID `json:"issuer_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetIssuer) QueryName() string {
	return QueryGetIssuer
}

// Validate implements cqrs.Validatable.
func (q GetIssuer) Validate() error {
	if q.IssuerID.IsZero() {
		return cqrs.ErrQueryValidation("GetIssuer.Validate", "issuer_id is required")
	}
	return nil
}

// ============================================================================
// GetIssuerByOrganization
// ============================================================================

// GetIssuerByOrganization retrieves an issuer by organization ID.
type GetIssuerByOrganization struct {
	OrganizationID string `json:"organization_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetIssuerByOrganization) QueryName() string {
	return QueryGetIssuerByOrganization
}

// Validate implements cqrs.Validatable.
func (q GetIssuerByOrganization) Validate() error {
	if q.OrganizationID == "" {
		return cqrs.ErrQueryValidation("GetIssuerByOrganization.Validate", "organization_id is required")
	}
	return nil
}

// ============================================================================
// ListIssuers
// ============================================================================

// ListIssuers lists issuers with optional status filter.
type ListIssuers struct {
	Status *string `json:"status" validate:"omitempty"`
	Limit  int     `json:"limit" validate:"omitempty"`
	Offset int     `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListIssuers) QueryName() string {
	return QueryListIssuers
}

// Validate implements cqrs.Validatable.
func (q ListIssuers) Validate() error {
	return nil
}

// ============================================================================
// GetTemplate
// ============================================================================

// GetTemplate retrieves a single template by ID.
type GetTemplate struct {
	TemplateID types.ID `json:"template_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetTemplate) QueryName() string {
	return QueryGetTemplate
}

// Validate implements cqrs.Validatable.
func (q GetTemplate) Validate() error {
	if q.TemplateID.IsZero() {
		return cqrs.ErrQueryValidation("GetTemplate.Validate", "template_id is required")
	}
	return nil
}

// ============================================================================
// ListTemplatesByIssuer
// ============================================================================

// ListTemplatesByIssuer lists templates for a given issuer.
type ListTemplatesByIssuer struct {
	IssuerID types.ID `json:"issuer_id" validate:"required"`
	Limit    int      `json:"limit" validate:"omitempty"`
	Offset   int      `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListTemplatesByIssuer) QueryName() string {
	return QueryListTemplatesByIssuer
}

// Validate implements cqrs.Validatable.
func (q ListTemplatesByIssuer) Validate() error {
	if q.IssuerID.IsZero() {
		return cqrs.ErrQueryValidation("ListTemplatesByIssuer.Validate", "issuer_id is required")
	}
	return nil
}
