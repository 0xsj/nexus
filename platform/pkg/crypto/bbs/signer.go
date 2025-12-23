package bbs

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
)

// Signer signs data using BBS+ signatures.
// Enables signing multiple messages that can be selectively disclosed.
type Signer struct {
	privateKey []byte
	publicKey  []byte
}

// Ensure Signer implements crypto.Signer.
var _ crypto.Signer = (*Signer)(nil)

// NewSigner creates a new BBS+ signer from a key pair.
// Phase 2: Not yet implemented.
func NewSigner(kp *KeyPair) (*Signer, error) {
	const op = "bbs.NewSigner"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// NewSignerFromPrivateKey creates a signer from raw private key bytes.
// Phase 2: Not yet implemented.
func NewSignerFromPrivateKey(privateKey []byte) (*Signer, error) {
	const op = "bbs.NewSignerFromPrivateKey"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// Sign signs a single message.
// For BBS+, prefer SignMultiple for selective disclosure support.
// Phase 2: Not yet implemented.
func (s *Signer) Sign(data []byte) ([]byte, error) {
	const op = "bbs.Signer.Sign"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// SignMultiple signs multiple messages in a single signature.
// Each message can later be selectively disclosed.
// Phase 2: Not yet implemented.
func (s *Signer) SignMultiple(messages [][]byte) ([]byte, error) {
	const op = "bbs.Signer.SignMultiple"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// Algorithm returns the signing algorithm.
func (s *Signer) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmBLS12381
}

// PublicKey returns the public key corresponding to the signing key.
func (s *Signer) PublicKey() []byte {
	return s.publicKey
}

// ============================================================================
// Verifier
// ============================================================================

// Verifier verifies BBS+ signatures and derived proofs.
type Verifier struct {
	publicKey []byte
}

// Ensure Verifier implements crypto.Verifier.
var _ crypto.Verifier = (*Verifier)(nil)

// NewVerifier creates a new BBS+ verifier from a public key.
// Phase 2: Not yet implemented.
func NewVerifier(publicKey []byte) (*Verifier, error) {
	const op = "bbs.NewVerifier"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// NewVerifierFromKeyPair creates a verifier from a key pair.
// Phase 2: Not yet implemented.
func NewVerifierFromKeyPair(kp *KeyPair) (*Verifier, error) {
	const op = "bbs.NewVerifierFromKeyPair"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// Verify verifies a signature against a single message.
// Phase 2: Not yet implemented.
func (v *Verifier) Verify(data, signature []byte) error {
	const op = "bbs.Verifier.Verify"
	return crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// VerifyMultiple verifies a signature against multiple messages.
// Phase 2: Not yet implemented.
func (v *Verifier) VerifyMultiple(messages [][]byte, signature []byte) error {
	const op = "bbs.Verifier.VerifyMultiple"
	return crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// VerifyProof verifies a derived zero-knowledge proof.
// Phase 2: Not yet implemented.
func (v *Verifier) VerifyProof(proof *Proof) error {
	const op = "bbs.Verifier.VerifyProof"
	return crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// Algorithm returns the verification algorithm.
func (v *Verifier) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmBLS12381
}

// PublicKey returns the public key used for verification.
func (v *Verifier) PublicKey() []byte {
	return v.publicKey
}
