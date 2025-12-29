package secp256k1

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
)

// Signer signs data using a secp256k1 private key.
// Used for Ethereum-compatible signatures.
type Signer struct {
	privateKey []byte
	publicKey  []byte
}

// Ensure Signer implements crypto.Signer.
var _ crypto.Signer = (*Signer)(nil)

// NewSigner creates a new secp256k1 signer from a key pair.
// Phase 2: Not yet implemented.
func NewSigner(kp *KeyPair) (*Signer, error) {
	const op = "secp256k1.NewSigner"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// NewSignerFromPrivateKey creates a signer from raw private key bytes.
// Phase 2: Not yet implemented.
func NewSignerFromPrivateKey(privateKey []byte) (*Signer, error) {
	const op = "secp256k1.NewSignerFromPrivateKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// Sign signs the data and returns the signature.
// Phase 2: Not yet implemented.
func (s *Signer) Sign(data []byte) ([]byte, error) {
	const op = "secp256k1.Signer.Sign"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// SignHash signs a pre-hashed message (32 bytes).
// Used for Ethereum-style signing where the message is already hashed.
// Phase 2: Not yet implemented.
func (s *Signer) SignHash(hash []byte) ([]byte, error) {
	const op = "secp256k1.Signer.SignHash"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// SignEthereum signs data with Ethereum's personal_sign prefix.
// Adds "\x19Ethereum Signed Message:\n" prefix before hashing.
// Phase 2: Not yet implemented.
func (s *Signer) SignEthereum(data []byte) ([]byte, error) {
	const op = "secp256k1.Signer.SignEthereum"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// Algorithm returns the signing algorithm.
func (s *Signer) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmSecp256k1
}

// PublicKey returns the public key corresponding to the signing key.
func (s *Signer) PublicKey() []byte {
	return s.publicKey
}

// ============================================================================
// Verifier
// ============================================================================

// Verifier verifies secp256k1 signatures.
type Verifier struct {
	publicKey []byte
}

// Ensure Verifier implements crypto.Verifier.
var _ crypto.Verifier = (*Verifier)(nil)

// NewVerifier creates a new secp256k1 verifier from a public key.
// Phase 2: Not yet implemented.
func NewVerifier(publicKey []byte) (*Verifier, error) {
	const op = "secp256k1.NewVerifier"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// NewVerifierFromKeyPair creates a verifier from a key pair.
// Phase 2: Not yet implemented.
func NewVerifierFromKeyPair(kp *KeyPair) (*Verifier, error) {
	const op = "secp256k1.NewVerifierFromKeyPair"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// Verify checks if the signature is valid for the data.
// Phase 2: Not yet implemented.
func (v *Verifier) Verify(data, signature []byte) error {
	const op = "secp256k1.Verifier.Verify"
	return crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// VerifyHash verifies a signature against a pre-hashed message.
// Phase 2: Not yet implemented.
func (v *Verifier) VerifyHash(hash, signature []byte) error {
	const op = "secp256k1.Verifier.VerifyHash"
	return crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// RecoverPublicKey recovers the public key from a signature and message hash.
// Useful for Ethereum ecrecover functionality.
// Phase 2: Not yet implemented.
func RecoverPublicKey(hash, signature []byte) ([]byte, error) {
	const op = "secp256k1.RecoverPublicKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmSecp256k1)
}

// Algorithm returns the verification algorithm.
func (v *Verifier) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmSecp256k1
}

// PublicKey returns the public key used for verification.
func (v *Verifier) PublicKey() []byte {
	return v.publicKey
}
