package aggregate

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/event"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Credential Status
// ============================================================================

// Status represents the current status of a credential.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusRevoked   Status = "revoked"
	StatusExpired   Status = "expired"
)

// IsValid returns true if the credential can be used for verification.
func (s Status) IsValid() bool {
	return s == StatusActive
}

// IsTerminal returns true if the credential cannot transition to another state.
func (s Status) IsTerminal() bool {
	return s == StatusRevoked || s == StatusExpired
}

// ============================================================================
// Credential Aggregate
// ============================================================================

// Credential is the event-sourced aggregate for verifiable credentials.
type Credential struct {
	eventsourcing.AggregateRoot

	// Identity
	credentialID   string
	credentialType string
	schemaID       string

	// Parties
	holderDID string
	issuerDID string

	// Claims
	claims map[string]any

	// Status
	status         Status
	suspendedUntil *time.Time
	suspensionInfo *SuspensionInfo
	revocationInfo *RevocationInfo

	// Timestamps
	requestedAt *time.Time
	issuedAt    *time.Time
	expiresAt   *time.Time

	// Credential data
	credentialJWT string
	evidence      []event.Evidence
}

// SuspensionInfo contains details about a credential suspension.
type SuspensionInfo struct {
	SuspendedBy string
	Reason      string
	SuspendedAt time.Time
}

// RevocationInfo contains details about a credential revocation.
type RevocationInfo struct {
	RevokedBy string
	Reason    string
	RevokedAt time.Time
}

// ============================================================================
// Factory
// ============================================================================

// NewCredential creates a new Credential aggregate.
func NewCredential(id string) *Credential {
	c := &Credential{
		claims: make(map[string]any),
	}
	c.InitAggregate(event.AggregateType, id)
	return c
}

// CredentialFactory creates Credential aggregates.
type CredentialFactory struct{}

// Create creates a new Credential aggregate with the given ID.
func (f *CredentialFactory) Create(id string) eventsourcing.Aggregate {
	return NewCredential(id)
}

// NewCredentialFactory creates a new CredentialFactory.
func NewCredentialFactory() *CredentialFactory {
	return &CredentialFactory{}
}

// ============================================================================
// Commands (Business Operations)
// ============================================================================

// Request initiates a credential request from a holder to an issuer.
func (c *Credential) Request(holderDID, issuerDID, credentialType string, claims map[string]any) error {
	if c.status != "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Request",
			"credential already exists",
		)
	}

	if holderDID == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Request",
			"holder DID is required",
		)
	}

	if issuerDID == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Request",
			"issuer DID is required",
		)
	}

	if credentialType == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Request",
			"credential type is required",
		)
	}

	e := event.NewCredentialRequested(
		c.AggregateID(),
		holderDID,
		issuerDID,
		credentialType,
		claims,
	)

	c.Raise(c, e)
	return nil
}

// Issue issues the credential to the holder.
func (c *Credential) Issue(issuerDID string, claims map[string]any, expiresAt *time.Time) error {
	// Can issue from pending state or directly (skip request)
	if c.status != "" && c.status != StatusPending {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Issue",
			"credential cannot be issued in current state: "+string(c.status),
		)
	}

	if issuerDID == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Issue",
			"issuer DID is required",
		)
	}

	if len(claims) == 0 {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Issue",
			"claims are required",
		)
	}

	// Use existing holder/type if from request, otherwise require them in claims
	holderDID := c.holderDID
	credentialType := c.credentialType

	if holderDID == "" {
		if h, ok := claims["holder_did"].(string); ok {
			holderDID = h
			delete(claims, "holder_did")
		} else {
			return eventsourcing.ErrAggregateValidation(
				"Credential.Issue",
				"holder DID is required",
			)
		}
	}

	if credentialType == "" {
		if t, ok := claims["credential_type"].(string); ok {
			credentialType = t
			delete(claims, "credential_type")
		} else {
			return eventsourcing.ErrAggregateValidation(
				"Credential.Issue",
				"credential type is required",
			)
		}
	}

	e := event.NewCredentialIssued(
		c.AggregateID(),
		holderDID,
		issuerDID,
		credentialType,
		claims,
		expiresAt,
	)

	c.Raise(c, e)
	return nil
}

// IssueWithJWT issues the credential with a pre-signed JWT.
func (c *Credential) IssueWithJWT(issuerDID string, claims map[string]any, expiresAt *time.Time, jwt string) error {
	if err := c.Issue(issuerDID, claims, expiresAt); err != nil {
		return err
	}

	// Get the last event and add JWT
	changes := c.Changes()
	if len(changes) > 0 {
		if issued, ok := changes[len(changes)-1].(*event.CredentialIssued); ok {
			issued.WithCredentialJWT(jwt)
		}
	}

	return nil
}

// Revoke permanently revokes the credential.
func (c *Credential) Revoke(revokedBy, reason string) error {
	if c.status == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Revoke",
			"credential does not exist",
		)
	}

	if c.status == StatusRevoked {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Revoke",
			"credential is already revoked",
		)
	}

	if c.status == StatusExpired {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Revoke",
			"cannot revoke an expired credential",
		)
	}

	if revokedBy == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Revoke",
			"revokedBy is required",
		)
	}

	if reason == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Revoke",
			"reason is required",
		)
	}

	e := event.NewCredentialRevoked(c.AggregateID(), revokedBy, reason)
	c.Raise(c, e)
	return nil
}

// Suspend temporarily suspends the credential.
func (c *Credential) Suspend(suspendedBy, reason string, until *time.Time) error {
	if c.status != StatusActive {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Suspend",
			"only active credentials can be suspended",
		)
	}

	if suspendedBy == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Suspend",
			"suspendedBy is required",
		)
	}

	if reason == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Suspend",
			"reason is required",
		)
	}

	e := event.NewCredentialSuspended(c.AggregateID(), suspendedBy, reason, until)
	c.Raise(c, e)
	return nil
}

// Reinstate reinstates a suspended credential.
func (c *Credential) Reinstate(reinstatedBy, reason string) error {
	if c.status != StatusSuspended {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Reinstate",
			"only suspended credentials can be reinstated",
		)
	}

	if reinstatedBy == "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Reinstate",
			"reinstatedBy is required",
		)
	}

	e := event.NewCredentialReinstated(c.AggregateID(), reinstatedBy, reason)
	c.Raise(c, e)
	return nil
}

// Expire marks the credential as expired.
func (c *Credential) Expire() error {
	if c.status != StatusActive && c.status != StatusSuspended {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Expire",
			"credential cannot expire in current state: "+string(c.status),
		)
	}

	e := event.NewCredentialExpired(c.AggregateID(), time.Now().UTC())
	c.Raise(c, e)
	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update aggregate state.
func (c *Credential) ApplyEvent(e eventsourcing.Event) {
	switch evt := e.(type) {
	case *event.CredentialRequested:
		c.applyCredentialRequested(evt)
	case *event.CredentialIssued:
		c.applyCredentialIssued(evt)
	case *event.CredentialRevoked:
		c.applyCredentialRevoked(evt)
	case *event.CredentialSuspended:
		c.applyCredentialSuspended(evt)
	case *event.CredentialReinstated:
		c.applyCredentialReinstated(evt)
	case *event.CredentialExpired:
		c.applyCredentialExpired(evt)
	}
}

func (c *Credential) applyCredentialRequested(e *event.CredentialRequested) {
	c.credentialID = e.CredentialID
	c.holderDID = e.HolderDID
	c.issuerDID = e.IssuerDID
	c.credentialType = e.CredentialType
	c.claims = e.Claims
	c.status = StatusPending
	c.requestedAt = &e.RequestedAt
}

func (c *Credential) applyCredentialIssued(e *event.CredentialIssued) {
	c.credentialID = e.CredentialID
	c.holderDID = e.HolderDID
	c.issuerDID = e.IssuerDID
	c.credentialType = e.CredentialType
	c.claims = e.Claims
	c.status = StatusActive
	c.issuedAt = &e.IssuedAt
	c.expiresAt = e.ExpiresAt
	c.credentialJWT = e.CredentialJWT
	c.schemaID = e.SchemaID
	c.evidence = e.Evidence
}

func (c *Credential) applyCredentialRevoked(e *event.CredentialRevoked) {
	c.status = StatusRevoked
	c.revocationInfo = &RevocationInfo{
		RevokedBy: e.RevokedBy,
		Reason:    e.Reason,
		RevokedAt: e.RevokedAt,
	}
}

func (c *Credential) applyCredentialSuspended(e *event.CredentialSuspended) {
	c.status = StatusSuspended
	c.suspendedUntil = e.SuspendedUntil
	c.suspensionInfo = &SuspensionInfo{
		SuspendedBy: e.SuspendedBy,
		Reason:      e.Reason,
		SuspendedAt: e.SuspendedAt,
	}
}

func (c *Credential) applyCredentialReinstated(e *event.CredentialReinstated) {
	c.status = StatusActive
	c.suspendedUntil = nil
	c.suspensionInfo = nil
}

func (c *Credential) applyCredentialExpired(e *event.CredentialExpired) {
	c.status = StatusExpired
}

// ============================================================================
// Getters (Read-Only Access)
// ============================================================================

// GetAggregateRoot returns the aggregate root.
func (c *Credential) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &c.AggregateRoot
}

// CredentialID returns the credential ID.
func (c *Credential) CredentialID() string {
	return c.credentialID
}

// CredentialType returns the credential type.
func (c *Credential) CredentialType() string {
	return c.credentialType
}

// SchemaID returns the schema ID.
func (c *Credential) SchemaID() string {
	return c.schemaID
}

// HolderDID returns the holder DID.
func (c *Credential) HolderDID() string {
	return c.holderDID
}

// IssuerDID returns the issuer DID.
func (c *Credential) IssuerDID() string {
	return c.issuerDID
}

// Claims returns a copy of the claims.
func (c *Credential) Claims() map[string]any {
	if c.claims == nil {
		return nil
	}
	copy := make(map[string]any, len(c.claims))
	for k, v := range c.claims {
		copy[k] = v
	}
	return copy
}

// Status returns the current status.
func (c *Credential) Status() Status {
	return c.status
}

// IsActive returns true if the credential is active.
func (c *Credential) IsActive() bool {
	return c.status == StatusActive
}

// IsSuspended returns true if the credential is suspended.
func (c *Credential) IsSuspended() bool {
	return c.status == StatusSuspended
}

// IsRevoked returns true if the credential is revoked.
func (c *Credential) IsRevoked() bool {
	return c.status == StatusRevoked
}

// IsExpired returns true if the credential has expired.
func (c *Credential) IsExpired() bool {
	if c.status == StatusExpired {
		return true
	}
	if c.expiresAt != nil && time.Now().After(*c.expiresAt) {
		return true
	}
	return false
}

// IsValid returns true if the credential can be used for verification.
func (c *Credential) IsValid() bool {
	return c.IsActive() && !c.IsExpired()
}

// RequestedAt returns when the credential was requested.
func (c *Credential) RequestedAt() *time.Time {
	return c.requestedAt
}

// IssuedAt returns when the credential was issued.
func (c *Credential) IssuedAt() *time.Time {
	return c.issuedAt
}

// ExpiresAt returns when the credential expires.
func (c *Credential) ExpiresAt() *time.Time {
	return c.expiresAt
}

// CredentialJWT returns the JWT representation.
func (c *Credential) CredentialJWT() string {
	return c.credentialJWT
}

// Evidence returns the credential evidence.
func (c *Credential) Evidence() []event.Evidence {
	return c.evidence
}

// SuspensionInfo returns suspension details if suspended.
func (c *Credential) SuspensionInfo() *SuspensionInfo {
	return c.suspensionInfo
}

// RevocationInfo returns revocation details if revoked.
func (c *Credential) RevocationInfo() *RevocationInfo {
	return c.revocationInfo
}

// ============================================================================
// Validation
// ============================================================================

// Validate validates the aggregate state.
func (c *Credential) Validate() error {
	if c.credentialID == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Validate", "credential ID is required")
	}

	if c.status == StatusActive || c.status == StatusSuspended || c.status == StatusRevoked || c.status == StatusExpired {
		if c.holderDID == "" {
			return eventsourcing.ErrAggregateValidation("Credential.Validate", "holder DID is required")
		}
		if c.issuerDID == "" {
			return eventsourcing.ErrAggregateValidation("Credential.Validate", "issuer DID is required")
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ eventsourcing.Aggregate            = (*Credential)(nil)
	_ eventsourcing.ValidatableAggregate = (*Credential)(nil)
	_ eventsourcing.AggregateFactory     = (*CredentialFactory)(nil)
)
