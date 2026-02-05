package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Verification is the aggregate root for verification attempts.
// It manages the lifecycle of verifying a user's external account via OAuth
// and issuing a credential based on the fetched data.
type Verification struct {
	eventsourcing.AggregateRoot

	id                VerificationID
	userID            string
	provider          ProviderType
	status            VerificationStatus
	oauthState        string
	credentialID      string
	verificationError VerificationError
	startedAt         time.Time
	completedAt       time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// StartVerification creates a new Verification aggregate for a verification attempt.
func StartVerification(id VerificationID, userID string, provider ProviderType, oauthState string) (*Verification, error) {
	if id.IsZero() {
		return nil, VerificationNotFound("Verification.Start", "")
	}

	if userID == "" {
		return nil, VerificationFailed("Verification.Start", "user ID is required")
	}

	if !provider.IsValid() {
		return nil, ProviderNotSupported("Verification.Start", provider.String())
	}

	if oauthState == "" {
		return nil, VerificationFailed("Verification.Start", "oauth state is required")
	}

	v := &Verification{}
	v.InitAggregate(AggregateTypeVerification, id.String())

	now := time.Now().UTC()

	v.Raise(v, &VerificationStartedEvent{
		BaseEvent:      newVerificationBaseEvent(id),
		VerificationID: id.String(),
		UserID:         userID,
		Provider:       provider.String(),
		OAuthState:     oauthState,
		StartedAt:      now,
	})

	return v, nil
}

// NewVerificationFromEvents reconstructs a Verification from events (for hydration).
func NewVerificationFromEvents(id string) *Verification {
	v := &Verification{}
	v.InitAggregate(AggregateTypeVerification, id)
	return v
}

// VerificationFactory creates a factory for Verification aggregates.
func VerificationFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewVerificationFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the verification's ID.
func (v *Verification) ID() VerificationID {
	return v.id
}

// UserID returns the user ID associated with this verification.
func (v *Verification) UserID() string {
	return v.userID
}

// Provider returns the verification provider type.
func (v *Verification) Provider() ProviderType {
	return v.provider
}

// Status returns the current verification status.
func (v *Verification) Status() VerificationStatus {
	return v.status
}

// OAuthState returns the OAuth state token.
func (v *Verification) OAuthState() string {
	return v.oauthState
}

// CredentialID returns the credential ID if one has been issued.
func (v *Verification) CredentialID() string {
	return v.credentialID
}

// VerificationError returns the error details if the verification failed.
func (v *Verification) VerificationError() VerificationError {
	return v.verificationError
}

// StartedAt returns when the verification was started.
func (v *Verification) StartedAt() time.Time {
	return v.startedAt
}

// CompletedAt returns when the verification completed (success or failure).
func (v *Verification) CompletedAt() time.Time {
	return v.completedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// RecordOAuthCallback records that an OAuth callback has been received.
func (v *Verification) RecordOAuthCallback() error {
	if v.status != VerificationStatusOAuthStarted {
		return VerificationFailed("Verification.RecordOAuthCallback",
			"cannot record OAuth callback in current status: "+v.status.String())
	}

	v.Raise(v, &OAuthCallbackReceivedEvent{
		BaseEvent:      newVerificationBaseEvent(v.id),
		VerificationID: v.id.String(),
		ReceivedAt:     time.Now().UTC(),
	})

	return nil
}

// RecordDataFetched records that provider data has been successfully fetched.
func (v *Verification) RecordDataFetched() error {
	if v.status != VerificationStatusOAuthCompleted {
		return VerificationFailed("Verification.RecordDataFetched",
			"cannot record data fetched in current status: "+v.status.String())
	}

	v.Raise(v, &DataFetchCompletedEvent{
		BaseEvent:      newVerificationBaseEvent(v.id),
		VerificationID: v.id.String(),
		CompletedAt:    time.Now().UTC(),
	})

	return nil
}

// RecordDataFetchFailed records that the provider data fetch failed.
func (v *Verification) RecordDataFetchFailed(code, message string) error {
	if v.status != VerificationStatusOAuthCompleted {
		return VerificationFailed("Verification.RecordDataFetchFailed",
			"cannot record data fetch failure in current status: "+v.status.String())
	}

	now := time.Now().UTC()

	v.Raise(v, &DataFetchFailedEvent{
		BaseEvent:      newVerificationBaseEvent(v.id),
		VerificationID: v.id.String(),
		ErrorCode:      code,
		ErrorMessage:   message,
		FailedAt:       now,
	})

	return nil
}

// RecordCredentialIssued records that a credential has been issued for this verification.
func (v *Verification) RecordCredentialIssued(credentialID string) error {
	if v.status != VerificationStatusDataFetched {
		return VerificationFailed("Verification.RecordCredentialIssued",
			"cannot record credential issued in current status: "+v.status.String())
	}

	now := time.Now().UTC()

	v.Raise(v, &CredentialIssueRequestedEvent{
		BaseEvent:      newVerificationBaseEvent(v.id),
		VerificationID: v.id.String(),
		CredentialID:   credentialID,
		RequestedAt:    now,
	})

	v.Raise(v, &VerificationCompletedEvent{
		BaseEvent:      newVerificationBaseEvent(v.id),
		VerificationID: v.id.String(),
		CredentialID:   credentialID,
		CompletedAt:    now,
	})

	return nil
}

// Fail marks the verification as failed.
func (v *Verification) Fail(code, message string) error {
	if v.status.IsTerminal() {
		return VerificationAlreadyCompleted("Verification.Fail", v.id.String())
	}

	v.Raise(v, &VerificationFailedEvent{
		BaseEvent:      newVerificationBaseEvent(v.id),
		VerificationID: v.id.String(),
		ErrorCode:      code,
		ErrorMessage:   message,
		FailedAt:       time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (v *Verification) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *VerificationStartedEvent:
		v.onVerificationStarted(e)
	case *OAuthCallbackReceivedEvent:
		v.onOAuthCallbackReceived(e)
	case *DataFetchCompletedEvent:
		v.onDataFetchCompleted(e)
	case *DataFetchFailedEvent:
		v.onDataFetchFailed(e)
	case *CredentialIssueRequestedEvent:
		v.onCredentialIssueRequested(e)
	case *VerificationCompletedEvent:
		v.onVerificationCompleted(e)
	case *VerificationFailedEvent:
		v.onVerificationFailed(e)
	}
}

func (v *Verification) onVerificationStarted(e *VerificationStartedEvent) {
	var err error
	v.id, err = ParseVerificationID(e.VerificationID)
	if err != nil {
		panic("corrupt event store: VerificationStarted has invalid VerificationID: " + e.VerificationID)
	}
	v.userID = e.UserID
	provider, err := ParseProviderType(e.Provider)
	if err != nil {
		panic("corrupt event store: VerificationStarted has invalid Provider: " + e.Provider)
	}
	v.provider = provider
	v.oauthState = e.OAuthState
	v.status = VerificationStatusOAuthStarted
	v.startedAt = e.StartedAt
}

func (v *Verification) onOAuthCallbackReceived(e *OAuthCallbackReceivedEvent) {
	v.status = VerificationStatusOAuthCompleted
}

func (v *Verification) onDataFetchCompleted(e *DataFetchCompletedEvent) {
	v.status = VerificationStatusDataFetched
}

func (v *Verification) onDataFetchFailed(e *DataFetchFailedEvent) {
	v.status = VerificationStatusFailed
	v.verificationError = NewVerificationError(e.ErrorCode, e.ErrorMessage, nil)
	v.completedAt = e.FailedAt
}

func (v *Verification) onCredentialIssueRequested(e *CredentialIssueRequestedEvent) {
	v.credentialID = e.CredentialID
}

func (v *Verification) onVerificationCompleted(e *VerificationCompletedEvent) {
	v.status = VerificationStatusCredentialIssued
	v.credentialID = e.CredentialID
	v.completedAt = e.CompletedAt
}

func (v *Verification) onVerificationFailed(e *VerificationFailedEvent) {
	v.status = VerificationStatusFailed
	v.verificationError = NewVerificationError(e.ErrorCode, e.ErrorMessage, nil)
	v.completedAt = e.FailedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (v *Verification) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &v.AggregateRoot
}
