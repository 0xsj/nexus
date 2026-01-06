package domain

import (
	"encoding/hex"
	"strings"
	"time"
)

// ============================================================================
// Signature
// ============================================================================

// Signature represents a cryptographic signature from a wallet.
type Signature struct {
	// bytes is the raw signature bytes.
	bytes []byte

	// hex is the hex-encoded signature string.
	hex string

	// algorithm is the signature algorithm used.
	algorithm SignatureAlgorithm

	// createdAt is when the signature was created.
	createdAt time.Time
}

// ============================================================================
// Signature Algorithm
// ============================================================================

// SignatureAlgorithm represents the algorithm used to create a signature.
type SignatureAlgorithm string

const (
	// SignatureAlgorithmSecp256k1 is the secp256k1 ECDSA algorithm (EVM).
	SignatureAlgorithmSecp256k1 SignatureAlgorithm = "secp256k1"

	// SignatureAlgorithmEd25519 is the Ed25519 algorithm (Solana).
	SignatureAlgorithmEd25519 SignatureAlgorithm = "ed25519"

	// SignatureAlgorithmSecp256r1 is the secp256r1 algorithm (some Cosmos).
	SignatureAlgorithmSecp256r1 SignatureAlgorithm = "secp256r1"
)

// String returns the string representation of the algorithm.
func (a SignatureAlgorithm) String() string {
	return string(a)
}

// IsValid returns true if the algorithm is valid.
func (a SignatureAlgorithm) IsValid() bool {
	switch a {
	case SignatureAlgorithmSecp256k1, SignatureAlgorithmEd25519, SignatureAlgorithmSecp256r1:
		return true
	default:
		return false
	}
}

// AlgorithmForFamily returns the default signature algorithm for a chain family.
func AlgorithmForFamily(family ChainFamily) SignatureAlgorithm {
	switch family {
	case ChainFamilyEVM:
		return SignatureAlgorithmSecp256k1
	case ChainFamilySolana:
		return SignatureAlgorithmEd25519
	case ChainFamilyCosmos:
		return SignatureAlgorithmSecp256k1 // Most Cosmos chains use secp256k1
	default:
		return SignatureAlgorithmSecp256k1
	}
}

// ============================================================================
// Constructors
// ============================================================================

// NewSignature creates a new Signature from hex-encoded string.
func NewSignature(hexStr string, algorithm SignatureAlgorithm) (Signature, error) {
	const op = "Signature.New"

	if hexStr == "" {
		return Signature{}, ErrInvalidSignature(op, "signature cannot be empty")
	}

	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")

	// Decode hex
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return Signature{}, ErrInvalidSignature(op, "invalid hex encoding")
	}

	// Validate length based on algorithm
	if err := validateSignatureLength(bytes, algorithm); err != nil {
		return Signature{}, err
	}

	return Signature{
		bytes:     bytes,
		hex:       hexStr,
		algorithm: algorithm,
		createdAt: time.Now(),
	}, nil
}

// NewSignatureFromBytes creates a new Signature from raw bytes.
func NewSignatureFromBytes(bytes []byte, algorithm SignatureAlgorithm) (Signature, error) {
	const op = "Signature.NewFromBytes"

	if len(bytes) == 0 {
		return Signature{}, ErrInvalidSignature(op, "signature cannot be empty")
	}

	// Validate length based on algorithm
	if err := validateSignatureLength(bytes, algorithm); err != nil {
		return Signature{}, err
	}

	return Signature{
		bytes:     bytes,
		hex:       hex.EncodeToString(bytes),
		algorithm: algorithm,
		createdAt: time.Now(),
	}, nil
}

// NewSignatureUnchecked creates a Signature without validation.
// Use only when the signature is known to be valid.
func NewSignatureUnchecked(bytes []byte, hexStr string, algorithm SignatureAlgorithm) Signature {
	return Signature{
		bytes:     bytes,
		hex:       hexStr,
		algorithm: algorithm,
		createdAt: time.Now(),
	}
}

// ============================================================================
// Getters
// ============================================================================

// Bytes returns the raw signature bytes.
func (s Signature) Bytes() []byte {
	result := make([]byte, len(s.bytes))
	copy(result, s.bytes)
	return result
}

// Hex returns the hex-encoded signature string (without 0x prefix).
func (s Signature) Hex() string {
	return s.hex
}

// HexPrefixed returns the hex-encoded signature string with 0x prefix.
func (s Signature) HexPrefixed() string {
	return "0x" + s.hex
}

// String returns the hex-encoded signature string with 0x prefix.
func (s Signature) String() string {
	return s.HexPrefixed()
}

// Algorithm returns the signature algorithm.
func (s Signature) Algorithm() SignatureAlgorithm {
	return s.algorithm
}

// CreatedAt returns when the signature was created.
func (s Signature) CreatedAt() time.Time {
	return s.createdAt
}

// Len returns the length of the signature in bytes.
func (s Signature) Len() int {
	return len(s.bytes)
}

// ============================================================================
// Predicates
// ============================================================================

// IsZero returns true if the signature is empty/uninitialized.
func (s Signature) IsZero() bool {
	return len(s.bytes) == 0
}

// IsEVM returns true if this is an EVM signature (secp256k1).
func (s Signature) IsEVM() bool {
	return s.algorithm == SignatureAlgorithmSecp256k1
}

// IsSolana returns true if this is a Solana signature (Ed25519).
func (s Signature) IsSolana() bool {
	return s.algorithm == SignatureAlgorithmEd25519
}

// ============================================================================
// EVM Signature Components
// ============================================================================

// EVMSignatureComponents contains the r, s, v components of an EVM signature.
type EVMSignatureComponents struct {
	R []byte
	S []byte
	V byte
}

// EVMComponents extracts r, s, v from an EVM signature.
// Returns error if not an EVM signature or invalid length.
func (s Signature) EVMComponents() (EVMSignatureComponents, error) {
	const op = "Signature.EVMComponents"

	if s.algorithm != SignatureAlgorithmSecp256k1 {
		return EVMSignatureComponents{}, ErrInvalidSignature(op, "not an EVM signature")
	}

	if len(s.bytes) != 65 {
		return EVMSignatureComponents{}, ErrInvalidSignature(op, "invalid EVM signature length")
	}

	return EVMSignatureComponents{
		R: s.bytes[0:32],
		S: s.bytes[32:64],
		V: s.bytes[64],
	}, nil
}

// ============================================================================
// Validation
// ============================================================================

// validateSignatureLength validates signature length based on algorithm.
func validateSignatureLength(bytes []byte, algorithm SignatureAlgorithm) error {
	const op = "validateSignatureLength"

	switch algorithm {
	case SignatureAlgorithmSecp256k1:
		// EVM signatures are 65 bytes (r: 32, s: 32, v: 1)
		if len(bytes) != 65 {
			return ErrInvalidSignature(op, "secp256k1 signature must be 65 bytes")
		}
	case SignatureAlgorithmEd25519:
		// Ed25519 signatures are 64 bytes
		if len(bytes) != 64 {
			return ErrInvalidSignature(op, "ed25519 signature must be 64 bytes")
		}
	case SignatureAlgorithmSecp256r1:
		// secp256r1 signatures are typically 64-72 bytes (DER encoded)
		if len(bytes) < 64 || len(bytes) > 72 {
			return ErrInvalidSignature(op, "secp256r1 signature must be 64-72 bytes")
		}
	}

	return nil
}

// ============================================================================
// Signed Message
// ============================================================================

// SignedMessage represents a message that has been signed by a wallet.
type SignedMessage struct {
	// Message is the original message that was signed.
	Message string

	// Signature is the cryptographic signature.
	Signature Signature

	// Address is the address that signed the message.
	Address Address

	// Nonce is the unique nonce used in the message (for replay protection).
	Nonce string

	// Timestamp is when the message was signed.
	Timestamp time.Time

	// ExpiresAt is when the signature expires (optional).
	ExpiresAt *time.Time
}

// IsExpired returns true if the signed message has expired.
func (m SignedMessage) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt)
}

// IsValid returns true if the signed message is valid (not expired, has signature).
func (m SignedMessage) IsValid() bool {
	if m.Signature.IsZero() {
		return false
	}
	if m.Address.IsZero() {
		return false
	}
	if m.IsExpired() {
		return false
	}
	return true
}

// ============================================================================
// SIWE Message
// ============================================================================

// SIWEMessage represents a Sign-In With Ethereum (EIP-4361) message.
type SIWEMessage struct {
	// Domain is the domain that requested the signing.
	Domain string

	// Address is the Ethereum address performing the signing.
	Address Address

	// Statement is an optional human-readable statement.
	Statement string

	// URI is the URI of the resource that requested the signing.
	URI string

	// Version is the SIWE message version (currently "1").
	Version string

	// ChainID is the EIP-155 chain ID.
	ChainID ChainID

	// Nonce is a random string for replay protection.
	Nonce string

	// IssuedAt is when the message was issued.
	IssuedAt time.Time

	// ExpirationTime is when the message expires (optional).
	ExpirationTime *time.Time

	// NotBefore is the earliest time the message is valid (optional).
	NotBefore *time.Time

	// RequestID is an optional request ID.
	RequestID string

	// Resources is a list of resources the user is requesting access to.
	Resources []string
}

// IsExpired returns true if the SIWE message has expired.
func (m SIWEMessage) IsExpired() bool {
	if m.ExpirationTime == nil {
		return false
	}
	return time.Now().After(*m.ExpirationTime)
}

// IsValidTime returns true if the current time is within the valid window.
func (m SIWEMessage) IsValidTime() bool {
	now := time.Now()

	if m.NotBefore != nil && now.Before(*m.NotBefore) {
		return false
	}

	if m.ExpirationTime != nil && now.After(*m.ExpirationTime) {
		return false
	}

	return true
}
