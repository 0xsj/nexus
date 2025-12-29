package ed25519

import (
	"crypto/ed25519"
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/crypto"
)

// Signer signs data using an Ed25519 private key.
type Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// Ensure Signer implements crypto.Signer.
var _ crypto.Signer = (*Signer)(nil)

// NewSigner creates a new Ed25519 signer from a key pair.
func NewSigner(kp *KeyPair) (*Signer, error) {
	const op = "ed25519.NewSigner"

	if !kp.HasPrivateKey() {
		return nil, crypto.ErrInvalidKey(op, "private key required for signing")
	}

	return &Signer{
		privateKey: kp.privateKey,
		publicKey:  kp.publicKey,
	}, nil
}

// NewSignerFromPrivateKey creates a signer from raw private key bytes.
func NewSignerFromPrivateKey(privateKey []byte) (*Signer, error) {
	const op = "ed25519.NewSignerFromPrivateKey"

	if len(privateKey) != PrivateKeySize {
		return nil, crypto.ErrInvalidKey(op, fmt.Sprintf(
			"invalid private key size: expected %d, got %d",
			PrivateKeySize, len(privateKey),
		))
	}

	privKey := ed25519.PrivateKey(privateKey)
	pubKey := privKey.Public().(ed25519.PublicKey)

	return &Signer{
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

// Sign signs the data and returns the signature.
func (s *Signer) Sign(data []byte) ([]byte, error) {
	const op = "ed25519.Signer.Sign"

	if s.privateKey == nil {
		return nil, crypto.ErrSigningFailed(op, fmt.Errorf("private key not available"))
	}

	signature := ed25519.Sign(s.privateKey, data)
	return signature, nil
}

// Algorithm returns the signing algorithm.
func (s *Signer) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmEd25519
}

// PublicKey returns the public key corresponding to the signing key.
func (s *Signer) PublicKey() []byte {
	return s.publicKey
}

// ============================================================================
// Verifier
// ============================================================================

// Verifier verifies Ed25519 signatures.
type Verifier struct {
	publicKey ed25519.PublicKey
}

// Ensure Verifier implements crypto.Verifier.
var _ crypto.Verifier = (*Verifier)(nil)

// NewVerifier creates a new Ed25519 verifier from a public key.
func NewVerifier(publicKey []byte) (*Verifier, error) {
	const op = "ed25519.NewVerifier"

	if len(publicKey) != PublicKeySize {
		return nil, crypto.ErrInvalidKey(op, fmt.Sprintf(
			"invalid public key size: expected %d, got %d",
			PublicKeySize, len(publicKey),
		))
	}

	return &Verifier{
		publicKey: ed25519.PublicKey(publicKey),
	}, nil
}

// NewVerifierFromKeyPair creates a verifier from a key pair.
func NewVerifierFromKeyPair(kp *KeyPair) *Verifier {
	return &Verifier{
		publicKey: kp.publicKey,
	}
}

// Verify checks if the signature is valid for the data.
// Returns nil if valid, error if invalid.
func (v *Verifier) Verify(data, signature []byte) error {
	const op = "ed25519.Verifier.Verify"

	if len(signature) != SignatureSize {
		return crypto.ErrInvalidSignature(op, fmt.Sprintf(
			"invalid signature size: expected %d, got %d",
			SignatureSize, len(signature),
		))
	}

	if !ed25519.Verify(v.publicKey, data, signature) {
		return crypto.ErrVerificationFailed(op)
	}

	return nil
}

// Algorithm returns the verification algorithm.
func (v *Verifier) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmEd25519
}

// PublicKey returns the public key used for verification.
func (v *Verifier) PublicKey() []byte {
	return v.publicKey
}

// ============================================================================
// SignerVerifier
// ============================================================================

// SignerVerifier combines signing and verification capabilities.
type SignerVerifier struct {
	*Signer
	*Verifier
}

// Ensure SignerVerifier implements crypto.SignerVerifier.
var _ crypto.SignerVerifier = (*SignerVerifier)(nil)

// NewSignerVerifier creates a combined signer and verifier from a key pair.
func NewSignerVerifier(kp *KeyPair) (*SignerVerifier, error) {
	signer, err := NewSigner(kp)
	if err != nil {
		return nil, err
	}

	verifier := NewVerifierFromKeyPair(kp)

	return &SignerVerifier{
		Signer:   signer,
		Verifier: verifier,
	}, nil
}

// PublicKey returns the public key (resolves ambiguity between Signer and Verifier).
func (sv *SignerVerifier) PublicKey() []byte {
	return sv.Signer.PublicKey()
}

// Algorithm returns the algorithm (resolves ambiguity between Signer and Verifier).
func (sv *SignerVerifier) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmEd25519
}
