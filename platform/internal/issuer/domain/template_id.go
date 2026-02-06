package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// TemplateID uniquely identifies a template within the Issuer context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (IssuerID, CredentialID, etc.).
type TemplateID struct {
	value types.ID
}

// NewTemplateID generates a new unique TemplateID.
func NewTemplateID() TemplateID {
	return TemplateID{value: types.NewID()}
}

// ParseTemplateID parses a string into a TemplateID.
func ParseTemplateID(s string) (TemplateID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return TemplateID{}, pkgerrors.Validation("TemplateID.Parse", "invalid template ID format").
			WithMeta("value", s)
	}
	return TemplateID{value: id}, nil
}

// MustParseTemplateID parses a string into a TemplateID and panics if invalid.
// Only use for constants or tests.
func MustParseTemplateID(s string) TemplateID {
	id, err := ParseTemplateID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// TemplateIDFromTypesID creates a TemplateID from a types.ID.
func TemplateIDFromTypesID(id types.ID) TemplateID {
	return TemplateID{value: id}
}

// String returns the string representation of the TemplateID.
func (id TemplateID) String() string {
	return id.value.String()
}

// IsZero returns true if the TemplateID is the zero value.
func (id TemplateID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the TemplateID is valid (non-zero).
func (id TemplateID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two TemplateIDs are equal.
func (id TemplateID) Equals(other TemplateID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id TemplateID) TypesID() types.ID {
	return id.value
}
