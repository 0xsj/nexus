package domain

import (
	"testing"

	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// NewEntryID Tests
// ============================================================================

func TestNewEntryID_GeneratesValidID(t *testing.T) {
	id := NewEntryID()

	if id.IsZero() {
		t.Error("expected non-zero ID")
	}
	if !id.IsValid() {
		t.Error("expected valid ID")
	}
	if id.String() == "" {
		t.Error("expected non-empty string representation")
	}
}

func TestNewEntryID_GeneratesUniqueIDs(t *testing.T) {
	id1 := NewEntryID()
	id2 := NewEntryID()

	if id1.Equals(id2) {
		t.Error("expected unique IDs")
	}
}

// ============================================================================
// ParseEntryID Tests
// ============================================================================

func TestParseEntryID_ValidUUID(t *testing.T) {
	original := NewEntryID()
	parsed, err := ParseEntryID(original.String())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !parsed.Equals(original) {
		t.Errorf("parsed ID does not match original: got %v, want %v", parsed, original)
	}
}

func TestParseEntryID_EmptyString(t *testing.T) {
	_, err := ParseEntryID("")

	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestParseEntryID_InvalidFormat(t *testing.T) {
	_, err := ParseEntryID("not-a-uuid")

	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseEntryID_MalformedUUID(t *testing.T) {
	_, err := ParseEntryID("12345678-1234-1234-1234-12345678901") // one char short

	if err == nil {
		t.Error("expected error for malformed UUID")
	}
}

// ============================================================================
// MustParseEntryID Tests
// ============================================================================

func TestMustParseEntryID_ValidUUID(t *testing.T) {
	original := NewEntryID()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	parsed := MustParseEntryID(original.String())
	if !parsed.Equals(original) {
		t.Errorf("parsed ID does not match original")
	}
}

func TestMustParseEntryID_InvalidUUID_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid UUID")
		}
	}()

	MustParseEntryID("invalid")
}

// ============================================================================
// FromTypesID Tests
// ============================================================================

func TestFromTypesID_Success(t *testing.T) {
	typesID := types.NewID()
	entryID := FromTypesID(typesID)

	if entryID.String() != typesID.String() {
		t.Errorf("ID mismatch: got %v, want %v", entryID.String(), typesID.String())
	}
	if entryID.Value() != typesID {
		t.Error("underlying value mismatch")
	}
}

// ============================================================================
// IsZero / IsValid Tests
// ============================================================================

func TestEntryID_IsZero(t *testing.T) {
	var zeroID EntryID

	if !zeroID.IsZero() {
		t.Error("expected zero value to be zero")
	}
	if zeroID.IsValid() {
		t.Error("expected zero value to be invalid")
	}
}

func TestEntryID_IsValid(t *testing.T) {
	id := NewEntryID()

	if id.IsZero() {
		t.Error("expected new ID to be non-zero")
	}
	if !id.IsValid() {
		t.Error("expected new ID to be valid")
	}
}

// ============================================================================
// Equals Tests
// ============================================================================

func TestEntryID_Equals_Same(t *testing.T) {
	id := NewEntryID()

	if !id.Equals(id) {
		t.Error("expected ID to equal itself")
	}
}

func TestEntryID_Equals_ParsedCopy(t *testing.T) {
	original := NewEntryID()
	parsed, _ := ParseEntryID(original.String())

	if !original.Equals(parsed) {
		t.Error("expected parsed ID to equal original")
	}
}

func TestEntryID_Equals_Different(t *testing.T) {
	id1 := NewEntryID()
	id2 := NewEntryID()

	if id1.Equals(id2) {
		t.Error("expected different IDs to not be equal")
	}
}

func TestEntryID_Equals_ZeroValues(t *testing.T) {
	var zero1, zero2 EntryID

	if !zero1.Equals(zero2) {
		t.Error("expected zero values to be equal")
	}
}

// ============================================================================
// String Tests
// ============================================================================

func TestEntryID_String_NonEmpty(t *testing.T) {
	id := NewEntryID()

	s := id.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(s) != 36 {
		t.Errorf("expected 36 character UUID string, got %d", len(s))
	}
}

func TestEntryID_String_ZeroValue(t *testing.T) {
	var id EntryID

	s := id.String()
	// Zero UUID: 00000000-0000-0000-0000-000000000000
	if s != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("unexpected zero value string: %v", s)
	}
}
