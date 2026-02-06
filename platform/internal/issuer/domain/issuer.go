package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Issuer is the aggregate root for B2B credential issuers.
// It manages the lifecycle of an organization's issuer profile including
// registration, activation, suspension, and branding.
type Issuer struct {
	eventsourcing.AggregateRoot

	id               IssuerID
	organizationID   string
	name             string
	description      string
	did              string
	webhookURL       string
	apiKeyHash       string
	status           IssuerStatus
	branding         IssuerBranding
	createdAt        time.Time
	activatedAt      time.Time
	suspendedAt      time.Time
	suspensionReason string
	updatedAt        time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// RegisterIssuer creates a new Issuer aggregate for registration.
func RegisterIssuer(id IssuerID, orgID, name, description, webhookURL string) (*Issuer, error) {
	if id.IsZero() {
		return nil, IssuerInvalid("Issuer.Register", "issuer ID is required")
	}

	if orgID == "" {
		return nil, IssuerInvalid("Issuer.Register", "organization ID is required")
	}

	if name == "" {
		return nil, IssuerInvalid("Issuer.Register", "name is required")
	}

	iss := &Issuer{}
	iss.InitAggregate(AggregateTypeIssuer, id.String())

	now := time.Now().UTC()

	iss.Raise(iss, &IssuerRegisteredEvent{
		BaseEvent:      newIssuerBaseEvent(id),
		IssuerID:       id.String(),
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		WebhookURL:     webhookURL,
		CreatedAt:      now,
	})

	return iss, nil
}

// NewIssuerFromEvents reconstructs an Issuer from events (for hydration).
func NewIssuerFromEvents(id string) *Issuer {
	iss := &Issuer{}
	iss.InitAggregate(AggregateTypeIssuer, id)
	return iss
}

// IssuerFactory creates a factory for Issuer aggregates.
func IssuerFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewIssuerFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the issuer's ID.
func (iss *Issuer) ID() IssuerID {
	return iss.id
}

// OrganizationID returns the issuer's organization ID.
func (iss *Issuer) OrganizationID() string {
	return iss.organizationID
}

// Name returns the issuer's name.
func (iss *Issuer) Name() string {
	return iss.name
}

// Description returns the issuer's description.
func (iss *Issuer) Description() string {
	return iss.description
}

// DID returns the issuer's DID.
func (iss *Issuer) DID() string {
	return iss.did
}

// WebhookURL returns the issuer's webhook URL.
func (iss *Issuer) WebhookURL() string {
	return iss.webhookURL
}

// APIKeyHash returns the issuer's API key hash.
func (iss *Issuer) APIKeyHash() string {
	return iss.apiKeyHash
}

// Status returns the issuer's status.
func (iss *Issuer) Status() IssuerStatus {
	return iss.status
}

// Branding returns the issuer's branding.
func (iss *Issuer) Branding() IssuerBranding {
	return iss.branding
}

// CreatedAt returns when the issuer was created.
func (iss *Issuer) CreatedAt() time.Time {
	return iss.createdAt
}

// ActivatedAt returns when the issuer was activated.
func (iss *Issuer) ActivatedAt() time.Time {
	return iss.activatedAt
}

// SuspendedAt returns when the issuer was suspended.
func (iss *Issuer) SuspendedAt() time.Time {
	return iss.suspendedAt
}

// SuspensionReason returns the reason the issuer was suspended.
func (iss *Issuer) SuspensionReason() string {
	return iss.suspensionReason
}

// UpdatedAt returns when the issuer was last updated.
func (iss *Issuer) UpdatedAt() time.Time {
	return iss.updatedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Activate activates a pending issuer with a DID, API key hash, and branding.
func (iss *Issuer) Activate(did, apiKeyHash string, branding IssuerBranding) error {
	if !iss.status.CanTransitionTo(IssuerStatusActive) {
		return IssuerAlreadyActive("Issuer.Activate", iss.id.String()).
			WithMessage("cannot activate issuer in current status: " + iss.status.String())
	}

	if did == "" {
		return IssuerInvalid("Issuer.Activate", "DID is required for activation")
	}

	if apiKeyHash == "" {
		return IssuerInvalid("Issuer.Activate", "API key hash is required for activation")
	}

	now := time.Now().UTC()

	iss.Raise(iss, &IssuerActivatedEvent{
		BaseEvent:   newIssuerBaseEvent(iss.id),
		IssuerID:    iss.id.String(),
		DID:         did,
		APIKeyHash:  apiKeyHash,
		Branding:    branding.ToMap(),
		ActivatedAt: now,
	})

	return nil
}

// Suspend suspends an active issuer.
func (iss *Issuer) Suspend(reason string) error {
	if !iss.status.CanTransitionTo(IssuerStatusSuspended) {
		return IssuerNotActive("Issuer.Suspend", iss.id.String()).
			WithMessage("cannot suspend issuer in current status: " + iss.status.String())
	}

	if reason == "" {
		return IssuerInvalid("Issuer.Suspend", "suspension reason is required")
	}

	now := time.Now().UTC()

	iss.Raise(iss, &IssuerSuspendedEvent{
		BaseEvent:   newIssuerBaseEvent(iss.id),
		IssuerID:    iss.id.String(),
		Reason:      reason,
		SuspendedAt: now,
	})

	return nil
}

// UpdateBranding updates the issuer's branding.
func (iss *Issuer) UpdateBranding(branding IssuerBranding) error {
	if !iss.status.IsActive() {
		return IssuerNotActive("Issuer.UpdateBranding", iss.id.String()).
			WithMessage("cannot update branding for issuer in current status: " + iss.status.String())
	}

	now := time.Now().UTC()

	iss.Raise(iss, &BrandingUpdatedEvent{
		BaseEvent: newIssuerBaseEvent(iss.id),
		IssuerID:  iss.id.String(),
		Branding:  branding.ToMap(),
		UpdatedAt: now,
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (iss *Issuer) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *IssuerRegisteredEvent:
		iss.onIssuerRegistered(e)
	case *IssuerActivatedEvent:
		iss.onIssuerActivated(e)
	case *IssuerSuspendedEvent:
		iss.onIssuerSuspended(e)
	case *BrandingUpdatedEvent:
		iss.onBrandingUpdated(e)
	}
}

func (iss *Issuer) onIssuerRegistered(e *IssuerRegisteredEvent) {
	var err error
	iss.id, err = ParseIssuerID(e.IssuerID)
	if err != nil {
		panic("corrupt event store: IssuerRegistered has invalid IssuerID: " + e.IssuerID)
	}
	iss.organizationID = e.OrganizationID
	iss.name = e.Name
	iss.description = e.Description
	iss.webhookURL = e.WebhookURL
	iss.status = IssuerStatusPending
	iss.createdAt = e.CreatedAt
	iss.updatedAt = e.CreatedAt
}

func (iss *Issuer) onIssuerActivated(e *IssuerActivatedEvent) {
	iss.did = e.DID
	iss.apiKeyHash = e.APIKeyHash
	iss.branding = BrandingFromMap(e.Branding)
	iss.status = IssuerStatusActive
	iss.activatedAt = e.ActivatedAt
	iss.updatedAt = e.ActivatedAt
}

func (iss *Issuer) onIssuerSuspended(e *IssuerSuspendedEvent) {
	iss.status = IssuerStatusSuspended
	iss.suspendedAt = e.SuspendedAt
	iss.suspensionReason = e.Reason
	iss.updatedAt = e.SuspendedAt
}

func (iss *Issuer) onBrandingUpdated(e *BrandingUpdatedEvent) {
	iss.branding = BrandingFromMap(e.Branding)
	iss.updatedAt = e.UpdatedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (iss *Issuer) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &iss.AggregateRoot
}
