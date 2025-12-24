package presentation

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Signer signs Verifiable Presentations.
type Signer interface {
	// Sign signs a presentation and returns the signed bytes.
	Sign(presentation *Presentation, opts SigningOptions) ([]byte, error)

	// SignWithRequest signs a presentation using parameters from a request.
	SignWithRequest(presentation *Presentation, request *Request) ([]byte, error)

	// HolderDID returns the DID of the holder.
	HolderDID() did.DID

	// Algorithm returns the signing algorithm.
	Algorithm() crypto.Algorithm

	// Format returns the output format (JWT or JSON-LD).
	Format() vc.Format
}

// SigningOptions configures presentation signing.
type SigningOptions struct {
	// Challenge is the verifier-provided challenge.
	Challenge string

	// Domain is the intended audience domain.
	Domain string

	// Nonce is an optional additional nonce.
	Nonce string

	// Created overrides the proof creation time.
	// If zero, uses current time.
	Created int64
}

// DefaultSigningOptions returns default signing options.
func DefaultSigningOptions() SigningOptions {
	return SigningOptions{}
}

// WithChallenge sets the challenge.
func (o SigningOptions) WithChallenge(challenge string) SigningOptions {
	o.Challenge = challenge
	return o
}

// WithDomain sets the domain.
func (o SigningOptions) WithDomain(domain string) SigningOptions {
	o.Domain = domain
	return o
}

// WithNonce sets the nonce.
func (o SigningOptions) WithNonce(nonce string) SigningOptions {
	o.Nonce = nonce
	return o
}

// WithCreated sets the creation timestamp.
func (o SigningOptions) WithCreated(timestamp int64) SigningOptions {
	o.Created = timestamp
	return o
}

// OptionsFromRequest creates signing options from a presentation request.
func OptionsFromRequest(r *Request) SigningOptions {
	return SigningOptions{
		Challenge: r.Challenge,
		Domain:    r.Domain,
	}
}

// ============================================================================
// Signer Configuration
// ============================================================================

// SignerConfig configures a presentation signer.
type SignerConfig struct {
	// HolderDID is the DID of the holder.
	HolderDID did.DID

	// KeyPair is the holder's key pair.
	KeyPair crypto.KeyPair

	// Signer is the cryptographic signer.
	Signer crypto.Signer

	// Format is the output format.
	Format vc.Format

	// VerificationMethod is the key ID for the proof.
	// If empty, derived from HolderDID.
	VerificationMethod string
}

// Validate checks if the config is valid.
func (c SignerConfig) Validate() error {
	const op = "SignerConfig.Validate"

	if c.HolderDID.IsZero() {
		return ErrInvalidHolder(op, "holder DID is required")
	}

	if c.KeyPair == nil {
		return ErrSigningFailed(op, errKeyPairRequired)
	}

	if c.Signer == nil {
		return ErrSigningFailed(op, errSignerRequired)
	}

	return nil
}

// GetVerificationMethod returns the verification method ID.
func (c SignerConfig) GetVerificationMethod() string {
	if c.VerificationMethod != "" {
		return c.VerificationMethod
	}
	// Default: DID + key fragment
	return c.HolderDID.String() + "#" + c.HolderDID.MethodSpecificID()
}

// Sentinel errors for config validation.
var (
	errKeyPairRequired = &configError{msg: "key pair is required"}
	errSignerRequired  = &configError{msg: "signer is required"}
)

type configError struct {
	msg string
}

func (e *configError) Error() string {
	return e.msg
}
