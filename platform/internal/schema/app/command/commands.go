package command

import (
	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandRegisterSchema       = "schema.RegisterSchema"
	CommandAddSchemaVersion     = "schema.AddSchemaVersion"
	CommandDeprecateSchema      = "schema.DeprecateSchema"
	CommandActivateSchema       = "schema.ActivateSchema"
	CommandUpdateSchemaMetadata = "schema.UpdateSchemaMetadata"
)

// ============================================================================
// RegisterSchema
// ============================================================================

// RegisterSchema creates a new schema.
type RegisterSchema struct {
	// SchemaID is the unique identifier for the new schema.
	SchemaID types.ID `json:"schema_id" validate:"required"`

	// SchemaType is the type name (e.g., "GitHubContributor").
	// Must be unique across all schemas.
	SchemaType string `json:"schema_type" validate:"required"`

	// Name is the human-readable name.
	Name string `json:"name" validate:"required"`

	// Description is the schema description.
	Description string `json:"description" validate:"omitempty"`

	// Version is the initial version (typically "1.0.0").
	Version domain.SchemaVersion `json:"version" validate:"required"`

	// Claims are the initial claim definitions.
	Claims []domain.ClaimDefinition `json:"claims" validate:"required,min=1"`

	// IssuerID is the optional issuer for custom schemas.
	// If nil, this is a built-in schema.
	IssuerID *types.ID `json:"issuer_id" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c RegisterSchema) CommandName() string {
	return CommandRegisterSchema
}

// Validate implements cqrs.Validatable.
func (c RegisterSchema) Validate() error {
	if c.SchemaID.IsZero() {
		return cqrs.ErrCommandValidation("RegisterSchema.Validate", "schema_id is required")
	}
	if c.SchemaType == "" {
		return cqrs.ErrCommandValidation("RegisterSchema.Validate", "schema_type is required")
	}
	if c.Name == "" {
		return cqrs.ErrCommandValidation("RegisterSchema.Validate", "name is required")
	}
	if c.Version.IsZero() {
		return cqrs.ErrCommandValidation("RegisterSchema.Validate", "version is required")
	}
	if len(c.Claims) == 0 {
		return cqrs.ErrCommandValidation("RegisterSchema.Validate", "at least one claim is required")
	}
	return nil
}

// RegisterSchemaResult is the result data for RegisterSchema.
type RegisterSchemaResult struct {
	SchemaID   string `json:"schema_id"`
	SchemaType string `json:"schema_type"`
	Version    string `json:"version"`
}

// ============================================================================
// AddSchemaVersion
// ============================================================================

// AddSchemaVersion adds a new version to an existing schema.
type AddSchemaVersion struct {
	// SchemaID is the schema to add a version to.
	SchemaID types.ID `json:"schema_id" validate:"required"`

	// Version is the new version.
	Version domain.SchemaVersion `json:"version" validate:"required"`

	// Claims are the claim definitions for this version.
	Claims []domain.ClaimDefinition `json:"claims" validate:"required,min=1"`

	// ChangeSummary describes what changed in this version.
	ChangeSummary string `json:"change_summary" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c AddSchemaVersion) CommandName() string {
	return CommandAddSchemaVersion
}

// Validate implements cqrs.Validatable.
func (c AddSchemaVersion) Validate() error {
	if c.SchemaID.IsZero() {
		return cqrs.ErrCommandValidation("AddSchemaVersion.Validate", "schema_id is required")
	}
	if c.Version.IsZero() {
		return cqrs.ErrCommandValidation("AddSchemaVersion.Validate", "version is required")
	}
	if len(c.Claims) == 0 {
		return cqrs.ErrCommandValidation("AddSchemaVersion.Validate", "at least one claim is required")
	}
	return nil
}

// AddSchemaVersionResult is the result data for AddSchemaVersion.
type AddSchemaVersionResult struct {
	SchemaID        string `json:"schema_id"`
	Version         string `json:"version"`
	PreviousVersion string `json:"previous_version"`
}

// ============================================================================
// DeprecateSchema
// ============================================================================

// DeprecateSchema marks a schema as deprecated.
type DeprecateSchema struct {
	// SchemaID is the schema to deprecate.
	SchemaID types.ID `json:"schema_id" validate:"required"`

	// Reason explains why the schema is being deprecated.
	Reason string `json:"reason" validate:"omitempty,max=500"`

	// ReplacementSchemaID is the optional ID of the replacement schema.
	ReplacementSchemaID *types.ID `json:"replacement_schema_id" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c DeprecateSchema) CommandName() string {
	return CommandDeprecateSchema
}

// Validate implements cqrs.Validatable.
func (c DeprecateSchema) Validate() error {
	if c.SchemaID.IsZero() {
		return cqrs.ErrCommandValidation("DeprecateSchema.Validate", "schema_id is required")
	}
	return nil
}

// DeprecateSchemaResult is the result data for DeprecateSchema.
type DeprecateSchemaResult struct {
	SchemaID            string `json:"schema_id"`
	ReplacementSchemaID string `json:"replacement_schema_id,omitempty"`
}

// ============================================================================
// ActivateSchema
// ============================================================================

// ActivateSchema reactivates a deprecated schema.
type ActivateSchema struct {
	// SchemaID is the schema to activate.
	SchemaID types.ID `json:"schema_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ActivateSchema) CommandName() string {
	return CommandActivateSchema
}

// Validate implements cqrs.Validatable.
func (c ActivateSchema) Validate() error {
	if c.SchemaID.IsZero() {
		return cqrs.ErrCommandValidation("ActivateSchema.Validate", "schema_id is required")
	}
	return nil
}

// ActivateSchemaResult is the result data for ActivateSchema.
type ActivateSchemaResult struct {
	SchemaID string `json:"schema_id"`
}

// ============================================================================
// UpdateSchemaMetadata
// ============================================================================

// UpdateSchemaMetadata updates a schema's name and/or description.
type UpdateSchemaMetadata struct {
	// SchemaID is the schema to update.
	SchemaID types.ID `json:"schema_id" validate:"required"`

	// Name is the new name (empty to leave unchanged).
	Name string `json:"name" validate:"omitempty,max=255"`

	// Description is the new description (empty to leave unchanged).
	Description string `json:"description" validate:"omitempty,max=1000"`
}

// CommandName implements cqrs.Command.
func (c UpdateSchemaMetadata) CommandName() string {
	return CommandUpdateSchemaMetadata
}

// Validate implements cqrs.Validatable.
func (c UpdateSchemaMetadata) Validate() error {
	if c.SchemaID.IsZero() {
		return cqrs.ErrCommandValidation("UpdateSchemaMetadata.Validate", "schema_id is required")
	}
	if c.Name == "" && c.Description == "" {
		return cqrs.ErrCommandValidation("UpdateSchemaMetadata.Validate", "at least one field must be provided")
	}
	return nil
}

// UpdateSchemaMetadataResult is the result data for UpdateSchemaMetadata.
type UpdateSchemaMetadataResult struct {
	SchemaID    string `json:"schema_id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
