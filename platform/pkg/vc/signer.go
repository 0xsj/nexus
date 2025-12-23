package vc

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/did"
)

// Signer signs Verifiable Credentials.
type Signer interface {
	// Sign signs a credential and returns the signed representation.
	// For JWT format, returns the JWT string.
	// For JSON-LD format, returns the credential with embedded proof.
	Sign(credential *Credential) ([]byte, error)

	// Format returns the output format (JWT, JSON-LD).
	Format() Format

	// IssuerDID returns the DID of the issuer.
	IssuerDID() did.DID

	// Algorithm returns the signing algorithm.
	Algorithm() crypto.Algorithm
}

// SignerConfig configures a credential signer.
type SignerConfig struct {
	// IssuerDID is the DID of the credential issuer.
	IssuerDID did.DID

	// KeyPair is the signing key pair.
	KeyPair crypto.KeyPair

	// Signer is the cryptographic signer.
	Signer crypto.Signer

	// Format is the output format.
	Format Format

	// VerificationMethodID is the ID of the verification method.
	// If empty, defaults to IssuerDID#<key-fragment>
	VerificationMethodID string

	// IssuerName is the human-readable issuer name (optional).
	IssuerName string

	// IssuerURL is the issuer's website (optional).
	IssuerURL string
}

// Validate checks if the signer config is valid.
func (c *SignerConfig) Validate() error {
	const op = "SignerConfig.Validate"

	if c.IssuerDID.IsZero() {
		return ErrInvalidIssuer(op, "issuer DID is required")
	}

	if c.KeyPair == nil {
		return ErrInvalidCredential(op, "key pair is required")
	}

	if c.Signer == nil {
		return ErrInvalidCredential(op, "signer is required")
	}

	if err := c.Format.Validate(); err != nil {
		return ErrUnsupportedFormat(op, c.Format.String())
	}

	return nil
}

// VerificationMethod returns the verification method ID.
func (c *SignerConfig) VerificationMethod() string {
	if c.VerificationMethodID != "" {
		return c.VerificationMethodID
	}
	// Default: DID with key fragment
	return c.IssuerDID.Fragment(c.IssuerDID.MethodSpecificID())
}

// SigningOptions configures individual signing operations.
type SigningOptions struct {
	// Created overrides the proof creation time.
	Created *string

	// Nonce is an optional nonce for replay protection.
	Nonce string

	// Domain is an optional domain binding.
	Domain string

	// Challenge is an optional challenge for authentication.
	Challenge string

	// ProofPurpose overrides the default proof purpose.
	ProofPurpose string
}

// DefaultSigningOptions returns default signing options.
func DefaultSigningOptions() SigningOptions {
	return SigningOptions{
		ProofPurpose: ProofPurposeAssertionMethod,
	}
}

// WithNonce sets the nonce.
func (o SigningOptions) WithNonce(nonce string) SigningOptions {
	o.Nonce = nonce
	return o
}

// WithDomain sets the domain.
func (o SigningOptions) WithDomain(domain string) SigningOptions {
	o.Domain = domain
	return o
}

// WithChallenge sets the challenge.
func (o SigningOptions) WithChallenge(challenge string) SigningOptions {
	o.Challenge = challenge
	return o
}

// WithProofPurpose sets the proof purpose.
func (o SigningOptions) WithProofPurpose(purpose string) SigningOptions {
	o.ProofPurpose = purpose
	return o
}

// ============================================================================
// Signer Factory
// ============================================================================

// SignerFactory creates credential signers.
type SignerFactory interface {
	// CreateSigner creates a signer for the specified format.
	CreateSigner(config SignerConfig) (Signer, error)

	// SupportedFormats returns the formats this factory supports.
	SupportedFormats() []Format
}

// ============================================================================
// Multi-Signer
// ============================================================================

// MultiSigner can sign credentials in multiple formats.
type MultiSigner struct {
	signers map[Format]Signer
}

// NewMultiSigner creates a new multi-format signer.
func NewMultiSigner() *MultiSigner {
	return &MultiSigner{
		signers: make(map[Format]Signer),
	}
}

// Register registers a signer for a format.
func (m *MultiSigner) Register(signer Signer) {
	m.signers[signer.Format()] = signer
}

// Sign signs a credential in the specified format.
func (m *MultiSigner) Sign(credential *Credential, format Format) ([]byte, error) {
	const op = "MultiSigner.Sign"

	signer, ok := m.signers[format]
	if !ok {
		return nil, ErrUnsupportedFormat(op, format.String())
	}

	return signer.Sign(credential)
}

// SupportedFormats returns all registered formats.
func (m *MultiSigner) SupportedFormats() []Format {
	formats := make([]Format, 0, len(m.signers))
	for f := range m.signers {
		formats = append(formats, f)
	}
	return formats
}

// SupportsFormat returns true if the format is registered.
func (m *MultiSigner) SupportsFormat(format Format) bool {
	_, ok := m.signers[format]
	return ok
}
