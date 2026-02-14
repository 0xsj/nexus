package projections

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
)

// schemaReaderQueries defines the query methods needed by SchemaReader.
type schemaReaderQueries interface {
	GetSchemaProjectionByType(ctx context.Context, schemaType string) (generated.CredentialSchemaProjection, error)
}

// SchemaReader implements domain.SchemaResolver using the local schema projection.
type SchemaReader struct {
	queries schemaReaderQueries
}

// Compile-time check.
var _ domain.SchemaResolver = (*SchemaReader)(nil)

// NewSchemaReader creates a new SchemaReader.
func NewSchemaReader(queries *generated.Queries) *SchemaReader {
	return &SchemaReader{queries: queries}
}

// GetSchema returns the schema ID for a given credential type.
func (r *SchemaReader) GetSchema(ctx context.Context, credType domain.CredentialType) (string, error) {
	proj, err := r.queries.GetSchemaProjectionByType(ctx, string(credType))
	if err != nil {
		return "", fmt.Errorf("schema not found for type: %s", credType)
	}
	return proj.SchemaID, nil
}

// ValidateClaims validates claims against the schema for a given credential type.
func (r *SchemaReader) ValidateClaims(ctx context.Context, credType domain.CredentialType, claims domain.Claims) error {
	proj, err := r.queries.GetSchemaProjectionByType(ctx, string(credType))
	if err != nil {
		return fmt.Errorf("schema not found for type: %s", credType)
	}
	if proj.Status != "active" {
		return fmt.Errorf("schema %s is not active: %s", credType, proj.Status)
	}

	// Deserialize schema claim definitions and validate required claims are present.
	var schemaClaims []struct {
		Key      string `json:"key"`
		Required bool   `json:"required"`
	}
	if err := json.Unmarshal(proj.Claims, &schemaClaims); err != nil {
		return nil // If claims can't be parsed, skip field-level validation
	}
	for _, sc := range schemaClaims {
		if sc.Required {
			if _, ok := claims.Get(sc.Key); !ok {
				return fmt.Errorf("required claim %q missing", sc.Key)
			}
		}
	}

	return nil
}
