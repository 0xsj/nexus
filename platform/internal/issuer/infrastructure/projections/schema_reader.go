package projections

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres/generated"
)

// schemaReaderQueries defines the query methods needed by SchemaReader.
type schemaReaderQueries interface {
	SchemaProjectionExistsByType(ctx context.Context, schemaType string) (bool, error)
}

// SchemaReader implements domain.SchemaReader using the local schema projection.
type SchemaReader struct {
	queries schemaReaderQueries
}

// Compile-time check.
var _ domain.SchemaReader = (*SchemaReader)(nil)

// NewSchemaReader creates a new SchemaReader.
func NewSchemaReader(queries *generated.Queries) *SchemaReader {
	return &SchemaReader{queries: queries}
}

// GetSchema verifies that a schema type exists.
func (r *SchemaReader) GetSchema(ctx context.Context, schemaType string) error {
	exists, err := r.queries.SchemaProjectionExistsByType(ctx, schemaType)
	if err != nil {
		return fmt.Errorf("checking schema existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("schema not found: %s", schemaType)
	}
	return nil
}

// SchemaExists checks if a schema type exists.
func (r *SchemaReader) SchemaExists(ctx context.Context, schemaType string) (bool, error) {
	return r.queries.SchemaProjectionExistsByType(ctx, schemaType)
}
