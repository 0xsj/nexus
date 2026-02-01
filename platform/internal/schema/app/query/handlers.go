package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/schema/domain"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Schema context.
type Handlers struct {
	schemaRepo   domain.SchemaRepository
	schemaLookup domain.SchemaLookup
	logger       log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	schemaRepo domain.SchemaRepository,
	schemaLookup domain.SchemaLookup,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		schemaRepo:   schemaRepo,
		schemaLookup: schemaLookup,
		logger:       logger,
	}
}

// ============================================================================
// Schema Query Handlers
// ============================================================================

// HandleGetSchema handles the GetSchema query.
func (h *Handlers) HandleGetSchema(ctx context.Context, q GetSchema) (*SchemaView, error) {
	const op = "Handlers.HandleGetSchema"

	schema, err := h.schemaRepo.GetByID(ctx, q.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapSchemaToView(schema), nil
}

// HandleGetSchemaByType handles the GetSchemaByType query.
func (h *Handlers) HandleGetSchemaByType(ctx context.Context, q GetSchemaByType) (*SchemaView, error) {
	const op = "Handlers.HandleGetSchemaByType"

	schema, err := h.schemaRepo.GetByType(ctx, q.SchemaType)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return h.mapSchemaToView(schema), nil
}

// HandleListSchemas handles the ListSchemas query.
func (h *Handlers) HandleListSchemas(ctx context.Context, q ListSchemas) (*SchemaListView, error) {
	const op = "Handlers.HandleListSchemas"

	// Build filter from query
	filter := domain.NewSchemaFilter()

	if q.Status != nil {
		status := domain.SchemaStatus(*q.Status)
		filter = filter.WithStatus(status)
	}

	if q.IssuerID != nil {
		filter = filter.WithIssuerID(*q.IssuerID)
	}

	if q.BuiltInOnly != nil {
		if *q.BuiltInOnly {
			filter = filter.WithBuiltInOnly()
		} else {
			filter = filter.WithCustomOnly()
		}
	}

	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	filter = filter.WithLimit(pageSize)

	// TODO: Handle cursor-based pagination
	// if q.Cursor != nil {
	//     offset := decodeCursor(*q.Cursor)
	//     filter = filter.WithOffset(offset)
	// }

	schemas, err := h.schemaRepo.List(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.schemaRepo.Count(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]SchemaSummaryView, len(schemas))
	for i, schema := range schemas {
		summaries[i] = h.mapSchemaToSummary(schema)
	}

	// TODO: Generate next cursor if more results
	var nextCursor *string

	return &SchemaListView{
		Schemas:    summaries,
		NextCursor: nextCursor,
		TotalCount: int(totalCount),
	}, nil
}

// HandleSearchSchemas handles the SearchSchemas query.
func (h *Handlers) HandleSearchSchemas(ctx context.Context, q SearchSchemas) (*SchemaListView, error) {
	const op = "Handlers.HandleSearchSchemas"

	filter := domain.NewSchemaFilter().
		WithSearchQuery(q.Query)

	if q.Status != nil {
		status := domain.SchemaStatus(*q.Status)
		filter = filter.WithStatus(status)
	}

	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	filter = filter.WithLimit(pageSize)

	schemas, err := h.schemaRepo.List(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.schemaRepo.Count(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]SchemaSummaryView, len(schemas))
	for i, schema := range schemas {
		summaries[i] = h.mapSchemaToSummary(schema)
	}

	var nextCursor *string

	return &SchemaListView{
		Schemas:    summaries,
		NextCursor: nextCursor,
		TotalCount: int(totalCount),
	}, nil
}

// ============================================================================
// Version Query Handlers
// ============================================================================

// HandleGetSchemaVersion handles the GetSchemaVersion query.
func (h *Handlers) HandleGetSchemaVersion(ctx context.Context, q GetSchemaVersion) (*SchemaVersionView, error) {
	const op = "Handlers.HandleGetSchemaVersion"

	version, err := domain.ParseSchemaVersion(q.Version)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	schema, err := h.schemaRepo.GetByID(ctx, q.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Check if requested version matches current version
	// Note: For full version history support, you'd need a separate version repository
	if !schema.CurrentVersion().Equals(version) {
		return nil, domain.VersionNotFound(op, q.SchemaID.String(), q.Version)
	}

	return &SchemaVersionView{
		SchemaID:   schema.ID().String(),
		SchemaType: schema.SchemaType(),
		Name:       schema.Name(),
		Version:    schema.CurrentVersion().String(),
		Status:     schema.Status().String(),
		Claims:     h.mapClaimsToViews(schema.Claims()),
		IsLatest:   true,
		CreatedAt:  schema.CreatedAt(),
	}, nil
}

// HandleListSchemaVersions handles the ListSchemaVersions query.
func (h *Handlers) HandleListSchemaVersions(ctx context.Context, q ListSchemaVersions) (*VersionListView, error) {
	const op = "Handlers.HandleListSchemaVersions"

	schema, err := h.schemaRepo.GetByID(ctx, q.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Note: For full version history, you'd need event sourcing replay or a version table
	// Currently, we only have the current version
	versions := []VersionInfo{
		{
			Version:   schema.CurrentVersion().String(),
			IsLatest:  true,
			CreatedAt: schema.CreatedAt(),
		},
	}

	return &VersionListView{
		SchemaID:       schema.ID().String(),
		SchemaType:     schema.SchemaType(),
		CurrentVersion: schema.CurrentVersion().String(),
		Versions:       versions,
	}, nil
}

// ============================================================================
// Lookup Query Handlers
// ============================================================================

// HandleSchemaExists handles the SchemaExists query.
func (h *Handlers) HandleSchemaExists(ctx context.Context, q SchemaExists) (*SchemaExistsView, error) {
	const op = "Handlers.HandleSchemaExists"

	// Check by ID if provided
	if q.SchemaID != nil && !q.SchemaID.IsZero() {
		exists, err := h.schemaRepo.Exists(ctx, *q.SchemaID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}

		result := &SchemaExistsView{
			Exists:   exists,
			SchemaID: q.SchemaID.String(),
		}

		if exists {
			schema, err := h.schemaRepo.GetByID(ctx, *q.SchemaID)
			if err == nil {
				result.SchemaType = schema.SchemaType()
			}
		}

		return result, nil
	}

	// Check by type if provided
	if q.SchemaType != nil && *q.SchemaType != "" {
		exists, err := h.schemaLookup.ExistsByType(ctx, *q.SchemaType)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}

		result := &SchemaExistsView{
			Exists:     exists,
			SchemaType: *q.SchemaType,
		}

		if exists {
			schemaID, err := h.schemaLookup.GetIDByType(ctx, *q.SchemaType)
			if err == nil {
				result.SchemaID = schemaID.String()
			}
		}

		return result, nil
	}

	return &SchemaExistsView{Exists: false}, nil
}

// HandleResolveSchemaType handles the ResolveSchemaType query.
func (h *Handlers) HandleResolveSchemaType(ctx context.Context, q ResolveSchemaType) (*SchemaTypeResolutionView, error) {
	const op = "Handlers.HandleResolveSchemaType"

	schema, err := h.schemaRepo.GetByType(ctx, q.SchemaType)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &SchemaTypeResolutionView{
		SchemaType:     schema.SchemaType(),
		SchemaID:       schema.ID().String(),
		CurrentVersion: schema.CurrentVersion().String(),
		Status:         schema.Status().String(),
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapSchemaToView maps a Schema aggregate to SchemaView.
func (h *Handlers) mapSchemaToView(schema *domain.Schema) *SchemaView {
	issuerID := ""
	if schema.IssuerID() != nil {
		issuerID = schema.IssuerID().String()
	}

	return &SchemaView{
		SchemaID:       schema.ID().String(),
		SchemaType:     schema.SchemaType(),
		Name:           schema.Name(),
		Description:    schema.Description(),
		CurrentVersion: schema.CurrentVersion().String(),
		Status:         schema.Status().String(),
		IssuerID:       issuerID,
		Claims:         h.mapClaimsToViews(schema.Claims()),
		CreatedAt:      schema.CreatedAt(),
		UpdatedAt:      schema.UpdatedAt(),
	}
}

// mapSchemaToSummary maps a Schema aggregate to SchemaSummaryView.
func (h *Handlers) mapSchemaToSummary(schema *domain.Schema) SchemaSummaryView {
	return SchemaSummaryView{
		SchemaID:       schema.ID().String(),
		SchemaType:     schema.SchemaType(),
		Name:           schema.Name(),
		CurrentVersion: schema.CurrentVersion().String(),
		Status:         schema.Status().String(),
		ClaimCount:     schema.ClaimCount(),
		CreatedAt:      schema.CreatedAt(),
	}
}

// mapClaimsToViews maps claim definitions to ClaimViews.
func (h *Handlers) mapClaimsToViews(claims []domain.ClaimDefinition) []ClaimView {
	views := make([]ClaimView, len(claims))
	for i, claim := range claims {
		views[i] = h.mapClaimToView(claim)
	}
	return views
}

// mapClaimToView maps a ClaimDefinition to ClaimView.
func (h *Handlers) mapClaimToView(claim domain.ClaimDefinition) ClaimView {
	view := ClaimView{
		Key:         claim.Key(),
		DataType:    claim.DataType().String(),
		Required:    claim.IsRequired(),
		DisplayName: claim.DisplayName(),
		Description: claim.Description(),
		Disclosable: claim.IsDisclosable(),
		Order:       claim.Order(),
	}

	// Map constraints if present
	if claim.ClaimType().HasConstraints() {
		constraints := claim.ClaimType().Constraints()
		view.Constraints = &ConstraintsView{
			AllowedValues: constraints.AllowedValues,
			Min:           constraints.Min,
			Max:           constraints.Max,
			MinLength:     constraints.MinLength,
			MaxLength:     constraints.MaxLength,
			Pattern:       constraints.Pattern,
			Format:        constraints.Format,
		}
	}

	return view
}
