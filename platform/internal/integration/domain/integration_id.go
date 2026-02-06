package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// IntegrationID uniquely identifies an integration within the Integration context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, CredentialID, etc.).
type IntegrationID struct {
	value types.ID
}

// NewIntegrationID generates a new unique IntegrationID.
func NewIntegrationID() IntegrationID {
	return IntegrationID{value: types.NewID()}
}

// ParseIntegrationID parses a string into an IntegrationID.
func ParseIntegrationID(s string) (IntegrationID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return IntegrationID{}, pkgerrors.Validation("IntegrationID.Parse", "invalid integration ID format").
			WithMeta("value", s)
	}
	return IntegrationID{value: id}, nil
}

// MustParseIntegrationID parses a string into an IntegrationID and panics if invalid.
// Only use for constants or tests.
func MustParseIntegrationID(s string) IntegrationID {
	id, err := ParseIntegrationID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// IntegrationIDFromTypesID creates an IntegrationID from a types.ID.
func IntegrationIDFromTypesID(id types.ID) IntegrationID {
	return IntegrationID{value: id}
}

// String returns the string representation of the IntegrationID.
func (id IntegrationID) String() string {
	return id.value.String()
}

// IsZero returns true if the IntegrationID is the zero value.
func (id IntegrationID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the IntegrationID is valid (non-zero).
func (id IntegrationID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two IntegrationIDs are equal.
func (id IntegrationID) Equals(other IntegrationID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id IntegrationID) TypesID() types.ID {
	return id.value
}
