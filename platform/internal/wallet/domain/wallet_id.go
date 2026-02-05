package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// WalletID uniquely identifies a wallet within the Wallet context.
// Wraps pkg/types.ID to provide type safety and prevent mixing
// with other ID types (UserID, SessionID, etc.).
type WalletID struct {
	value types.ID
}

// NewWalletID generates a new unique WalletID.
func NewWalletID() WalletID {
	return WalletID{value: types.NewID()}
}

// ParseWalletID parses a string into a WalletID.
func ParseWalletID(s string) (WalletID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return WalletID{}, pkgerrors.Validation("WalletID.Parse", "invalid wallet ID format").
			WithMeta("value", s)
	}
	return WalletID{value: id}, nil
}

// MustParseWalletID parses a string into a WalletID and panics if invalid.
// Only use for constants or tests.
func MustParseWalletID(s string) WalletID {
	id, err := ParseWalletID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// WalletIDFromTypesID creates a WalletID from a types.ID.
func WalletIDFromTypesID(id types.ID) WalletID {
	return WalletID{value: id}
}

// String returns the string representation of the WalletID.
func (id WalletID) String() string {
	return id.value.String()
}

// IsZero returns true if the WalletID is the zero value.
func (id WalletID) IsZero() bool {
	return id.value.IsZero()
}

// IsValid returns true if the WalletID is valid (non-zero).
func (id WalletID) IsValid() bool {
	return id.value.IsValid()
}

// Equals checks if two WalletIDs are equal.
func (id WalletID) Equals(other WalletID) bool {
	return id.value.Equals(other.value)
}

// TypesID returns the underlying types.ID.
func (id WalletID) TypesID() types.ID {
	return id.value
}
