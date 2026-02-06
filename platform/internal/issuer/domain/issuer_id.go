package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// IssuerID uniquely identifies an issuer within the Issuer context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (CredentialID, TemplateID, etc.).
type IssuerID struct {
	value types.ID
}

// NewIssuerID generates a new unique IssuerID.
func NewIssuerID() IssuerID {
	return IssuerID{value: types.NewID()}
}

// ParseIssuerID parses a string into an IssuerID.
func ParseIssuerID(s string) (IssuerID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return IssuerID{}, pkgerrors.Validation("IssuerID.Parse", "invalid issuer ID format").
			WithMeta("value", s)
	}
	return IssuerID{value: id}, nil
}

// MustParseIssuerID parses a string into an IssuerID and panics if invalid.
// Only use for constants or tests.
func MustParseIssuerID(s string) IssuerID {
	id, err := ParseIssuerID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// IssuerIDFromTypesID creates an IssuerID from a types.ID.
func IssuerIDFromTypesID(id types.ID) IssuerID {
	return IssuerID{value: id}
}

// String returns the string representation of the IssuerID.
func (id IssuerID) String() string {
	return id.value.String()
}

// IsZero returns true if the IssuerID is the zero value.
func (id IssuerID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the IssuerID is valid (non-zero).
func (id IssuerID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two IssuerIDs are equal.
func (id IssuerID) Equals(other IssuerID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id IssuerID) TypesID() types.ID {
	return id.value
}
