package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeVerification = "Verification"
)

// Event type constants
const (
	EventTypeVerificationStarted      = "Verification.Started"
	EventTypeOAuthCallbackReceived    = "Verification.OAuthCallbackReceived"
	EventTypeDataFetchCompleted       = "Verification.DataFetchCompleted"
	EventTypeDataFetchFailed          = "Verification.DataFetchFailed"
	EventTypeCredentialIssueRequested = "Verification.CredentialIssueRequested"
	EventTypeVerificationCompleted    = "Verification.Completed"
	EventTypeVerificationFailed       = "Verification.Failed"
)

// ============================================================================
// Verification Events
// ============================================================================

// VerificationStartedEvent is emitted when a new verification is initiated.
type VerificationStartedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	UserID         string    `json:"user_id"`
	Provider       string    `json:"provider"`
	OAuthState     string    `json:"oauth_state"`
	StartedAt      time.Time `json:"started_at"`
}

// EventType returns the event type.
func (e VerificationStartedEvent) EventType() string {
	return EventTypeVerificationStarted
}

// OAuthCallbackReceivedEvent is emitted when an OAuth callback is received.
type OAuthCallbackReceivedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	ReceivedAt     time.Time `json:"received_at"`
}

// EventType returns the event type.
func (e OAuthCallbackReceivedEvent) EventType() string {
	return EventTypeOAuthCallbackReceived
}

// DataFetchCompletedEvent is emitted when provider data is successfully fetched.
type DataFetchCompletedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	CompletedAt    time.Time `json:"completed_at"`
}

// EventType returns the event type.
func (e DataFetchCompletedEvent) EventType() string {
	return EventTypeDataFetchCompleted
}

// DataFetchFailedEvent is emitted when provider data fetch fails.
type DataFetchFailedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	ErrorCode      string    `json:"error_code"`
	ErrorMessage   string    `json:"error_message"`
	FailedAt       time.Time `json:"failed_at"`
}

// EventType returns the event type.
func (e DataFetchFailedEvent) EventType() string {
	return EventTypeDataFetchFailed
}

// CredentialIssueRequestedEvent is emitted when credential issuance is requested.
type CredentialIssueRequestedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	CredentialID   string    `json:"credential_id"`
	RequestedAt    time.Time `json:"requested_at"`
}

// EventType returns the event type.
func (e CredentialIssueRequestedEvent) EventType() string {
	return EventTypeCredentialIssueRequested
}

// VerificationCompletedEvent is emitted when a verification completes successfully.
type VerificationCompletedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	CredentialID   string    `json:"credential_id"`
	CompletedAt    time.Time `json:"completed_at"`
}

// EventType returns the event type.
func (e VerificationCompletedEvent) EventType() string {
	return EventTypeVerificationCompleted
}

// VerificationFailedEvent is emitted when a verification fails.
type VerificationFailedEvent struct {
	eventsourcing.BaseEvent

	VerificationID string    `json:"verification_id"`
	ErrorCode      string    `json:"error_code"`
	ErrorMessage   string    `json:"error_message"`
	FailedAt       time.Time `json:"failed_at"`
}

// EventType returns the event type.
func (e VerificationFailedEvent) EventType() string {
	return EventTypeVerificationFailed
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterVerificationEvents registers all Verification domain events with the event registry.
func RegisterVerificationEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeVerificationStarted, func() eventsourcing.Event { return &VerificationStartedEvent{} })
	registry.Register(EventTypeOAuthCallbackReceived, func() eventsourcing.Event { return &OAuthCallbackReceivedEvent{} })
	registry.Register(EventTypeDataFetchCompleted, func() eventsourcing.Event { return &DataFetchCompletedEvent{} })
	registry.Register(EventTypeDataFetchFailed, func() eventsourcing.Event { return &DataFetchFailedEvent{} })
	registry.Register(EventTypeCredentialIssueRequested, func() eventsourcing.Event { return &CredentialIssueRequestedEvent{} })
	registry.Register(EventTypeVerificationCompleted, func() eventsourcing.Event { return &VerificationCompletedEvent{} })
	registry.Register(EventTypeVerificationFailed, func() eventsourcing.Event { return &VerificationFailedEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterVerificationEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newVerificationBaseEvent creates a base event for verification aggregate.
func newVerificationBaseEvent(verificationID VerificationID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeVerification, verificationID.String())
}
