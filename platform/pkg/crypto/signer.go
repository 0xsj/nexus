package crypto

// Signer signs data using a private key.
type Signer interface {
	// Sign signs the data and returns the signature.
	Sign(data []byte) ([]byte, error)

	// Algorithm returns the signing algorithm.
	Algorithm() Algorithm

	// PublicKey returns the public key corresponding to the signing key.
	PublicKey() []byte
}

// Verifier verifies signatures using a public key.
type Verifier interface {
	// Verify checks if the signature is valid for the data.
	// Returns nil if valid, error if invalid.
	Verify(data, signature []byte) error

	// Algorithm returns the verification algorithm.
	Algorithm() Algorithm

	// PublicKey returns the public key used for verification.
	PublicKey() []byte
}

// SignerVerifier combines signing and verification capabilities.
type SignerVerifier interface {
	Signer
	Verifier
}

// SignerFactory creates signers and verifiers.
type SignerFactory interface {
	// NewSigner creates a signer from a key pair.
	NewSigner(kp KeyPair) (Signer, error)

	// NewVerifier creates a verifier from a public key.
	NewVerifier(publicKey []byte) (Verifier, error)

	// Algorithm returns the algorithm this factory handles.
	Algorithm() Algorithm
}

// SignatureInfo contains metadata about a signature.
type SignatureInfo struct {
	// Algorithm used for signing.
	Algorithm Algorithm

	// PublicKey that can verify the signature.
	PublicKey []byte

	// Signature bytes.
	Signature []byte
}

// SignOptions configures signing behavior.
type SignOptions struct {
	// Detached indicates whether to produce a detached signature.
	// If true, the signature does not include the original data.
	Detached bool

	// Context is optional context data bound to the signature.
	// Used by some algorithms (e.g., Ed25519ctx).
	Context []byte
}

// DefaultSignOptions returns default signing options.
func DefaultSignOptions() SignOptions {
	return SignOptions{
		Detached: true,
		Context:  nil,
	}
}

// VerifyOptions configures verification behavior.
type VerifyOptions struct {
	// Context is optional context data that was bound to the signature.
	Context []byte
}

// DefaultVerifyOptions returns default verification options.
func DefaultVerifyOptions() VerifyOptions {
	return VerifyOptions{
		Context: nil,
	}
}
