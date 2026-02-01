package domain

import (
	"testing"
)

// ============================================================================
// NewEventType Tests
// ============================================================================

func TestNewEventType_Valid(t *testing.T) {
	tests := []string{
		"credential.issued",
		"user.registered",
		"wallet.linked",
		"identity.session_created",
		"organization.member_added",
		"a.b",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			et, err := NewEventType(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if et.String() != input {
				t.Errorf("EventType.String() = %v, want %v", et.String(), input)
			}
			if et.IsZero() {
				t.Error("expected non-zero EventType")
			}
			if !et.IsValid() {
				t.Error("expected valid EventType")
			}
		})
	}
}

func TestNewEventType_Empty(t *testing.T) {
	_, err := NewEventType("")

	if err == nil {
		t.Error("expected error for empty EventType")
	}
}

func TestNewEventType_NoDot(t *testing.T) {
	invalidTypes := []string{
		"credential",
		"issued",
		"nodothere",
	}

	for _, input := range invalidTypes {
		t.Run(input, func(t *testing.T) {
			_, err := NewEventType(input)
			if err == nil {
				t.Errorf("expected error for EventType without dot: %q", input)
			}
		})
	}
}

func TestNewEventType_EmptyParts(t *testing.T) {
	invalidTypes := []string{
		".issued",
		"credential.",
		".",
	}

	for _, input := range invalidTypes {
		t.Run(input, func(t *testing.T) {
			_, err := NewEventType(input)
			if err == nil {
				t.Errorf("expected error for EventType with empty parts: %q", input)
			}
		})
	}
}

// ============================================================================
// MustNewEventType Tests
// ============================================================================

func TestMustNewEventType_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	et := MustNewEventType("credential.issued")
	if et.String() != "credential.issued" {
		t.Errorf("unexpected EventType: %v", et.String())
	}
}

func TestMustNewEventType_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid EventType")
		}
	}()

	MustNewEventType("invalid")
}

// ============================================================================
// IsZero / IsValid Tests
// ============================================================================

func TestEventType_IsZero(t *testing.T) {
	var zeroET EventType

	if !zeroET.IsZero() {
		t.Error("expected zero value to be zero")
	}
	if zeroET.IsValid() {
		t.Error("expected zero value to be invalid")
	}
}

func TestEventType_IsValid(t *testing.T) {
	et, _ := NewEventType("credential.issued")

	if et.IsZero() {
		t.Error("expected non-zero EventType")
	}
	if !et.IsValid() {
		t.Error("expected valid EventType")
	}
}

// ============================================================================
// Equals Tests
// ============================================================================

func TestEventType_Equals(t *testing.T) {
	et1, _ := NewEventType("credential.issued")
	et2, _ := NewEventType("credential.issued")
	et3, _ := NewEventType("credential.revoked")

	if !et1.Equals(et2) {
		t.Error("expected equal EventTypes to be equal")
	}
	if et1.Equals(et3) {
		t.Error("expected different EventTypes to not be equal")
	}
}

func TestEventType_Equals_ZeroValues(t *testing.T) {
	var zero1, zero2 EventType

	if !zero1.Equals(zero2) {
		t.Error("expected zero values to be equal")
	}
}

// ============================================================================
// Context Tests
// ============================================================================

func TestEventType_Context(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"credential.issued", "credential"},
		{"user.registered", "user"},
		{"identity.session_created", "identity"},
		{"organization.member_added", "organization"},
		{"a.b", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			et, _ := NewEventType(tt.input)
			got := et.Context()
			if got != tt.want {
				t.Errorf("EventType(%q).Context() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEventType_Context_ZeroValue(t *testing.T) {
	var et EventType

	if et.Context() != "" {
		t.Errorf("expected empty context for zero EventType, got %v", et.Context())
	}
}

// ============================================================================
// Name Tests
// ============================================================================

func TestEventType_Name(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"credential.issued", "issued"},
		{"user.registered", "registered"},
		{"identity.session_created", "session_created"},
		{"organization.member_added", "member_added"},
		{"a.b", "b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			et, _ := NewEventType(tt.input)
			got := et.Name()
			if got != tt.want {
				t.Errorf("EventType(%q).Name() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEventType_Name_ZeroValue(t *testing.T) {
	var et EventType

	if et.Name() != "" {
		t.Errorf("expected empty name for zero EventType, got %v", et.Name())
	}
}

func TestEventType_Name_MultipleDots(t *testing.T) {
	et, _ := NewEventType("context.event.with.dots")

	// Should return everything after the first dot
	if et.Name() != "event.with.dots" {
		t.Errorf("EventType.Name() = %v, want %v", et.Name(), "event.with.dots")
	}
}

// ============================================================================
// Matches Tests
// ============================================================================

func TestEventType_Matches_Exact(t *testing.T) {
	et, _ := NewEventType("credential.issued")

	if !et.Matches("credential.issued") {
		t.Error("expected exact match to succeed")
	}
	if et.Matches("credential.revoked") {
		t.Error("expected different event to not match")
	}
}

func TestEventType_Matches_Wildcard(t *testing.T) {
	et, _ := NewEventType("credential.issued")

	if !et.Matches("*") {
		t.Error("expected '*' wildcard to match everything")
	}
}

func TestEventType_Matches_ContextWildcard(t *testing.T) {
	tests := []struct {
		eventType string
		pattern   string
		want      bool
	}{
		{"credential.issued", "credential.*", true},
		{"credential.revoked", "credential.*", true},
		{"credential.updated", "credential.*", true},
		{"user.registered", "credential.*", false},
		{"user.registered", "user.*", true},
		{"identity.session_created", "identity.*", true},
	}

	for _, tt := range tests {
		t.Run(tt.eventType+"_"+tt.pattern, func(t *testing.T) {
			et, _ := NewEventType(tt.eventType)
			got := et.Matches(tt.pattern)
			if got != tt.want {
				t.Errorf("EventType(%q).Matches(%q) = %v, want %v", tt.eventType, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestEventType_Matches_NoMatch(t *testing.T) {
	et, _ := NewEventType("credential.issued")

	if et.Matches("user.registered") {
		t.Error("expected no match for different event type")
	}
	if et.Matches("user.*") {
		t.Error("expected no match for different context wildcard")
	}
}

// ============================================================================
// Constants Tests
// ============================================================================

func TestEventTypeConstants_Valid(t *testing.T) {
	constants := []string{
		EventTypeUserRegistered,
		EventTypeUserDeleted,
		EventTypeSessionCreated,
		EventTypeSessionRevoked,
		EventTypeDIDLinked,
		EventTypeDIDUnlinked,
		EventTypeAPIKeyCreated,
		EventTypeAPIKeyRevoked,
		EventTypeWalletLinked,
		EventTypeWalletUnlinked,
		EventTypeSignatureVerified,
		EventTypeVerificationInitiated,
		EventTypeVerificationCompleted,
		EventTypeVerificationFailed,
		EventTypeOAuthTokensReceived,
		EventTypeOAuthTokensRevoked,
		EventTypeProviderDataFetched,
		EventTypeCredentialIssued,
		EventTypeCredentialRevoked,
		EventTypeCredentialExpired,
		EventTypeCredentialUpdated,
		EventTypePresentationCreated,
		EventTypePresentationShared,
		EventTypePresentationVerified,
		EventTypeShareLinkCreated,
		EventTypeShareLinkAccessed,
		EventTypeShareLinkRevoked,
		EventTypeVouchGiven,
		EventTypeVouchReceived,
		EventTypeVouchRevoked,
		EventTypeReputationUpdated,
		EventTypeMemberAdded,
		EventTypeMemberRemoved,
		EventTypeRoleChanged,
		EventTypeIssuerRegistered,
		EventTypeIssuerSuspended,
	}

	for _, c := range constants {
		t.Run(c, func(t *testing.T) {
			et, err := NewEventType(c)
			if err != nil {
				t.Errorf("constant %q is not a valid EventType: %v", c, err)
			}
			if et.Context() == "" {
				t.Errorf("constant %q has empty context", c)
			}
			if et.Name() == "" {
				t.Errorf("constant %q has empty name", c)
			}
		})
	}
}

// ============================================================================
// Roundtrip Tests
// ============================================================================

func TestEventType_Roundtrip(t *testing.T) {
	original, _ := NewEventType("credential.issued")

	// Create from string representation
	str := original.String()
	parsed, err := NewEventType(str)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !parsed.Equals(original) {
		t.Errorf("roundtrip failed: got %v, want %v", parsed, original)
	}
}
