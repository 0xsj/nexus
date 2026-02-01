// Package query contains the query definitions and handlers for the Schema context.
package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Query Name Constants
// ============================================================================

const (
	// Schema queries
	QueryGetSchema       = "schema.GetSchema"
	QueryGetSchemaByType = "schema.GetSchemaByType"
	QueryListSchemas     = "schema.ListSchemas"
	QuerySearchSchemas   = "schema.SearchSchemas"

	// Version queries
	QueryGetSchemaVersion   = "schema.GetSchemaVersion"
	QueryListSchemaVersions = "schema.ListSchemaVersions"

	// Lookup queries
	QuerySchemaExists      = "schema.SchemaExists"
	QueryResolveSchemaType = "schema.ResolveSchemaType"
)

// ============================================================================
// Schema Queries
// ============================================================================

// GetSchema retrieves a schema by ID.
type GetSchema struct {
	SchemaID types.ID `json:"schema_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetSchema) QueryName() string {
	return QueryGetSchema
}

// Validate implements cqrs.Validatable.
func (q GetSchema) Validate() error {
	if q.SchemaID.IsZero() {
		return cqrs.ErrQueryValidation("GetSchema.Validate", "schema_id is required")
	}
	return nil
}

// GetSchemaByType retrieves a schema by its type name.
type GetSchemaByType struct {
	SchemaType string `json:"schema_type" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetSchemaByType) QueryName() string {
	return QueryGetSchemaByType
}

// Validate implements cqrs.Validatable.
func (q GetSchemaByType) Validate() error {
	if q.SchemaType == "" {
		return cqrs.ErrQueryValidation("GetSchemaByType.Validate", "schema_type is required")
	}
	return nil
}

// ListSchemas lists schemas with pagination and filtering.
type ListSchemas struct {
	Status      *string   `json:"status" validate:"omitempty,oneof=active deprecated"`
	IssuerID    *types.ID `json:"issuer_id" validate:"omitempty"`
	BuiltInOnly *bool     `json:"built_in_only" validate:"omitempty"`
	PageSize    int       `json:"page_size" validate:"omitempty,min=1,max=100"`
	Cursor      *string   `json:"cursor" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListSchemas) QueryName() string {
	return QueryListSchemas
}

// Validate implements cqrs.Validatable.
func (q ListSchemas) Validate() error {
	if q.PageSize < 0 {
		return cqrs.ErrQueryValidation("ListSchemas.Validate", "page_size must be non-negative")
	}
	if q.PageSize > 100 {
		return cqrs.ErrQueryValidation("ListSchemas.Validate", "page_size must be 100 or less")
	}
	return nil
}

// SearchSchemas searches schemas by name or description.
type SearchSchemas struct {
	Query    string  `json:"query" validate:"required,min=1,max=100"`
	Status   *string `json:"status" validate:"omitempty,oneof=active deprecated"`
	PageSize int     `json:"page_size" validate:"omitempty,min=1,max=100"`
	Cursor   *string `json:"cursor" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q SearchSchemas) QueryName() string {
	return QuerySearchSchemas
}

// Validate implements cqrs.Validatable.
func (q SearchSchemas) Validate() error {
	if q.Query == "" {
		return cqrs.ErrQueryValidation("SearchSchemas.Validate", "query is required")
	}
	if len(q.Query) > 100 {
		return cqrs.ErrQueryValidation("SearchSchemas.Validate", "query must be 100 characters or less")
	}
	if q.PageSize < 0 {
		return cqrs.ErrQueryValidation("SearchSchemas.Validate", "page_size must be non-negative")
	}
	if q.PageSize > 100 {
		return cqrs.ErrQueryValidation("SearchSchemas.Validate", "page_size must be 100 or less")
	}
	return nil
}

// ============================================================================
// Version Queries
// ============================================================================

// GetSchemaVersion retrieves a specific version of a schema.
type GetSchemaVersion struct {
	SchemaID types.ID `json:"schema_id" validate:"required"`
	Version  string   `json:"version" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetSchemaVersion) QueryName() string {
	return QueryGetSchemaVersion
}

// Validate implements cqrs.Validatable.
func (q GetSchemaVersion) Validate() error {
	if q.SchemaID.IsZero() {
		return cqrs.ErrQueryValidation("GetSchemaVersion.Validate", "schema_id is required")
	}
	if q.Version == "" {
		return cqrs.ErrQueryValidation("GetSchemaVersion.Validate", "version is required")
	}
	return nil
}

// ListSchemaVersions lists all versions of a schema.
type ListSchemaVersions struct {
	SchemaID types.ID `json:"schema_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q ListSchemaVersions) QueryName() string {
	return QueryListSchemaVersions
}

// Validate implements cqrs.Validatable.
func (q ListSchemaVersions) Validate() error {
	if q.SchemaID.IsZero() {
		return cqrs.ErrQueryValidation("ListSchemaVersions.Validate", "schema_id is required")
	}
	return nil
}

// ============================================================================
// Lookup Queries
// ============================================================================

// SchemaExists checks if a schema exists by ID or type.
type SchemaExists struct {
	SchemaID   *types.ID `json:"schema_id" validate:"omitempty"`
	SchemaType *string   `json:"schema_type" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q SchemaExists) QueryName() string {
	return QuerySchemaExists
}

// Validate implements cqrs.Validatable.
func (q SchemaExists) Validate() error {
	if (q.SchemaID == nil || q.SchemaID.IsZero()) && (q.SchemaType == nil || *q.SchemaType == "") {
		return cqrs.ErrQueryValidation("SchemaExists.Validate", "either schema_id or schema_type is required")
	}
	return nil
}

// ResolveSchemaType resolves a schema type to its ID and current version.
type ResolveSchemaType struct {
	SchemaType string `json:"schema_type" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q ResolveSchemaType) QueryName() string {
	return QueryResolveSchemaType
}

// Validate implements cqrs.Validatable.
func (q ResolveSchemaType) Validate() error {
	if q.SchemaType == "" {
		return cqrs.ErrQueryValidation("ResolveSchemaType.Validate", "schema_type is required")
	}
	return nil
}
