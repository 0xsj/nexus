package secp256k1

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
)

const (
	// PublicKeySize is the size of a compressed secp256k1 public key in bytes.
	PublicKeySize = 33

	// PublicKeyUncompressedSize is the size of an uncompressed public key.
	PublicKeyUncompressedSize = 65

	// PrivateKeySize is the size of a secp256k1 private key in bytes.
	PrivateKeySize = 32

	// SignatureSize is the size of a secp256k1 signature in bytes (r + s).
	SignatureSize = 64
)

// KeyPair is a secp256k1 key pair.
// Used for Ethereum wallet compatibility (did:pkh, did:ethr).
type KeyPair struct {
	publicKey  []byte
	privateKey []byte
}

// Ensure KeyPair implements crypto.KeyPair.
var _ crypto.KeyPair = (*KeyPair)(nil)

// Generate creates a new random secp256k1 key pair.
// Phase 2: Not yet implemented.
func Generate() (*KeyPair, error) {
	const op = "secp256k1.Generate"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// FromPrivateKey reconstructs a key pair from a private key.
// Phase 2: Not yet implemented.
func FromPrivateKey(privateKey []byte) (*KeyPair, error) {
	const op = "secp256k1.FromPrivateKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// FromSeed generates a deterministic key pair from a 32-byte seed.
// Phase 2: Not yet implemented.
func FromSeed(seed []byte) (*KeyPair, error) {
	const op = "secp256k1.FromSeed"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// FromPublicKey creates a key pair with only the public key.
// Phase 2: Not yet implemented.
func FromPublicKey(publicKey []byte) (*KeyPair, error) {
	const op = "secp256k1.FromPublicKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// Algorithm returns the cryptographic algorithm.
func (kp *KeyPair) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmSecp256k1
}

// PublicKey returns the raw public key bytes (compressed).
func (kp *KeyPair) PublicKey() []byte {
	return kp.publicKey
}

// PrivateKey returns the raw private key bytes.
func (kp *KeyPair) PrivateKey() []byte {
	return kp.privateKey
}

// PublicKeyBase58 returns the public key as base58 encoded string.
func (kp *KeyPair) PublicKeyBase58() string {
	// Phase 2: implement base58 encoding
	return ""
}

// PublicKeyMultibase returns the public key as multibase encoded string.
func (kp *KeyPair) PublicKeyMultibase() string {
	// Phase 2: implement multibase encoding with secp256k1 multicodec prefix
	return ""
}

// HasPrivateKey returns true if the private key is available.
func (kp *KeyPair) HasPrivateKey() bool {
	return len(kp.privateKey) > 0
}

// EthereumAddress returns the Ethereum address derived from the public key.
// Phase 2: implement keccak256 hash of uncompressed public key.
func (kp *KeyPair) EthereumAddress() string {
	return ""
}
