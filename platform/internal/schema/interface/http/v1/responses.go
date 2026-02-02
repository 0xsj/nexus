package v1

import (
	"time"
)

// ============================================================================
// Schema Responses
// ============================================================================

// SchemaResponse represents a full schema in API responses.
type SchemaResponse struct {
	SchemaID       string          `json:"schema_id"`
	SchemaType     string          `json:"schema_type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	CurrentVersion string          `json:"current_version"`
	Status         string          `json:"status"`
	IssuerID       string          `json:"issuer_id,omitempty"`
	Claims         []ClaimResponse `json:"claims"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// SchemaSummaryResponse represents a lightweight schema in list responses.
type SchemaSummaryResponse struct {
	SchemaID       string    `json:"schema_id"`
	SchemaType     string    `json:"schema_type"`
	Name           string    `json:"name"`
	CurrentVersion string    `json:"current_version"`
	Status         string    `json:"status"`
	ClaimCount     int       `json:"claim_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// SchemaListResponse represents a paginated list of schemas.
type SchemaListResponse struct {
	Schemas    []SchemaSummaryResponse `json:"schemas"`
	TotalCount int                     `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	HasMore    bool                    `json:"has_more"`
}

// ============================================================================
// Claim Responses
// ============================================================================

// ClaimResponse represents a claim definition in API responses.
type ClaimResponse struct {
	Key         string               `json:"key"`
	DataType    string               `json:"data_type"`
	Required    bool                 `json:"required"`
	DisplayName string               `json:"display_name,omitempty"`
	Description string               `json:"description,omitempty"`
	Disclosable bool                 `json:"disclosable"`
	Order       int                  `json:"order"`
	Constraints *ConstraintsResponse `json:"constraints,omitempty"`
}

// ConstraintsResponse represents claim constraints in API responses.
type ConstraintsResponse struct {
	AllowedValues []string `json:"allowed_values,omitempty"`
	Min           *int64   `json:"min,omitempty"`
	Max           *int64   `json:"max,omitempty"`
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	Pattern       *string  `json:"pattern,omitempty"`
	Format        *string  `json:"format,omitempty"`
}

// ============================================================================
// Version Responses
// ============================================================================

// SchemaVersionResponse represents a specific version of a schema.
type SchemaVersionResponse struct {
	SchemaID   string          `json:"schema_id"`
	SchemaType string          `json:"schema_type"`
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Status     string          `json:"status"`
	Claims     []ClaimResponse `json:"claims"`
	IsLatest   bool            `json:"is_latest"`
	CreatedAt  time.Time       `json:"created_at"`
}

// VersionInfoResponse represents version info in list responses.
type VersionInfoResponse struct {
	Version   string    `json:"version"`
	IsLatest  bool      `json:"is_latest"`
	CreatedAt time.Time `json:"created_at"`
}

// VersionListResponse represents a list of versions for a schema.
type VersionListResponse struct {
	SchemaID       string                `json:"schema_id"`
	SchemaType     string                `json:"schema_type"`
	CurrentVersion string                `json:"current_version"`
	Versions       []VersionInfoResponse `json:"versions"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// RegisterSchemaResponse represents the result of registering a schema.
type RegisterSchemaResponse struct {
	SchemaID   string    `json:"schema_id"`
	SchemaType string    `json:"schema_type"`
	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

// AddVersionResponse represents the result of adding a version.
type AddVersionResponse struct {
	SchemaID        string `json:"schema_id"`
	Version         string `json:"version"`
	PreviousVersion string `json:"previous_version"`
}

// DeprecateSchemaResponse represents the result of deprecating a schema.
type DeprecateSchemaResponse struct {
	SchemaID            string    `json:"schema_id"`
	ReplacementSchemaID string    `json:"replacement_schema_id,omitempty"`
	DeprecatedAt        time.Time `json:"deprecated_at"`
}

// ActivateSchemaResponse represents the result of activating a schema.
type ActivateSchemaResponse struct {
	SchemaID    string    `json:"schema_id"`
	ActivatedAt time.Time `json:"activated_at"`
}

// UpdateMetadataResponse represents the result of updating metadata.
type UpdateMetadataResponse struct {
	SchemaID    string    `json:"schema_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================================================
// Lookup Responses
// ============================================================================

// SchemaExistsResponse represents the result of an existence check.
type SchemaExistsResponse struct {
	Exists     bool   `json:"exists"`
	SchemaID   string `json:"schema_id,omitempty"`
	SchemaType string `json:"schema_type,omitempty"`
}

// ResolveSchemaTypeResponse represents the result of resolving a schema type.
type ResolveSchemaTypeResponse struct {
	SchemaType     string `json:"schema_type"`
	SchemaID       string `json:"schema_id"`
	CurrentVersion string `json:"current_version"`
	Status         string `json:"status"`
}
