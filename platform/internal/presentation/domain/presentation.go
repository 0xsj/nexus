package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Presentation is the aggregate root for verifiable presentations.
// It manages the lifecycle of a Verifiable Presentation including
// creation, revocation, and disclosure policy enforcement.
type Presentation struct {
	eventsourcing.AggregateRoot

	id               PresentationID
	holderDID        string
	credentialIDs    []string
	disclosurePolicy DisclosurePolicy
	vpJWT            string
	purpose          string
	status           PresentationStatus
	createdAt        time.Time
	revokedAt        time.Time
	revocationReason string
	updatedAt        time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// CreatePresentation creates a new Presentation aggregate.
func CreatePresentation(
	id PresentationID,
	holderDID string,
	credentialIDs []string,
	disclosurePolicy DisclosurePolicy,
	purpose string,
) (*Presentation, error) {
	if id.IsZero() {
		return nil, PresentationInvalid("Presentation.Create", "presentation ID is required")
	}

	if holderDID == "" {
		return nil, PresentationInvalid("Presentation.Create", "holder DID is required")
	}

	if len(credentialIDs) == 0 {
		return nil, PresentationInvalid("Presentation.Create", "at least one credential ID is required")
	}

	if disclosurePolicy.IsZero() {
		return nil, PresentationInvalid("Presentation.Create", "disclosure policy is required")
	}

	p := &Presentation{}
	p.InitAggregate(AggregateTypePresentation, id.String())

	now := time.Now().UTC()

	// Defensive copy of credential IDs
	credIDs := make([]string, len(credentialIDs))
	copy(credIDs, credentialIDs)

	p.Raise(p, &PresentationCreatedEvent{
		BaseEvent:        newPresentationBaseEvent(id),
		PresentationID:   id.String(),
		HolderDID:        holderDID,
		CredentialIDs:    credIDs,
		DisclosurePolicy: disclosurePolicy.ToMap(),
		Purpose:          purpose,
		CreatedAt:        now,
	})

	return p, nil
}

// NewPresentationFromEvents reconstructs a Presentation from events (for hydration).
func NewPresentationFromEvents(id string) *Presentation {
	p := &Presentation{}
	p.InitAggregate(AggregateTypePresentation, id)
	return p
}

// PresentationFactory creates a factory for Presentation aggregates.
func PresentationFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewPresentationFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the presentation's ID.
func (p *Presentation) ID() PresentationID {
	return p.id
}

// HolderDID returns the holder's DID.
func (p *Presentation) HolderDID() string {
	return p.holderDID
}

// CredentialIDs returns a copy of the credential IDs.
func (p *Presentation) CredentialIDs() []string {
	if p.credentialIDs == nil {
		return nil
	}
	copied := make([]string, len(p.credentialIDs))
	copy(copied, p.credentialIDs)
	return copied
}

// DisclosurePolicy returns the disclosure policy.
func (p *Presentation) DisclosurePolicy() DisclosurePolicy {
	return p.disclosurePolicy
}

// VPJWT returns the signed VP JWT representation.
func (p *Presentation) VPJWT() string {
	return p.vpJWT
}

// Purpose returns the purpose of the presentation.
func (p *Presentation) Purpose() string {
	return p.purpose
}

// Status returns the presentation's status.
func (p *Presentation) Status() PresentationStatus {
	return p.status
}

// CreatedAt returns when the presentation was created.
func (p *Presentation) CreatedAt() time.Time {
	return p.createdAt
}

// RevokedAt returns when the presentation was revoked.
func (p *Presentation) RevokedAt() time.Time {
	return p.revokedAt
}

// RevocationReason returns the reason for revocation.
func (p *Presentation) RevocationReason() string {
	return p.revocationReason
}

// UpdatedAt returns when the presentation was last updated.
func (p *Presentation) UpdatedAt() time.Time {
	return p.updatedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Revoke revokes an active presentation.
func (p *Presentation) Revoke(reason string) error {
	if !p.status.CanTransitionTo(PresentationStatusRevoked) {
		return PresentationRevoked("Presentation.Revoke", p.id.String()).
			WithMessage("cannot revoke presentation in current status: " + p.status.String())
	}

	p.Raise(p, &PresentationRevokedEvent{
		BaseEvent:      newPresentationBaseEvent(p.id),
		PresentationID: p.id.String(),
		Reason:         reason,
		RevokedAt:      time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (p *Presentation) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *PresentationCreatedEvent:
		p.onPresentationCreated(e)
	case *PresentationRevokedEvent:
		p.onPresentationRevoked(e)
	}
}

func (p *Presentation) onPresentationCreated(e *PresentationCreatedEvent) {
	var err error
	p.id, err = ParsePresentationID(e.PresentationID)
	if err != nil {
		panic("corrupt event store: PresentationCreated has invalid PresentationID: " + e.PresentationID)
	}

	p.holderDID = e.HolderDID

	p.credentialIDs = make([]string, len(e.CredentialIDs))
	copy(p.credentialIDs, e.CredentialIDs)

	p.disclosurePolicy, err = DisclosurePolicyFromMap(e.DisclosurePolicy)
	if err != nil {
		panic("corrupt event store: PresentationCreated has invalid DisclosurePolicy: " + err.Error())
	}

	p.vpJWT = e.VPJWT
	p.purpose = e.Purpose
	p.status = PresentationStatusActive
	p.createdAt = e.CreatedAt
	p.updatedAt = e.CreatedAt
}

func (p *Presentation) onPresentationRevoked(e *PresentationRevokedEvent) {
	p.status = PresentationStatusRevoked
	p.revokedAt = e.RevokedAt
	p.revocationReason = e.Reason
	p.updatedAt = e.RevokedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (p *Presentation) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &p.AggregateRoot
}
