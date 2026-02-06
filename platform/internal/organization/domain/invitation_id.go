package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// InvitationID uniquely identifies an invitation.
type InvitationID struct {
	value types.ID
}

// NewInvitationID generates a new unique InvitationID.
func NewInvitationID() InvitationID {
	return InvitationID{value: types.NewID()}
}

// ParseInvitationID parses a string into an InvitationID.
func ParseInvitationID(s string) (InvitationID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return InvitationID{}, pkgerrors.Validation("InvitationID.Parse", "invalid invitation ID format").
			WithMeta("value", s)
	}
	return InvitationID{value: id}, nil
}

// String returns the string representation.
func (id InvitationID) String() string { return id.value.String() }

// IsZero returns true if the ID is the zero value.
func (id InvitationID) IsZero() bool { return id.value.IsZero() }

// Equals checks if two IDs are equal.
func (id InvitationID) Equals(other InvitationID) bool { return id.value.Equals(other.value) }
