package v1

// ============================================================================
// Schema Requests
// ============================================================================

// RegisterSchemaRequest represents a request to register a new schema.
type RegisterSchemaRequest struct {
	SchemaType  string                   `json:"schema_type" validate:"required"`
	Name        string                   `json:"name" validate:"required"`
	Description string                   `json:"description,omitempty"`
	Version     string                   `json:"version" validate:"required"`
	Claims      []ClaimDefinitionRequest `json:"claims" validate:"required,min=1"`
	IssuerID    *string                  `json:"issuer_id,omitempty"`
}

// ClaimDefinitionRequest represents a claim definition in a request.
type ClaimDefinitionRequest struct {
	Key         string              `json:"key" validate:"required"`
	DataType    string              `json:"data_type" validate:"required"`
	Required    bool                `json:"required,omitempty"`
	DisplayName string              `json:"display_name,omitempty"`
	Description string              `json:"description,omitempty"`
	Disclosable *bool               `json:"disclosable,omitempty"`
	Order       int                 `json:"order,omitempty"`
	Constraints *ConstraintsRequest `json:"constraints,omitempty"`
}

// ConstraintsRequest represents claim constraints in a request.
type ConstraintsRequest struct {
	AllowedValues []string `json:"allowed_values,omitempty"`
	Min           *int64   `json:"min,omitempty"`
	Max           *int64   `json:"max,omitempty"`
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	Pattern       *string  `json:"pattern,omitempty"`
	Format        *string  `json:"format,omitempty"`
}

// ============================================================================
// Version Requests
// ============================================================================

// AddSchemaVersionRequest represents a request to add a new version.
type AddSchemaVersionRequest struct {
	Version       string                   `json:"version" validate:"required"`
	Claims        []ClaimDefinitionRequest `json:"claims" validate:"required,min=1"`
	ChangeSummary string                   `json:"change_summary,omitempty"`
}

// ============================================================================
// Status Requests
// ============================================================================

// DeprecateSchemaRequest represents a request to deprecate a schema.
type DeprecateSchemaRequest struct {
	Reason              string  `json:"reason,omitempty"`
	ReplacementSchemaID *string `json:"replacement_schema_id,omitempty"`
}

// ============================================================================
// Metadata Requests
// ============================================================================

// UpdateSchemaMetadataRequest represents a request to update schema metadata.
type UpdateSchemaMetadataRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ============================================================================
// List/Search Requests
// ============================================================================

// ListSchemasRequest represents query parameters for listing schemas.
type ListSchemasRequest struct {
	Status      string `json:"status,omitempty"`
	IssuerID    string `json:"issuer_id,omitempty"`
	BuiltInOnly *bool  `json:"built_in_only,omitempty"`
	Search      string `json:"search,omitempty"`
	OrderBy     string `json:"order_by,omitempty"`
	OrderDir    string `json:"order_dir,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}
