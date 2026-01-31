package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Event Type Constants
// ============================================================================

const (
	// AggregateTypeSchema is the aggregate type for Schema.
	AggregateTypeSchema = "Schema"

	// Event type names
	EventTypeSchemaRegistered      = "Schema.Registered"
	EventTypeSchemaVersionAdded    = "Schema.VersionAdded"
	EventTypeSchemaDeprecated      = "Schema.Deprecated"
	EventTypeSchemaActivated       = "Schema.Activated"
	EventTypeSchemaMetadataUpdated = "Schema.MetadataUpdated"
)

// ============================================================================
// SchemaRegisteredEvent
// ============================================================================

// SchemaRegisteredEvent is raised when a new schema is created.
type SchemaRegisteredEvent struct {
	eventsourcing.BaseEvent

	// SchemaID is the unique identifier of the schema.
	SchemaID string `json:"schema_id"`

	// SchemaType is the type name (e.g., "GitHubContributor").
	SchemaType string `json:"schema_type"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Description is the schema description.
	Description string `json:"description"`

	// Version is the initial version.
	Version string `json:"version"`

	// Claims contains the initial claim definitions.
	Claims []ClaimDefinitionData `json:"claims"`

	// IssuerID is the optional issuer who owns this schema (for custom schemas).
	IssuerID string `json:"issuer_id,omitempty"`

	// RegisteredAt is when the schema was registered.
	RegisteredAt time.Time `json:"registered_at"`
}

// EventType returns the event type name.
func (e SchemaRegisteredEvent) EventType() string {
	return EventTypeSchemaRegistered
}

// NewSchemaRegisteredEvent creates a new SchemaRegisteredEvent.
func NewSchemaRegisteredEvent(
	schemaID string,
	schemaType string,
	name string,
	description string,
	version SchemaVersion,
	claims []ClaimDefinition,
	issuerID string,
) SchemaRegisteredEvent {
	claimData := make([]ClaimDefinitionData, len(claims))
	for i, c := range claims {
		claimData[i] = ClaimDefinitionToData(c)
	}

	return SchemaRegisteredEvent{
		BaseEvent:    eventsourcing.NewBaseEvent(AggregateTypeSchema, schemaID),
		SchemaID:     schemaID,
		SchemaType:   schemaType,
		Name:         name,
		Description:  description,
		Version:      version.String(),
		Claims:       claimData,
		IssuerID:     issuerID,
		RegisteredAt: time.Now().UTC(),
	}
}

// ============================================================================
// SchemaVersionAddedEvent
// ============================================================================

// SchemaVersionAddedEvent is raised when a new version is added to an existing schema.
type SchemaVersionAddedEvent struct {
	eventsourcing.BaseEvent

	// SchemaID is the unique identifier of the schema.
	SchemaID string `json:"schema_id"`

	// Version is the new version.
	Version string `json:"version"`

	// PreviousVersion is the version this was based on.
	PreviousVersion string `json:"previous_version"`

	// Claims contains the claim definitions for this version.
	Claims []ClaimDefinitionData `json:"claims"`

	// ChangeSummary describes what changed in this version.
	ChangeSummary string `json:"change_summary,omitempty"`

	// AddedAt is when the version was added.
	AddedAt time.Time `json:"added_at"`
}

// EventType returns the event type name.
func (e SchemaVersionAddedEvent) EventType() string {
	return EventTypeSchemaVersionAdded
}

// NewSchemaVersionAddedEvent creates a new SchemaVersionAddedEvent.
func NewSchemaVersionAddedEvent(
	schemaID string,
	version SchemaVersion,
	previousVersion SchemaVersion,
	claims []ClaimDefinition,
	changeSummary string,
) SchemaVersionAddedEvent {
	claimData := make([]ClaimDefinitionData, len(claims))
	for i, c := range claims {
		claimData[i] = ClaimDefinitionToData(c)
	}

	return SchemaVersionAddedEvent{
		BaseEvent:       eventsourcing.NewBaseEvent(AggregateTypeSchema, schemaID),
		SchemaID:        schemaID,
		Version:         version.String(),
		PreviousVersion: previousVersion.String(),
		Claims:          claimData,
		ChangeSummary:   changeSummary,
		AddedAt:         time.Now().UTC(),
	}
}

// ============================================================================
// SchemaDeprecatedEvent
// ============================================================================

// SchemaDeprecatedEvent is raised when a schema is marked as deprecated.
type SchemaDeprecatedEvent struct {
	eventsourcing.BaseEvent

	// SchemaID is the unique identifier of the schema.
	SchemaID string `json:"schema_id"`

	// Reason explains why the schema was deprecated.
	Reason string `json:"reason,omitempty"`

	// ReplacementSchemaID is the ID of the schema that replaces this one.
	ReplacementSchemaID string `json:"replacement_schema_id,omitempty"`

	// DeprecatedAt is when the schema was deprecated.
	DeprecatedAt time.Time `json:"deprecated_at"`
}

// EventType returns the event type name.
func (e SchemaDeprecatedEvent) EventType() string {
	return EventTypeSchemaDeprecated
}

// NewSchemaDeprecatedEvent creates a new SchemaDeprecatedEvent.
func NewSchemaDeprecatedEvent(
	schemaID string,
	reason string,
	replacementSchemaID string,
) SchemaDeprecatedEvent {
	return SchemaDeprecatedEvent{
		BaseEvent:           eventsourcing.NewBaseEvent(AggregateTypeSchema, schemaID),
		SchemaID:            schemaID,
		Reason:              reason,
		ReplacementSchemaID: replacementSchemaID,
		DeprecatedAt:        time.Now().UTC(),
	}
}

// ============================================================================
// SchemaActivatedEvent
// ============================================================================

// SchemaActivatedEvent is raised when a schema is (re)activated.
type SchemaActivatedEvent struct {
	eventsourcing.BaseEvent

	// SchemaID is the unique identifier of the schema.
	SchemaID string `json:"schema_id"`

	// ActivatedAt is when the schema was activated.
	ActivatedAt time.Time `json:"activated_at"`
}

// EventType returns the event type name.
func (e SchemaActivatedEvent) EventType() string {
	return EventTypeSchemaActivated
}

// NewSchemaActivatedEvent creates a new SchemaActivatedEvent.
func NewSchemaActivatedEvent(schemaID string) SchemaActivatedEvent {
	return SchemaActivatedEvent{
		BaseEvent:   eventsourcing.NewBaseEvent(AggregateTypeSchema, schemaID),
		SchemaID:    schemaID,
		ActivatedAt: time.Now().UTC(),
	}
}

// ============================================================================
// SchemaMetadataUpdatedEvent
// ============================================================================

// SchemaMetadataUpdatedEvent is raised when schema metadata (name/description) is changed.
type SchemaMetadataUpdatedEvent struct {
	eventsourcing.BaseEvent

	// SchemaID is the unique identifier of the schema.
	SchemaID string `json:"schema_id"`

	// Name is the updated name (empty if unchanged).
	Name string `json:"name,omitempty"`

	// Description is the updated description (empty if unchanged).
	Description string `json:"description,omitempty"`

	// UpdatedAt is when the metadata was updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// EventType returns the event type name.
func (e SchemaMetadataUpdatedEvent) EventType() string {
	return EventTypeSchemaMetadataUpdated
}

// NewSchemaMetadataUpdatedEvent creates a new SchemaMetadataUpdatedEvent.
func NewSchemaMetadataUpdatedEvent(
	schemaID string,
	name string,
	description string,
) SchemaMetadataUpdatedEvent {
	return SchemaMetadataUpdatedEvent{
		BaseEvent:   eventsourcing.NewBaseEvent(AggregateTypeSchema, schemaID),
		SchemaID:    schemaID,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now().UTC(),
	}
}

// ============================================================================
// Claim Definition Data (for serialization)
// ============================================================================

// ClaimDefinitionData is a serializable representation of ClaimDefinition.
// Used in events to avoid complex nested types.
type ClaimDefinitionData struct {
	Key         string                `json:"key"`
	DataType    string                `json:"data_type"`
	Required    bool                  `json:"required"`
	DisplayName string                `json:"display_name,omitempty"`
	Description string                `json:"description,omitempty"`
	Disclosable bool                  `json:"disclosable"`
	Order       int                   `json:"order"`
	Constraints *ClaimConstraintsData `json:"constraints,omitempty"`
}

// ClaimConstraintsData is a serializable representation of ClaimConstraints.
type ClaimConstraintsData struct {
	AllowedValues []string `json:"allowed_values,omitempty"`
	Min           *int64   `json:"min,omitempty"`
	Max           *int64   `json:"max,omitempty"`
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	Pattern       *string  `json:"pattern,omitempty"`
	Format        *string  `json:"format,omitempty"`
}

// ClaimDefinitionToData converts a ClaimDefinition to ClaimDefinitionData.
func ClaimDefinitionToData(cd ClaimDefinition) ClaimDefinitionData {
	data := ClaimDefinitionData{
		Key:         cd.Key(),
		DataType:    cd.DataType().String(),
		Required:    cd.IsRequired(),
		DisplayName: cd.displayName,
		Description: cd.description,
		Disclosable: cd.IsDisclosable(),
		Order:       cd.Order(),
	}

	// Add constraints if present
	c := cd.ClaimType().Constraints()
	if cd.ClaimType().HasConstraints() {
		data.Constraints = &ClaimConstraintsData{
			AllowedValues: c.AllowedValues,
			Min:           c.Min,
			Max:           c.Max,
			MinLength:     c.MinLength,
			MaxLength:     c.MaxLength,
			Pattern:       c.Pattern,
			Format:        c.Format,
		}
	}

	return data
}

// ToClaimDefinition converts ClaimDefinitionData back to a ClaimDefinition.
func (d ClaimDefinitionData) ToClaimDefinition() (ClaimDefinition, error) {
	dataType, err := ParseDataType(d.DataType)
	if err != nil {
		return ClaimDefinition{}, err
	}

	// Build claim type options from constraints
	var opts []ClaimTypeOption
	if d.Constraints != nil {
		c := d.Constraints
		if len(c.AllowedValues) > 0 {
			opts = append(opts, WithAllowedValues(c.AllowedValues...))
		}
		if c.Min != nil {
			opts = append(opts, WithMin(*c.Min))
		}
		if c.Max != nil {
			opts = append(opts, WithMax(*c.Max))
		}
		if c.MinLength != nil {
			opts = append(opts, WithMinLength(*c.MinLength))
		}
		if c.MaxLength != nil {
			opts = append(opts, WithMaxLength(*c.MaxLength))
		}
		if c.Pattern != nil {
			opts = append(opts, WithPattern(*c.Pattern))
		}
		if c.Format != nil {
			opts = append(opts, WithFormat(*c.Format))
		}
	}

	claimType, err := NewClaimType(dataType, opts...)
	if err != nil {
		return ClaimDefinition{}, err
	}

	// Build claim definition options
	var cdOpts []ClaimDefinitionOption
	if d.Required {
		cdOpts = append(cdOpts, Required())
	}
	if d.DisplayName != "" {
		cdOpts = append(cdOpts, WithDisplayName(d.DisplayName))
	}
	if d.Description != "" {
		cdOpts = append(cdOpts, WithDescription(d.Description))
	}
	if !d.Disclosable {
		cdOpts = append(cdOpts, NonDisclosable())
	}
	if d.Order != 0 {
		cdOpts = append(cdOpts, WithOrder(d.Order))
	}

	return NewClaimDefinition(d.Key, claimType, cdOpts...)
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterSchemaEvents registers all schema events with the given registry.
func RegisterSchemaEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeSchemaRegistered, func() eventsourcing.Event {
		return &SchemaRegisteredEvent{}
	})
	registry.Register(EventTypeSchemaVersionAdded, func() eventsourcing.Event {
		return &SchemaVersionAddedEvent{}
	})
	registry.Register(EventTypeSchemaDeprecated, func() eventsourcing.Event {
		return &SchemaDeprecatedEvent{}
	})
	registry.Register(EventTypeSchemaActivated, func() eventsourcing.Event {
		return &SchemaActivatedEvent{}
	})
	registry.Register(EventTypeSchemaMetadataUpdated, func() eventsourcing.Event {
		return &SchemaMetadataUpdatedEvent{}
	})
}

// init registers schema events with the default registry.
func init() {
	RegisterSchemaEvents(eventsourcing.DefaultRegistry)
}
