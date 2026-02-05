package domain

import (
	"strings"
	"unicode/utf8"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// WalletLabel constraints
const (
	WalletLabelMaxLength = 50
)

// WalletLabel represents an optional user-assigned label for a wallet.
// Labels are used to help users distinguish between multiple wallets.
type WalletLabel struct {
	value string
}

// NewWalletLabel creates a new WalletLabel from a string.
// Trims whitespace and validates that the label does not exceed the maximum length.
// An empty label is allowed.
func NewWalletLabel(s string) (WalletLabel, error) {
	s = strings.TrimSpace(s)

	length := utf8.RuneCountInString(s)
	if length > WalletLabelMaxLength {
		return WalletLabel{}, pkgerrors.Validation("WalletLabel.New", "wallet label too long").
			WithMeta("max_length", WalletLabelMaxLength).
			WithMeta("actual_length", length)
	}

	return WalletLabel{value: s}, nil
}

// String returns the label string.
func (l WalletLabel) String() string {
	return l.value
}

// IsZero returns true if the label is empty.
func (l WalletLabel) IsZero() bool {
	return l.value == ""
}

// Equals checks if two wallet labels are equal.
func (l WalletLabel) Equals(other WalletLabel) bool {
	return l.value == other.value
}
