package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// VerificationID uniquely identifies a verification within the Verification context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, SessionID, CredentialID, etc.).
type VerificationID struct {
	value types.ID
}

// NewVerificationID generates a new unique VerificationID.
func NewVerificationID() VerificationID {
	return VerificationID{value: types.NewID()}
}

// ParseVerificationID parses a string into a VerificationID.
func ParseVerificationID(s string) (VerificationID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return VerificationID{}, pkgerrors.Validation("VerificationID.Parse", "invalid verification ID format").
			WithMeta("value", s)
	}
	return VerificationID{value: id}, nil
}

// MustParseVerificationID parses a string into a VerificationID and panics if invalid.
// Only use for constants or tests.
func MustParseVerificationID(s string) VerificationID {
	id, err := ParseVerificationID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// VerificationIDFromTypesID creates a VerificationID from a types.ID.
func VerificationIDFromTypesID(id types.ID) VerificationID {
	return VerificationID{value: id}
}

// String returns the string representation of the VerificationID.
func (id VerificationID) String() string {
	return id.value.String()
}

// IsZero returns true if the VerificationID is the zero value.
func (id VerificationID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the VerificationID is valid (non-zero).
func (id VerificationID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two VerificationIDs are equal.
func (id VerificationID) Equals(other VerificationID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id VerificationID) TypesID() types.ID {
	return id.value
}
