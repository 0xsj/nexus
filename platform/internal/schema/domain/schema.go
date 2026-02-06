package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Schema Status
// ============================================================================

// SchemaStatus represents the lifecycle status of a schema.
type SchemaStatus string

const (
	// SchemaStatusActive indicates the schema is active and can be used.
	SchemaStatusActive SchemaStatus = "active"

	// SchemaStatusDeprecated indicates the schema is deprecated.
	SchemaStatusDeprecated SchemaStatus = "deprecated"
)

// String returns the string representation of the status.
func (s SchemaStatus) String() string {
	return string(s)
}

// IsValid returns true if the status is valid.
func (s SchemaStatus) IsValid() bool {
	return s == SchemaStatusActive || s == SchemaStatusDeprecated
}

// ============================================================================
// Schema Aggregate
// ============================================================================

// Schema is the aggregate root for credential schema management.
// It defines what kinds of credentials can be issued and their claim structure.
type Schema struct {
	eventsourcing.AggregateRoot

	id             types.ID
	schemaType     string
	name           string
	description    string
	currentVersion SchemaVersion
	claims         []ClaimDefinition
	status         SchemaStatus
	issuerID       *types.ID
	createdAt      time.Time
	updatedAt      time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// NewSchema creates a new Schema aggregate (for hydration).
func NewSchema(id types.ID) *Schema {
	s := &Schema{
		id: id,
	}
	s.AggregateRoot.InitAggregate(AggregateTypeSchema, id.String())
	return s
}

// RegisterSchema creates and registers a new schema.
func RegisterSchema(
	id types.ID,
	schemaType string,
	name string,
	description string,
	version SchemaVersion,
	claims []ClaimDefinition,
	issuerID *types.ID,
) (*Schema, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, SchemaInvalid("RegisterSchema", "id is required")
	}
	if schemaType == "" {
		return nil, SchemaInvalid("RegisterSchema", "schema type is required")
	}
	if name == "" {
		return nil, SchemaInvalid("RegisterSchema", "name is required")
	}
	if version.IsZero() {
		return nil, SchemaInvalid("RegisterSchema", "version is required")
	}
	if len(claims) == 0 {
		return nil, SchemaInvalid("RegisterSchema", "at least one claim is required")
	}

	// Validate claims have unique keys
	if err := validateUniqueClaims(claims); err != nil {
		return nil, err
	}

	// Create aggregate
	s := NewSchema(id)

	// Resolve issuer ID string
	issuerIDStr := ""
	if issuerID != nil {
		issuerIDStr = issuerID.String()
	}

	// Raise event
	event := NewSchemaRegisteredEvent(
		id.String(),
		schemaType,
		name,
		description,
		version,
		claims,
		issuerIDStr,
	)
	s.AggregateRoot.Raise(s, event)

	return s, nil
}

// ============================================================================
// Command Methods
// ============================================================================

// AddVersion adds a new version to the schema.
func (s *Schema) AddVersion(
	version SchemaVersion,
	claims []ClaimDefinition,
	changeSummary string,
) error {
	if s.status == SchemaStatusDeprecated {
		return SchemaDeprecated("Schema.AddVersion", s.id.String())
	}

	if !version.IsNewerThan(s.currentVersion) {
		return VersionAlreadyExists("Schema.AddVersion", s.id.String(), version.String())
	}

	if len(claims) == 0 {
		return SchemaInvalid("Schema.AddVersion", "at least one claim is required")
	}

	if err := validateUniqueClaims(claims); err != nil {
		return err
	}

	event := NewSchemaVersionAddedEvent(
		s.id.String(),
		version,
		s.currentVersion,
		claims,
		changeSummary,
	)
	s.AggregateRoot.Raise(s, event)

	return nil
}

// Deprecate marks the schema as deprecated.
func (s *Schema) Deprecate(reason string, replacementSchemaID *types.ID) error {
	if s.status == SchemaStatusDeprecated {
		return SchemaDeprecated("Schema.Deprecate", s.id.String())
	}

	replacementIDStr := ""
	if replacementSchemaID != nil {
		replacementIDStr = replacementSchemaID.String()
	}

	event := NewSchemaDeprecatedEvent(
		s.id.String(),
		reason,
		replacementIDStr,
	)
	s.AggregateRoot.Raise(s, event)

	return nil
}

// Activate reactivates a deprecated schema.
func (s *Schema) Activate() error {
	if s.status == SchemaStatusActive {
		return SchemaInvalid("Schema.Activate", "schema is already active")
	}

	event := NewSchemaActivatedEvent(s.id.String())
	s.AggregateRoot.Raise(s, event)

	return nil
}

// UpdateMetadata updates the schema name and/or description.
func (s *Schema) UpdateMetadata(name string, description string) error {
	if name == "" && description == "" {
		return SchemaInvalid("Schema.UpdateMetadata", "at least one field must be provided")
	}

	event := NewSchemaMetadataUpdatedEvent(
		s.id.String(),
		name,
		description,
	)
	s.AggregateRoot.Raise(s, event)

	return nil
}

// ============================================================================
// Query Methods
// ============================================================================

// ID returns the schema ID.
func (s *Schema) ID() types.ID {
	return s.id
}

// SchemaType returns the schema type name.
func (s *Schema) SchemaType() string {
	return s.schemaType
}

// Name returns the human-readable name.
func (s *Schema) Name() string {
	return s.name
}

// Description returns the schema description.
func (s *Schema) Description() string {
	return s.description
}

// CurrentVersion returns the current version.
func (s *Schema) CurrentVersion() SchemaVersion {
	return s.currentVersion
}

// Claims returns the current claim definitions.
func (s *Schema) Claims() []ClaimDefinition {
	result := make([]ClaimDefinition, len(s.claims))
	copy(result, s.claims)
	return result
}

// Status returns the schema status.
func (s *Schema) Status() SchemaStatus {
	return s.status
}

// IssuerID returns the optional issuer ID.
func (s *Schema) IssuerID() *types.ID {
	return s.issuerID
}

// CreatedAt returns the creation timestamp.
func (s *Schema) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the last update timestamp.
func (s *Schema) UpdatedAt() time.Time {
	return s.updatedAt
}

// IsActive returns true if the schema is active.
func (s *Schema) IsActive() bool {
	return s.status == SchemaStatusActive
}

// IsDeprecated returns true if the schema is deprecated.
func (s *Schema) IsDeprecated() bool {
	return s.status == SchemaStatusDeprecated
}

// IsBuiltIn returns true if this is a built-in schema (no issuer).
func (s *Schema) IsBuiltIn() bool {
	return s.issuerID == nil
}

// IsCustom returns true if this is a custom issuer schema.
func (s *Schema) IsCustom() bool {
	return s.issuerID != nil
}

// GetClaim returns a claim by key.
func (s *Schema) GetClaim(key string) (ClaimDefinition, bool) {
	for _, claim := range s.claims {
		if claim.Key() == key {
			return claim, true
		}
	}
	return ClaimDefinition{}, false
}

// HasClaim returns true if a claim with the given key exists.
func (s *Schema) HasClaim(key string) bool {
	_, found := s.GetClaim(key)
	return found
}

// RequiredClaims returns all required claims.
func (s *Schema) RequiredClaims() []ClaimDefinition {
	result := make([]ClaimDefinition, 0)
	for _, claim := range s.claims {
		if claim.IsRequired() {
			result = append(result, claim)
		}
	}
	return result
}

// OptionalClaims returns all optional claims.
func (s *Schema) OptionalClaims() []ClaimDefinition {
	result := make([]ClaimDefinition, 0)
	for _, claim := range s.claims {
		if claim.IsOptional() {
			result = append(result, claim)
		}
	}
	return result
}

// ClaimCount returns the number of claims.
func (s *Schema) ClaimCount() int {
	return len(s.claims)
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update aggregate state.
func (s *Schema) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *SchemaRegisteredEvent:
		s.applySchemaRegistered(e)
	case SchemaRegisteredEvent:
		s.applySchemaRegistered(&e)
	case *SchemaVersionAddedEvent:
		s.applySchemaVersionAdded(e)
	case SchemaVersionAddedEvent:
		s.applySchemaVersionAdded(&e)
	case *SchemaDeprecatedEvent:
		s.applySchemaDeprecated(e)
	case SchemaDeprecatedEvent:
		s.applySchemaDeprecated(&e)
	case *SchemaActivatedEvent:
		s.applySchemaActivated(e)
	case SchemaActivatedEvent:
		s.applySchemaActivated(&e)
	case *SchemaMetadataUpdatedEvent:
		s.applySchemaMetadataUpdated(e)
	case SchemaMetadataUpdatedEvent:
		s.applySchemaMetadataUpdated(&e)
	}
}

func (s *Schema) applySchemaRegistered(e *SchemaRegisteredEvent) {
	var err error
	s.id, err = types.ParseID(e.SchemaID)
	if err != nil {
		panic("corrupt event store: SchemaRegistered has invalid SchemaID: " + e.SchemaID)
	}
	s.schemaType = e.SchemaType
	s.name = e.Name
	s.description = e.Description
	s.currentVersion, err = ParseSchemaVersion(e.Version)
	if err != nil {
		panic("corrupt event store: SchemaRegistered has invalid Version: " + e.Version)
	}
	s.status = SchemaStatusActive
	s.createdAt = e.RegisteredAt
	s.updatedAt = e.RegisteredAt

	if e.IssuerID != "" {
		issuerID, err := types.ParseID(e.IssuerID)
		if err != nil {
			panic("corrupt event store: SchemaRegistered has invalid IssuerID: " + e.IssuerID)
		}
		s.issuerID = &issuerID
	}

	s.claims = make([]ClaimDefinition, 0, len(e.Claims))
	for _, claimData := range e.Claims {
		claim, err := claimData.ToClaimDefinition()
		if err != nil {
			panic("corrupt event store: SchemaRegistered has invalid ClaimDefinition: " + err.Error())
		}
		s.claims = append(s.claims, claim)
	}
}

func (s *Schema) applySchemaVersionAdded(e *SchemaVersionAddedEvent) {
	var err error
	s.currentVersion, err = ParseSchemaVersion(e.Version)
	if err != nil {
		panic("corrupt event store: SchemaVersionAdded has invalid Version: " + e.Version)
	}
	s.updatedAt = e.AddedAt

	s.claims = make([]ClaimDefinition, 0, len(e.Claims))
	for _, claimData := range e.Claims {
		claim, err := claimData.ToClaimDefinition()
		if err != nil {
			panic("corrupt event store: SchemaVersionAdded has invalid ClaimDefinition: " + err.Error())
		}
		s.claims = append(s.claims, claim)
	}
}

func (s *Schema) applySchemaDeprecated(e *SchemaDeprecatedEvent) {
	s.status = SchemaStatusDeprecated
	s.updatedAt = e.DeprecatedAt
}

func (s *Schema) applySchemaActivated(e *SchemaActivatedEvent) {
	s.status = SchemaStatusActive
	s.updatedAt = e.ActivatedAt
}

func (s *Schema) applySchemaMetadataUpdated(e *SchemaMetadataUpdatedEvent) {
	if e.Name != "" {
		s.name = e.Name
	}
	if e.Description != "" {
		s.description = e.Description
	}
	s.updatedAt = e.UpdatedAt
}

// ============================================================================
// Aggregate Interface Implementation
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (s *Schema) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &s.AggregateRoot
}

// ============================================================================
// Factory
// ============================================================================

// SchemaFactory creates Schema aggregates for the repository.
type SchemaFactory struct{}

// NewSchemaFactory creates a new SchemaFactory.
func NewSchemaFactory() *SchemaFactory {
	return &SchemaFactory{}
}

// Create creates a new Schema aggregate with the given ID.
func (f *SchemaFactory) Create(aggregateID string) eventsourcing.Aggregate {
	id, _ := types.ParseID(aggregateID)
	return NewSchema(id)
}

// ============================================================================
// Helpers
// ============================================================================

// validateUniqueClaims ensures all claims have unique keys.
func validateUniqueClaims(claims []ClaimDefinition) error {
	seen := make(map[string]struct{})
	for _, claim := range claims {
		if _, exists := seen[claim.Key()]; exists {
			return ClaimAlreadyExists("validateUniqueClaims", "", claim.Key())
		}
		seen[claim.Key()] = struct{}{}
	}
	return nil
}
