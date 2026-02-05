package domain

import (
	"encoding/hex"
	"strings"
)

// WalletAddress represents a validated blockchain wallet address.
// Currently supports EVM-compatible addresses (0x-prefixed, 20-byte hex).
type WalletAddress struct {
	value string
}

// NewWalletAddress creates a new WalletAddress from a string.
// Validates that the address is a valid EVM address: starts with "0x",
// is 42 characters total, and the remaining 40 characters are valid hex.
// The address is normalized to lowercase.
func NewWalletAddress(s string) (WalletAddress, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return WalletAddress{}, InvalidAddress("WalletAddress.New", s, "address cannot be empty")
	}

	if !strings.HasPrefix(s, "0x") {
		return WalletAddress{}, InvalidAddress("WalletAddress.New", s, "address must start with 0x")
	}

	if len(s) != 42 {
		return WalletAddress{}, InvalidAddress("WalletAddress.New", s, "address must be 42 characters")
	}

	// Validate hex portion (after 0x prefix)
	hexPart := s[2:]
	if _, err := hex.DecodeString(hexPart); err != nil {
		return WalletAddress{}, InvalidAddress("WalletAddress.New", s, "address contains invalid hex characters")
	}

	// Normalize to lowercase
	return WalletAddress{value: strings.ToLower(s)}, nil
}

// String returns the string representation of the wallet address.
func (a WalletAddress) String() string {
	return a.value
}

// IsZero returns true if the wallet address is the zero value.
func (a WalletAddress) IsZero() bool {
	return a.value == ""
}

// IsValid returns true if the wallet address is valid (non-empty).
func (a WalletAddress) IsValid() bool {
	return a.value != ""
}

// Equals checks if two wallet addresses are equal.
func (a WalletAddress) Equals(other WalletAddress) bool {
	return a.value == other.value
}
