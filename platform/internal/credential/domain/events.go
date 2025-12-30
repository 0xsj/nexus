package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Aggregate Type
// ============================================================================

const (
	// AggregateType is the type identifier for credential aggregates.
	AggregateType = "credential"
)

// ============================================================================
// Event Types
// ============================================================================

const (
	EventTypeCredentialRequested  = "credential.requested"
	EventTypeCredentialIssued     = "credential.issued"
	EventTypeCredentialRevoked    = "credential.revoked"
	EventTypeCredentialSuspended  = "credential.suspended"
	EventTypeCredentialReinstated = "credential.reinstated"
	EventTypeCredentialExpired    = "credential.expired"
)

// ============================================================================
// Credential Requested Event
// ============================================================================

// CredentialRequested is raised when a credential is requested by a holder.
type CredentialRequested struct {
	eventsourcing.BaseEvent
	HolderDID      string         `json:"holder_did"`
	IssuerDID      string         `json:"issuer_did"`
	CredentialType string         `json:"credential_type"`
	Claims         map[string]any `json:"claims,omitempty"`
	RequestedAt    time.Time      `json:"requested_at"`
}

// EventType returns the event type.
func (e *CredentialRequested) EventType() string {
	return EventTypeCredentialRequested
}

// NewCredentialRequested creates a new CredentialRequested event.
func NewCredentialRequested(
	aggregateID string,
	holderDID string,
	issuerDID string,
	credentialType string,
	claims map[string]any,
) *CredentialRequested {
	return &CredentialRequested{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		HolderDID:      holderDID,
		IssuerDID:      issuerDID,
		CredentialType: credentialType,
		Claims:         claims,
		RequestedAt:    time.Now().UTC(),
	}
}

// ============================================================================
// Credential Issued Event
// ============================================================================

// CredentialIssued is raised when a credential is issued by an issuer.
type CredentialIssued struct {
	eventsourcing.BaseEvent
	IssuerDID      string         `json:"issuer_did"`
	HolderDID      string         `json:"holder_did"`
	CredentialType string         `json:"credential_type"`
	SchemaID       string         `json:"schema_id,omitempty"`
	Claims         map[string]any `json:"claims"`
	SignedVC       string         `json:"signed_vc,omitempty"`
	IssuedAt       time.Time      `json:"issued_at"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
}

// EventType returns the event type.
func (e *CredentialIssued) EventType() string {
	return EventTypeCredentialIssued
}

// NewCredentialIssued creates a new CredentialIssued event.
func NewCredentialIssued(
	aggregateID string,
	issuerDID string,
	holderDID string,
	credentialType string,
	claims map[string]any,
	expiresAt *time.Time,
) *CredentialIssued {
	return &CredentialIssued{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		IssuerDID:      issuerDID,
		HolderDID:      holderDID,
		CredentialType: credentialType,
		Claims:         claims,
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      expiresAt,
	}
}

// WithSchemaID sets the schema ID.
func (e *CredentialIssued) WithSchemaID(schemaID string) *CredentialIssued {
	e.SchemaID = schemaID
	return e
}

// WithSignedVC sets the signed verifiable credential JWT.
func (e *CredentialIssued) WithSignedVC(signedVC string) *CredentialIssued {
	e.SignedVC = signedVC
	return e
}

// ============================================================================
// Credential Revoked Event
// ============================================================================

// CredentialRevoked is raised when a credential is permanently revoked.
type CredentialRevoked struct {
	eventsourcing.BaseEvent
	RevokedBy string    `json:"revoked_by"`
	Reason    string    `json:"reason"`
	RevokedAt time.Time `json:"revoked_at"`
}

// EventType returns the event type.
func (e *CredentialRevoked) EventType() string {
	return EventTypeCredentialRevoked
}

// NewCredentialRevoked creates a new CredentialRevoked event.
func NewCredentialRevoked(aggregateID string, revokedBy string, reason string) *CredentialRevoked {
	return &CredentialRevoked{
		BaseEvent: eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		RevokedBy: revokedBy,
		Reason:    reason,
		RevokedAt: time.Now().UTC(),
	}
}

// ============================================================================
// Credential Suspended Event
// ============================================================================

// CredentialSuspended is raised when a credential is temporarily suspended.
type CredentialSuspended struct {
	eventsourcing.BaseEvent
	SuspendedBy string     `json:"suspended_by"`
	Reason      string     `json:"reason"`
	SuspendedAt time.Time  `json:"suspended_at"`
	Until       *time.Time `json:"until,omitempty"`
}

// EventType returns the event type.
func (e *CredentialSuspended) EventType() string {
	return EventTypeCredentialSuspended
}

// NewCredentialSuspended creates a new CredentialSuspended event.
func NewCredentialSuspended(
	aggregateID string,
	suspendedBy string,
	reason string,
	until *time.Time,
) *CredentialSuspended {
	return &CredentialSuspended{
		BaseEvent:   eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		SuspendedBy: suspendedBy,
		Reason:      reason,
		SuspendedAt: time.Now().UTC(),
		Until:       until,
	}
}

// ============================================================================
// Credential Reinstated Event
// ============================================================================

// CredentialReinstated is raised when a suspended credential is reinstated.
type CredentialReinstated struct {
	eventsourcing.BaseEvent
	ReinstatedBy string    `json:"reinstated_by"`
	Reason       string    `json:"reason,omitempty"`
	ReinstatedAt time.Time `json:"reinstated_at"`
}

// EventType returns the event type.
func (e *CredentialReinstated) EventType() string {
	return EventTypeCredentialReinstated
}

// NewCredentialReinstated creates a new CredentialReinstated event.
func NewCredentialReinstated(aggregateID string, reinstatedBy string, reason string) *CredentialReinstated {
	return &CredentialReinstated{
		BaseEvent:    eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		ReinstatedBy: reinstatedBy,
		Reason:       reason,
		ReinstatedAt: time.Now().UTC(),
	}
}

// ============================================================================
// Credential Expired Event
// ============================================================================

// CredentialExpired is raised when a credential reaches its expiration date.
type CredentialExpired struct {
	eventsourcing.BaseEvent
	ExpiredAt time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e *CredentialExpired) EventType() string {
	return EventTypeCredentialExpired
}

// NewCredentialExpired creates a new CredentialExpired event.
func NewCredentialExpired(aggregateID string) *CredentialExpired {
	return &CredentialExpired{
		BaseEvent: eventsourcing.NewBaseEvent(AggregateType, aggregateID),
		ExpiredAt: time.Now().UTC(),
	}
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ eventsourcing.Event = (*CredentialRequested)(nil)
	_ eventsourcing.Event = (*CredentialIssued)(nil)
	_ eventsourcing.Event = (*CredentialRevoked)(nil)
	_ eventsourcing.Event = (*CredentialSuspended)(nil)
	_ eventsourcing.Event = (*CredentialReinstated)(nil)
	_ eventsourcing.Event = (*CredentialExpired)(nil)
)
