package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Credential Aggregate
// ============================================================================

// Credential is the aggregate root for credential management.
type Credential struct {
	eventsourcing.AggregateRoot

	// Identity
	holderDID      string
	issuerDID      string
	credentialType string
	schemaID       string

	// State
	status Status
	claims map[string]any

	// Timestamps
	issuedAt  *time.Time
	expiresAt *time.Time

	// Lifecycle info
	revocationInfo *RevocationInfo
	suspensionInfo *SuspensionInfo
}

// NewCredential creates a new Credential aggregate.
func NewCredential(id string) *Credential {
	c := &Credential{}
	c.InitAggregate(AggregateType, id)
	return c
}

// GetAggregateRoot returns the embedded AggregateRoot.
// Required for eventsourcing.Hydrate to work.
func (c *Credential) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &c.AggregateRoot
}

// ============================================================================
// Accessors
// ============================================================================

// HolderDID returns the holder's DID.
func (c *Credential) HolderDID() string {
	return c.holderDID
}

// IssuerDID returns the issuer's DID.
func (c *Credential) IssuerDID() string {
	return c.issuerDID
}

// CredentialType returns the credential type.
func (c *Credential) CredentialType() string {
	return c.credentialType
}

// SchemaID returns the schema ID.
func (c *Credential) SchemaID() string {
	return c.schemaID
}

// Status returns the current status.
func (c *Credential) Status() Status {
	return c.status
}

// Claims returns the credential claims.
func (c *Credential) Claims() map[string]any {
	if c.claims == nil {
		return nil
	}
	// Return a copy to prevent external modification
	cp := make(map[string]any, len(c.claims))
	for k, v := range c.claims {
		cp[k] = v
	}
	return cp
}

// IssuedAt returns when the credential was issued.
func (c *Credential) IssuedAt() *time.Time {
	return c.issuedAt
}

// ExpiresAt returns when the credential expires.
func (c *Credential) ExpiresAt() *time.Time {
	return c.expiresAt
}

// RevocationInfo returns revocation details if revoked.
func (c *Credential) RevocationInfo() *RevocationInfo {
	return c.revocationInfo
}

// SuspensionInfo returns suspension details if suspended.
func (c *Credential) SuspensionInfo() *SuspensionInfo {
	return c.suspensionInfo
}

// IsActive returns true if the credential is currently active.
func (c *Credential) IsActive() bool {
	return c.status == StatusActive
}

// IsRevoked returns true if the credential has been revoked.
func (c *Credential) IsRevoked() bool {
	return c.status == StatusRevoked
}

// IsSuspended returns true if the credential is suspended.
func (c *Credential) IsSuspended() bool {
	return c.status == StatusSuspended
}

// IsExpired returns true if the credential has expired.
func (c *Credential) IsExpired() bool {
	if c.status == StatusExpired {
		return true
	}
	if c.expiresAt != nil && time.Now().UTC().After(*c.expiresAt) {
		return true
	}
	return false
}

// ============================================================================
// Commands (Domain Logic)
// ============================================================================

// Request initiates a credential request from a holder.
func (c *Credential) Request(holderDID, issuerDID, credentialType string, claims map[string]any) error {
	// Validate state
	if c.status != "" {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Request",
			"credential already exists",
		)
	}

	// Validate inputs
	if holderDID == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Request", "holder DID is required")
	}
	if issuerDID == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Request", "issuer DID is required")
	}
	if credentialType == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Request", "credential type is required")
	}

	// Raise event
	event := NewCredentialRequested(c.AggregateID(), holderDID, issuerDID, credentialType, claims)
	c.Raise(c, event)

	return nil
}

// Issue issues the credential.
func (c *Credential) Issue(issuerDID string, claims map[string]any, expiresAt *time.Time) error {
	// Validate state transitions
	if c.status != "" && c.status != StatusPending {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Issue",
			"credential cannot be issued in current state: "+string(c.status),
		)
	}

	// Validate inputs
	if issuerDID == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Issue", "issuer DID is required")
	}
	if len(claims) == 0 {
		return eventsourcing.ErrAggregateValidation("Credential.Issue", "claims are required")
	}

	// Determine holder DID and credential type
	holderDID := c.holderDID
	credentialType := c.credentialType

	// For direct issuance (no prior request), extract from claims
	if holderDID == "" {
		if hd, ok := claims["holder_did"].(string); ok {
			holderDID = hd
		} else {
			return eventsourcing.ErrAggregateValidation("Credential.Issue", "holder DID is required")
		}
	}
	if credentialType == "" {
		if ct, ok := claims["credential_type"].(string); ok {
			credentialType = ct
		} else {
			return eventsourcing.ErrAggregateValidation("Credential.Issue", "credential type is required")
		}
	}

	// Raise event
	event := NewCredentialIssued(c.AggregateID(), issuerDID, holderDID, credentialType, claims, expiresAt)
	c.Raise(c, event)

	return nil
}

// Revoke permanently revokes the credential.
func (c *Credential) Revoke(revokedBy, reason string) error {
	// Validate state
	if c.status == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Revoke", "credential does not exist")
	}
	if c.status == StatusRevoked {
		return eventsourcing.ErrAggregateValidation("Credential.Revoke", "credential is already revoked")
	}
	if !c.status.CanTransitionTo(StatusRevoked) {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Revoke",
			"cannot revoke credential in current state: "+string(c.status),
		)
	}

	// Validate inputs
	if revokedBy == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Revoke", "revoked_by is required")
	}
	if reason == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Revoke", "reason is required")
	}

	// Raise event
	event := NewCredentialRevoked(c.AggregateID(), revokedBy, reason)
	c.Raise(c, event)

	return nil
}

// Suspend temporarily suspends the credential.
func (c *Credential) Suspend(suspendedBy, reason string, until *time.Time) error {
	// Validate state
	if c.status != StatusActive {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Suspend",
			"only active credentials can be suspended",
		)
	}

	// Validate inputs
	if suspendedBy == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Suspend", "suspended_by is required")
	}
	if reason == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Suspend", "reason is required")
	}

	// Raise event
	event := NewCredentialSuspended(c.AggregateID(), suspendedBy, reason, until)
	c.Raise(c, event)

	return nil
}

// Reinstate reinstates a suspended credential.
func (c *Credential) Reinstate(reinstatedBy, reason string) error {
	// Validate state
	if c.status != StatusSuspended {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Reinstate",
			"only suspended credentials can be reinstated",
		)
	}

	// Validate inputs
	if reinstatedBy == "" {
		return eventsourcing.ErrAggregateValidation("Credential.Reinstate", "reinstated_by is required")
	}

	// Raise event
	event := NewCredentialReinstated(c.AggregateID(), reinstatedBy, reason)
	c.Raise(c, event)

	return nil
}

// Expire marks the credential as expired.
func (c *Credential) Expire() error {
	// Validate state
	if c.status != StatusActive {
		return eventsourcing.ErrAggregateValidation(
			"Credential.Expire",
			"only active credentials can expire",
		)
	}

	// Raise event
	event := NewCredentialExpired(c.AggregateID())
	c.Raise(c, event)

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update aggregate state.
func (c *Credential) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *CredentialRequested:
		c.applyCredentialRequested(e)
	case *CredentialIssued:
		c.applyCredentialIssued(e)
	case *CredentialRevoked:
		c.applyCredentialRevoked(e)
	case *CredentialSuspended:
		c.applyCredentialSuspended(e)
	case *CredentialReinstated:
		c.applyCredentialReinstated(e)
	case *CredentialExpired:
		c.applyCredentialExpired(e)
	}
}

func (c *Credential) applyCredentialRequested(e *CredentialRequested) {
	c.holderDID = e.HolderDID
	c.issuerDID = e.IssuerDID
	c.credentialType = e.CredentialType
	c.claims = e.Claims
	c.status = StatusPending
}

func (c *Credential) applyCredentialIssued(e *CredentialIssued) {
	c.issuerDID = e.IssuerDID
	c.holderDID = e.HolderDID
	c.credentialType = e.CredentialType
	c.schemaID = e.SchemaID
	c.claims = e.Claims
	c.issuedAt = &e.IssuedAt
	c.expiresAt = e.ExpiresAt
	c.status = StatusActive
}

func (c *Credential) applyCredentialRevoked(e *CredentialRevoked) {
	c.status = StatusRevoked
	c.revocationInfo = &RevocationInfo{
		RevokedBy: e.RevokedBy,
		Reason:    e.Reason,
		RevokedAt: e.RevokedAt,
	}
	c.suspensionInfo = nil // Clear any suspension
}

func (c *Credential) applyCredentialSuspended(e *CredentialSuspended) {
	c.status = StatusSuspended
	c.suspensionInfo = &SuspensionInfo{
		SuspendedBy: e.SuspendedBy,
		Reason:      e.Reason,
		SuspendedAt: e.SuspendedAt,
		Until:       e.Until,
	}
}

func (c *Credential) applyCredentialReinstated(e *CredentialReinstated) {
	c.status = StatusActive
	c.suspensionInfo = nil
}

func (c *Credential) applyCredentialExpired(e *CredentialExpired) {
	c.status = StatusExpired
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ eventsourcing.Aggregate = (*Credential)(nil)
