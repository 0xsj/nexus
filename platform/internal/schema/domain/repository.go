package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Repository Interface
// ============================================================================

// SchemaRepository defines the persistence interface for Schema aggregates.
// Implementations handle storage concerns (PostgreSQL, event store, etc.).
type SchemaRepository interface {
	// Save persists a schema aggregate.
	// For event-sourced implementations, this appends new events.
	Save(ctx context.Context, schema *Schema) error

	// GetByID loads a schema by its unique identifier.
	// Returns ErrSchemaNotFound if the schema does not exist.
	GetByID(ctx context.Context, id types.ID) (*Schema, error)

	// GetByType loads a schema by its type name.
	// Returns the latest version of the schema.
	// Returns ErrSchemaNotFound if no schema with the type exists.
	GetByType(ctx context.Context, schemaType string) (*Schema, error)

	// GetByTypeAndVersion loads a specific version of a schema.
	// Returns ErrSchemaNotFound if the schema type does not exist.
	// Returns ErrVersionNotFound if the specific version does not exist.
	GetByTypeAndVersion(ctx context.Context, schemaType string, version SchemaVersion) (*Schema, error)

	// List returns schemas matching the given filter.
	List(ctx context.Context, filter SchemaFilter) ([]*Schema, error)

	// Count returns the total number of schemas matching the filter.
	Count(ctx context.Context, filter SchemaFilter) (int64, error)

	// Exists checks if a schema with the given ID exists.
	Exists(ctx context.Context, id types.ID) (bool, error)

	// ExistsByType checks if a schema with the given type name exists.
	ExistsByType(ctx context.Context, schemaType string) (bool, error)
}

// ============================================================================
// Filter
// ============================================================================

// SchemaFilter defines filtering options for listing schemas.
type SchemaFilter struct {
	// Status filters by schema status.
	// nil means all statuses.
	Status *SchemaStatus

	// IssuerID filters by issuer.
	// nil means all issuers (including built-in).
	IssuerID *types.ID

	// BuiltInOnly filters to only built-in schemas (no issuer).
	// nil means all, true means built-in only, false means custom only.
	BuiltInOnly *bool

	// SchemaTypes filters by specific schema types.
	// Empty slice means all types.
	SchemaTypes []string

	// SearchQuery is an optional text search on name/description.
	SearchQuery string

	// Limit is the maximum number of results to return.
	// 0 means use default limit.
	Limit int

	// Offset is the number of results to skip (for pagination).
	Offset int

	// OrderBy specifies the field to order by.
	// Empty means default ordering (by created_at desc).
	OrderBy SchemaOrderBy

	// OrderDir specifies the order direction.
	OrderDir OrderDirection
}

// SchemaOrderBy defines fields that can be used for ordering.
type SchemaOrderBy string

const (
	SchemaOrderByCreatedAt SchemaOrderBy = "created_at"
	SchemaOrderByUpdatedAt SchemaOrderBy = "updated_at"
	SchemaOrderByName      SchemaOrderBy = "name"
	SchemaOrderByType      SchemaOrderBy = "schema_type"
)

// OrderDirection defines sort order direction.
type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "asc"
	OrderDirectionDesc OrderDirection = "desc"
)

// ============================================================================
// Filter Builder
// ============================================================================

// NewSchemaFilter creates a new SchemaFilter with default values.
func NewSchemaFilter() SchemaFilter {
	return SchemaFilter{
		Limit:    20,
		Offset:   0,
		OrderBy:  SchemaOrderByCreatedAt,
		OrderDir: OrderDirectionDesc,
	}
}

// WithStatus sets the status filter.
func (f SchemaFilter) WithStatus(status SchemaStatus) SchemaFilter {
	f.Status = &status
	return f
}

// WithActiveOnly filters to only active schemas.
func (f SchemaFilter) WithActiveOnly() SchemaFilter {
	status := SchemaStatusActive
	f.Status = &status
	return f
}

// WithDeprecatedOnly filters to only deprecated schemas.
func (f SchemaFilter) WithDeprecatedOnly() SchemaFilter {
	status := SchemaStatusDeprecated
	f.Status = &status
	return f
}

// WithIssuerID sets the issuer ID filter.
func (f SchemaFilter) WithIssuerID(issuerID types.ID) SchemaFilter {
	f.IssuerID = &issuerID
	return f
}

// WithBuiltInOnly filters to only built-in schemas.
func (f SchemaFilter) WithBuiltInOnly() SchemaFilter {
	builtIn := true
	f.BuiltInOnly = &builtIn
	return f
}

// WithCustomOnly filters to only custom (issuer-defined) schemas.
func (f SchemaFilter) WithCustomOnly() SchemaFilter {
	builtIn := false
	f.BuiltInOnly = &builtIn
	return f
}

// WithSchemaTypes sets the schema types filter.
func (f SchemaFilter) WithSchemaTypes(types ...string) SchemaFilter {
	f.SchemaTypes = types
	return f
}

// WithSearchQuery sets the search query.
func (f SchemaFilter) WithSearchQuery(query string) SchemaFilter {
	f.SearchQuery = query
	return f
}

// WithLimit sets the limit.
func (f SchemaFilter) WithLimit(limit int) SchemaFilter {
	f.Limit = limit
	return f
}

// WithOffset sets the offset.
func (f SchemaFilter) WithOffset(offset int) SchemaFilter {
	f.Offset = offset
	return f
}

// WithPagination sets both limit and offset.
func (f SchemaFilter) WithPagination(limit, offset int) SchemaFilter {
	f.Limit = limit
	f.Offset = offset
	return f
}

// WithOrderBy sets the order by field and direction.
func (f SchemaFilter) WithOrderBy(orderBy SchemaOrderBy, dir OrderDirection) SchemaFilter {
	f.OrderBy = orderBy
	f.OrderDir = dir
	return f
}

// ============================================================================
// Read Model Repository (Optional CQRS)
// ============================================================================

// SchemaReadRepository defines a read-optimized repository for queries.
// Use this for CQRS implementations where reads and writes are separated.
type SchemaReadRepository interface {
	// GetByID loads a schema view by ID.
	GetByID(ctx context.Context, id types.ID) (*SchemaView, error)

	// GetByType loads a schema view by type.
	GetByType(ctx context.Context, schemaType string) (*SchemaView, error)

	// List returns schema views matching the filter.
	List(ctx context.Context, filter SchemaFilter) ([]*SchemaView, error)

	// Count returns the total count matching the filter.
	Count(ctx context.Context, filter SchemaFilter) (int64, error)
}

// SchemaView is a read-optimized projection of a Schema.
// Used for query responses without full aggregate reconstruction.
type SchemaView struct {
	ID             types.ID       `json:"id"`
	SchemaType     string         `json:"schema_type"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	CurrentVersion string         `json:"current_version"`
	Status         SchemaStatus   `json:"status"`
	IssuerID       *types.ID      `json:"issuer_id,omitempty"`
	ClaimCount     int            `json:"claim_count"`
	Claims         []ClaimSummary `json:"claims,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// ClaimSummary is a lightweight representation of a claim for views.
type ClaimSummary struct {
	Key         string `json:"key"`
	DataType    string `json:"data_type"`
	Required    bool   `json:"required"`
	DisplayName string `json:"display_name,omitempty"`
}
