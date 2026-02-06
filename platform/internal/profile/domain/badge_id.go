package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// BadgeID uniquely identifies a badge within the Profile context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (ProfileID, UserID, etc.).
type BadgeID struct {
	value types.ID
}

// NewBadgeID generates a new unique BadgeID.
func NewBadgeID() BadgeID {
	return BadgeID{value: types.NewID()}
}

// ParseBadgeID parses a string into a BadgeID.
func ParseBadgeID(s string) (BadgeID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return BadgeID{}, pkgerrors.Validation("BadgeID.Parse", "invalid badge ID format").
			WithMeta("value", s)
	}
	return BadgeID{value: id}, nil
}

// MustParseBadgeID parses a string into a BadgeID and panics if invalid.
// Only use for constants or tests.
func MustParseBadgeID(s string) BadgeID {
	id, err := ParseBadgeID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// BadgeIDFromTypesID creates a BadgeID from a types.ID.
func BadgeIDFromTypesID(id types.ID) BadgeID {
	return BadgeID{value: id}
}

// String returns the string representation of the BadgeID.
func (id BadgeID) String() string {
	return id.value.String()
}

// IsZero returns true if the BadgeID is the zero value.
func (id BadgeID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the BadgeID is valid (non-zero).
func (id BadgeID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two BadgeIDs are equal.
func (id BadgeID) Equals(other BadgeID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id BadgeID) TypesID() types.ID {
	return id.value
}
