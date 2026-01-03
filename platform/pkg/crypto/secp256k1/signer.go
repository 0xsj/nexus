package secp256k1

import (
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/0xsj/nexus/platform/pkg/crypto"
)

// ============================================================================
// Signer
// ============================================================================

// Signer signs data using a secp256k1 private key.
// Used for Ethereum-compatible signatures.
type Signer struct {
	privKey *secp256k1.PrivateKey
	pubKey  *secp256k1.PublicKey
}

// Ensure Signer implements crypto.Signer.
var _ crypto.Signer = (*Signer)(nil)

// NewSigner creates a new secp256k1 signer from a key pair.
func NewSigner(kp *KeyPair) (*Signer, error) {
	const op = "secp256k1.NewSigner"

	if kp == nil {
		return nil, crypto.ErrInvalidKey(op, "key pair is nil")
	}

	if !kp.HasPrivateKey() {
		return nil, crypto.ErrInvalidKey(op, "key pair has no private key")
	}

	return &Signer{
		privKey: kp.privKey,
		pubKey:  kp.pubKey,
	}, nil
}

// NewSignerFromPrivateKey creates a signer from raw private key bytes.
func NewSignerFromPrivateKey(privateKey []byte) (*Signer, error) {
	const op = "secp256k1.NewSignerFromPrivateKey"

	if len(privateKey) != PrivateKeySize {
		return nil, crypto.ErrInvalidKey(op, "private key must be 32 bytes")
	}

	privKey := secp256k1.PrivKeyFromBytes(privateKey)

	return &Signer{
		privKey: privKey,
		pubKey:  privKey.PubKey(),
	}, nil
}

// Sign signs the data and returns the signature.
// The data is hashed with Keccak256 before signing.
// Returns a 65-byte signature [R || S || V] where V is 0 or 1.
func (s *Signer) Sign(data []byte) ([]byte, error) {
	hash := Keccak256(data)
	return s.SignHash(hash)
}

// SignHash signs a pre-hashed message (32 bytes).
// Returns a 65-byte signature [R || S || V] where V is 0 or 1.
func (s *Signer) SignHash(hash []byte) ([]byte, error) {
	const op = "secp256k1.Signer.SignHash"

	if len(hash) != 32 {
		return nil, crypto.ErrSigningFailed(op, fmt.Errorf("hash must be 32 bytes, got %d", len(hash)))
	}

	// SignCompact returns [V || R || S] where V is 27 or 28
	compactSig := ecdsa.SignCompact(s.privKey, hash, false)

	// Convert to Ethereum format [R || S || V] where V is 0 or 1
	v := compactSig[0] - 27
	r := compactSig[1:33]
	ss := compactSig[33:65]

	result := make([]byte, SignatureWithRecoverySize)
	copy(result[0:32], r)
	copy(result[32:64], ss)
	result[64] = v

	return result, nil
}

// SignEthereum signs data with Ethereum's personal_sign prefix.
// Adds "\x19Ethereum Signed Message:\n<len>" prefix before hashing.
// Returns a 65-byte signature [R || S || V] where V is 27 or 28.
func (s *Signer) SignEthereum(data []byte) ([]byte, error) {
	hash := HashEthereumMessage(data)

	sig, err := s.SignHash(hash)
	if err != nil {
		return nil, err
	}

	// Convert V from 0/1 to 27/28 for Ethereum
	sig[64] += 27

	return sig, nil
}

// Algorithm returns the signing algorithm.
func (s *Signer) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmSecp256k1
}

// PublicKey returns the public key corresponding to the signing key (uncompressed, 65 bytes).
func (s *Signer) PublicKey() []byte {
	return s.pubKey.SerializeUncompressed()
}

// PublicKeyCompressed returns the compressed public key (33 bytes).
func (s *Signer) PublicKeyCompressed() []byte {
	return s.pubKey.SerializeCompressed()
}

// EthereumAddress returns the Ethereum address for this signer.
func (s *Signer) EthereumAddress() string {
	pubKeyBytes := s.pubKey.SerializeUncompressed()[1:] // Remove 0x04 prefix
	hash := Keccak256(pubKeyBytes)
	return "0x" + hexEncode(hash[12:])
}

// ============================================================================
// Verifier
// ============================================================================

// Verifier verifies secp256k1 signatures.
type Verifier struct {
	pubKey *secp256k1.PublicKey
}

// Ensure Verifier implements crypto.Verifier.
var _ crypto.Verifier = (*Verifier)(nil)

// NewVerifier creates a new secp256k1 verifier from a public key.
// Accepts both compressed (33 bytes) and uncompressed (65 bytes) formats.
func NewVerifier(publicKey []byte) (*Verifier, error) {
	const op = "secp256k1.NewVerifier"

	if len(publicKey) != PublicKeySize && len(publicKey) != PublicKeyUncompressedSize {
		return nil, crypto.ErrInvalidKey(op, fmt.Sprintf("invalid public key length: %d", len(publicKey)))
	}

	pubKey, err := secp256k1.ParsePubKey(publicKey)
	if err != nil {
		return nil, crypto.ErrInvalidKey(op, "invalid public key: "+err.Error())
	}

	return &Verifier{
		pubKey: pubKey,
	}, nil
}

// NewVerifierFromKeyPair creates a verifier from a key pair.
func NewVerifierFromKeyPair(kp *KeyPair) (*Verifier, error) {
	const op = "secp256k1.NewVerifierFromKeyPair"

	if kp == nil {
		return nil, crypto.ErrInvalidKey(op, "key pair is nil")
	}

	if kp.pubKey == nil {
		return nil, crypto.ErrInvalidKey(op, "key pair has no public key")
	}

	return &Verifier{
		pubKey: kp.pubKey,
	}, nil
}

// Verify checks if the signature is valid for the data.
// The data is hashed with Keccak256 before verification.
func (v *Verifier) Verify(data, signature []byte) error {
	hash := Keccak256(data)
	return v.VerifyHash(hash, signature)
}

// VerifyHash verifies a signature against a pre-hashed message.
// Signature should be 64 bytes [R || S] or 65 bytes [R || S || V].
func (v *Verifier) VerifyHash(hash, signature []byte) error {
	const op = "secp256k1.Verifier.VerifyHash"

	if len(hash) != 32 {
		return crypto.ErrVerificationFailed(op)
	}

	var r, ss []byte
	switch len(signature) {
	case SignatureSize:
		r = signature[0:32]
		ss = signature[32:64]
	case SignatureWithRecoverySize:
		r = signature[0:32]
		ss = signature[32:64]
		// V byte is ignored for direct verification
	default:
		return crypto.ErrVerificationFailed(op)
	}

	// Parse R and S
	var rScalar, sScalar secp256k1.ModNScalar
	if overflow := rScalar.SetByteSlice(r); overflow {
		return crypto.ErrVerificationFailed(op)
	}
	if overflow := sScalar.SetByteSlice(ss); overflow {
		return crypto.ErrVerificationFailed(op)
	}

	// Create and verify signature
	sig := ecdsa.NewSignature(&rScalar, &sScalar)
	if !sig.Verify(hash, v.pubKey) {
		return crypto.ErrVerificationFailed(op)
	}

	return nil
}

// Algorithm returns the verification algorithm.
func (v *Verifier) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmSecp256k1
}

// PublicKey returns the public key used for verification (uncompressed, 65 bytes).
func (v *Verifier) PublicKey() []byte {
	return v.pubKey.SerializeUncompressed()
}

// PublicKeyCompressed returns the compressed public key (33 bytes).
func (v *Verifier) PublicKeyCompressed() []byte {
	return v.pubKey.SerializeCompressed()
}

// ============================================================================
// Recovery Functions
// ============================================================================

// RecoverPublicKey recovers the public key from a signature and message hash.
// Signature must be 65 bytes [R || S || V] where V is 0/1 or 27/28.
// Returns the uncompressed public key (65 bytes).
func RecoverPublicKey(hash, signature []byte) ([]byte, error) {
	const op = "secp256k1.RecoverPublicKey"

	if len(hash) != 32 {
		return nil, crypto.ErrVerificationFailed(op)
	}

	if len(signature) != SignatureWithRecoverySize {
		return nil, crypto.ErrVerificationFailed(op)
	}

	// Extract R, S, V
	r := signature[0:32]
	s := signature[32:64]
	v := signature[64]

	// Normalize V to 0/1
	if v >= 27 {
		v -= 27
	}
	if v > 1 {
		return nil, crypto.ErrVerificationFailed(op)
	}

	// Convert to compact format [V || R || S] for RecoverCompact
	// RecoverCompact expects V as 27 or 28
	compactSig := make([]byte, SignatureWithRecoverySize)
	compactSig[0] = v + 27
	copy(compactSig[1:33], r)
	copy(compactSig[33:65], s)

	// Recover public key
	pubKey, _, err := ecdsa.RecoverCompact(compactSig, hash)
	if err != nil {
		return nil, crypto.ErrVerificationFailed(op)
	}

	return pubKey.SerializeUncompressed(), nil
}

// RecoverAddress recovers the Ethereum address from a signature and message hash.
// Signature must be 65 bytes [R || S || V] where V is 0/1 or 27/28.
func RecoverAddress(hash, signature []byte) (string, error) {
	const op = "secp256k1.RecoverAddress"

	pubKeyBytes, err := RecoverPublicKey(hash, signature)
	if err != nil {
		return "", err
	}

	// Remove 0x04 prefix from uncompressed public key
	pubKeyNoPrefix := pubKeyBytes[1:]

	// Keccak256 hash
	addressHash := Keccak256(pubKeyNoPrefix)

	// Take last 20 bytes
	address := addressHash[12:]

	return "0x" + hexEncode(address), nil
}

// RecoverAddressFromEthereumSignature recovers address from an Ethereum-style signature.
// The signature should have V as 27 or 28.
func RecoverAddressFromEthereumSignature(hash, signature []byte) (string, error) {
	if len(signature) != SignatureWithRecoverySize {
		return "", crypto.ErrVerificationFailed("secp256k1.RecoverAddressFromEthereumSignature")
	}

	// Make a copy and normalize V
	sigCopy := make([]byte, SignatureWithRecoverySize)
	copy(sigCopy, signature)

	v := sigCopy[64]
	if v >= 27 {
		sigCopy[64] = v - 27
	}

	return RecoverAddress(hash, sigCopy)
}
