package command

import (
	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// RegisterSchema
// ============================================================================

// RegisterSchema creates a new schema.
type RegisterSchema struct {
	// SchemaID is the unique identifier for the new schema.
	SchemaID types.ID

	// SchemaType is the type name (e.g., "GitHubContributor").
	// Must be unique across all schemas.
	SchemaType string

	// Name is the human-readable name.
	Name string

	// Description is the schema description.
	Description string

	// Version is the initial version (typically "1.0.0").
	Version domain.SchemaVersion

	// Claims are the initial claim definitions.
	Claims []domain.ClaimDefinition

	// IssuerID is the optional issuer for custom schemas.
	// If nil, this is a built-in schema.
	IssuerID *types.ID
}

// CommandName implements cqrs.Command.
func (c RegisterSchema) CommandName() string {
	return "schema.register"
}

// Validate implements cqrs.Validatable.
func (c RegisterSchema) Validate() error {
	if c.SchemaID.IsZero() {
		return domain.SchemaInvalid("RegisterSchema.Validate", "schema_id is required")
	}
	if c.SchemaType == "" {
		return domain.SchemaInvalid("RegisterSchema.Validate", "schema_type is required")
	}
	if c.Name == "" {
		return domain.SchemaInvalid("RegisterSchema.Validate", "name is required")
	}
	if c.Version.IsZero() {
		return domain.SchemaInvalid("RegisterSchema.Validate", "version is required")
	}
	if len(c.Claims) == 0 {
		return domain.SchemaInvalid("RegisterSchema.Validate", "at least one claim is required")
	}
	return nil
}

// ============================================================================
// AddSchemaVersion
// ============================================================================

// AddSchemaVersion adds a new version to an existing schema.
type AddSchemaVersion struct {
	// SchemaID is the schema to add a version to.
	SchemaID types.ID

	// Version is the new version.
	Version domain.SchemaVersion

	// Claims are the claim definitions for this version.
	Claims []domain.ClaimDefinition

	// ChangeSummary describes what changed in this version.
	ChangeSummary string
}

// CommandName implements cqrs.Command.
func (c AddSchemaVersion) CommandName() string {
	return "schema.add_version"
}

// Validate implements cqrs.Validatable.
func (c AddSchemaVersion) Validate() error {
	if c.SchemaID.IsZero() {
		return domain.SchemaInvalid("AddSchemaVersion.Validate", "schema_id is required")
	}
	if c.Version.IsZero() {
		return domain.SchemaInvalid("AddSchemaVersion.Validate", "version is required")
	}
	if len(c.Claims) == 0 {
		return domain.SchemaInvalid("AddSchemaVersion.Validate", "at least one claim is required")
	}
	return nil
}

// ============================================================================
// DeprecateSchema
// ============================================================================

// DeprecateSchema marks a schema as deprecated.
type DeprecateSchema struct {
	// SchemaID is the schema to deprecate.
	SchemaID types.ID

	// Reason explains why the schema is being deprecated.
	Reason string

	// ReplacementSchemaID is the optional ID of the replacement schema.
	ReplacementSchemaID *types.ID
}

// CommandName implements cqrs.Command.
func (c DeprecateSchema) CommandName() string {
	return "schema.deprecate"
}

// Validate implements cqrs.Validatable.
func (c DeprecateSchema) Validate() error {
	if c.SchemaID.IsZero() {
		return domain.SchemaInvalid("DeprecateSchema.Validate", "schema_id is required")
	}
	return nil
}

// ============================================================================
// ActivateSchema
// ============================================================================

// ActivateSchema reactivates a deprecated schema.
type ActivateSchema struct {
	// SchemaID is the schema to activate.
	SchemaID types.ID
}

// CommandName implements cqrs.Command.
func (c ActivateSchema) CommandName() string {
	return "schema.activate"
}

// Validate implements cqrs.Validatable.
func (c ActivateSchema) Validate() error {
	if c.SchemaID.IsZero() {
		return domain.SchemaInvalid("ActivateSchema.Validate", "schema_id is required")
	}
	return nil
}

// ============================================================================
// UpdateSchemaMetadata
// ============================================================================

// UpdateSchemaMetadata updates a schema's name and/or description.
type UpdateSchemaMetadata struct {
	// SchemaID is the schema to update.
	SchemaID types.ID

	// Name is the new name (empty to leave unchanged).
	Name string

	// Description is the new description (empty to leave unchanged).
	Description string
}

// CommandName implements cqrs.Command.
func (c UpdateSchemaMetadata) CommandName() string {
	return "schema.update_metadata"
}

// Validate implements cqrs.Validatable.
func (c UpdateSchemaMetadata) Validate() error {
	if c.SchemaID.IsZero() {
		return domain.SchemaInvalid("UpdateSchemaMetadata.Validate", "schema_id is required")
	}
	if c.Name == "" && c.Description == "" {
		return domain.SchemaInvalid("UpdateSchemaMetadata.Validate", "at least one field must be provided")
	}
	return nil
}
