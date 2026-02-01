package query

import "time"

// ============================================================================
// Schema Views
// ============================================================================

// SchemaView is the full schema read model.
type SchemaView struct {
	SchemaID       string      `json:"schema_id"`
	SchemaType     string      `json:"schema_type"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	CurrentVersion string      `json:"current_version"`
	Status         string      `json:"status"`
	IssuerID       string      `json:"issuer_id,omitempty"`
	Claims         []ClaimView `json:"claims"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// SchemaSummaryView is a lightweight schema representation for lists.
type SchemaSummaryView struct {
	SchemaID       string    `json:"schema_id"`
	SchemaType     string    `json:"schema_type"`
	Name           string    `json:"name"`
	CurrentVersion string    `json:"current_version"`
	Status         string    `json:"status"`
	ClaimCount     int       `json:"claim_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// SchemaListView is a paginated list of schemas.
type SchemaListView struct {
	Schemas    []SchemaSummaryView `json:"schemas"`
	NextCursor *string             `json:"next_cursor,omitempty"`
	TotalCount int                 `json:"total_count"`
}

// ============================================================================
// Claim Views
// ============================================================================

// ClaimView represents a claim definition in a schema.
type ClaimView struct {
	Key         string           `json:"key"`
	DataType    string           `json:"data_type"`
	Required    bool             `json:"required"`
	DisplayName string           `json:"display_name,omitempty"`
	Description string           `json:"description,omitempty"`
	Disclosable bool             `json:"disclosable"`
	Order       int              `json:"order"`
	Constraints *ConstraintsView `json:"constraints,omitempty"`
}

// ConstraintsView represents claim constraints.
type ConstraintsView struct {
	AllowedValues []string `json:"allowed_values,omitempty"`
	Min           *int64   `json:"min,omitempty"`
	Max           *int64   `json:"max,omitempty"`
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	Pattern       *string  `json:"pattern,omitempty"`
	Format        *string  `json:"format,omitempty"`
}

// ============================================================================
// Version Views
// ============================================================================

// SchemaVersionView represents a specific version of a schema.
type SchemaVersionView struct {
	SchemaID   string      `json:"schema_id"`
	SchemaType string      `json:"schema_type"`
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Status     string      `json:"status"`
	Claims     []ClaimView `json:"claims"`
	IsLatest   bool        `json:"is_latest"`
	CreatedAt  time.Time   `json:"created_at"`
}

// VersionListView is a list of versions for a schema.
type VersionListView struct {
	SchemaID       string        `json:"schema_id"`
	SchemaType     string        `json:"schema_type"`
	CurrentVersion string        `json:"current_version"`
	Versions       []VersionInfo `json:"versions"`
}

// VersionInfo is a lightweight version representation.
type VersionInfo struct {
	Version   string    `json:"version"`
	IsLatest  bool      `json:"is_latest"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// Lookup Views
// ============================================================================

// SchemaExistsView is the result of checking schema existence.
type SchemaExistsView struct {
	Exists     bool   `json:"exists"`
	SchemaID   string `json:"schema_id,omitempty"`
	SchemaType string `json:"schema_type,omitempty"`
}

// SchemaTypeResolutionView is the result of resolving a schema type to ID.
type SchemaTypeResolutionView struct {
	SchemaType     string `json:"schema_type"`
	SchemaID       string `json:"schema_id"`
	CurrentVersion string `json:"current_version"`
	Status         string `json:"status"`
}
