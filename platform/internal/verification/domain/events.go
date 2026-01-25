package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Aggregate Type
// ============================================================================

const (
	// AggregateType is the type identifier for verification aggregates.
	AggregateType = "verification"
)

// ============================================================================
// Event Types
// ============================================================================

const (
	EventTypeVerificationInitiated  = "integration.verification.initiated"
	EventTypeVerificationAuthorized = "integration.verification.authorized"
	EventTypeVerificationFetching   = "integration.verification.fetching"
	EventTypeVerificationCompleted  = "integration.verification.completed"
	EventTypeVerificationFailed     = "integration.verification.failed"
	EventTypeVerificationExpired    = "integration.verification.expired"
)

// ============================================================================
// Verification Initiated Event
// ============================================================================

// VerificationInitiatedEvent is raised when a verification flow is started.
type VerificationInitiatedEvent struct {
	eventsourcing.BaseEvent

	UserID         string    `json:"user_id"`
	Provider       string    `json:"provider"`
	CredentialType string    `json:"credential_type"`
	OAuthState     string    `json:"oauth_state"`
	RedirectURL    string    `json:"redirect_url,omitempty"`
	InitiatedAt    time.Time `json:"initiated_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// EventType returns the event type.
func (e *VerificationInitiatedEvent) EventType() string {
	return EventTypeVerificationInitiated
}

// NewVerificationInitiatedEvent creates a new VerificationInitiatedEvent.
func NewVerificationInitiatedEvent(
	aggregateID string,
	userID string,
	provider Provider,
	credentialType CredentialType,
	oauthState string,
	redirectURL string,
	expiresAt time.Time,
) *VerificationInitiatedEvent {
	return &VerificationInitiatedEvent{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		UserID:         userID,
		Provider:       provider.String(),
		CredentialType: credentialType.String(),
		OAuthState:     oauthState,
		RedirectURL:    redirectURL,
		InitiatedAt:    time.Now().UTC(),
		ExpiresAt:      expiresAt,
	}
}

// ============================================================================
// Verification Authorized Event
// ============================================================================

// VerificationAuthorizedEvent is raised when OAuth authorization succeeds.
type VerificationAuthorizedEvent struct {
	eventsourcing.BaseEvent

	AuthorizationCode string    `json:"authorization_code"`
	AuthorizedAt      time.Time `json:"authorized_at"`
}

// EventType returns the event type.
func (e *VerificationAuthorizedEvent) EventType() string {
	return EventTypeVerificationAuthorized
}

// NewVerificationAuthorizedEvent creates a new VerificationAuthorizedEvent.
func NewVerificationAuthorizedEvent(aggregateID string, authorizationCode string) *VerificationAuthorizedEvent {
	return &VerificationAuthorizedEvent{
		BaseEvent:         eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		AuthorizationCode: authorizationCode,
		AuthorizedAt:      time.Now().UTC(),
	}
}

// ============================================================================
// Verification Fetching Event
// ============================================================================

// VerificationFetchingEvent is raised when provider data fetch begins.
type VerificationFetchingEvent struct {
	eventsourcing.BaseEvent

	AccessTokenHash string    `json:"access_token_hash"` // Hash for audit, not the actual token
	StartedAt       time.Time `json:"started_at"`
}

// EventType returns the event type.
func (e *VerificationFetchingEvent) EventType() string {
	return EventTypeVerificationFetching
}

// NewVerificationFetchingEvent creates a new VerificationFetchingEvent.
func NewVerificationFetchingEvent(aggregateID string, accessTokenHash string) *VerificationFetchingEvent {
	return &VerificationFetchingEvent{
		BaseEvent:       eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		AccessTokenHash: accessTokenHash,
		StartedAt:       time.Now().UTC(),
	}
}

// ============================================================================
// Verification Completed Event
// ============================================================================

// VerificationCompletedEvent is raised when verification completes successfully.
type VerificationCompletedEvent struct {
	eventsourcing.BaseEvent

	ProviderUserID string         `json:"provider_user_id"`
	Username       string         `json:"username"`
	CredentialID   string         `json:"credential_id"`
	Claims         map[string]any `json:"claims"`
	CompletedAt    time.Time      `json:"completed_at"`
}

// EventType returns the event type.
func (e *VerificationCompletedEvent) EventType() string {
	return EventTypeVerificationCompleted
}

// NewVerificationCompletedEvent creates a new VerificationCompletedEvent.
func NewVerificationCompletedEvent(
	aggregateID string,
	providerUserID string,
	username string,
	credentialID string,
	claims map[string]any,
) *VerificationCompletedEvent {
	return &VerificationCompletedEvent{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		ProviderUserID: providerUserID,
		Username:       username,
		CredentialID:   credentialID,
		Claims:         claims,
		CompletedAt:    time.Now().UTC(),
	}
}

// ============================================================================
// Verification Failed Event
// ============================================================================

// VerificationFailedEvent is raised when verification fails.
type VerificationFailedEvent struct {
	eventsourcing.BaseEvent

	Reason    string    `json:"reason"`
	ErrorCode string    `json:"error_code,omitempty"`
	FailedAt  time.Time `json:"failed_at"`
}

// EventType returns the event type.
func (e *VerificationFailedEvent) EventType() string {
	return EventTypeVerificationFailed
}

// NewVerificationFailedEvent creates a new VerificationFailedEvent.
func NewVerificationFailedEvent(aggregateID string, reason string, errorCode string) *VerificationFailedEvent {
	return &VerificationFailedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		Reason:    reason,
		ErrorCode: errorCode,
		FailedAt:  time.Now().UTC(),
	}
}

// ============================================================================
// Verification Expired Event
// ============================================================================

// VerificationExpiredEvent is raised when a verification flow expires.
type VerificationExpiredEvent struct {
	eventsourcing.BaseEvent

	ExpiredAt time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e *VerificationExpiredEvent) EventType() string {
	return EventTypeVerificationExpired
}

// NewVerificationExpiredEvent creates a new VerificationExpiredEvent.
func NewVerificationExpiredEvent(aggregateID string) *VerificationExpiredEvent {
	return &VerificationExpiredEvent{
		BaseEvent: eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		ExpiredAt: time.Now().UTC(),
	}
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterEvents registers all integration events with the given registry.
func RegisterEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeVerificationInitiated, func() eventsourcing.Event {
		return &VerificationInitiatedEvent{}
	})
	registry.Register(EventTypeVerificationAuthorized, func() eventsourcing.Event {
		return &VerificationAuthorizedEvent{}
	})
	registry.Register(EventTypeVerificationFetching, func() eventsourcing.Event {
		return &VerificationFetchingEvent{}
	})
	registry.Register(EventTypeVerificationCompleted, func() eventsourcing.Event {
		return &VerificationCompletedEvent{}
	})
	registry.Register(EventTypeVerificationFailed, func() eventsourcing.Event {
		return &VerificationFailedEvent{}
	})
	registry.Register(EventTypeVerificationExpired, func() eventsourcing.Event {
		return &VerificationExpiredEvent{}
	})
}

// NewRegistry creates a new event registry with all integration events registered.
func NewRegistry() *eventsourcing.EventRegistry {
	registry := eventsourcing.NewEventRegistry()
	RegisterEvents(registry)
	return registry
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ eventsourcing.Event = (*VerificationInitiatedEvent)(nil)
	_ eventsourcing.Event = (*VerificationAuthorizedEvent)(nil)
	_ eventsourcing.Event = (*VerificationFetchingEvent)(nil)
	_ eventsourcing.Event = (*VerificationCompletedEvent)(nil)
	_ eventsourcing.Event = (*VerificationFailedEvent)(nil)
	_ eventsourcing.Event = (*VerificationExpiredEvent)(nil)
)
