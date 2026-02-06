package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// MemberID uniquely identifies a membership record.
type MemberID struct {
	value types.ID
}

// NewMemberID generates a new unique MemberID.
func NewMemberID() MemberID {
	return MemberID{value: types.NewID()}
}

// ParseMemberID parses a string into a MemberID.
func ParseMemberID(s string) (MemberID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return MemberID{}, pkgerrors.Validation("MemberID.Parse", "invalid member ID format").
			WithMeta("value", s)
	}
	return MemberID{value: id}, nil
}

// String returns the string representation.
func (id MemberID) String() string { return id.value.String() }

// IsZero returns true if the ID is the zero value.
func (id MemberID) IsZero() bool { return id.value.IsZero() }

// Equals checks if two IDs are equal.
func (id MemberID) Equals(other MemberID) bool { return id.value.Equals(other.value) }
