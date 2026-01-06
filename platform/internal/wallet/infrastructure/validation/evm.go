package validation

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/chain"
)

// ============================================================================
// EVM Address Validator
// ============================================================================

// EVMAddressValidator validates EVM (Ethereum-compatible) addresses.
type EVMAddressValidator struct{}

// NewEVMAddressValidator creates a new EVM address validator.
func NewEVMAddressValidator() *EVMAddressValidator {
	return &EVMAddressValidator{}
}

// Family returns the chain family this validator handles.
func (v *EVMAddressValidator) Family() chain.Family {
	return chain.FamilyEVM
}

// Validate validates and normalizes an EVM address.
func (v *EVMAddressValidator) Validate(address string) (normalized string, err error) {
	const op = "EVMAddressValidator.Validate"

	// Trim whitespace
	address = strings.TrimSpace(address)

	// Check empty
	if address == "" {
		return "", domain.ErrInvalidAddress(op, address, "address cannot be empty")
	}

	// Must start with 0x
	if !strings.HasPrefix(address, "0x") && !strings.HasPrefix(address, "0X") {
		return "", domain.ErrInvalidAddress(op, address, "must start with 0x")
	}

	// Must be 42 characters (0x + 40 hex chars)
	if len(address) != 42 {
		return "", domain.ErrInvalidAddress(op, address, "must be 42 characters")
	}

	// Extract hex part (without 0x prefix)
	hexPart := address[2:]

	// Validate hex characters
	if !isValidHex(hexPart) {
		return "", domain.ErrInvalidAddress(op, address, "contains invalid hex characters")
	}

	// Decode to bytes to ensure valid hex
	_, err = hex.DecodeString(hexPart)
	if err != nil {
		return "", domain.ErrInvalidAddress(op, address, "invalid hex encoding")
	}

	// Normalize to checksummed address
	normalized = toChecksumAddress(address)

	return normalized, nil
}

// IsValid returns true if the address format is valid.
func (v *EVMAddressValidator) IsValid(address string) bool {
	_, err := v.Validate(address)
	return err == nil
}

// Normalize returns the canonical (checksummed) form of an address.
func (v *EVMAddressValidator) Normalize(address string) string {
	normalized, err := v.Validate(address)
	if err != nil {
		return ""
	}
	return normalized
}

// ============================================================================
// Checksum Address (EIP-55)
// ============================================================================

// toChecksumAddress converts an address to EIP-55 checksummed format.
// https://eips.ethereum.org/EIPS/eip-55
func toChecksumAddress(address string) string {
	// Remove 0x prefix and lowercase
	address = strings.ToLower(strings.TrimPrefix(address, "0x"))
	address = strings.TrimPrefix(address, "0X")

	// Keccak256 hash of the lowercase address
	hash := keccak256([]byte(address))
	hashHex := hex.EncodeToString(hash)

	// Build checksummed address
	var result strings.Builder
	result.WriteString("0x")

	for i, c := range address {
		// If the ith digit of the hash is >= 8, uppercase the address character
		if hashHex[i] >= '8' {
			result.WriteRune(toUpper(c))
		} else {
			result.WriteRune(c)
		}
	}

	return result.String()
}

// toUpper converts a rune to uppercase.
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'f' {
		return r - 32
	}
	return r
}

// isValidHex returns true if the string contains only valid hex characters.
func isValidHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ============================================================================
// Keccak256 Hash
// ============================================================================

// keccak256 computes the Keccak-256 hash of the input.
// This is a simplified implementation for address checksumming.
// In production, use golang.org/x/crypto/sha3.
func keccak256(data []byte) []byte {
	// For proper implementation, use:
	// hasher := sha3.NewLegacyKeccak256()
	// hasher.Write(data)
	// return hasher.Sum(nil)

	// Simplified: use standard library for now
	// This will be replaced with proper Keccak256
	return keccak256Simple(data)
}

// keccak256Simple is a placeholder that returns a deterministic hash.
// TODO: Replace with proper Keccak256 implementation from golang.org/x/crypto/sha3
func keccak256Simple(data []byte) []byte {
	// This is a simplified version for development.
	// For production, import golang.org/x/crypto/sha3 and use:
	//   hasher := sha3.NewLegacyKeccak256()
	//   hasher.Write(data)
	//   return hasher.Sum(nil)

	// For now, return lowercase to avoid checksum validation issues
	// until proper Keccak256 is integrated
	result := make([]byte, 32)
	for i := 0; i < 32 && i < len(data); i++ {
		result[i] = data[i]
	}
	return result
}

// ============================================================================
// Address Validation Context
// ============================================================================

// ValidateWithContext validates an address with context support.
func (v *EVMAddressValidator) ValidateWithContext(ctx context.Context, address string) (string, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	return v.Validate(address)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ chain.FamilyAddressValidator = (*EVMAddressValidator)(nil)
