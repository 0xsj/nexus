package event

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Aggregate Constants
// ============================================================================

const (
	// AggregateType is the type name for credential aggregates.
	AggregateType = "Credential"
)

// ============================================================================
// Event Type Constants
// ============================================================================

const (
	TypeCredentialRequested  = "Credential.Requested"
	TypeCredentialIssued     = "Credential.Issued"
	TypeCredentialRevoked    = "Credential.Revoked"
	TypeCredentialSuspended  = "Credential.Suspended"
	TypeCredentialReinstated = "Credential.Reinstated"
	TypeCredentialExpired    = "Credential.Expired"
)

// ============================================================================
// CredentialRequested Event
// ============================================================================

// CredentialRequested is emitted when a holder requests a credential from an issuer.
type CredentialRequested struct {
	eventsourcing.BaseEvent

	// CredentialID is the unique identifier for this credential.
	CredentialID string `json:"credential_id"`

	// HolderDID is the DID of the credential holder (subject).
	HolderDID string `json:"holder_did"`

	// IssuerDID is the DID of the credential issuer.
	IssuerDID string `json:"issuer_did"`

	// CredentialType is the type of credential being requested.
	CredentialType string `json:"credential_type"`

	// Claims are the requested claims/attributes.
	Claims map[string]any `json:"claims,omitempty"`

	// RequestedAt is when the request was made.
	RequestedAt time.Time `json:"requested_at"`
}

// NewCredentialRequested creates a new CredentialRequested event.
func NewCredentialRequested(
	credentialID string,
	holderDID string,
	issuerDID string,
	credentialType string,
	claims map[string]any,
) *CredentialRequested {
	return &CredentialRequested{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, credentialID),
		CredentialID:   credentialID,
		HolderDID:      holderDID,
		IssuerDID:      issuerDID,
		CredentialType: credentialType,
		Claims:         claims,
		RequestedAt:    time.Now().UTC(),
	}
}

// EventType returns the event type.
func (e *CredentialRequested) EventType() string {
	return TypeCredentialRequested
}

// ============================================================================
// CredentialIssued Event
// ============================================================================

// CredentialIssued is emitted when a credential is issued to a holder.
type CredentialIssued struct {
	eventsourcing.BaseEvent

	// CredentialID is the unique identifier for this credential.
	CredentialID string `json:"credential_id"`

	// HolderDID is the DID of the credential holder (subject).
	HolderDID string `json:"holder_did"`

	// IssuerDID is the DID of the credential issuer.
	IssuerDID string `json:"issuer_did"`

	// CredentialType is the type of credential.
	CredentialType string `json:"credential_type"`

	// Claims are the credential claims/attributes.
	Claims map[string]any `json:"claims"`

	// IssuedAt is when the credential was issued.
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt is when the credential expires (optional).
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CredentialJWT is the signed JWT representation (if using JWT-VC format).
	CredentialJWT string `json:"credential_jwt,omitempty"`

	// SchemaID is the optional schema identifier for the credential.
	SchemaID string `json:"schema_id,omitempty"`

	// Evidence contains optional evidence supporting the credential.
	Evidence []Evidence `json:"evidence,omitempty"`
}

// Evidence represents supporting evidence for a credential.
type Evidence struct {
	ID              string    `json:"id,omitempty"`
	Type            string    `json:"type"`
	Verifier        string    `json:"verifier,omitempty"`
	VerificationURL string    `json:"verification_url,omitempty"`
	VerifiedAt      time.Time `json:"verified_at,omitempty"`
}

// NewCredentialIssued creates a new CredentialIssued event.
func NewCredentialIssued(
	credentialID string,
	holderDID string,
	issuerDID string,
	credentialType string,
	claims map[string]any,
	expiresAt *time.Time,
) *CredentialIssued {
	return &CredentialIssued{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, credentialID),
		CredentialID:   credentialID,
		HolderDID:      holderDID,
		IssuerDID:      issuerDID,
		CredentialType: credentialType,
		Claims:         claims,
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      expiresAt,
	}
}

// EventType returns the event type.
func (e *CredentialIssued) EventType() string {
	return TypeCredentialIssued
}

// WithCredentialJWT sets the JWT representation.
func (e *CredentialIssued) WithCredentialJWT(jwt string) *CredentialIssued {
	e.CredentialJWT = jwt
	return e
}

// WithSchemaID sets the schema identifier.
func (e *CredentialIssued) WithSchemaID(schemaID string) *CredentialIssued {
	e.SchemaID = schemaID
	return e
}

// WithEvidence sets the evidence.
func (e *CredentialIssued) WithEvidence(evidence []Evidence) *CredentialIssued {
	e.Evidence = evidence
	return e
}

// ============================================================================
// CredentialRevoked Event
// ============================================================================

// CredentialRevoked is emitted when a credential is permanently revoked.
type CredentialRevoked struct {
	eventsourcing.BaseEvent

	// CredentialID is the unique identifier for this credential.
	CredentialID string `json:"credential_id"`

	// RevokedBy is the DID of the entity that revoked the credential.
	RevokedBy string `json:"revoked_by"`

	// Reason is the reason for revocation.
	Reason string `json:"reason"`

	// RevokedAt is when the credential was revoked.
	RevokedAt time.Time `json:"revoked_at"`
}

// NewCredentialRevoked creates a new CredentialRevoked event.
func NewCredentialRevoked(credentialID string, revokedBy string, reason string) *CredentialRevoked {
	return &CredentialRevoked{
		BaseEvent:    eventsourcing.NewBaseEvent(AggregateType, credentialID),
		CredentialID: credentialID,
		RevokedBy:    revokedBy,
		Reason:       reason,
		RevokedAt:    time.Now().UTC(),
	}
}

// EventType returns the event type.
func (e *CredentialRevoked) EventType() string {
	return TypeCredentialRevoked
}

// ============================================================================
// CredentialSuspended Event
// ============================================================================

// CredentialSuspended is emitted when a credential is temporarily suspended.
type CredentialSuspended struct {
	eventsourcing.BaseEvent

	// CredentialID is the unique identifier for this credential.
	CredentialID string `json:"credential_id"`

	// SuspendedBy is the DID of the entity that suspended the credential.
	SuspendedBy string `json:"suspended_by"`

	// Reason is the reason for suspension.
	Reason string `json:"reason"`

	// SuspendedAt is when the credential was suspended.
	SuspendedAt time.Time `json:"suspended_at"`

	// SuspendedUntil is when the suspension ends (optional).
	SuspendedUntil *time.Time `json:"suspended_until,omitempty"`
}

// NewCredentialSuspended creates a new CredentialSuspended event.
func NewCredentialSuspended(credentialID string, suspendedBy string, reason string, until *time.Time) *CredentialSuspended {
	return &CredentialSuspended{
		BaseEvent:      eventsourcing.NewBaseEvent(AggregateType, credentialID),
		CredentialID:   credentialID,
		SuspendedBy:    suspendedBy,
		Reason:         reason,
		SuspendedAt:    time.Now().UTC(),
		SuspendedUntil: until,
	}
}

// EventType returns the event type.
func (e *CredentialSuspended) EventType() string {
	return TypeCredentialSuspended
}

// ============================================================================
// CredentialReinstated Event
// ============================================================================

// CredentialReinstated is emitted when a suspended credential is reinstated.
type CredentialReinstated struct {
	eventsourcing.BaseEvent

	// CredentialID is the unique identifier for this credential.
	CredentialID string `json:"credential_id"`

	// ReinstatedBy is the DID of the entity that reinstated the credential.
	ReinstatedBy string `json:"reinstated_by"`

	// Reason is the reason for reinstatement.
	Reason string `json:"reason,omitempty"`

	// ReinstatedAt is when the credential was reinstated.
	ReinstatedAt time.Time `json:"reinstated_at"`
}

// NewCredentialReinstated creates a new CredentialReinstated event.
func NewCredentialReinstated(credentialID string, reinstatedBy string, reason string) *CredentialReinstated {
	return &CredentialReinstated{
		BaseEvent:    eventsourcing.NewBaseEvent(AggregateType, credentialID),
		CredentialID: credentialID,
		ReinstatedBy: reinstatedBy,
		Reason:       reason,
		ReinstatedAt: time.Now().UTC(),
	}
}

// EventType returns the event type.
func (e *CredentialReinstated) EventType() string {
	return TypeCredentialReinstated
}

// ============================================================================
// CredentialExpired Event
// ============================================================================

// CredentialExpired is emitted when a credential reaches its expiration date.
type CredentialExpired struct {
	eventsourcing.BaseEvent

	// CredentialID is the unique identifier for this credential.
	CredentialID string `json:"credential_id"`

	// ExpiredAt is when the credential expired.
	ExpiredAt time.Time `json:"expired_at"`
}

// NewCredentialExpired creates a new CredentialExpired event.
func NewCredentialExpired(credentialID string, expiredAt time.Time) *CredentialExpired {
	return &CredentialExpired{
		BaseEvent:    eventsourcing.NewBaseEvent(AggregateType, credentialID),
		CredentialID: credentialID,
		ExpiredAt:    expiredAt,
	}
}

// EventType returns the event type.
func (e *CredentialExpired) EventType() string {
	return TypeCredentialExpired
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
