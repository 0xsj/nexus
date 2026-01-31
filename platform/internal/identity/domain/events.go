package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeUser    = "User"
	AggregateTypeSession = "Session"
)

// Event type constants
const (
	// User events
	EventTypeUserRegistered         = "User.Registered"
	EventTypeUserActivated          = "User.Activated"
	EventTypeUserSuspended          = "User.Suspended"
	EventTypeUserReactivated        = "User.Reactivated"
	EventTypeUserDeleted            = "User.Deleted"
	EventTypeUserDisplayNameChanged = "User.DisplayNameChanged"
	EventTypeUserEmailChanged       = "User.EmailChanged"
	EventTypeUserDIDAdded           = "User.DIDAdded"
	EventTypeUserDIDRemoved         = "User.DIDRemoved"
	EventTypeUserOAuthLinked        = "User.OAuthLinked"
	EventTypeUserOAuthUnlinked      = "User.OAuthUnlinked"

	// Session events
	EventTypeSessionStarted   = "Session.Started"
	EventTypeSessionRefreshed = "Session.Refreshed"
	EventTypeSessionRevoked   = "Session.Revoked"
	EventTypeSessionExpired   = "Session.Expired"
)

// ============================================================================
// User Events
// ============================================================================

// UserRegisteredEvent is emitted when a new user registers.
type UserRegisteredEvent struct {
	eventsourcing.BaseEvent

	UserID       string    `json:"user_id"`
	Email        string    `json:"email,omitempty"`
	DisplayName  string    `json:"display_name,omitempty"`
	AuthMethod   string    `json:"auth_method"`
	PrimaryDID   string    `json:"primary_did"`
	RegisteredAt time.Time `json:"registered_at"`
}

// EventType returns the event type.
func (e UserRegisteredEvent) EventType() string {
	return EventTypeUserRegistered
}

// UserActivatedEvent is emitted when a user account is activated.
type UserActivatedEvent struct {
	eventsourcing.BaseEvent

	UserID      string    `json:"user_id"`
	ActivatedAt time.Time `json:"activated_at"`
}

// EventType returns the event type.
func (e UserActivatedEvent) EventType() string {
	return EventTypeUserActivated
}

// UserSuspendedEvent is emitted when a user account is suspended.
type UserSuspendedEvent struct {
	eventsourcing.BaseEvent

	UserID      string    `json:"user_id"`
	Reason      string    `json:"reason,omitempty"`
	SuspendedAt time.Time `json:"suspended_at"`
}

// EventType returns the event type.
func (e UserSuspendedEvent) EventType() string {
	return EventTypeUserSuspended
}

// UserReactivatedEvent is emitted when a suspended user account is reactivated.
type UserReactivatedEvent struct {
	eventsourcing.BaseEvent

	UserID        string    `json:"user_id"`
	ReactivatedAt time.Time `json:"reactivated_at"`
}

// EventType returns the event type.
func (e UserReactivatedEvent) EventType() string {
	return EventTypeUserReactivated
}

// UserDeletedEvent is emitted when a user account is soft-deleted.
type UserDeletedEvent struct {
	eventsourcing.BaseEvent

	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// EventType returns the event type.
func (e UserDeletedEvent) EventType() string {
	return EventTypeUserDeleted
}

// UserDisplayNameChangedEvent is emitted when a user changes their display name.
type UserDisplayNameChangedEvent struct {
	eventsourcing.BaseEvent

	UserID         string    `json:"user_id"`
	OldDisplayName string    `json:"old_display_name,omitempty"`
	NewDisplayName string    `json:"new_display_name"`
	ChangedAt      time.Time `json:"changed_at"`
}

// EventType returns the event type.
func (e UserDisplayNameChangedEvent) EventType() string {
	return EventTypeUserDisplayNameChanged
}

// UserEmailChangedEvent is emitted when a user changes their email address.
type UserEmailChangedEvent struct {
	eventsourcing.BaseEvent

	UserID    string    `json:"user_id"`
	OldEmail  string    `json:"old_email,omitempty"`
	NewEmail  string    `json:"new_email"`
	ChangedAt time.Time `json:"changed_at"`
}

// EventType returns the event type.
func (e UserEmailChangedEvent) EventType() string {
	return EventTypeUserEmailChanged
}

// UserDIDAddedEvent is emitted when a DID is added to a user's identity.
type UserDIDAddedEvent struct {
	eventsourcing.BaseEvent

	UserID  string    `json:"user_id"`
	DID     string    `json:"did"`
	AddedAt time.Time `json:"added_at"`
}

// EventType returns the event type.
func (e UserDIDAddedEvent) EventType() string {
	return EventTypeUserDIDAdded
}

// UserDIDRemovedEvent is emitted when a DID is removed from a user's identity.
type UserDIDRemovedEvent struct {
	eventsourcing.BaseEvent

	UserID    string    `json:"user_id"`
	DID       string    `json:"did"`
	RemovedAt time.Time `json:"removed_at"`
}

// EventType returns the event type.
func (e UserDIDRemovedEvent) EventType() string {
	return EventTypeUserDIDRemoved
}

// UserOAuthLinkedEvent is emitted when an OAuth account is linked to a user.
type UserOAuthLinkedEvent struct {
	eventsourcing.BaseEvent

	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	ExternalID string    `json:"external_id"`
	Email      string    `json:"email,omitempty"`
	LinkedAt   time.Time `json:"linked_at"`
}

// EventType returns the event type.
func (e UserOAuthLinkedEvent) EventType() string {
	return EventTypeUserOAuthLinked
}

// UserOAuthUnlinkedEvent is emitted when an OAuth account is unlinked from a user.
type UserOAuthUnlinkedEvent struct {
	eventsourcing.BaseEvent

	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	ExternalID string    `json:"external_id"`
	UnlinkedAt time.Time `json:"unlinked_at"`
}

// EventType returns the event type.
func (e UserOAuthUnlinkedEvent) EventType() string {
	return EventTypeUserOAuthUnlinked
}

// ============================================================================
// Session Events
// ============================================================================

// SessionStartedEvent is emitted when a new session is created.
type SessionStartedEvent struct {
	eventsourcing.BaseEvent

	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	AuthMethod string    `json:"auth_method"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	StartedAt  time.Time `json:"started_at"`
}

// EventType returns the event type.
func (e SessionStartedEvent) EventType() string {
	return EventTypeSessionStarted
}

// SessionRefreshedEvent is emitted when a session is refreshed (token rotated).
type SessionRefreshedEvent struct {
	eventsourcing.BaseEvent

	SessionID    string    `json:"session_id"`
	OldTokenHash string    `json:"old_token_hash"`
	NewTokenHash string    `json:"new_token_hash"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshedAt  time.Time `json:"refreshed_at"`
}

// EventType returns the event type.
func (e SessionRefreshedEvent) EventType() string {
	return EventTypeSessionRefreshed
}

// SessionRevokedEvent is emitted when a session is explicitly revoked.
type SessionRevokedEvent struct {
	eventsourcing.BaseEvent

	SessionID string    `json:"session_id"`
	Reason    string    `json:"reason,omitempty"`
	RevokedAt time.Time `json:"revoked_at"`
}

// EventType returns the event type.
func (e SessionRevokedEvent) EventType() string {
	return EventTypeSessionRevoked
}

// SessionExpiredEvent is emitted when a session expires.
type SessionExpiredEvent struct {
	eventsourcing.BaseEvent

	SessionID string    `json:"session_id"`
	ExpiredAt time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e SessionExpiredEvent) EventType() string {
	return EventTypeSessionExpired
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterEvents registers all Identity domain events with the event registry.
func RegisterEvents(registry *eventsourcing.EventRegistry) {
	// User events
	registry.Register(EventTypeUserRegistered, func() eventsourcing.Event { return &UserRegisteredEvent{} })
	registry.Register(EventTypeUserActivated, func() eventsourcing.Event { return &UserActivatedEvent{} })
	registry.Register(EventTypeUserSuspended, func() eventsourcing.Event { return &UserSuspendedEvent{} })
	registry.Register(EventTypeUserReactivated, func() eventsourcing.Event { return &UserReactivatedEvent{} })
	registry.Register(EventTypeUserDeleted, func() eventsourcing.Event { return &UserDeletedEvent{} })
	registry.Register(EventTypeUserDisplayNameChanged, func() eventsourcing.Event { return &UserDisplayNameChangedEvent{} })
	registry.Register(EventTypeUserEmailChanged, func() eventsourcing.Event { return &UserEmailChangedEvent{} })
	registry.Register(EventTypeUserDIDAdded, func() eventsourcing.Event { return &UserDIDAddedEvent{} })
	registry.Register(EventTypeUserDIDRemoved, func() eventsourcing.Event { return &UserDIDRemovedEvent{} })
	registry.Register(EventTypeUserOAuthLinked, func() eventsourcing.Event { return &UserOAuthLinkedEvent{} })
	registry.Register(EventTypeUserOAuthUnlinked, func() eventsourcing.Event { return &UserOAuthUnlinkedEvent{} })

	// Session events
	registry.Register(EventTypeSessionStarted, func() eventsourcing.Event { return &SessionStartedEvent{} })
	registry.Register(EventTypeSessionRefreshed, func() eventsourcing.Event { return &SessionRefreshedEvent{} })
	registry.Register(EventTypeSessionRevoked, func() eventsourcing.Event { return &SessionRevokedEvent{} })
	registry.Register(EventTypeSessionExpired, func() eventsourcing.Event { return &SessionExpiredEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newUserBaseEvent creates a base event for user aggregate.
func newUserBaseEvent(userID UserID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeUser, userID.String())
}

// newSessionBaseEvent creates a base event for session aggregate.
func newSessionBaseEvent(sessionID SessionID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeSession, sessionID.String())
}
