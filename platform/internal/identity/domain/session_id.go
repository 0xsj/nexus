package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// SessionID uniquely identifies a session within the Identity context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, CredentialID, etc.).
type SessionID struct {
	value types.ID
}

// NewSessionID generates a new unique SessionID.
func NewSessionID() SessionID {
	return SessionID{value: types.NewID()}
}

// ParseSessionID parses a string into a SessionID.
func ParseSessionID(s string) (SessionID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return SessionID{}, pkgerrors.Validation("SessionID.Parse", "invalid session ID format").
			WithMeta("value", s)
	}
	return SessionID{value: id}, nil
}

// MustParseSessionID parses a string into a SessionID and panics if invalid.
// Only use for constants or tests.
func MustParseSessionID(s string) SessionID {
	id, err := ParseSessionID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// SessionIDFromTypesID creates a SessionID from a types.ID.
func SessionIDFromTypesID(id types.ID) SessionID {
	return SessionID{value: id}
}

// String returns the string representation of the SessionID.
func (id SessionID) String() string {
	return id.value.String()
}

// IsZero returns true if the SessionID is the zero value.
func (id SessionID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the SessionID is valid (non-zero).
func (id SessionID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two SessionIDs are equal.
func (id SessionID) Equals(other SessionID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id SessionID) TypesID() types.ID {
	return id.value
}
