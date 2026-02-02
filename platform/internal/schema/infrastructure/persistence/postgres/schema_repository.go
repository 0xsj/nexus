package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/0xsj/nexus/platform/internal/schema/domain"
	generated "github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres/generated"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Compile-time interface checks
// ============================================================================

var (
	_ domain.SchemaRepository = (*Repository)(nil)
	_ domain.SchemaLookup     = (*Repository)(nil)
)

// ============================================================================
// Repository Implementation
// ============================================================================

// Repository implements domain.SchemaRepository using PostgreSQL.
type Repository struct {
	queries *generated.Queries
}

// NewRepository creates a new Repository.
func NewRepository(queries *generated.Queries) *Repository {
	return &Repository{
		queries: queries,
	}
}

// ============================================================================
// Write Operations
// ============================================================================

// Save persists a schema aggregate.
func (r *Repository) Save(ctx context.Context, schema *domain.Schema) error {
	const op = "postgres.Repository.Save"

	// Check if schema exists
	exists, err := r.queries.SchemaExistsByID(ctx, schema.ID().UUID())
	if err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	if exists {
		return r.update(ctx, op, schema)
	}

	return r.insert(ctx, op, schema)
}

// insert creates a new schema record.
func (r *Repository) insert(ctx context.Context, op string, schema *domain.Schema) error {
	// Insert schema
	params := ToInsertSchemaParams(schema)
	if err := r.queries.InsertSchema(ctx, params); err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	// Insert claims
	now := time.Now().UTC()
	for _, claim := range schema.Claims() {
		claimParams, err := ToInsertSchemaClaimParams(schema.ID(), claim, now)
		if err != nil {
			return pkgerrors.Wrap(err, op)
		}
		if err := r.queries.InsertSchemaClaim(ctx, claimParams); err != nil {
			return pkgerrors.Infrastructure(op, err)
		}
	}

	// Insert initial version
	versionParams, err := ToInsertSchemaVersionParams(
		schema.ID(),
		schema.CurrentVersion(),
		schema.Claims(),
		"Initial version",
		now,
	)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}
	if err := r.queries.InsertSchemaVersion(ctx, versionParams); err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	return nil
}

// update updates an existing schema record.
func (r *Repository) update(ctx context.Context, op string, schema *domain.Schema) error {
	// Update schema
	params := ToUpdateSchemaParams(schema)
	if err := r.queries.UpdateSchema(ctx, params); err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	// Check if version changed (new version added)
	versionExists, err := r.queries.SchemaVersionExists(ctx, generated.SchemaVersionExistsParams{
		SchemaID: schema.ID().UUID(),
		Version:  schema.CurrentVersion().String(),
	})
	if err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	if !versionExists {
		// New version - delete old claims and insert new ones
		if err := r.queries.DeleteSchemaClaimsBySchemaID(ctx, schema.ID().UUID()); err != nil {
			return pkgerrors.Infrastructure(op, err)
		}

		now := time.Now().UTC()
		for _, claim := range schema.Claims() {
			claimParams, err := ToInsertSchemaClaimParams(schema.ID(), claim, now)
			if err != nil {
				return pkgerrors.Wrap(err, op)
			}
			if err := r.queries.InsertSchemaClaim(ctx, claimParams); err != nil {
				return pkgerrors.Infrastructure(op, err)
			}
		}

		// Insert new version record
		versionParams, err := ToInsertSchemaVersionParams(
			schema.ID(),
			schema.CurrentVersion(),
			schema.Claims(),
			"", // Change summary would come from the event
			now,
		)
		if err != nil {
			return pkgerrors.Wrap(err, op)
		}
		if err := r.queries.InsertSchemaVersion(ctx, versionParams); err != nil {
			return pkgerrors.Infrastructure(op, err)
		}
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// GetByID loads a schema by its unique identifier.
func (r *Repository) GetByID(ctx context.Context, id types.ID) (*domain.Schema, error) {
	const op = "postgres.Repository.GetByID"

	row, err := r.queries.GetSchemaByID(ctx, id.UUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.SchemaNotFound(op, id.String())
		}
		return nil, pkgerrors.Infrastructure(op, err)
	}

	claimRows, err := r.queries.GetSchemaClaimsBySchemaID(ctx, id.UUID())
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	return ToDomainSchema(row, claimRows)
}

// GetByType loads a schema by its type name.
func (r *Repository) GetByType(ctx context.Context, schemaType string) (*domain.Schema, error) {
	const op = "postgres.Repository.GetByType"

	row, err := r.queries.GetSchemaByType(ctx, schemaType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.SchemaNotFound(op, schemaType)
		}
		return nil, pkgerrors.Infrastructure(op, err)
	}

	claimRows, err := r.queries.GetSchemaClaimsBySchemaID(ctx, row.ID)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	return ToDomainSchema(row, claimRows)
}

// GetByTypeAndVersion loads a specific version of a schema.
func (r *Repository) GetByTypeAndVersion(ctx context.Context, schemaType string, version domain.SchemaVersion) (*domain.Schema, error) {
	const op = "postgres.Repository.GetByTypeAndVersion"

	// First get the schema by type
	row, err := r.queries.GetSchemaByType(ctx, schemaType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.SchemaNotFound(op, schemaType)
		}
		return nil, pkgerrors.Infrastructure(op, err)
	}

	// Check if the requested version exists
	versionRow, err := r.queries.GetSchemaVersion(ctx, generated.GetSchemaVersionParams{
		SchemaID: row.ID,
		Version:  version.String(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.VersionNotFound(op, row.ID.String(), version.String())
		}
		return nil, pkgerrors.Infrastructure(op, err)
	}

	// If it's the current version, return normally
	if row.CurrentVersion == version.String() {
		claimRows, err := r.queries.GetSchemaClaimsBySchemaID(ctx, row.ID)
		if err != nil {
			return nil, pkgerrors.Infrastructure(op, err)
		}
		return ToDomainSchema(row, claimRows)
	}

	// For historical versions, reconstruct from snapshot
	return r.reconstructFromSnapshot(ctx, op, row, versionRow)
}

// reconstructFromSnapshot rebuilds a schema from a version snapshot.
func (r *Repository) reconstructFromSnapshot(ctx context.Context, op string, schemaRow generated.Schema, versionRow generated.SchemaVersion) (*domain.Schema, error) {
	// Parse claims from snapshot
	var claimData []domain.ClaimDefinitionData
	if err := json.Unmarshal(versionRow.ClaimsSnapshot, &claimData); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	claims := make([]domain.ClaimDefinition, 0, len(claimData))
	for _, cd := range claimData {
		claim, err := cd.ToClaimDefinition()
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		claims = append(claims, claim)
	}

	// Build claim definition data for reconstitution
	claimDataForEvent := make([]domain.ClaimDefinitionData, len(claims))
	for i, c := range claims {
		claimDataForEvent[i] = domain.ClaimDefinitionToData(c)
	}

	issuerIDStr := ""
	if schemaRow.IssuerID.Valid {
		issuerIDStr = uuidToString(schemaRow.IssuerID.Bytes)
	}

	schemaID := mustParseTypesID(schemaRow.ID.String())
	schema := domain.NewSchema(schemaID)

	// Apply with the historical version
	registeredEvent := domain.SchemaRegisteredEvent{
		SchemaID:     schemaID.String(),
		SchemaType:   schemaRow.SchemaType,
		Name:         schemaRow.Name,
		Description:  schemaRow.Description,
		Version:      versionRow.Version, // Use the historical version
		Claims:       claimDataForEvent,
		IssuerID:     issuerIDStr,
		RegisteredAt: versionRow.CreatedAt,
	}
	schema.ApplyEvent(registeredEvent)
	schema.ClearChanges()

	return schema, nil
}

// List returns schemas matching the given filter.
func (r *Repository) List(ctx context.Context, filter domain.SchemaFilter) ([]*domain.Schema, error) {
	const op = "postgres.Repository.List"

	params := ToListSchemasParams(filter)
	rows, err := r.queries.ListSchemas(ctx, params)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	schemas := make([]*domain.Schema, 0, len(rows))
	for _, row := range rows {
		claimRows, err := r.queries.GetSchemaClaimsBySchemaID(ctx, row.ID)
		if err != nil {
			return nil, pkgerrors.Infrastructure(op, err)
		}

		schema, err := ToDomainSchema(row, claimRows)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// Count returns the total number of schemas matching the filter.
func (r *Repository) Count(ctx context.Context, filter domain.SchemaFilter) (int64, error) {
	const op = "postgres.Repository.Count"

	params := ToCountSchemasParams(filter)
	count, err := r.queries.CountSchemas(ctx, params)
	if err != nil {
		return 0, pkgerrors.Infrastructure(op, err)
	}

	return count, nil
}

// Exists checks if a schema with the given ID exists.
func (r *Repository) Exists(ctx context.Context, id types.ID) (bool, error) {
	const op = "postgres.Repository.Exists"

	exists, err := r.queries.SchemaExistsByID(ctx, id.UUID())
	if err != nil {
		return false, pkgerrors.Infrastructure(op, err)
	}

	return exists, nil
}

// ExistsByType checks if a schema with the given type name exists.
func (r *Repository) ExistsByType(ctx context.Context, schemaType string) (bool, error) {
	const op = "postgres.Repository.ExistsByType"

	exists, err := r.queries.SchemaExistsByType(ctx, schemaType)
	if err != nil {
		return false, pkgerrors.Infrastructure(op, err)
	}

	return exists, nil
}

// ============================================================================
// SchemaLookup Implementation
// ============================================================================

// GetIDByType returns the schema ID for a given type, if it exists.
func (r *Repository) GetIDByType(ctx context.Context, schemaType string) (types.ID, error) {
	const op = "postgres.Repository.GetIDByType"

	id, err := r.queries.GetSchemaIDByType(ctx, schemaType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return types.ID{}, domain.SchemaNotFound(op, schemaType)
		}
		return types.ID{}, pkgerrors.Infrastructure(op, err)
	}

	return mustParseTypesID(id.String()), nil
}
