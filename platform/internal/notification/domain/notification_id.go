package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// NotificationID uniquely identifies a notification within the Notification context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, CredentialID, etc.).
type NotificationID struct {
	value types.ID
}

// NewNotificationID generates a new unique NotificationID.
func NewNotificationID() NotificationID {
	return NotificationID{value: types.NewID()}
}

// ParseNotificationID parses a string into a NotificationID.
func ParseNotificationID(s string) (NotificationID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return NotificationID{}, pkgerrors.Validation("NotificationID.Parse", "invalid notification ID format").
			WithMeta("value", s)
	}
	return NotificationID{value: id}, nil
}

// MustParseNotificationID parses a string into a NotificationID and panics if invalid.
// Only use for constants or tests.
func MustParseNotificationID(s string) NotificationID {
	id, err := ParseNotificationID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// NotificationIDFromTypesID creates a NotificationID from a types.ID.
func NotificationIDFromTypesID(id types.ID) NotificationID {
	return NotificationID{value: id}
}

// String returns the string representation of the NotificationID.
func (id NotificationID) String() string {
	return id.value.String()
}

// IsZero returns true if the NotificationID is the zero value.
func (id NotificationID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the NotificationID is valid (non-zero).
func (id NotificationID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two NotificationIDs are equal.
func (id NotificationID) Equals(other NotificationID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id NotificationID) TypesID() types.ID {
	return id.value
}
