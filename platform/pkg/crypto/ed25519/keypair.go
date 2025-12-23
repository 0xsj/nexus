package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/crypto"
)

const (
	// PublicKeySize is the size of an Ed25519 public key in bytes.
	PublicKeySize = ed25519.PublicKeySize // 32

	// PrivateKeySize is the size of an Ed25519 private key in bytes.
	PrivateKeySize = ed25519.PrivateKeySize // 64

	// SeedSize is the size of an Ed25519 seed in bytes.
	SeedSize = ed25519.SeedSize // 32

	// SignatureSize is the size of an Ed25519 signature in bytes.
	SignatureSize = ed25519.SignatureSize // 64
)

// KeyPair is an Ed25519 key pair.
type KeyPair struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// Ensure KeyPair implements crypto.KeyPair.
var _ crypto.KeyPair = (*KeyPair)(nil)

// Generate creates a new random Ed25519 key pair.
func Generate() (*KeyPair, error) {
	const op = "ed25519.Generate"

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, crypto.ErrKeyGenerationFailed(op, err)
	}

	return &KeyPair{
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

// FromPrivateKey reconstructs a key pair from a private key.
// The private key must be 64 bytes (seed + public key).
func FromPrivateKey(privateKey []byte) (*KeyPair, error) {
	const op = "ed25519.FromPrivateKey"

	if len(privateKey) != PrivateKeySize {
		return nil, crypto.ErrInvalidKey(op, fmt.Sprintf(
			"invalid private key size: expected %d, got %d",
			PrivateKeySize, len(privateKey),
		))
	}

	privKey := ed25519.PrivateKey(privateKey)
	pubKey := privKey.Public().(ed25519.PublicKey)

	return &KeyPair{
		publicKey:  pubKey,
		privateKey: privKey,
	}, nil
}

// FromSeed generates a deterministic key pair from a 32-byte seed.
func FromSeed(seed []byte) (*KeyPair, error) {
	const op = "ed25519.FromSeed"

	if len(seed) != SeedSize {
		return nil, crypto.ErrInvalidKey(op, fmt.Sprintf(
			"invalid seed size: expected %d, got %d",
			SeedSize, len(seed),
		))
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &KeyPair{
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

// FromPublicKey creates a key pair with only the public key.
// This can only be used for verification, not signing.
func FromPublicKey(publicKey []byte) (*KeyPair, error) {
	const op = "ed25519.FromPublicKey"

	if len(publicKey) != PublicKeySize {
		return nil, crypto.ErrInvalidKey(op, fmt.Sprintf(
			"invalid public key size: expected %d, got %d",
			PublicKeySize, len(publicKey),
		))
	}

	return &KeyPair{
		publicKey:  ed25519.PublicKey(publicKey),
		privateKey: nil,
	}, nil
}

// Algorithm returns the cryptographic algorithm.
func (kp *KeyPair) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmEd25519
}

// PublicKey returns the raw public key bytes.
func (kp *KeyPair) PublicKey() []byte {
	return kp.publicKey
}

// PrivateKey returns the raw private key bytes.
func (kp *KeyPair) PrivateKey() []byte {
	return kp.privateKey
}

// Seed returns the 32-byte seed from which the private key was derived.
// Returns nil if the key pair was created from public key only.
func (kp *KeyPair) Seed() []byte {
	if kp.privateKey == nil {
		return nil
	}
	return kp.privateKey.Seed()
}

// HasPrivateKey returns true if the private key is available.
func (kp *KeyPair) HasPrivateKey() bool {
	return kp.privateKey != nil
}

// PublicKeyBase58 returns the public key as base58 encoded string.
func (kp *KeyPair) PublicKeyBase58() string {
	return base58Encode(kp.publicKey)
}

// PublicKeyMultibase returns the public key as multibase encoded string.
// Uses base58-btc encoding with multicodec prefix for Ed25519.
func (kp *KeyPair) PublicKeyMultibase() string {
	// Multicodec prefix for ed25519-pub: 0xed01
	prefixed := append([]byte{0xed, 0x01}, kp.publicKey...)
	// Multibase prefix 'z' indicates base58-btc
	return "z" + base58Encode(prefixed)
}

// PublicKeyBase64URL returns the public key as URL-safe base64 encoded string.
// Used in JWK representations.
func (kp *KeyPair) PublicKeyBase64URL() string {
	return base64.RawURLEncoding.EncodeToString(kp.publicKey)
}

// ============================================================================
// Base58 Encoding (Bitcoin alphabet)
// ============================================================================

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Encode encodes bytes to base58 string.
func base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	// Count leading zeros
	zeros := 0
	for _, b := range input {
		if b != 0 {
			break
		}
		zeros++
	}

	// Allocate enough space
	size := len(input)*138/100 + 1
	buf := make([]byte, size)

	for _, b := range input {
		carry := int(b)
		for i := size - 1; i >= 0; i-- {
			carry += 256 * int(buf[i])
			buf[i] = byte(carry % 58)
			carry /= 58
		}
	}

	// Skip leading zeros in base58 result
	i := 0
	for i < len(buf) && buf[i] == 0 {
		i++
	}

	// Build result
	result := make([]byte, zeros+len(buf)-i)
	for j := 0; j < zeros; j++ {
		result[j] = '1'
	}
	for j := zeros; i < len(buf); i, j = i+1, j+1 {
		result[j] = base58Alphabet[buf[i]]
	}

	return string(result)
}

// base58Decode decodes a base58 string to bytes.
func base58Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	// Build alphabet index
	alphabetIdx := make(map[rune]int)
	for i, c := range base58Alphabet {
		alphabetIdx[c] = i
	}

	// Count leading '1's (zeros in result)
	zeros := 0
	for _, c := range input {
		if c != '1' {
			break
		}
		zeros++
	}

	// Allocate enough space
	size := len(input)*733/1000 + 1
	buf := make([]byte, size)

	for _, c := range input {
		idx, ok := alphabetIdx[c]
		if !ok {
			return nil, fmt.Errorf("invalid base58 character: %c", c)
		}

		carry := idx
		for i := size - 1; i >= 0; i-- {
			carry += 58 * int(buf[i])
			buf[i] = byte(carry % 256)
			carry /= 256
		}
	}

	// Skip leading zeros in buf
	i := 0
	for i < len(buf) && buf[i] == 0 {
		i++
	}

	// Build result with leading zeros
	result := make([]byte, zeros+len(buf)-i)
	copy(result[zeros:], buf[i:])

	return result, nil
}
