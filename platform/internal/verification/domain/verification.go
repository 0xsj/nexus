package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Verification Aggregate
// ============================================================================

// Verification is the aggregate root for verification flows.
// It tracks the lifecycle of verifying a user's identity with an external provider
// and issuing a verifiable credential based on that verification.
type Verification struct {
	eventsourcing.AggregateRoot

	// Identity
	userID         string
	provider       Provider
	credentialType CredentialType

	// OAuth state
	oauthState  string
	redirectURL string

	// Status
	status    VerificationStatus
	expiresAt time.Time

	// Provider data (populated after successful fetch)
	providerUserID string
	username       string
	profile        *ProviderProfile

	// Result
	credentialID string
	claims       map[string]any

	// Failure info
	failureReason string
	failureCode   string

	// Timestamps
	initiatedAt  time.Time
	authorizedAt *time.Time
	completedAt  *time.Time
	failedAt     *time.Time
}

// NewVerification creates a new Verification aggregate.
func NewVerification(id string) *Verification {
	v := &Verification{}
	v.InitAggregate(AggregateType, id)
	return v
}

// GetAggregateRoot returns the embedded AggregateRoot.
// Required for eventsourcing.Hydrate to work.
func (v *Verification) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &v.AggregateRoot
}

// ============================================================================
// Accessors
// ============================================================================

// UserID returns the user ID.
func (v *Verification) UserID() string {
	return v.userID
}

// Provider returns the provider.
func (v *Verification) Provider() Provider {
	return v.provider
}

// CredentialType returns the credential type.
func (v *Verification) CredentialType() CredentialType {
	return v.credentialType
}

// OAuthState returns the OAuth state.
func (v *Verification) OAuthState() string {
	return v.oauthState
}

// RedirectURL returns the redirect URL.
func (v *Verification) RedirectURL() string {
	return v.redirectURL
}

// Status returns the current status.
func (v *Verification) Status() VerificationStatus {
	return v.status
}

// ExpiresAt returns the expiration time.
func (v *Verification) ExpiresAt() time.Time {
	return v.expiresAt
}

// ProviderUserID returns the provider user ID.
func (v *Verification) ProviderUserID() string {
	return v.providerUserID
}

// Username returns the username from the provider.
func (v *Verification) Username() string {
	return v.username
}

// Profile returns the full provider profile.
func (v *Verification) Profile() *ProviderProfile {
	return v.profile
}

// CredentialID returns the issued credential ID.
func (v *Verification) CredentialID() string {
	return v.credentialID
}

// Claims returns the credential claims.
func (v *Verification) Claims() map[string]any {
	if v.claims == nil {
		return nil
	}
	// Return a copy to prevent external modification
	cp := make(map[string]any, len(v.claims))
	for k, v := range v.claims {
		cp[k] = v
	}
	return cp
}

// FailureReason returns the failure reason.
func (v *Verification) FailureReason() string {
	return v.failureReason
}

// FailureCode returns the failure error code.
func (v *Verification) FailureCode() string {
	return v.failureCode
}

// InitiatedAt returns when the verification was initiated.
func (v *Verification) InitiatedAt() time.Time {
	return v.initiatedAt
}

// AuthorizedAt returns when OAuth authorization completed.
func (v *Verification) AuthorizedAt() *time.Time {
	return v.authorizedAt
}

// CompletedAt returns when the verification completed.
func (v *Verification) CompletedAt() *time.Time {
	return v.completedAt
}

// FailedAt returns when the verification failed.
func (v *Verification) FailedAt() *time.Time {
	return v.failedAt
}

// ============================================================================
// State Checks
// ============================================================================

// IsExpired returns true if the verification has expired.
func (v *Verification) IsExpired() bool {
	if v.status == StatusExpired {
		return true
	}
	return time.Now().After(v.expiresAt)
}

// IsActive returns true if the verification is still in progress.
func (v *Verification) IsActive() bool {
	return v.status.IsActive() && !v.IsExpired()
}

// IsCompleted returns true if the verification completed successfully.
func (v *Verification) IsCompleted() bool {
	return v.status == StatusCompleted
}

// IsFailed returns true if the verification failed.
func (v *Verification) IsFailed() bool {
	return v.status == StatusFailed
}

// IsTerminal returns true if the verification is in a terminal state.
func (v *Verification) IsTerminal() bool {
	return v.status.IsTerminal()
}

// ============================================================================
// Commands (Domain Logic)
// ============================================================================

// Initiate starts a new verification flow.
func (v *Verification) Initiate(
	userID string,
	provider Provider,
	credentialType CredentialType,
	oauthState string,
	redirectURL string,
	ttl time.Duration,
) error {
	const op = "Verification.Initiate"

	// Validate not already initialized
	if v.status != "" {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusPending))
	}

	// Validate inputs
	if userID == "" {
		return eventsourcing.ErrAggregateValidation(op, "user ID is required")
	}
	if !provider.IsValid() {
		return ErrProviderNotSupported(op, provider.String())
	}
	if !credentialType.IsValid() {
		return eventsourcing.ErrAggregateValidation(op, "invalid credential type")
	}
	if oauthState == "" {
		return eventsourcing.ErrAggregateValidation(op, "OAuth state is required")
	}

	expiresAt := time.Now().Add(ttl)

	event := NewVerificationInitiatedEvent(
		v.AggregateID(),
		userID,
		provider,
		credentialType,
		oauthState,
		redirectURL,
		expiresAt,
	)
	v.Raise(v, event)

	return nil
}

// Authorize marks the OAuth authorization as successful.
func (v *Verification) Authorize(authorizationCode string) error {
	const op = "Verification.Authorize"

	// Validate state transition
	if !v.status.CanTransitionTo(StatusAuthorized) {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusAuthorized))
	}

	// Check expiration
	if v.IsExpired() {
		return ErrVerificationExpired(op, v.AggregateID())
	}

	// Validate inputs
	if authorizationCode == "" {
		return eventsourcing.ErrAggregateValidation(op, "authorization code is required")
	}

	event := NewVerificationAuthorizedEvent(v.AggregateID(), authorizationCode)
	v.Raise(v, event)

	return nil
}

// UpdateOAuthState updates the OAuth state value.
// This is used when the verification is created before the OAuth state.
func (v *Verification) UpdateOAuthState(oauthState string) error {
	const op = "Verification.UpdateOAuthState"

	if v.status != StatusPending {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusPending))
	}

	if oauthState == "" {
		return eventsourcing.ErrAggregateValidation(op, "OAuth state is required")
	}

	// Directly update - no event needed for this internal state change
	v.oauthState = oauthState
	return nil
}

// StartFetching marks that provider data fetch has begun.
func (v *Verification) StartFetching(accessTokenHash string) error {
	const op = "Verification.StartFetching"

	// Validate state transition
	if !v.status.CanTransitionTo(StatusFetching) {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusFetching))
	}

	// Check expiration
	if v.IsExpired() {
		return ErrVerificationExpired(op, v.AggregateID())
	}

	event := NewVerificationFetchingEvent(v.AggregateID(), accessTokenHash)
	v.Raise(v, event)

	return nil
}

// Complete marks the verification as successfully completed.
func (v *Verification) Complete(
	providerUserID string,
	username string,
	credentialID string,
	claims map[string]any,
) error {
	const op = "Verification.Complete"

	// Validate state transition
	if !v.status.CanTransitionTo(StatusCompleted) {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusCompleted))
	}

	// Validate inputs
	if providerUserID == "" {
		return eventsourcing.ErrAggregateValidation(op, "provider user ID is required")
	}
	if credentialID == "" {
		return eventsourcing.ErrAggregateValidation(op, "credential ID is required")
	}

	event := NewVerificationCompletedEvent(
		v.AggregateID(),
		providerUserID,
		username,
		credentialID,
		claims,
	)
	v.Raise(v, event)

	return nil
}

// Fail marks the verification as failed.
func (v *Verification) Fail(reason string, errorCode string) error {
	const op = "Verification.Fail"

	// Can fail from any non-terminal state
	if v.status.IsTerminal() {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusFailed))
	}

	// Validate inputs
	if reason == "" {
		return eventsourcing.ErrAggregateValidation(op, "failure reason is required")
	}

	event := NewVerificationFailedEvent(v.AggregateID(), reason, errorCode)
	v.Raise(v, event)

	return nil
}

// Expire marks the verification as expired.
func (v *Verification) Expire() error {
	const op = "Verification.Expire"

	// Can only expire from active states
	if v.status.IsTerminal() {
		return ErrVerificationInvalidState(op, string(v.status), string(StatusExpired))
	}

	event := NewVerificationExpiredEvent(v.AggregateID())
	v.Raise(v, event)

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update aggregate state.
func (v *Verification) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *VerificationInitiatedEvent:
		v.applyInitiated(e)
	case *VerificationAuthorizedEvent:
		v.applyAuthorized(e)
	case *VerificationFetchingEvent:
		v.applyFetching(e)
	case *VerificationCompletedEvent:
		v.applyCompleted(e)
	case *VerificationFailedEvent:
		v.applyFailed(e)
	case *VerificationExpiredEvent:
		v.applyExpired(e)
	}
}

func (v *Verification) applyInitiated(e *VerificationInitiatedEvent) {
	v.userID = e.UserID
	v.provider = Provider(e.Provider)
	v.credentialType = CredentialType(e.CredentialType)
	v.oauthState = e.OAuthState
	v.redirectURL = e.RedirectURL
	v.status = StatusPending
	v.initiatedAt = e.InitiatedAt
	v.expiresAt = e.ExpiresAt
}

func (v *Verification) applyAuthorized(e *VerificationAuthorizedEvent) {
	v.status = StatusAuthorized
	v.authorizedAt = &e.AuthorizedAt
}

func (v *Verification) applyFetching(e *VerificationFetchingEvent) {
	v.status = StatusFetching
}

func (v *Verification) applyCompleted(e *VerificationCompletedEvent) {
	v.status = StatusCompleted
	v.providerUserID = e.ProviderUserID
	v.username = e.Username
	v.credentialID = e.CredentialID
	v.claims = e.Claims
	v.completedAt = &e.CompletedAt
}

func (v *Verification) applyFailed(e *VerificationFailedEvent) {
	v.status = StatusFailed
	v.failureReason = e.Reason
	v.failureCode = e.ErrorCode
	v.failedAt = &e.FailedAt
}

func (v *Verification) applyExpired(e *VerificationExpiredEvent) {
	v.status = StatusExpired
}

// ============================================================================
// Reconstruction (for repositories)
// ============================================================================

// ReconstructVerification creates a Verification from stored data.
// This bypasses event sourcing and should only be used by repositories.
func ReconstructVerification(
	id string,
	userID string,
	provider Provider,
	credentialType CredentialType,
	oauthState string,
	redirectURL string,
	status VerificationStatus,
	expiresAt time.Time,
	providerUserID string,
	username string,
	credentialID string,
	claims map[string]any,
	failureReason string,
	failureCode string,
	initiatedAt time.Time,
	authorizedAt *time.Time,
	completedAt *time.Time,
	failedAt *time.Time,
	version int,
) *Verification {
	v := &Verification{
		userID:         userID,
		provider:       provider,
		credentialType: credentialType,
		oauthState:     oauthState,
		redirectURL:    redirectURL,
		status:         status,
		expiresAt:      expiresAt,
		providerUserID: providerUserID,
		username:       username,
		credentialID:   credentialID,
		claims:         claims,
		failureReason:  failureReason,
		failureCode:    failureCode,
		initiatedAt:    initiatedAt,
		authorizedAt:   authorizedAt,
		completedAt:    completedAt,
		failedAt:       failedAt,
	}
	v.InitAggregate(AggregateType, id)
	v.SetVersion(version)
	return v
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ eventsourcing.Aggregate = (*Verification)(nil)
