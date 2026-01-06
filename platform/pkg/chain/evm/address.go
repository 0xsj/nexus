package evm

import (
	"encoding/hex"
	"strings"

	"github.com/0xsj/nexus/platform/pkg/chain"
	"golang.org/x/crypto/sha3"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// AddressLength is the length of an EVM address in bytes.
	AddressLength = 20

	// AddressHexLength is the length of an EVM address in hex characters (without 0x prefix).
	AddressHexLength = 40

	// AddressPrefixedHexLength is the length of an EVM address with 0x prefix.
	AddressPrefixedHexLength = 42
)

// ============================================================================
// AddressValidator
// ============================================================================

// AddressValidator validates and normalizes EVM addresses.
type AddressValidator struct{}

// NewAddressValidator creates a new EVM address validator.
func NewAddressValidator() *AddressValidator {
	return &AddressValidator{}
}

// Family returns the chain family this validator handles.
func (v *AddressValidator) Family() chain.Family {
	return chain.FamilyEVM
}

// Validate validates and normalizes an EVM address.
// Returns the checksummed address (EIP-55) or an error.
func (v *AddressValidator) Validate(address string) (string, error) {
	const op = "evm.AddressValidator.Validate"

	// Check for empty address
	if address == "" {
		return "", chain.ErrInvalidAddressFor(op, address, "address is empty")
	}

	// Normalize: remove 0x prefix and convert to lowercase
	addr := strings.TrimPrefix(address, "0x")
	addr = strings.TrimPrefix(addr, "0X")
	addr = strings.ToLower(addr)

	// Check length
	if len(addr) != AddressHexLength {
		return "", chain.ErrInvalidAddressFor(op, address, "invalid length")
	}

	// Validate hex characters
	if _, err := hex.DecodeString(addr); err != nil {
		return "", chain.ErrInvalidAddressFor(op, address, "invalid hex characters")
	}

	// Check for zero address
	if isZeroAddress(addr) {
		return "", chain.ErrInvalidAddressFor(op, address, "zero address not allowed")
	}

	// Return checksummed address
	return toChecksumAddress(addr), nil
}

// IsValid returns true if the address format is valid.
func (v *AddressValidator) IsValid(address string) bool {
	_, err := v.Validate(address)
	return err == nil
}

// Normalize returns the checksummed form of an address.
// Returns empty string if the address is invalid.
func (v *AddressValidator) Normalize(address string) string {
	normalized, err := v.Validate(address)
	if err != nil {
		return ""
	}
	return normalized
}

// ============================================================================
// EIP-55 Checksum
// ============================================================================

// toChecksumAddress converts a lowercase hex address to EIP-55 checksummed format.
// Input must be 40 lowercase hex characters (no 0x prefix).
func toChecksumAddress(addr string) string {
	// Hash the lowercase address
	hash := keccak256([]byte(addr))

	// Build checksummed address
	var result strings.Builder
	result.WriteString("0x")

	for i, c := range addr {
		if c >= '0' && c <= '9' {
			// Numbers are never checksummed
			result.WriteByte(byte(c))
		} else {
			// Get the corresponding nibble from the hash
			// Each byte in hash covers 2 characters
			hashByte := hash[i/2]
			var nibble byte
			if i%2 == 0 {
				nibble = hashByte >> 4
			} else {
				nibble = hashByte & 0x0F
			}

			// If nibble >= 8, uppercase the character
			if nibble >= 8 {
				result.WriteByte(byte(c) - 32) // lowercase to uppercase
			} else {
				result.WriteByte(byte(c))
			}
		}
	}

	return result.String()
}

// VerifyChecksum verifies that an address has a valid EIP-55 checksum.
// Returns true if the checksum is valid or if the address is all lowercase/uppercase.
func VerifyChecksum(address string) bool {
	// Remove prefix
	addr := strings.TrimPrefix(address, "0x")
	addr = strings.TrimPrefix(addr, "0X")

	if len(addr) != AddressHexLength {
		return false
	}

	// All lowercase or all uppercase is valid (no checksum)
	if addr == strings.ToLower(addr) || addr == strings.ToUpper(addr) {
		return true
	}

	// Mixed case must match checksum
	checksummed := toChecksumAddress(strings.ToLower(addr))
	return checksummed == "0x"+addr
}

// ============================================================================
// Helpers
// ============================================================================

// keccak256 computes the Keccak-256 hash of the input.
func keccak256(data []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return h.Sum(nil)
}

// isZeroAddress checks if the address is the zero address.
func isZeroAddress(addr string) bool {
	for _, c := range addr {
		if c != '0' {
			return false
		}
	}
	return true
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ chain.FamilyAddressValidator = (*AddressValidator)(nil)
