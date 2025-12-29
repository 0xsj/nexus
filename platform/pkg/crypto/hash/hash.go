package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Hash represents a cryptographic hash value.
type Hash struct {
	value []byte
}

// SHA256Size is the size of a SHA256 hash in bytes.
const SHA256Size = sha256.Size // 32

// SHA256 computes the SHA256 hash of the data.
func SHA256(data []byte) Hash {
	h := sha256.Sum256(data)
	return Hash{value: h[:]}
}

// SHA256String computes the SHA256 hash of a string.
func SHA256String(s string) Hash {
	return SHA256([]byte(s))
}

// SHA256Multi computes the SHA256 hash of multiple byte slices.
// Useful for hashing multiple fields without concatenation allocation.
func SHA256Multi(parts ...[]byte) Hash {
	h := sha256.New()
	for _, part := range parts {
		h.Write(part)
	}
	return Hash{value: h.Sum(nil)}
}

// FromBytes creates a Hash from raw bytes.
func FromBytes(b []byte) (Hash, error) {
	if len(b) != SHA256Size {
		return Hash{}, fmt.Errorf("invalid hash size: expected %d, got %d", SHA256Size, len(b))
	}
	value := make([]byte, SHA256Size)
	copy(value, b)
	return Hash{value: value}, nil
}

// FromHex creates a Hash from a hex-encoded string.
func FromHex(s string) (Hash, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return Hash{}, fmt.Errorf("invalid hex encoding: %w", err)
	}
	return FromBytes(b)
}

// MustFromHex creates a Hash from hex string, panics if invalid.
// Only use for constants or tests.
func MustFromHex(s string) Hash {
	h, err := FromHex(s)
	if err != nil {
		panic(fmt.Sprintf("invalid hash hex: %v", err))
	}
	return h
}

// Bytes returns the raw hash bytes.
func (h Hash) Bytes() []byte {
	return h.value
}

// Hex returns the hash as a hex-encoded string.
func (h Hash) Hex() string {
	return hex.EncodeToString(h.value)
}

// String returns the hash as a hex-encoded string.
func (h Hash) String() string {
	return h.Hex()
}

// IsZero returns true if the hash is empty/unset.
func (h Hash) IsZero() bool {
	return len(h.value) == 0
}

// IsValid returns true if the hash has valid length.
func (h Hash) IsValid() bool {
	return len(h.value) == SHA256Size
}

// Equal compares two hashes for equality.
func (h Hash) Equal(other Hash) bool {
	if len(h.value) != len(other.value) {
		return false
	}
	for i := range h.value {
		if h.value[i] != other.value[i] {
			return false
		}
	}
	return true
}

// Bytes32 returns the hash as a [32]byte array.
// Useful for smart contract interactions.
func (h Hash) Bytes32() [32]byte {
	var b [32]byte
	copy(b[:], h.value)
	return b
}

// ============================================================================
// Credential Hashing
// ============================================================================

// CredentialHash computes a hash suitable for on-chain anchoring.
// Combines credential ID, issuer, subject, and issuance date.
func CredentialHash(credentialID, issuerDID, subjectDID string, issuanceDate int64) Hash {
	return SHA256Multi(
		[]byte(credentialID),
		[]byte(issuerDID),
		[]byte(subjectDID),
		int64ToBytes(issuanceDate),
	)
}

// int64ToBytes converts an int64 to big-endian bytes.
func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(n & 0xff)
		n >>= 8
	}
	return b
}
