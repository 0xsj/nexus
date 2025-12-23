package crypto

import "fmt"

// Algorithm identifies a cryptographic algorithm.
// Used to select key generation and signing implementations.
type Algorithm string

const (
	// AlgorithmEd25519 is the EdDSA signature algorithm using Curve25519.
	// Primary algorithm for DIDs and VC signing.
	// Fast, secure, small signatures (64 bytes).
	AlgorithmEd25519 Algorithm = "Ed25519"

	// AlgorithmSecp256k1 is the ECDSA algorithm using secp256k1 curve.
	// Used for Ethereum wallet compatibility (did:pkh, did:ethr).
	// Phase 2 implementation.
	AlgorithmSecp256k1 Algorithm = "secp256k1"

	// AlgorithmBLS12381 is the BLS signature algorithm.
	// Used for BBS+ signatures enabling selective disclosure.
	// Phase 2 implementation.
	AlgorithmBLS12381 Algorithm = "BLS12-381"
)

// String returns the string representation
func (a Algorithm) String() string {
	return string(a)
}

// Validate checks if the algorithm is supported
func (a Algorithm) Validate() error {
	switch a {
	case AlgorithmEd25519, AlgorithmSecp256k1, AlgorithmBLS12381:
		return nil
	default:
		return fmt.Errorf("unsupported algorithm: %s", a)
	}
}

// IsSupported returns true if the algorithm is currently implemented.
func (a Algorithm) IsSupported() bool {
	switch a {
	case AlgorithmEd25519:
		return true
	case AlgorithmSecp256k1, AlgorithmBLS12381:
		// Phase 2
		return false
	default:
		return false
	}
}

// KeySize returns the private key size in bytes.
func (a Algorithm) KeySize() int {
	switch a {
	case AlgorithmEd25519:
		return 32
	case AlgorithmSecp256k1:
		return 32
	case AlgorithmBLS12381:
		return 32
	default:
		return 0
	}
}

// PublicKeySize returns the public key size in bytes.
func (a Algorithm) PublicKeySize() int {
	switch a {
	case AlgorithmEd25519:
		return 32
	case AlgorithmSecp256k1:
		return 33 // compressed
	case AlgorithmBLS12381:
		return 48
	default:
		return 0
	}
}

// SignatureSize returns the signature size in bytes.
func (a Algorithm) SignatureSize() int {
	switch a {
	case AlgorithmEd25519:
		return 64
	case AlgorithmSecp256k1:
		return 64 // r + s
	case AlgorithmBLS12381:
		return 96
	default:
		return 0
	}
}

// JWAName returns the JSON Web Algorithm (JWA) name.
// Used in JWT headers and JWK representations.
func (a Algorithm) JWAName() string {
	switch a {
	case AlgorithmEd25519:
		return "EdDSA"
	case AlgorithmSecp256k1:
		return "ES256K"
	case AlgorithmBLS12381:
		return "BLS"
	default:
		return ""
	}
}

// MulticodecPrefix returns the multicodec prefix for the algorithm.
// Used in did:key encoding.
func (a Algorithm) MulticodecPrefix() []byte {
	switch a {
	case AlgorithmEd25519:
		// 0xed01 - ed25519-pub
		return []byte{0xed, 0x01}
	case AlgorithmSecp256k1:
		// 0xe701 - secp256k1-pub
		return []byte{0xe7, 0x01}
	case AlgorithmBLS12381:
		// 0xeb01 - bls12_381-g2-pub
		return []byte{0xeb, 0x01}
	default:
		return nil
	}
}
