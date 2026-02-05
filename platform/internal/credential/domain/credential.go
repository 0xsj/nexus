package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Credential is the aggregate root for verifiable credentials.
// It manages the lifecycle of a W3C Verifiable Credential including
// issuance, revocation, and expiration.
type Credential struct {
	eventsourcing.AggregateRoot

	id               CredentialID
	credType         CredentialType
	issuerDID        string
	subjectDID       string
	claims           Claims
	issuedAt         time.Time
	expiresAt        time.Time
	status           CredentialStatus
	revokedAt        time.Time
	revocationReason string
	jwt              string
	verificationID   string
}

// ============================================================================
// Constructors
// ============================================================================

// IssueCredential creates a new Credential aggregate for issuance.
func IssueCredential(
	id CredentialID,
	credType CredentialType,
	issuerDID string,
	subjectDID string,
	claims Claims,
	expiresAt time.Time,
	verificationID string,
	jwt string,
) (*Credential, error) {
	if id.IsZero() {
		return nil, CredentialInvalid("Credential.Issue", "credential ID is required")
	}

	if !credType.IsValid() {
		return nil, CredentialInvalid("Credential.Issue", "invalid credential type")
	}

	if issuerDID == "" {
		return nil, CredentialInvalid("Credential.Issue", "issuer DID is required")
	}

	if subjectDID == "" {
		return nil, CredentialInvalid("Credential.Issue", "subject DID is required")
	}

	if claims.IsZero() {
		return nil, ClaimsValidationFailed("Credential.Issue", "claims cannot be empty")
	}

	c := &Credential{}
	c.InitAggregate(AggregateTypeCredential, id.String())

	now := time.Now().UTC()

	c.Raise(c, &CredentialIssuedEvent{
		BaseEvent:      newCredentialBaseEvent(id),
		CredentialID:   id.String(),
		CredentialType: credType.String(),
		IssuerDID:      issuerDID,
		SubjectDID:     subjectDID,
		Claims:         claims.ToMap(),
		IssuedAt:       now,
		ExpiresAt:      expiresAt,
		VerificationID: verificationID,
		JWT:            jwt,
	})

	return c, nil
}

// NewCredentialFromEvents reconstructs a Credential from events (for hydration).
func NewCredentialFromEvents(id string) *Credential {
	c := &Credential{}
	c.InitAggregate(AggregateTypeCredential, id)
	return c
}

// CredentialFactory creates a factory for Credential aggregates.
func CredentialFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewCredentialFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the credential's ID.
func (c *Credential) ID() CredentialID {
	return c.id
}

// Type returns the credential's type.
func (c *Credential) Type() CredentialType {
	return c.credType
}

// IssuerDID returns the issuer's DID.
func (c *Credential) IssuerDID() string {
	return c.issuerDID
}

// SubjectDID returns the subject's DID.
func (c *Credential) SubjectDID() string {
	return c.subjectDID
}

// Claims returns the credential's claims.
func (c *Credential) Claims() Claims {
	return c.claims
}

// IssuedAt returns when the credential was issued.
func (c *Credential) IssuedAt() time.Time {
	return c.issuedAt
}

// ExpiresAt returns when the credential expires.
func (c *Credential) ExpiresAt() time.Time {
	return c.expiresAt
}

// Status returns the credential's status.
func (c *Credential) Status() CredentialStatus {
	return c.status
}

// RevokedAt returns when the credential was revoked.
func (c *Credential) RevokedAt() time.Time {
	return c.revokedAt
}

// RevocationReason returns the reason for revocation.
func (c *Credential) RevocationReason() string {
	return c.revocationReason
}

// JWT returns the signed JWT representation.
func (c *Credential) JWT() string {
	return c.jwt
}

// VerificationID returns the verification that triggered this credential.
func (c *Credential) VerificationID() string {
	return c.verificationID
}

// ============================================================================
// Command Methods
// ============================================================================

// Revoke revokes an active credential.
func (c *Credential) Revoke(reason string) error {
	if !c.status.CanTransitionTo(CredentialStatusRevoked) {
		return CredentialRevoked("Credential.Revoke", c.id.String()).
			WithMessage("cannot revoke credential in current status: " + c.status.String())
	}

	c.Raise(c, &CredentialRevokedEvent{
		BaseEvent:    newCredentialBaseEvent(c.id),
		CredentialID: c.id.String(),
		Reason:       reason,
		RevokedAt:    time.Now().UTC(),
	})

	return nil
}

// Expire marks an active credential as expired.
func (c *Credential) Expire() error {
	if !c.status.CanTransitionTo(CredentialStatusExpired) {
		return CredentialExpired("Credential.Expire", c.id.String()).
			WithMessage("cannot expire credential in current status: " + c.status.String())
	}

	c.Raise(c, &CredentialExpiredEvent{
		BaseEvent:    newCredentialBaseEvent(c.id),
		CredentialID: c.id.String(),
		ExpiredAt:    time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (c *Credential) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *CredentialIssuedEvent:
		c.onCredentialIssued(e)
	case *CredentialRevokedEvent:
		c.onCredentialRevoked(e)
	case *CredentialExpiredEvent:
		c.onCredentialExpired(e)
	}
}

func (c *Credential) onCredentialIssued(e *CredentialIssuedEvent) {
	var err error
	c.id, err = ParseCredentialID(e.CredentialID)
	if err != nil {
		panic("corrupt event store: CredentialIssued has invalid CredentialID: " + e.CredentialID)
	}
	c.credType, err = ParseCredentialType(e.CredentialType)
	if err != nil {
		panic("corrupt event store: CredentialIssued has invalid CredentialType: " + e.CredentialType)
	}
	c.claims, err = NewClaims(e.Claims)
	if err != nil {
		panic("corrupt event store: CredentialIssued has invalid Claims")
	}
	c.issuerDID = e.IssuerDID
	c.subjectDID = e.SubjectDID
	c.issuedAt = e.IssuedAt
	c.expiresAt = e.ExpiresAt
	c.status = CredentialStatusActive
	c.verificationID = e.VerificationID
	c.jwt = e.JWT
}

func (c *Credential) onCredentialRevoked(e *CredentialRevokedEvent) {
	c.status = CredentialStatusRevoked
	c.revokedAt = e.RevokedAt
	c.revocationReason = e.Reason
}

func (c *Credential) onCredentialExpired(e *CredentialExpiredEvent) {
	c.status = CredentialStatusExpired
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (c *Credential) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &c.AggregateRoot
}
