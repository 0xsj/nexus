package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ProfileID uniquely identifies a profile within the Profile context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, BadgeID, etc.).
type ProfileID struct {
	value types.ID
}

// NewProfileID generates a new unique ProfileID.
func NewProfileID() ProfileID {
	return ProfileID{value: types.NewID()}
}

// ParseProfileID parses a string into a ProfileID.
func ParseProfileID(s string) (ProfileID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return ProfileID{}, pkgerrors.Validation("ProfileID.Parse", "invalid profile ID format").
			WithMeta("value", s)
	}
	return ProfileID{value: id}, nil
}

// MustParseProfileID parses a string into a ProfileID and panics if invalid.
// Only use for constants or tests.
func MustParseProfileID(s string) ProfileID {
	id, err := ParseProfileID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// ProfileIDFromTypesID creates a ProfileID from a types.ID.
func ProfileIDFromTypesID(id types.ID) ProfileID {
	return ProfileID{value: id}
}

// String returns the string representation of the ProfileID.
func (id ProfileID) String() string {
	return id.value.String()
}

// IsZero returns true if the ProfileID is the zero value.
func (id ProfileID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the ProfileID is valid (non-zero).
func (id ProfileID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two ProfileIDs are equal.
func (id ProfileID) Equals(other ProfileID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id ProfileID) TypesID() types.ID {
	return id.value
}
