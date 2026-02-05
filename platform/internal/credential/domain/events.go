package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeCredential = "Credential"
)

// Event type constants
const (
	EventTypeCredentialIssued  = "Credential.Issued"
	EventTypeCredentialRevoked = "Credential.Revoked"
	EventTypeCredentialExpired = "Credential.Expired"
)

// ============================================================================
// Credential Events
// ============================================================================

// CredentialIssuedEvent is emitted when a new credential is issued.
type CredentialIssuedEvent struct {
	eventsourcing.BaseEvent

	CredentialID   string         `json:"credential_id"`
	CredentialType string         `json:"credential_type"`
	IssuerDID      string         `json:"issuer_did"`
	SubjectDID     string         `json:"subject_did"`
	Claims         map[string]any `json:"claims"`
	IssuedAt       time.Time      `json:"issued_at"`
	ExpiresAt      time.Time      `json:"expires_at"`
	VerificationID string         `json:"verification_id,omitempty"`
	JWT            string         `json:"jwt,omitempty"`
}

// EventType returns the event type.
func (e CredentialIssuedEvent) EventType() string {
	return EventTypeCredentialIssued
}

// CredentialRevokedEvent is emitted when a credential is revoked.
type CredentialRevokedEvent struct {
	eventsourcing.BaseEvent

	CredentialID string    `json:"credential_id"`
	Reason       string    `json:"reason"`
	RevokedAt    time.Time `json:"revoked_at"`
}

// EventType returns the event type.
func (e CredentialRevokedEvent) EventType() string {
	return EventTypeCredentialRevoked
}

// CredentialExpiredEvent is emitted when a credential expires.
type CredentialExpiredEvent struct {
	eventsourcing.BaseEvent

	CredentialID string    `json:"credential_id"`
	ExpiredAt    time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e CredentialExpiredEvent) EventType() string {
	return EventTypeCredentialExpired
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterCredentialEvents registers all Credential domain events with the event registry.
func RegisterCredentialEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeCredentialIssued, func() eventsourcing.Event { return &CredentialIssuedEvent{} })
	registry.Register(EventTypeCredentialRevoked, func() eventsourcing.Event { return &CredentialRevokedEvent{} })
	registry.Register(EventTypeCredentialExpired, func() eventsourcing.Event { return &CredentialExpiredEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterCredentialEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newCredentialBaseEvent creates a base event for credential aggregate.
func newCredentialBaseEvent(id CredentialID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeCredential, id.String())
}
