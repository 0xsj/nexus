package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Template is the aggregate root for credential templates.
// It defines the structure, claim mappings, and default values
// for credentials that an issuer can issue.
type Template struct {
	eventsourcing.AggregateRoot

	id             TemplateID
	issuerID       IssuerID
	name           string
	description    string
	schemaType     string
	claimMappings  map[string]string
	defaultValues  map[string]any
	expirationDays int
	autoApprove    bool
	status         TemplateStatus
	version        int
	createdAt      time.Time
	updatedAt      time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// CreateTemplate creates a new Template aggregate.
func CreateTemplate(
	id TemplateID,
	issuerID IssuerID,
	name, description, schemaType string,
	claimMappings map[string]string,
	defaultValues map[string]any,
	expirationDays int,
	autoApprove bool,
) (*Template, error) {
	if id.IsZero() {
		return nil, TemplateInvalid("Template.Create", "template ID is required")
	}

	if issuerID.IsZero() {
		return nil, TemplateInvalid("Template.Create", "issuer ID is required")
	}

	if name == "" {
		return nil, TemplateInvalid("Template.Create", "name is required")
	}

	if schemaType == "" {
		return nil, TemplateInvalid("Template.Create", "schema type is required")
	}

	t := &Template{}
	t.InitAggregate(AggregateTypeTemplate, id.String())

	now := time.Now().UTC()

	// Defensive copy of maps
	mappingsCopy := copyStringMap(claimMappings)
	defaultsCopy := copyAnyMap(defaultValues)

	t.Raise(t, &TemplateCreatedEvent{
		BaseEvent:      newTemplateBaseEvent(id),
		TemplateID:     id.String(),
		IssuerID:       issuerID.String(),
		Name:           name,
		Description:    description,
		SchemaType:     schemaType,
		ClaimMappings:  mappingsCopy,
		DefaultValues:  defaultsCopy,
		ExpirationDays: expirationDays,
		AutoApprove:    autoApprove,
		CreatedAt:      now,
	})

	return t, nil
}

// NewTemplateFromEvents reconstructs a Template from events (for hydration).
func NewTemplateFromEvents(id string) *Template {
	t := &Template{}
	t.InitAggregate(AggregateTypeTemplate, id)
	return t
}

// TemplateFactory creates a factory for Template aggregates.
func TemplateFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewTemplateFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the template's ID.
func (t *Template) ID() TemplateID {
	return t.id
}

// IssuerID returns the template's issuer ID.
func (t *Template) IssuerID() IssuerID {
	return t.issuerID
}

// Name returns the template's name.
func (t *Template) Name() string {
	return t.name
}

// Description returns the template's description.
func (t *Template) Description() string {
	return t.description
}

// SchemaType returns the template's schema type.
func (t *Template) SchemaType() string {
	return t.schemaType
}

// ClaimMappings returns a defensive copy of the template's claim mappings.
func (t *Template) ClaimMappings() map[string]string {
	return copyStringMap(t.claimMappings)
}

// DefaultValues returns a defensive copy of the template's default values.
func (t *Template) DefaultValues() map[string]any {
	return copyAnyMap(t.defaultValues)
}

// ExpirationDays returns the template's expiration days.
func (t *Template) ExpirationDays() int {
	return t.expirationDays
}

// AutoApprove returns whether the template auto-approves credentials.
func (t *Template) AutoApprove() bool {
	return t.autoApprove
}

// Status returns the template's status.
func (t *Template) Status() TemplateStatus {
	return t.status
}

// TemplateVersion returns the template's version number.
func (t *Template) TemplateVersion() int {
	return t.version
}

// CreatedAt returns when the template was created.
func (t *Template) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns when the template was last updated.
func (t *Template) UpdatedAt() time.Time {
	return t.updatedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Update updates the template's mutable fields.
func (t *Template) Update(
	name, description string,
	claimMappings map[string]string,
	defaultValues map[string]any,
	expirationDays int,
	autoApprove bool,
) error {
	if !t.status.IsActive() {
		return TemplateArchived("Template.Update", t.id.String()).
			WithMessage("cannot update template in current status: " + t.status.String())
	}

	if name == "" {
		return TemplateInvalid("Template.Update", "name is required")
	}

	now := time.Now().UTC()
	newVersion := t.version + 1

	mappingsCopy := copyStringMap(claimMappings)
	defaultsCopy := copyAnyMap(defaultValues)

	t.Raise(t, &TemplateUpdatedEvent{
		BaseEvent:      newTemplateBaseEvent(t.id),
		TemplateID:     t.id.String(),
		Name:           name,
		Description:    description,
		ClaimMappings:  mappingsCopy,
		DefaultValues:  defaultsCopy,
		ExpirationDays: expirationDays,
		AutoApprove:    autoApprove,
		Version:        newVersion,
		UpdatedAt:      now,
	})

	return nil
}

// Archive archives the template.
func (t *Template) Archive() error {
	if !t.status.CanTransitionTo(TemplateStatusArchived) {
		return TemplateArchived("Template.Archive", t.id.String()).
			WithMessage("cannot archive template in current status: " + t.status.String())
	}

	now := time.Now().UTC()

	t.Raise(t, &TemplateArchivedEvent{
		BaseEvent:  newTemplateBaseEvent(t.id),
		TemplateID: t.id.String(),
		ArchivedAt: now,
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (t *Template) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *TemplateCreatedEvent:
		t.onTemplateCreated(e)
	case *TemplateUpdatedEvent:
		t.onTemplateUpdated(e)
	case *TemplateArchivedEvent:
		t.onTemplateArchived(e)
	}
}

func (t *Template) onTemplateCreated(e *TemplateCreatedEvent) {
	var err error
	t.id, err = ParseTemplateID(e.TemplateID)
	if err != nil {
		panic("corrupt event store: TemplateCreated has invalid TemplateID: " + e.TemplateID)
	}
	t.issuerID, err = ParseIssuerID(e.IssuerID)
	if err != nil {
		panic("corrupt event store: TemplateCreated has invalid IssuerID: " + e.IssuerID)
	}
	t.name = e.Name
	t.description = e.Description
	t.schemaType = e.SchemaType
	t.claimMappings = copyStringMap(e.ClaimMappings)
	t.defaultValues = copyAnyMap(e.DefaultValues)
	t.expirationDays = e.ExpirationDays
	t.autoApprove = e.AutoApprove
	t.status = TemplateStatusActive
	t.version = 1
	t.createdAt = e.CreatedAt
	t.updatedAt = e.CreatedAt
}

func (t *Template) onTemplateUpdated(e *TemplateUpdatedEvent) {
	t.name = e.Name
	t.description = e.Description
	t.claimMappings = copyStringMap(e.ClaimMappings)
	t.defaultValues = copyAnyMap(e.DefaultValues)
	t.expirationDays = e.ExpirationDays
	t.autoApprove = e.AutoApprove
	t.version = e.Version
	t.updatedAt = e.UpdatedAt
}

func (t *Template) onTemplateArchived(e *TemplateArchivedEvent) {
	t.status = TemplateStatusArchived
	t.updatedAt = e.ArchivedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (t *Template) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &t.AggregateRoot
}

// ============================================================================
// Helpers
// ============================================================================

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	copied := make(map[string]string, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

func copyAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	copied := make(map[string]any, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}
