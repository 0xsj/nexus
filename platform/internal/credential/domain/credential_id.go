package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// CredentialID uniquely identifies a credential within the Credential context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, SessionID, etc.).
type CredentialID struct {
	value types.ID
}

// NewCredentialID generates a new unique CredentialID.
func NewCredentialID() CredentialID {
	return CredentialID{value: types.NewID()}
}

// ParseCredentialID parses a string into a CredentialID.
func ParseCredentialID(s string) (CredentialID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return CredentialID{}, pkgerrors.Validation("CredentialID.Parse", "invalid credential ID format").
			WithMeta("value", s)
	}
	return CredentialID{value: id}, nil
}

// MustParseCredentialID parses a string into a CredentialID and panics if invalid.
// Only use for constants or tests.
func MustParseCredentialID(s string) CredentialID {
	id, err := ParseCredentialID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// CredentialIDFromTypesID creates a CredentialID from a types.ID.
func CredentialIDFromTypesID(id types.ID) CredentialID {
	return CredentialID{value: id}
}

// String returns the string representation of the CredentialID.
func (id CredentialID) String() string {
	return id.value.String()
}

// IsZero returns true if the CredentialID is the zero value.
func (id CredentialID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the CredentialID is valid (non-zero).
func (id CredentialID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two CredentialIDs are equal.
func (id CredentialID) Equals(other CredentialID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id CredentialID) TypesID() types.ID {
	return id.value
}
