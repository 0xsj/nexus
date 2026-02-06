package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypePresentation = "Presentation"
	AggregateTypeShareLink    = "ShareLink"
)

// Event type constants
const (
	EventTypePresentationCreated = "Presentation.Created"
	EventTypePresentationRevoked = "Presentation.Revoked"
	EventTypeShareLinkCreated    = "ShareLink.Created"
	EventTypeShareLinkAccessed   = "ShareLink.Accessed"
	EventTypeShareLinkRevoked    = "ShareLink.Revoked"
	EventTypeShareLinkExpired    = "ShareLink.Expired"
)

// ============================================================================
// Presentation Events
// ============================================================================

// PresentationCreatedEvent is emitted when a new presentation is created.
type PresentationCreatedEvent struct {
	eventsourcing.BaseEvent

	PresentationID   string         `json:"presentation_id"`
	HolderDID        string         `json:"holder_did"`
	CredentialIDs    []string       `json:"credential_ids"`
	DisclosurePolicy map[string]any `json:"disclosure_policy"`
	VPJWT            string         `json:"vp_jwt,omitempty"`
	Purpose          string         `json:"purpose,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

// EventType returns the event type.
func (e PresentationCreatedEvent) EventType() string {
	return EventTypePresentationCreated
}

// PresentationRevokedEvent is emitted when a presentation is revoked.
type PresentationRevokedEvent struct {
	eventsourcing.BaseEvent

	PresentationID string    `json:"presentation_id"`
	Reason         string    `json:"reason"`
	RevokedAt      time.Time `json:"revoked_at"`
}

// EventType returns the event type.
func (e PresentationRevokedEvent) EventType() string {
	return EventTypePresentationRevoked
}

// ============================================================================
// ShareLink Events
// ============================================================================

// ShareLinkCreatedEvent is emitted when a new share link is created.
type ShareLinkCreatedEvent struct {
	eventsourcing.BaseEvent

	ShareLinkID    string    `json:"share_link_id"`
	PresentationID string    `json:"presentation_id"`
	Token          string    `json:"token"`
	ExpiresAt      time.Time `json:"expires_at,omitempty"`
	MaxViews       int       `json:"max_views"`
	PinHash        string    `json:"pin_hash,omitempty"`
	Audience       string    `json:"audience,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// EventType returns the event type.
func (e ShareLinkCreatedEvent) EventType() string {
	return EventTypeShareLinkCreated
}

// ShareLinkAccessedEvent is emitted when a share link is accessed.
type ShareLinkAccessedEvent struct {
	eventsourcing.BaseEvent

	ShareLinkID string    `json:"share_link_id"`
	VerifierDID string    `json:"verifier_did,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	AccessedAt  time.Time `json:"accessed_at"`
}

// EventType returns the event type.
func (e ShareLinkAccessedEvent) EventType() string {
	return EventTypeShareLinkAccessed
}

// ShareLinkRevokedEvent is emitted when a share link is revoked.
type ShareLinkRevokedEvent struct {
	eventsourcing.BaseEvent

	ShareLinkID string    `json:"share_link_id"`
	RevokedAt   time.Time `json:"revoked_at"`
}

// EventType returns the event type.
func (e ShareLinkRevokedEvent) EventType() string {
	return EventTypeShareLinkRevoked
}

// ShareLinkExpiredEvent is emitted when a share link expires.
type ShareLinkExpiredEvent struct {
	eventsourcing.BaseEvent

	ShareLinkID string    `json:"share_link_id"`
	ExpiredAt   time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e ShareLinkExpiredEvent) EventType() string {
	return EventTypeShareLinkExpired
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterPresentationEvents registers all Presentation domain events with the event registry.
func RegisterPresentationEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypePresentationCreated, func() eventsourcing.Event { return &PresentationCreatedEvent{} })
	registry.Register(EventTypePresentationRevoked, func() eventsourcing.Event { return &PresentationRevokedEvent{} })
	registry.Register(EventTypeShareLinkCreated, func() eventsourcing.Event { return &ShareLinkCreatedEvent{} })
	registry.Register(EventTypeShareLinkAccessed, func() eventsourcing.Event { return &ShareLinkAccessedEvent{} })
	registry.Register(EventTypeShareLinkRevoked, func() eventsourcing.Event { return &ShareLinkRevokedEvent{} })
	registry.Register(EventTypeShareLinkExpired, func() eventsourcing.Event { return &ShareLinkExpiredEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterPresentationEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newPresentationBaseEvent creates a base event for presentation aggregate.
func newPresentationBaseEvent(id PresentationID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypePresentation, id.String())
}

// newShareLinkBaseEvent creates a base event for share link aggregate.
func newShareLinkBaseEvent(id ShareLinkID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeShareLink, id.String())
}
