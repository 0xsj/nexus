package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ShareLinkID uniquely identifies a share link within the Presentation context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (PresentationID, UserID, etc.).
type ShareLinkID struct {
	value types.ID
}

// NewShareLinkID generates a new unique ShareLinkID.
func NewShareLinkID() ShareLinkID {
	return ShareLinkID{value: types.NewID()}
}

// ParseShareLinkID parses a string into a ShareLinkID.
func ParseShareLinkID(s string) (ShareLinkID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return ShareLinkID{}, pkgerrors.Validation("ShareLinkID.Parse", "invalid share link ID format").
			WithMeta("value", s)
	}
	return ShareLinkID{value: id}, nil
}

// MustParseShareLinkID parses a string into a ShareLinkID and panics if invalid.
// Only use for constants or tests.
func MustParseShareLinkID(s string) ShareLinkID {
	id, err := ParseShareLinkID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// ShareLinkIDFromTypesID creates a ShareLinkID from a types.ID.
func ShareLinkIDFromTypesID(id types.ID) ShareLinkID {
	return ShareLinkID{value: id}
}

// String returns the string representation of the ShareLinkID.
func (id ShareLinkID) String() string {
	return id.value.String()
}

// IsZero returns true if the ShareLinkID is the zero value.
func (id ShareLinkID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the ShareLinkID is valid (non-zero).
func (id ShareLinkID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two ShareLinkIDs are equal.
func (id ShareLinkID) Equals(other ShareLinkID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id ShareLinkID) TypesID() types.ID {
	return id.value
}
