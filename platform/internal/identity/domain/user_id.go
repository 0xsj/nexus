package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// UserID uniquely identifies a user within the Identity context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (SessionID, CredentialID, etc.).
type UserID struct {
	value types.ID
}

// NewUserID generates a new unique UserID.
func NewUserID() UserID {
	return UserID{value: types.NewID()}
}

// ParseUserID parses a string into a UserID.
func ParseUserID(s string) (UserID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return UserID{}, pkgerrors.Validation("UserID.Parse", "invalid user ID format").
			WithMeta("value", s)
	}
	return UserID{value: id}, nil
}

// MustParseUserID parses a string into a UserID and panics if invalid.
// Only use for constants or tests.
func MustParseUserID(s string) UserID {
	id, err := ParseUserID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// UserIDFromTypesID creates a UserID from a types.ID.
func UserIDFromTypesID(id types.ID) UserID {
	return UserID{value: id}
}

// String returns the string representation of the UserID.
func (id UserID) String() string {
	return id.value.String()
}

// IsZero returns true if the UserID is the zero value.
func (id UserID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the UserID is valid (non-zero).
func (id UserID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two UserIDs are equal.
func (id UserID) Equals(other UserID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id UserID) TypesID() types.ID {
	return id.value
}
