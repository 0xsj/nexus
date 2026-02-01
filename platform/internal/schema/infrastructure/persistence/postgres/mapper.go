package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/schema/app/query"
	"github.com/0xsj/nexus/platform/internal/schema/domain"
	generated "github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Domain to Database Mapping
// ============================================================================

// ToInsertSchemaParams converts a domain Schema to database insert params.
func ToInsertSchemaParams(schema *domain.Schema) generated.InsertSchemaParams {
	var issuerID pgtype.UUID
	if schema.IssuerID() != nil {
		issuerID = pgtype.UUID{
			Bytes: schema.IssuerID().UUID(),
			Valid: true,
		}
	}

	return generated.InsertSchemaParams{
		ID:             schema.ID().UUID(),
		SchemaType:     schema.SchemaType(),
		Name:           schema.Name(),
		Description:    schema.Description(),
		CurrentVersion: schema.CurrentVersion().String(),
		Status:         schema.Status().String(),
		IssuerID:       issuerID,
		CreatedAt:      schema.CreatedAt(),
		UpdatedAt:      schema.UpdatedAt(),
	}
}

// ToUpdateSchemaParams converts a domain Schema to database update params.
func ToUpdateSchemaParams(schema *domain.Schema) generated.UpdateSchemaParams {
	return generated.UpdateSchemaParams{
		ID:             schema.ID().UUID(),
		Name:           schema.Name(),
		Description:    schema.Description(),
		CurrentVersion: schema.CurrentVersion().String(),
		Status:         schema.Status().String(),
		UpdatedAt:      schema.UpdatedAt(),
	}
}

// ToInsertSchemaClaimParams converts a domain ClaimDefinition to database insert params.
func ToInsertSchemaClaimParams(schemaID types.ID, claim domain.ClaimDefinition, now time.Time) (generated.InsertSchemaClaimParams, error) {
	var constraints []byte
	if claim.ClaimType().HasConstraints() {
		c := claim.ClaimType().Constraints()
		var err error
		constraints, err = json.Marshal(c)
		if err != nil {
			return generated.InsertSchemaClaimParams{}, err
		}
	}

	var displayName *string
	if dn := claim.DisplayName(); dn != "" && dn != claim.Key() {
		displayName = &dn
	}

	var description *string
	if desc := claim.Description(); desc != "" {
		description = &desc
	}

	return generated.InsertSchemaClaimParams{
		SchemaID:    schemaID.UUID(),
		Key:         claim.Key(),
		DataType:    claim.DataType().String(),
		Required:    claim.IsRequired(),
		DisplayName: displayName,
		Description: description,
		Disclosable: claim.IsDisclosable(),
		SortOrder:   int32(claim.Order()),
		Constraints: constraints,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ToInsertSchemaVersionParams converts version info to database insert params.
func ToInsertSchemaVersionParams(schemaID types.ID, version domain.SchemaVersion, claims []domain.ClaimDefinition, changeSummary string, now time.Time) (generated.InsertSchemaVersionParams, error) {
	claimsSnapshot, err := claimsToJSON(claims)
	if err != nil {
		return generated.InsertSchemaVersionParams{}, err
	}

	var summary *string
	if changeSummary != "" {
		summary = &changeSummary
	}

	return generated.InsertSchemaVersionParams{
		SchemaID:       schemaID.UUID(),
		Version:        version.String(),
		ChangeSummary:  summary,
		ClaimsSnapshot: claimsSnapshot,
		CreatedAt:      now,
	}, nil
}

// ============================================================================
// Database to Domain Mapping
// ============================================================================

// ToDomainSchema converts database rows to a domain Schema.
// Note: This reconstitutes the schema by replaying events or direct state assignment.
// For proper event sourcing, use the event store. This is for state-based persistence.
func ToDomainSchema(row generated.Schema, claimRows []generated.SchemaClaim) (*domain.Schema, error) {
	schemaID := mustParseTypesID(row.ID.String())

	version, err := domain.ParseSchemaVersion(row.CurrentVersion)
	if err != nil {
		return nil, err
	}

	var issuerID *types.ID
	if row.IssuerID.Valid {
		id := mustParseTypesID(uuidToString(row.IssuerID.Bytes))
		issuerID = &id
	}

	claims, err := toDomainClaims(claimRows)
	if err != nil {
		return nil, err
	}

	// Build a SchemaRegisteredEvent to reconstitute via event application
	claimData := make([]domain.ClaimDefinitionData, len(claims))
	for i, c := range claims {
		claimData[i] = domain.ClaimDefinitionToData(c)
	}

	issuerIDStr := ""
	if issuerID != nil {
		issuerIDStr = issuerID.String()
	}

	// Create schema and apply reconstitution event
	schema := domain.NewSchema(schemaID)

	// Apply registered event to set initial state
	registeredEvent := domain.SchemaRegisteredEvent{
		SchemaID:     schemaID.String(),
		SchemaType:   row.SchemaType,
		Name:         row.Name,
		Description:  row.Description,
		Version:      row.CurrentVersion,
		Claims:       claimData,
		IssuerID:     issuerIDStr,
		RegisteredAt: row.CreatedAt,
	}
	schema.ApplyEvent(registeredEvent)

	// If status is deprecated, apply deprecated event
	if row.Status == domain.SchemaStatusDeprecated.String() {
		deprecatedEvent := domain.SchemaDeprecatedEvent{
			SchemaID:     schemaID.String(),
			DeprecatedAt: row.UpdatedAt,
		}
		schema.ApplyEvent(deprecatedEvent)
	}

	// If version differs from what was registered, we need to update
	// This handles the case where version was updated after registration
	if row.CurrentVersion != version.String() {
		// The version is already correct from the registered event
		// If there were version updates, they would need separate tracking
	}

	// Clear changes since this is reconstitution, not new commands
	schema.ClearChanges()

	return schema, nil
}

// toDomainClaims converts database claim rows to domain ClaimDefinitions.
func toDomainClaims(rows []generated.SchemaClaim) ([]domain.ClaimDefinition, error) {
	claims := make([]domain.ClaimDefinition, 0, len(rows))

	for _, row := range rows {
		claim, err := toDomainClaim(row)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	return claims, nil
}

// toDomainClaim converts a database claim row to a domain ClaimDefinition.
func toDomainClaim(row generated.SchemaClaim) (domain.ClaimDefinition, error) {
	dataType, err := domain.ParseDataType(row.DataType)
	if err != nil {
		return domain.ClaimDefinition{}, err
	}

	// Build claim type options from constraints
	var claimTypeOpts []domain.ClaimTypeOption
	if len(row.Constraints) > 0 {
		var constraints domain.ClaimConstraints
		if err := json.Unmarshal(row.Constraints, &constraints); err != nil {
			return domain.ClaimDefinition{}, err
		}

		if len(constraints.AllowedValues) > 0 {
			claimTypeOpts = append(claimTypeOpts, domain.WithAllowedValues(constraints.AllowedValues...))
		}
		if constraints.Min != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMin(*constraints.Min))
		}
		if constraints.Max != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMax(*constraints.Max))
		}
		if constraints.MinLength != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMinLength(*constraints.MinLength))
		}
		if constraints.MaxLength != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithMaxLength(*constraints.MaxLength))
		}
		if constraints.Pattern != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithPattern(*constraints.Pattern))
		}
		if constraints.Format != nil {
			claimTypeOpts = append(claimTypeOpts, domain.WithFormat(*constraints.Format))
		}
	}

	claimType, err := domain.NewClaimType(dataType, claimTypeOpts...)
	if err != nil {
		return domain.ClaimDefinition{}, err
	}

	// Build claim definition options
	var claimOpts []domain.ClaimDefinitionOption
	if row.Required {
		claimOpts = append(claimOpts, domain.Required())
	}
	if row.DisplayName != nil && *row.DisplayName != "" {
		claimOpts = append(claimOpts, domain.WithDisplayName(*row.DisplayName))
	}
	if row.Description != nil && *row.Description != "" {
		claimOpts = append(claimOpts, domain.WithDescription(*row.Description))
	}
	if !row.Disclosable {
		claimOpts = append(claimOpts, domain.NonDisclosable())
	}
	if row.SortOrder != 0 {
		claimOpts = append(claimOpts, domain.WithOrder(int(row.SortOrder)))
	}

	return domain.NewClaimDefinition(row.Key, claimType, claimOpts...)
}

// ============================================================================
// Database to View Mapping
// ============================================================================

// ToSchemaView converts database rows to a query SchemaView.
func ToSchemaView(row generated.Schema, claimRows []generated.SchemaClaim) query.SchemaView {
	issuerID := ""
	if row.IssuerID.Valid {
		issuerID = uuidToString(row.IssuerID.Bytes)
	}

	return query.SchemaView{
		SchemaID:       row.ID.String(),
		SchemaType:     row.SchemaType,
		Name:           row.Name,
		Description:    row.Description,
		CurrentVersion: row.CurrentVersion,
		Status:         row.Status,
		IssuerID:       issuerID,
		Claims:         toClaimViews(claimRows),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

// ToSchemaSummaryView converts a database row to a query SchemaSummaryView.
func ToSchemaSummaryView(row generated.Schema, claimCount int) query.SchemaSummaryView {
	return query.SchemaSummaryView{
		SchemaID:       row.ID.String(),
		SchemaType:     row.SchemaType,
		Name:           row.Name,
		CurrentVersion: row.CurrentVersion,
		Status:         row.Status,
		ClaimCount:     claimCount,
		CreatedAt:      row.CreatedAt,
	}
}

// toClaimViews converts database claim rows to query ClaimViews.
func toClaimViews(rows []generated.SchemaClaim) []query.ClaimView {
	views := make([]query.ClaimView, len(rows))
	for i, row := range rows {
		views[i] = toClaimView(row)
	}
	return views
}

// toClaimView converts a database claim row to a query ClaimView.
func toClaimView(row generated.SchemaClaim) query.ClaimView {
	displayName := ""
	if row.DisplayName != nil {
		displayName = *row.DisplayName
	}

	description := ""
	if row.Description != nil {
		description = *row.Description
	}

	view := query.ClaimView{
		Key:         row.Key,
		DataType:    row.DataType,
		Required:    row.Required,
		DisplayName: displayName,
		Description: description,
		Disclosable: row.Disclosable,
		Order:       int(row.SortOrder),
	}

	// Parse constraints if present
	if len(row.Constraints) > 0 {
		var constraints query.ConstraintsView
		if err := json.Unmarshal(row.Constraints, &constraints); err == nil {
			view.Constraints = &constraints
		}
	}

	return view
}

// ToSchemaVersionView converts a database version row to a query SchemaVersionView.
func ToSchemaVersionView(schemaRow generated.Schema, versionRow generated.SchemaVersion, isLatest bool) query.SchemaVersionView {
	// Parse claims from snapshot
	var claims []query.ClaimView
	if len(versionRow.ClaimsSnapshot) > 0 {
		_ = json.Unmarshal(versionRow.ClaimsSnapshot, &claims)
	}

	return query.SchemaVersionView{
		SchemaID:   schemaRow.ID.String(),
		SchemaType: schemaRow.SchemaType,
		Name:       schemaRow.Name,
		Version:    versionRow.Version,
		Status:     schemaRow.Status,
		Claims:     claims,
		IsLatest:   isLatest,
		CreatedAt:  versionRow.CreatedAt,
	}
}

// ToVersionListView converts database version rows to a query VersionListView.
func ToVersionListView(schemaRow generated.Schema, versionRows []generated.SchemaVersion) query.VersionListView {
	versions := make([]query.VersionInfo, len(versionRows))
	for i, row := range versionRows {
		versions[i] = query.VersionInfo{
			Version:   row.Version,
			IsLatest:  row.Version == schemaRow.CurrentVersion,
			CreatedAt: row.CreatedAt,
		}
	}

	return query.VersionListView{
		SchemaID:       schemaRow.ID.String(),
		SchemaType:     schemaRow.SchemaType,
		CurrentVersion: schemaRow.CurrentVersion,
		Versions:       versions,
	}
}

// ============================================================================
// Filter Mapping
// ============================================================================

// ToListSchemasParams converts a domain filter to database list params.
func ToListSchemasParams(filter domain.SchemaFilter) generated.ListSchemasParams {
	status := ""
	if filter.Status != nil {
		status = filter.Status.String()
	}

	var issuerID uuid.UUID
	if filter.IssuerID != nil {
		issuerID = filter.IssuerID.UUID()
	}

	return generated.ListSchemasParams{
		Status:      status,
		IssuerID:    issuerID,
		BuiltInOnly: filter.BuiltInOnly != nil && *filter.BuiltInOnly,
		SearchQuery: filter.SearchQuery,
		OrderBy:     string(filter.OrderBy),
		OrderDir:    string(filter.OrderDir),
		PageOffset:  int32(filter.Offset),
		PageLimit:   int32(filter.Limit),
	}
}

// ToCountSchemasParams converts a domain filter to database count params.
func ToCountSchemasParams(filter domain.SchemaFilter) generated.CountSchemasParams {
	status := ""
	if filter.Status != nil {
		status = filter.Status.String()
	}

	var issuerID uuid.UUID
	if filter.IssuerID != nil {
		issuerID = filter.IssuerID.UUID()
	}

	return generated.CountSchemasParams{
		Status:      status,
		IssuerID:    issuerID,
		BuiltInOnly: filter.BuiltInOnly != nil && *filter.BuiltInOnly,
		SearchQuery: filter.SearchQuery,
	}
}

// ============================================================================
// Helpers
// ============================================================================

// claimsToJSON serializes claims to JSON for version snapshots.
func claimsToJSON(claims []domain.ClaimDefinition) (json.RawMessage, error) {
	claimData := make([]domain.ClaimDefinitionData, len(claims))
	for i, claim := range claims {
		claimData[i] = domain.ClaimDefinitionToData(claim)
	}
	return json.Marshal(claimData)
}

// mustParseTypesID parses a string to types.ID, panics on error.
func mustParseTypesID(s string) types.ID {
	id, err := types.ParseID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// uuidToString converts a [16]byte UUID to string.
func uuidToString(b [16]byte) string {
	u := uuid.UUID(b)
	return u.String()
}
