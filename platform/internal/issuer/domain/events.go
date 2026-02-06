package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeIssuer   = "Issuer"
	AggregateTypeTemplate = "Template"
)

// Event type constants
const (
	EventTypeIssuerRegistered = "Issuer.Registered"
	EventTypeIssuerActivated  = "Issuer.Activated"
	EventTypeIssuerSuspended  = "Issuer.Suspended"
	EventTypeBrandingUpdated  = "Issuer.BrandingUpdated"
	EventTypeTemplateCreated  = "Template.Created"
	EventTypeTemplateUpdated  = "Template.Updated"
	EventTypeTemplateArchived = "Template.Archived"
)

// ============================================================================
// Issuer Events
// ============================================================================

// IssuerRegisteredEvent is emitted when a new issuer is registered.
type IssuerRegisteredEvent struct {
	eventsourcing.BaseEvent

	IssuerID       string    `json:"issuer_id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	WebhookURL     string    `json:"webhook_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// EventType returns the event type.
func (e IssuerRegisteredEvent) EventType() string {
	return EventTypeIssuerRegistered
}

// IssuerActivatedEvent is emitted when an issuer is activated.
type IssuerActivatedEvent struct {
	eventsourcing.BaseEvent

	IssuerID    string         `json:"issuer_id"`
	DID         string         `json:"did"`
	APIKeyHash  string         `json:"api_key_hash"`
	Branding    map[string]any `json:"branding,omitempty"`
	ActivatedAt time.Time      `json:"activated_at"`
}

// EventType returns the event type.
func (e IssuerActivatedEvent) EventType() string {
	return EventTypeIssuerActivated
}

// IssuerSuspendedEvent is emitted when an issuer is suspended.
type IssuerSuspendedEvent struct {
	eventsourcing.BaseEvent

	IssuerID    string    `json:"issuer_id"`
	Reason      string    `json:"reason"`
	SuspendedAt time.Time `json:"suspended_at"`
}

// EventType returns the event type.
func (e IssuerSuspendedEvent) EventType() string {
	return EventTypeIssuerSuspended
}

// BrandingUpdatedEvent is emitted when an issuer's branding is updated.
type BrandingUpdatedEvent struct {
	eventsourcing.BaseEvent

	IssuerID  string         `json:"issuer_id"`
	Branding  map[string]any `json:"branding"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// EventType returns the event type.
func (e BrandingUpdatedEvent) EventType() string {
	return EventTypeBrandingUpdated
}

// ============================================================================
// Template Events
// ============================================================================

// TemplateCreatedEvent is emitted when a new template is created.
type TemplateCreatedEvent struct {
	eventsourcing.BaseEvent

	TemplateID     string            `json:"template_id"`
	IssuerID       string            `json:"issuer_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	SchemaType     string            `json:"schema_type"`
	ClaimMappings  map[string]string `json:"claim_mappings,omitempty"`
	DefaultValues  map[string]any    `json:"default_values,omitempty"`
	ExpirationDays int               `json:"expiration_days"`
	AutoApprove    bool              `json:"auto_approve"`
	CreatedAt      time.Time         `json:"created_at"`
}

// EventType returns the event type.
func (e TemplateCreatedEvent) EventType() string {
	return EventTypeTemplateCreated
}

// TemplateUpdatedEvent is emitted when a template is updated.
type TemplateUpdatedEvent struct {
	eventsourcing.BaseEvent

	TemplateID     string            `json:"template_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	ClaimMappings  map[string]string `json:"claim_mappings,omitempty"`
	DefaultValues  map[string]any    `json:"default_values,omitempty"`
	ExpirationDays int               `json:"expiration_days"`
	AutoApprove    bool              `json:"auto_approve"`
	Version        int               `json:"version"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// EventType returns the event type.
func (e TemplateUpdatedEvent) EventType() string {
	return EventTypeTemplateUpdated
}

// TemplateArchivedEvent is emitted when a template is archived.
type TemplateArchivedEvent struct {
	eventsourcing.BaseEvent

	TemplateID string    `json:"template_id"`
	ArchivedAt time.Time `json:"archived_at"`
}

// EventType returns the event type.
func (e TemplateArchivedEvent) EventType() string {
	return EventTypeTemplateArchived
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterIssuerEvents registers all Issuer domain events with the event registry.
func RegisterIssuerEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeIssuerRegistered, func() eventsourcing.Event { return &IssuerRegisteredEvent{} })
	registry.Register(EventTypeIssuerActivated, func() eventsourcing.Event { return &IssuerActivatedEvent{} })
	registry.Register(EventTypeIssuerSuspended, func() eventsourcing.Event { return &IssuerSuspendedEvent{} })
	registry.Register(EventTypeBrandingUpdated, func() eventsourcing.Event { return &BrandingUpdatedEvent{} })
	registry.Register(EventTypeTemplateCreated, func() eventsourcing.Event { return &TemplateCreatedEvent{} })
	registry.Register(EventTypeTemplateUpdated, func() eventsourcing.Event { return &TemplateUpdatedEvent{} })
	registry.Register(EventTypeTemplateArchived, func() eventsourcing.Event { return &TemplateArchivedEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterIssuerEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newIssuerBaseEvent creates a base event for issuer aggregate.
func newIssuerBaseEvent(id IssuerID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeIssuer, id.String())
}

// newTemplateBaseEvent creates a base event for template aggregate.
func newTemplateBaseEvent(id TemplateID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeTemplate, id.String())
}
