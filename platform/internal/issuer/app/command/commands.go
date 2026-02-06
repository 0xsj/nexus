package command

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandRegisterIssuer  = "issuer.RegisterIssuer"
	CommandActivateIssuer  = "issuer.ActivateIssuer"
	CommandSuspendIssuer   = "issuer.SuspendIssuer"
	CommandCreateTemplate  = "issuer.CreateTemplate"
	CommandUpdateTemplate  = "issuer.UpdateTemplate"
	CommandArchiveTemplate = "issuer.ArchiveTemplate"
)

// ============================================================================
// RegisterIssuer
// ============================================================================

// RegisterIssuer registers a new issuer for an organization.
type RegisterIssuer struct {
	OrganizationID string `json:"organization_id" validate:"required"`
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description" validate:"omitempty"`
	WebhookURL     string `json:"webhook_url" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c RegisterIssuer) CommandName() string {
	return CommandRegisterIssuer
}

// Validate implements cqrs.Validatable.
func (c RegisterIssuer) Validate() error {
	if c.OrganizationID == "" {
		return cqrs.ErrCommandValidation("RegisterIssuer.Validate", "organization_id is required")
	}
	if c.Name == "" {
		return cqrs.ErrCommandValidation("RegisterIssuer.Validate", "name is required")
	}
	return nil
}

// RegisterIssuerResult is the result data for RegisterIssuer.
type RegisterIssuerResult struct {
	IssuerID string `json:"issuer_id"`
	Status   string `json:"status"`
}

// ============================================================================
// ActivateIssuer
// ============================================================================

// ActivateIssuer activates a pending issuer.
type ActivateIssuer struct {
	IssuerID          types.ID `json:"issuer_id" validate:"required"`
	DID               string   `json:"did" validate:"required"`
	APIKey            string   `json:"api_key" validate:"required"`
	LogoURL           string   `json:"logo_url" validate:"omitempty"`
	PrimaryColor      string   `json:"primary_color" validate:"omitempty"`
	SecondaryColor    string   `json:"secondary_color" validate:"omitempty"`
	CertificateDesign string   `json:"certificate_design" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c ActivateIssuer) CommandName() string {
	return CommandActivateIssuer
}

// Validate implements cqrs.Validatable.
func (c ActivateIssuer) Validate() error {
	if c.IssuerID.IsZero() {
		return cqrs.ErrCommandValidation("ActivateIssuer.Validate", "issuer_id is required")
	}
	if c.DID == "" {
		return cqrs.ErrCommandValidation("ActivateIssuer.Validate", "did is required")
	}
	if c.APIKey == "" {
		return cqrs.ErrCommandValidation("ActivateIssuer.Validate", "api_key is required")
	}
	return nil
}

// ActivateIssuerResult is the result data for ActivateIssuer.
type ActivateIssuerResult struct {
	IssuerID string `json:"issuer_id"`
	Status   string `json:"status"`
}

// ============================================================================
// SuspendIssuer
// ============================================================================

// SuspendIssuer suspends an active issuer.
type SuspendIssuer struct {
	IssuerID types.ID `json:"issuer_id" validate:"required"`
	Reason   string   `json:"reason" validate:"required,max=500"`
}

// CommandName implements cqrs.Command.
func (c SuspendIssuer) CommandName() string {
	return CommandSuspendIssuer
}

// Validate implements cqrs.Validatable.
func (c SuspendIssuer) Validate() error {
	if c.IssuerID.IsZero() {
		return cqrs.ErrCommandValidation("SuspendIssuer.Validate", "issuer_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("SuspendIssuer.Validate", "reason is required")
	}
	if len(c.Reason) > 500 {
		return cqrs.ErrCommandValidation("SuspendIssuer.Validate", "reason must be 500 characters or less")
	}
	return nil
}

// SuspendIssuerResult is the result data for SuspendIssuer.
type SuspendIssuerResult struct {
	IssuerID string `json:"issuer_id"`
	Status   string `json:"status"`
}

// ============================================================================
// CreateTemplate
// ============================================================================

// CreateTemplate creates a new credential template for an issuer.
type CreateTemplate struct {
	IssuerID       types.ID          `json:"issuer_id" validate:"required"`
	Name           string            `json:"name" validate:"required"`
	Description    string            `json:"description" validate:"omitempty"`
	SchemaType     string            `json:"schema_type" validate:"required"`
	ClaimMappings  map[string]string `json:"claim_mappings" validate:"omitempty"`
	DefaultValues  map[string]any    `json:"default_values" validate:"omitempty"`
	ExpirationDays int               `json:"expiration_days" validate:"omitempty"`
	AutoApprove    bool              `json:"auto_approve" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c CreateTemplate) CommandName() string {
	return CommandCreateTemplate
}

// Validate implements cqrs.Validatable.
func (c CreateTemplate) Validate() error {
	if c.IssuerID.IsZero() {
		return cqrs.ErrCommandValidation("CreateTemplate.Validate", "issuer_id is required")
	}
	if c.Name == "" {
		return cqrs.ErrCommandValidation("CreateTemplate.Validate", "name is required")
	}
	if c.SchemaType == "" {
		return cqrs.ErrCommandValidation("CreateTemplate.Validate", "schema_type is required")
	}
	if c.ExpirationDays < 0 {
		return cqrs.ErrCommandValidation("CreateTemplate.Validate", "expiration_days must be non-negative")
	}
	return nil
}

// CreateTemplateResult is the result data for CreateTemplate.
type CreateTemplateResult struct {
	TemplateID string `json:"template_id"`
	IssuerID   string `json:"issuer_id"`
	Status     string `json:"status"`
}

// ============================================================================
// UpdateTemplate
// ============================================================================

// UpdateTemplate updates an existing credential template.
type UpdateTemplate struct {
	TemplateID     types.ID          `json:"template_id" validate:"required"`
	Name           string            `json:"name" validate:"required"`
	Description    string            `json:"description" validate:"omitempty"`
	ClaimMappings  map[string]string `json:"claim_mappings" validate:"omitempty"`
	DefaultValues  map[string]any    `json:"default_values" validate:"omitempty"`
	ExpirationDays int               `json:"expiration_days" validate:"omitempty"`
	AutoApprove    bool              `json:"auto_approve" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c UpdateTemplate) CommandName() string {
	return CommandUpdateTemplate
}

// Validate implements cqrs.Validatable.
func (c UpdateTemplate) Validate() error {
	if c.TemplateID.IsZero() {
		return cqrs.ErrCommandValidation("UpdateTemplate.Validate", "template_id is required")
	}
	if c.Name == "" {
		return cqrs.ErrCommandValidation("UpdateTemplate.Validate", "name is required")
	}
	if c.ExpirationDays < 0 {
		return cqrs.ErrCommandValidation("UpdateTemplate.Validate", "expiration_days must be non-negative")
	}
	return nil
}

// UpdateTemplateResult is the result data for UpdateTemplate.
type UpdateTemplateResult struct {
	TemplateID string `json:"template_id"`
	Version    int    `json:"version"`
	Status     string `json:"status"`
}

// ============================================================================
// ArchiveTemplate
// ============================================================================

// ArchiveTemplate archives a credential template.
type ArchiveTemplate struct {
	TemplateID types.ID `json:"template_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ArchiveTemplate) CommandName() string {
	return CommandArchiveTemplate
}

// Validate implements cqrs.Validatable.
func (c ArchiveTemplate) Validate() error {
	if c.TemplateID.IsZero() {
		return cqrs.ErrCommandValidation("ArchiveTemplate.Validate", "template_id is required")
	}
	return nil
}

// ArchiveTemplateResult is the result data for ArchiveTemplate.
type ArchiveTemplateResult struct {
	TemplateID string `json:"template_id"`
	Status     string `json:"status"`
}
