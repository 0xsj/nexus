package bbs

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
)

const (
	// PublicKeySize is the size of a BLS12-381 G2 public key in bytes.
	PublicKeySize = 96

	// PrivateKeySize is the size of a BLS12-381 private key in bytes.
	PrivateKeySize = 32

	// SignatureSize is the size of a BBS+ signature in bytes.
	SignatureSize = 112
)

// KeyPair is a BBS+ key pair using BLS12-381 curve.
// Used for selective disclosure and zero-knowledge proofs.
type KeyPair struct {
	publicKey  []byte
	privateKey []byte
}

// Ensure KeyPair implements crypto.KeyPair.
var _ crypto.KeyPair = (*KeyPair)(nil)

// Generate creates a new random BBS+ key pair.
// Phase 2: Not yet implemented.
func Generate() (*KeyPair, error) {
	const op = "bbs.Generate"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// FromPrivateKey reconstructs a key pair from a private key.
// Phase 2: Not yet implemented.
func FromPrivateKey(privateKey []byte) (*KeyPair, error) {
	const op = "bbs.FromPrivateKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// FromSeed generates a deterministic key pair from a seed.
// Phase 2: Not yet implemented.
func FromSeed(seed []byte) (*KeyPair, error) {
	const op = "bbs.FromSeed"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// FromPublicKey creates a key pair with only the public key.
// Phase 2: Not yet implemented.
func FromPublicKey(publicKey []byte) (*KeyPair, error) {
	const op = "bbs.FromPublicKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// Algorithm returns the cryptographic algorithm.
func (kp *KeyPair) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmBLS12381
}

// PublicKey returns the raw public key bytes.
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
	// Phase 2: implement multibase encoding with BLS12-381 multicodec prefix
	return ""
}

// HasPrivateKey returns true if the private key is available.
func (kp *KeyPair) HasPrivateKey() bool {
	return len(kp.privateKey) > 0
}
