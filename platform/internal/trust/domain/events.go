package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeVouch = "Vouch"
)

// Event type constants
const (
	EventTypeVouchGiven    = "Vouch.Given"
	EventTypeVouchAccepted = "Vouch.Accepted"
	EventTypeVouchRevoked  = "Vouch.Revoked"
	EventTypeVouchExpired  = "Vouch.Expired"
)

// ============================================================================
// Vouch Events
// ============================================================================

// VouchGivenEvent is emitted when a new vouch is given.
type VouchGivenEvent struct {
	eventsourcing.BaseEvent

	VouchID      string    `json:"vouch_id"`
	VoucherID    string    `json:"voucher_id"`
	VoucheeID    string    `json:"vouchee_id"`
	CredentialID string    `json:"credential_id,omitempty"`
	ClaimKey     string    `json:"claim_key,omitempty"`
	Relationship string    `json:"relationship"`
	Strength     int       `json:"strength"`
	Statement    string    `json:"statement"`
	Context      string    `json:"context"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// EventType returns the event type.
func (e VouchGivenEvent) EventType() string {
	return EventTypeVouchGiven
}

// VouchAcceptedEvent is emitted when a vouch is accepted by the vouchee.
type VouchAcceptedEvent struct {
	eventsourcing.BaseEvent

	VouchID    string    `json:"vouch_id"`
	AcceptedAt time.Time `json:"accepted_at"`
}

// EventType returns the event type.
func (e VouchAcceptedEvent) EventType() string {
	return EventTypeVouchAccepted
}

// VouchRevokedEvent is emitted when a vouch is revoked.
type VouchRevokedEvent struct {
	eventsourcing.BaseEvent

	VouchID   string    `json:"vouch_id"`
	Reason    string    `json:"reason"`
	RevokedAt time.Time `json:"revoked_at"`
}

// EventType returns the event type.
func (e VouchRevokedEvent) EventType() string {
	return EventTypeVouchRevoked
}

// VouchExpiredEvent is emitted when a vouch expires.
type VouchExpiredEvent struct {
	eventsourcing.BaseEvent

	VouchID   string    `json:"vouch_id"`
	ExpiredAt time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e VouchExpiredEvent) EventType() string {
	return EventTypeVouchExpired
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterVouchEvents registers all Vouch domain events with the event registry.
func RegisterVouchEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeVouchGiven, func() eventsourcing.Event { return &VouchGivenEvent{} })
	registry.Register(EventTypeVouchAccepted, func() eventsourcing.Event { return &VouchAcceptedEvent{} })
	registry.Register(EventTypeVouchRevoked, func() eventsourcing.Event { return &VouchRevokedEvent{} })
	registry.Register(EventTypeVouchExpired, func() eventsourcing.Event { return &VouchExpiredEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterVouchEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newVouchBaseEvent creates a base event for vouch aggregate.
func newVouchBaseEvent(id VouchID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeVouch, id.String())
}
