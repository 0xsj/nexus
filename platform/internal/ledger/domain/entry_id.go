package domain

import (
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// EntryID Value Object
// ============================================================================

// EntryID uniquely identifies an audit entry.
// Wraps pkg/types.ID (UUIDv7) for type safety within the Ledger context.
type EntryID struct {
	value types.ID
}

// NewEntryID generates a new unique EntryID.
func NewEntryID() EntryID {
	return EntryID{value: types.NewID()}
}

// ParseEntryID parses a string into an EntryID.
func ParseEntryID(s string) (EntryID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return EntryID{}, InvalidEntryID("EntryID.Parse", s, err.Error())
	}
	return EntryID{value: id}, nil
}

// MustParseEntryID parses a string into an EntryID and panics if invalid.
// Only use for constants or tests.
func MustParseEntryID(s string) EntryID {
	id, err := ParseEntryID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// FromTypesID creates an EntryID from a pkg/types.ID.
func FromTypesID(id types.ID) EntryID {
	return EntryID{value: id}
}

// String returns the string representation.
func (id EntryID) String() string {
	return id.value.String()
}

// IsZero returns true if the EntryID is the zero value.
func (id EntryID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the EntryID is valid (non-zero).
func (id EntryID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two EntryIDs are equal.
func (id EntryID) Equals(other EntryID) bool {
	return id.value.Equals(other.value)
}

// Value returns the underlying types.ID.
func (id EntryID) Value() types.ID {
	return id.value
}
