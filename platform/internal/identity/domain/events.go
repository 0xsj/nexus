package domain

import (
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Event Types
// ============================================================================

const (
	EventTypeUserCreated       = "identity.user.created"
	EventTypeUserLoggedIn      = "identity.user.logged_in"
	EventTypeUserSuspended     = "identity.user.suspended"
	EventTypeUserActivated     = "identity.user.activated"
	EventTypeIdentityLinked    = "identity.identity.linked"
	EventTypeEmailVerified     = "identity.email.verified"
	EventTypeWalletLinked      = "identity.wallet.linked"
	EventTypeWalletUnlinked    = "identity.wallet.unlinked"
	EventTypeOAuthLinked       = "identity.oauth.linked"
	EventTypeOAuthUnlinked     = "identity.oauth.unlinked"
	EventTypePrimaryDIDUpdated = "identity.primary_did.updated"
	EventTypeSessionCreated    = "identity.session.created"
	EventTypeSessionRevoked    = "identity.session.revoked"
	EventTypeAPIKeyCreated     = "identity.apikey.created"
	EventTypeAPIKeyRevoked     = "identity.apikey.revoked"
)

// ============================================================================
// User Events
// ============================================================================

// UserCreatedEvent is raised when a new user is created.
type UserCreatedEvent struct {
	eventsourcing.BaseEvent
	PrimaryDID string `json:"primary_did"`
}

// EventType returns the event type.
func (e *UserCreatedEvent) EventType() string { return EventTypeUserCreated }

// NewUserCreatedEvent creates a new UserCreatedEvent.
func NewUserCreatedEvent(userID, primaryDID string) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent:  eventsourcing.NewBaseEvent(UserAggregateType, userID),
		PrimaryDID: primaryDID,
	}
}

// UserLoggedInEvent is raised when a user logs in.
type UserLoggedInEvent struct {
	eventsourcing.BaseEvent
	Method string `json:"method"`
}

// EventType returns the event type.
func (e *UserLoggedInEvent) EventType() string { return EventTypeUserLoggedIn }

// NewUserLoggedInEvent creates a new UserLoggedInEvent.
func NewUserLoggedInEvent(userID, method string) *UserLoggedInEvent {
	return &UserLoggedInEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Method:    method,
	}
}

// UserSuspendedEvent is raised when a user is suspended.
type UserSuspendedEvent struct {
	eventsourcing.BaseEvent
	Reason string `json:"reason"`
}

// EventType returns the event type.
func (e *UserSuspendedEvent) EventType() string { return EventTypeUserSuspended }

// NewUserSuspendedEvent creates a new UserSuspendedEvent.
func NewUserSuspendedEvent(userID, reason string) *UserSuspendedEvent {
	return &UserSuspendedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Reason:    reason,
	}
}

// UserActivatedEvent is raised when a user is activated.
type UserActivatedEvent struct {
	eventsourcing.BaseEvent
}

// EventType returns the event type.
func (e *UserActivatedEvent) EventType() string { return EventTypeUserActivated }

// NewUserActivatedEvent creates a new UserActivatedEvent.
func NewUserActivatedEvent(userID string) *UserActivatedEvent {
	return &UserActivatedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
	}
}

// ============================================================================
// Identity Linking Events
// ============================================================================

// IdentityLinkedEvent is raised when an identity is linked to a user.
type IdentityLinkedEvent struct {
	eventsourcing.BaseEvent
	IdentityType string `json:"identity_type"`
	Value        string `json:"value"`
}

// EventType returns the event type.
func (e *IdentityLinkedEvent) EventType() string { return EventTypeIdentityLinked }

// NewIdentityLinkedEvent creates a new IdentityLinkedEvent.
func NewIdentityLinkedEvent(userID, identityType, value string) *IdentityLinkedEvent {
	return &IdentityLinkedEvent{
		BaseEvent:    eventsourcing.NewBaseEvent(UserAggregateType, userID),
		IdentityType: identityType,
		Value:        value,
	}
}

// EmailVerifiedEvent is raised when an email is verified.
type EmailVerifiedEvent struct {
	eventsourcing.BaseEvent
	Email string `json:"email"`
}

// EventType returns the event type.
func (e *EmailVerifiedEvent) EventType() string { return EventTypeEmailVerified }

// NewEmailVerifiedEvent creates a new EmailVerifiedEvent.
func NewEmailVerifiedEvent(userID, email string) *EmailVerifiedEvent {
	return &EmailVerifiedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Email:     email,
	}
}

// ============================================================================
// Wallet Events
// ============================================================================

// WalletLinkedEvent is raised when a wallet is linked to a user.
type WalletLinkedEvent struct {
	eventsourcing.BaseEvent
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// EventType returns the event type.
func (e *WalletLinkedEvent) EventType() string { return EventTypeWalletLinked }

// NewWalletLinkedEvent creates a new WalletLinkedEvent.
func NewWalletLinkedEvent(userID, address, chain string) *WalletLinkedEvent {
	return &WalletLinkedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Address:   address,
		Chain:     chain,
	}
}

// WalletUnlinkedEvent is raised when a wallet is unlinked from a user.
type WalletUnlinkedEvent struct {
	eventsourcing.BaseEvent
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

// EventType returns the event type.
func (e *WalletUnlinkedEvent) EventType() string { return EventTypeWalletUnlinked }

// NewWalletUnlinkedEvent creates a new WalletUnlinkedEvent.
func NewWalletUnlinkedEvent(userID, address, chain string) *WalletUnlinkedEvent {
	return &WalletUnlinkedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Address:   address,
		Chain:     chain,
	}
}

// ============================================================================
// OAuth Events
// ============================================================================

// OAuthLinkedEvent is raised when an OAuth connection is linked.
type OAuthLinkedEvent struct {
	eventsourcing.BaseEvent
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
}

// EventType returns the event type.
func (e *OAuthLinkedEvent) EventType() string { return EventTypeOAuthLinked }

// NewOAuthLinkedEvent creates a new OAuthLinkedEvent.
func NewOAuthLinkedEvent(userID, provider, providerUserID string) *OAuthLinkedEvent {
	return &OAuthLinkedEvent{
		BaseEvent:      eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Provider:       provider,
		ProviderUserID: providerUserID,
	}
}

// OAuthUnlinkedEvent is raised when an OAuth connection is unlinked.
type OAuthUnlinkedEvent struct {
	eventsourcing.BaseEvent
	Provider string `json:"provider"`
}

// EventType returns the event type.
func (e *OAuthUnlinkedEvent) EventType() string { return EventTypeOAuthUnlinked }

// NewOAuthUnlinkedEvent creates a new OAuthUnlinkedEvent.
func NewOAuthUnlinkedEvent(userID, provider string) *OAuthUnlinkedEvent {
	return &OAuthUnlinkedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		Provider:  provider,
	}
}

// ============================================================================
// DID Events
// ============================================================================

// PrimaryDIDUpdatedEvent is raised when the primary DID is updated.
type PrimaryDIDUpdatedEvent struct {
	eventsourcing.BaseEvent
	OldDID string `json:"old_did"`
	NewDID string `json:"new_did"`
}

// EventType returns the event type.
func (e *PrimaryDIDUpdatedEvent) EventType() string { return EventTypePrimaryDIDUpdated }

// NewPrimaryDIDUpdatedEvent creates a new PrimaryDIDUpdatedEvent.
func NewPrimaryDIDUpdatedEvent(userID, oldDID, newDID string) *PrimaryDIDUpdatedEvent {
	return &PrimaryDIDUpdatedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		OldDID:    oldDID,
		NewDID:    newDID,
	}
}

// ============================================================================
// Session Events
// ============================================================================

// SessionCreatedEvent is raised when a session is created.
type SessionCreatedEvent struct {
	eventsourcing.BaseEvent
	SessionID  string `json:"session_id"`
	AuthMethod string `json:"auth_method"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
}

// EventType returns the event type.
func (e *SessionCreatedEvent) EventType() string { return EventTypeSessionCreated }

// NewSessionCreatedEvent creates a new SessionCreatedEvent.
func NewSessionCreatedEvent(userID, sessionID, authMethod, userAgent, ipAddress string) *SessionCreatedEvent {
	return &SessionCreatedEvent{
		BaseEvent:  eventsourcing.NewBaseEvent(UserAggregateType, userID),
		SessionID:  sessionID,
		AuthMethod: authMethod,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}
}

// SessionRevokedEvent is raised when a session is revoked.
type SessionRevokedEvent struct {
	eventsourcing.BaseEvent
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
}

// EventType returns the event type.
func (e *SessionRevokedEvent) EventType() string { return EventTypeSessionRevoked }

// NewSessionRevokedEvent creates a new SessionRevokedEvent.
func NewSessionRevokedEvent(userID, sessionID, reason string) *SessionRevokedEvent {
	return &SessionRevokedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		SessionID: sessionID,
		Reason:    reason,
	}
}

// ============================================================================
// API Key Events
// ============================================================================

// APIKeyCreatedEvent is raised when an API key is created.
type APIKeyCreatedEvent struct {
	eventsourcing.BaseEvent
	KeyID  string   `json:"key_id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// EventType returns the event type.
func (e *APIKeyCreatedEvent) EventType() string { return EventTypeAPIKeyCreated }

// NewAPIKeyCreatedEvent creates a new APIKeyCreatedEvent.
func NewAPIKeyCreatedEvent(userID, keyID, name string, scopes []string) *APIKeyCreatedEvent {
	return &APIKeyCreatedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		KeyID:     keyID,
		Name:      name,
		Scopes:    scopes,
	}
}

// APIKeyRevokedEvent is raised when an API key is revoked.
type APIKeyRevokedEvent struct {
	eventsourcing.BaseEvent
	KeyID  string `json:"key_id"`
	Reason string `json:"reason"`
}

// EventType returns the event type.
func (e *APIKeyRevokedEvent) EventType() string { return EventTypeAPIKeyRevoked }

// NewAPIKeyRevokedEvent creates a new APIKeyRevokedEvent.
func NewAPIKeyRevokedEvent(userID, keyID, reason string) *APIKeyRevokedEvent {
	return &APIKeyRevokedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(UserAggregateType, userID),
		KeyID:     keyID,
		Reason:    reason,
	}
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterEvents registers all identity events with the given registry.
func RegisterEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeUserCreated, func() eventsourcing.Event { return &UserCreatedEvent{} })
	registry.Register(EventTypeUserLoggedIn, func() eventsourcing.Event { return &UserLoggedInEvent{} })
	registry.Register(EventTypeUserSuspended, func() eventsourcing.Event { return &UserSuspendedEvent{} })
	registry.Register(EventTypeUserActivated, func() eventsourcing.Event { return &UserActivatedEvent{} })
	registry.Register(EventTypeIdentityLinked, func() eventsourcing.Event { return &IdentityLinkedEvent{} })
	registry.Register(EventTypeEmailVerified, func() eventsourcing.Event { return &EmailVerifiedEvent{} })
	registry.Register(EventTypeWalletLinked, func() eventsourcing.Event { return &WalletLinkedEvent{} })
	registry.Register(EventTypeWalletUnlinked, func() eventsourcing.Event { return &WalletUnlinkedEvent{} })
	registry.Register(EventTypeOAuthLinked, func() eventsourcing.Event { return &OAuthLinkedEvent{} })
	registry.Register(EventTypeOAuthUnlinked, func() eventsourcing.Event { return &OAuthUnlinkedEvent{} })
	registry.Register(EventTypePrimaryDIDUpdated, func() eventsourcing.Event { return &PrimaryDIDUpdatedEvent{} })
	registry.Register(EventTypeSessionCreated, func() eventsourcing.Event { return &SessionCreatedEvent{} })
	registry.Register(EventTypeSessionRevoked, func() eventsourcing.Event { return &SessionRevokedEvent{} })
	registry.Register(EventTypeAPIKeyCreated, func() eventsourcing.Event { return &APIKeyCreatedEvent{} })
	registry.Register(EventTypeAPIKeyRevoked, func() eventsourcing.Event { return &APIKeyRevokedEvent{} })
}
