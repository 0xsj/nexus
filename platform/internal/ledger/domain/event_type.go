package domain

import (
	"errors"
	"strings"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// EventType Error Codes
// ============================================================================

const (
	CodeInvalidEventType pkgerrors.Code = "LEDGER_INVALID_EVENT_TYPE"
)

// ============================================================================
// EventType Sentinel Errors
// ============================================================================

var (
	ErrInvalidEventType = errors.New("invalid event type")
)

// ============================================================================
// EventType Error Constructors
// ============================================================================

// InvalidEventType creates an invalid event type error.
func InvalidEventType(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid event type: "+reason).
		WithCode(CodeInvalidEventType).
		WithMeta("value", value)
}

// ============================================================================
// EventType Value Object
// ============================================================================

// EventType represents the type of domain event that occurred.
// Format: "context.event_name" (e.g., "credential.issued", "user.registered")
type EventType struct {
	value string
}

// Well-known event type constants organized by context.
const (
	// Identity events
	EventTypeUserRegistered = "identity.user_registered"
	EventTypeUserDeleted    = "identity.user_deleted"
	EventTypeSessionCreated = "identity.session_created"
	EventTypeSessionRevoked = "identity.session_revoked"
	EventTypeDIDLinked      = "identity.did_linked"
	EventTypeDIDUnlinked    = "identity.did_unlinked"
	EventTypeAPIKeyCreated  = "identity.api_key_created"
	EventTypeAPIKeyRevoked  = "identity.api_key_revoked"

	// Wallet events
	EventTypeWalletLinked      = "wallet.linked"
	EventTypeWalletUnlinked    = "wallet.unlinked"
	EventTypeSignatureVerified = "wallet.signature_verified"

	// Verification events
	EventTypeVerificationInitiated = "verification.initiated"
	EventTypeVerificationCompleted = "verification.completed"
	EventTypeVerificationFailed    = "verification.failed"
	EventTypeOAuthTokensReceived   = "verification.oauth_tokens_received"
	EventTypeOAuthTokensRevoked    = "verification.oauth_tokens_revoked"
	EventTypeProviderDataFetched   = "verification.provider_data_fetched"

	// Credential events
	EventTypeCredentialIssued  = "credential.issued"
	EventTypeCredentialRevoked = "credential.revoked"
	EventTypeCredentialExpired = "credential.expired"
	EventTypeCredentialUpdated = "credential.updated"

	// Presentation events
	EventTypePresentationCreated  = "presentation.created"
	EventTypePresentationShared   = "presentation.shared"
	EventTypePresentationVerified = "presentation.verified"
	EventTypeShareLinkCreated     = "presentation.share_link_created"
	EventTypeShareLinkAccessed    = "presentation.share_link_accessed"
	EventTypeShareLinkRevoked     = "presentation.share_link_revoked"

	// Trust events
	EventTypeVouchGiven        = "trust.vouch_given"
	EventTypeVouchReceived     = "trust.vouch_received"
	EventTypeVouchRevoked      = "trust.vouch_revoked"
	EventTypeReputationUpdated = "trust.reputation_updated"

	// Organization events
	EventTypeMemberAdded      = "organization.member_added"
	EventTypeMemberRemoved    = "organization.member_removed"
	EventTypeRoleChanged      = "organization.role_changed"
	EventTypeIssuerRegistered = "organization.issuer_registered"
	EventTypeIssuerSuspended  = "organization.issuer_suspended"
)

// NewEventType creates a new EventType from a string.
// Validates format: must contain a dot separator between context and event name.
func NewEventType(value string) (EventType, error) {
	if value == "" {
		return EventType{}, InvalidEventType("domain.NewEventType", value, "cannot be empty")
	}

	if !strings.Contains(value, ".") {
		return EventType{}, InvalidEventType("domain.NewEventType", value, "must be in format 'context.event_name'")
	}

	parts := strings.SplitN(value, ".", 2)
	if parts[0] == "" || parts[1] == "" {
		return EventType{}, InvalidEventType("domain.NewEventType", value, "context and event name cannot be empty")
	}

	return EventType{value: value}, nil
}

// MustNewEventType creates a new EventType and panics if invalid.
// Only use for constants or tests.
func MustNewEventType(value string) EventType {
	et, err := NewEventType(value)
	if err != nil {
		panic(err)
	}
	return et
}

// String returns the string representation.
func (t EventType) String() string {
	return t.value
}

// IsZero returns true if the EventType is the zero value.
func (t EventType) IsZero() bool {
	return t.value == ""
}

// IsValid returns true if the EventType is valid (non-empty with proper format).
func (t EventType) IsValid() bool {
	if t.value == "" {
		return false
	}
	return strings.Contains(t.value, ".")
}

// Equals checks if two EventTypes are equal.
func (t EventType) Equals(other EventType) bool {
	return t.value == other.value
}

// Context returns the context portion of the event type.
// For "credential.issued", returns "credential".
func (t EventType) Context() string {
	if t.value == "" {
		return ""
	}
	parts := strings.SplitN(t.value, ".", 2)
	return parts[0]
}

// Name returns the event name portion of the event type.
// For "credential.issued", returns "issued".
func (t EventType) Name() string {
	if t.value == "" {
		return ""
	}
	parts := strings.SplitN(t.value, ".", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// Matches checks if the EventType matches a pattern.
// Supports wildcard "*" for context (e.g., "credential.*" matches all credential events).
func (t EventType) Matches(pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return t.Context() == prefix
	}

	return t.value == pattern
}
