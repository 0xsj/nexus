package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// PresentationID uniquely identifies a presentation within the Presentation context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, ShareLinkID, etc.).
type PresentationID struct {
	value types.ID
}

// NewPresentationID generates a new unique PresentationID.
func NewPresentationID() PresentationID {
	return PresentationID{value: types.NewID()}
}

// ParsePresentationID parses a string into a PresentationID.
func ParsePresentationID(s string) (PresentationID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return PresentationID{}, pkgerrors.Validation("PresentationID.Parse", "invalid presentation ID format").
			WithMeta("value", s)
	}
	return PresentationID{value: id}, nil
}

// MustParsePresentationID parses a string into a PresentationID and panics if invalid.
// Only use for constants or tests.
func MustParsePresentationID(s string) PresentationID {
	id, err := ParsePresentationID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// PresentationIDFromTypesID creates a PresentationID from a types.ID.
func PresentationIDFromTypesID(id types.ID) PresentationID {
	return PresentationID{value: id}
}

// String returns the string representation of the PresentationID.
func (id PresentationID) String() string {
	return id.value.String()
}

// IsZero returns true if the PresentationID is the zero value.
func (id PresentationID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the PresentationID is valid (non-zero).
func (id PresentationID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two PresentationIDs are equal.
func (id PresentationID) Equals(other PresentationID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id PresentationID) TypesID() types.ID {
	return id.value
}
