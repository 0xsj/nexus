package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Vouch is the aggregate root for peer endorsements.
// It manages the lifecycle of a vouch from one user to another,
// optionally tied to a specific credential or claim.
type Vouch struct {
	eventsourcing.AggregateRoot

	id               VouchID
	voucherID        string
	voucheeID        string
	credentialID     string
	claimKey         string
	relationship     RelationshipType
	strength         VouchStrength
	statement        string
	vouchContext     string
	status           VouchStatus
	expiresAt        time.Time
	createdAt        time.Time
	acceptedAt       time.Time
	revokedAt        time.Time
	revocationReason string
	updatedAt        time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// GiveVouch creates a new Vouch aggregate for giving a vouch.
func GiveVouch(
	id VouchID,
	voucherID string,
	voucheeID string,
	credentialID string,
	claimKey string,
	relationship RelationshipType,
	strength VouchStrength,
	statement string,
	vouchContext string,
	expiresAt time.Time,
) (*Vouch, error) {
	if id.IsZero() {
		return nil, VouchInvalid("Vouch.Give", "vouch ID is required")
	}

	if voucherID == "" {
		return nil, VouchInvalid("Vouch.Give", "voucher ID is required")
	}

	if voucheeID == "" {
		return nil, VouchInvalid("Vouch.Give", "vouchee ID is required")
	}

	if voucherID == voucheeID {
		return nil, SelfVouch("Vouch.Give", voucherID)
	}

	if !relationship.IsValid() {
		return nil, VouchInvalid("Vouch.Give", "invalid relationship type")
	}

	if !strength.IsValid() {
		return nil, VouchInvalid("Vouch.Give", "strength must be between 1 and 10")
	}

	v := &Vouch{}
	v.InitAggregate(AggregateTypeVouch, id.String())

	now := time.Now().UTC()

	v.Raise(v, &VouchGivenEvent{
		BaseEvent:    newVouchBaseEvent(id),
		VouchID:      id.String(),
		VoucherID:    voucherID,
		VoucheeID:    voucheeID,
		CredentialID: credentialID,
		ClaimKey:     claimKey,
		Relationship: relationship.String(),
		Strength:     strength.Value(),
		Statement:    statement,
		Context:      vouchContext,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	})

	return v, nil
}

// NewVouchFromEvents reconstructs a Vouch from events (for hydration).
func NewVouchFromEvents(id string) *Vouch {
	v := &Vouch{}
	v.InitAggregate(AggregateTypeVouch, id)
	return v
}

// VouchFactory creates a factory for Vouch aggregates.
func VouchFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewVouchFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the vouch's ID.
func (v *Vouch) ID() VouchID {
	return v.id
}

// VoucherID returns the voucher's user ID.
func (v *Vouch) VoucherID() string {
	return v.voucherID
}

// VoucheeID returns the vouchee's user ID.
func (v *Vouch) VoucheeID() string {
	return v.voucheeID
}

// CredentialID returns the optional credential ID the vouch is tied to.
func (v *Vouch) CredentialID() string {
	return v.credentialID
}

// ClaimKey returns the optional claim key the vouch is tied to.
func (v *Vouch) ClaimKey() string {
	return v.claimKey
}

// Relationship returns the relationship type between voucher and vouchee.
func (v *Vouch) Relationship() RelationshipType {
	return v.relationship
}

// Strength returns the vouch strength.
func (v *Vouch) Strength() VouchStrength {
	return v.strength
}

// Statement returns the vouch statement.
func (v *Vouch) Statement() string {
	return v.statement
}

// VouchContext returns the vouch context.
func (v *Vouch) VouchContext() string {
	return v.vouchContext
}

// Status returns the vouch's status.
func (v *Vouch) Status() VouchStatus {
	return v.status
}

// ExpiresAt returns when the vouch expires.
func (v *Vouch) ExpiresAt() time.Time {
	return v.expiresAt
}

// CreatedAt returns when the vouch was created.
func (v *Vouch) CreatedAt() time.Time {
	return v.createdAt
}

// AcceptedAt returns when the vouch was accepted.
func (v *Vouch) AcceptedAt() time.Time {
	return v.acceptedAt
}

// RevokedAt returns when the vouch was revoked.
func (v *Vouch) RevokedAt() time.Time {
	return v.revokedAt
}

// RevocationReason returns the reason for revocation.
func (v *Vouch) RevocationReason() string {
	return v.revocationReason
}

// UpdatedAt returns when the vouch was last updated.
func (v *Vouch) UpdatedAt() time.Time {
	return v.updatedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Accept accepts a pending vouch. Only the vouchee can accept.
func (v *Vouch) Accept(voucheeID string) error {
	if voucheeID != v.voucheeID {
		return VouchUnauthorized("Vouch.Accept", v.id.String()).
			WithMessage("only the vouchee can accept a vouch")
	}

	if !v.status.CanTransitionTo(VouchStatusAccepted) {
		return VouchAlreadyAccepted("Vouch.Accept", v.id.String()).
			WithMessage("cannot accept vouch in current status: " + v.status.String())
	}

	v.Raise(v, &VouchAcceptedEvent{
		BaseEvent:  newVouchBaseEvent(v.id),
		VouchID:    v.id.String(),
		AcceptedAt: time.Now().UTC(),
	})

	return nil
}

// Revoke revokes a vouch.
func (v *Vouch) Revoke(reason string) error {
	if !v.status.CanTransitionTo(VouchStatusRevoked) {
		return VouchRevoked("Vouch.Revoke", v.id.String()).
			WithMessage("cannot revoke vouch in current status: " + v.status.String())
	}

	v.Raise(v, &VouchRevokedEvent{
		BaseEvent: newVouchBaseEvent(v.id),
		VouchID:   v.id.String(),
		Reason:    reason,
		RevokedAt: time.Now().UTC(),
	})

	return nil
}

// Expire marks a vouch as expired.
func (v *Vouch) Expire() error {
	if !v.status.CanTransitionTo(VouchStatusExpired) {
		return VouchExpired("Vouch.Expire", v.id.String()).
			WithMessage("cannot expire vouch in current status: " + v.status.String())
	}

	v.Raise(v, &VouchExpiredEvent{
		BaseEvent: newVouchBaseEvent(v.id),
		VouchID:   v.id.String(),
		ExpiredAt: time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (v *Vouch) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *VouchGivenEvent:
		v.onVouchGiven(e)
	case *VouchAcceptedEvent:
		v.onVouchAccepted(e)
	case *VouchRevokedEvent:
		v.onVouchRevoked(e)
	case *VouchExpiredEvent:
		v.onVouchExpired(e)
	}
}

func (v *Vouch) onVouchGiven(e *VouchGivenEvent) {
	var err error
	v.id, err = ParseVouchID(e.VouchID)
	if err != nil {
		panic("corrupt event store: VouchGiven has invalid VouchID: " + e.VouchID)
	}
	v.voucherID = e.VoucherID
	v.voucheeID = e.VoucheeID
	v.credentialID = e.CredentialID
	v.claimKey = e.ClaimKey
	v.relationship, err = ParseRelationshipType(e.Relationship)
	if err != nil {
		panic("corrupt event store: VouchGiven has invalid Relationship: " + e.Relationship)
	}
	v.strength, err = NewVouchStrength(e.Strength)
	if err != nil {
		panic("corrupt event store: VouchGiven has invalid Strength")
	}
	v.statement = e.Statement
	v.vouchContext = e.Context
	v.expiresAt = e.ExpiresAt
	v.createdAt = e.CreatedAt
	v.status = VouchStatusPending
	v.updatedAt = e.CreatedAt
}

func (v *Vouch) onVouchAccepted(e *VouchAcceptedEvent) {
	v.status = VouchStatusAccepted
	v.acceptedAt = e.AcceptedAt
	v.updatedAt = e.AcceptedAt
}

func (v *Vouch) onVouchRevoked(e *VouchRevokedEvent) {
	v.status = VouchStatusRevoked
	v.revokedAt = e.RevokedAt
	v.revocationReason = e.Reason
	v.updatedAt = e.RevokedAt
}

func (v *Vouch) onVouchExpired(e *VouchExpiredEvent) {
	v.status = VouchStatusExpired
	v.updatedAt = e.ExpiredAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (v *Vouch) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &v.AggregateRoot
}
