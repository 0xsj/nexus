package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// VouchID uniquely identifies a vouch within the Trust context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, CredentialID, etc.).
type VouchID struct {
	value types.ID
}

// NewVouchID generates a new unique VouchID.
func NewVouchID() VouchID {
	return VouchID{value: types.NewID()}
}

// ParseVouchID parses a string into a VouchID.
func ParseVouchID(s string) (VouchID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return VouchID{}, pkgerrors.Validation("VouchID.Parse", "invalid vouch ID format").
			WithMeta("value", s)
	}
	return VouchID{value: id}, nil
}

// MustParseVouchID parses a string into a VouchID and panics if invalid.
// Only use for constants or tests.
func MustParseVouchID(s string) VouchID {
	id, err := ParseVouchID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// VouchIDFromTypesID creates a VouchID from a types.ID.
func VouchIDFromTypesID(id types.ID) VouchID {
	return VouchID{value: id}
}

// String returns the string representation of the VouchID.
func (id VouchID) String() string {
	return id.value.String()
}

// IsZero returns true if the VouchID is the zero value.
func (id VouchID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the VouchID is valid (non-zero).
func (id VouchID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two VouchIDs are equal.
func (id VouchID) Equals(other VouchID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id VouchID) TypesID() types.ID {
	return id.value
}
