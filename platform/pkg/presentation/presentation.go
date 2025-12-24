package presentation

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/types"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// PresentationType represents standard presentation types.
type PresentationType string

const (
	// TypeVerifiablePresentation is the base type for all VPs.
	TypeVerifiablePresentation PresentationType = "VerifiablePresentation"
)

// String returns the string representation.
func (p PresentationType) String() string {
	return string(p)
}

// Context URIs for JSON-LD presentations.
const (
	ContextCredentialsV1 = "https://www.w3.org/2018/credentials/v1"
	ContextCredentialsV2 = "https://www.w3.org/ns/credentials/v2"
)

// DefaultContext returns the default JSON-LD context for presentations.
func DefaultContext() []string {
	return []string{ContextCredentialsV1}
}

// Presentation represents a W3C Verifiable Presentation.
// A VP is how holders share credentials with verifiers.
type Presentation struct {
	// Context defines the JSON-LD context(s).
	Context []string `json:"@context"`

	// ID is the unique identifier for this presentation.
	ID string `json:"id,omitempty"`

	// Type specifies the presentation type(s).
	// Must include "VerifiablePresentation".
	Type []PresentationType `json:"type"`

	// Holder is the DID of the entity presenting the credentials.
	Holder did.DID `json:"holder"`

	// VerifiableCredential contains the credentials being presented.
	// Can be Credential objects or JWT strings.
	VerifiableCredential []vc.Credential `json:"verifiableCredential,omitempty"`

	// VerifiableCredentialJWT contains JWT-encoded credentials.
	// Used when presenting JWT-VCs.
	VerifiableCredentialJWT []string `json:"-"`

	// Proof contains the cryptographic proof (for JSON-LD format).
	// For JWT format, the proof is in the JWT signature.
	Proof *Proof `json:"proof,omitempty"`
}

// Proof represents a cryptographic proof for a Verifiable Presentation.
type Proof struct {
	// Type specifies the proof type.
	Type vc.ProofType `json:"type"`

	// Created is when the proof was created.
	Created time.Time `json:"created"`

	// VerificationMethod is the ID of the key used to create the proof.
	VerificationMethod string `json:"verificationMethod"`

	// ProofPurpose indicates the purpose of the proof.
	// Typically "authentication" for presentations.
	ProofPurpose string `json:"proofPurpose"`

	// ProofValue is the signature value.
	ProofValue string `json:"proofValue,omitempty"`

	// JWS is the JSON Web Signature (for JwtProof2020).
	JWS string `json:"jws,omitempty"`

	// Challenge is the verifier-provided challenge.
	Challenge string `json:"challenge,omitempty"`

	// Domain is the intended domain for this presentation.
	Domain string `json:"domain,omitempty"`

	// Nonce is an optional nonce for replay protection.
	Nonce string `json:"nonce,omitempty"`
}

// Proof purposes for presentations.
const (
	ProofPurposeAuthentication = "authentication"
)

// ============================================================================
// Presentation Builder
// ============================================================================

// New creates a new presentation with the given holder.
func New(holder did.DID) *Presentation {
	return &Presentation{
		Context: DefaultContext(),
		ID:      GeneratePresentationID(),
		Type:    []PresentationType{TypeVerifiablePresentation},
		Holder:  holder,
	}
}

// WithID sets the presentation ID.
func (p *Presentation) WithID(id string) *Presentation {
	p.ID = id
	return p
}

// WithCredentials adds credentials to the presentation.
func (p *Presentation) WithCredentials(credentials ...vc.Credential) *Presentation {
	p.VerifiableCredential = append(p.VerifiableCredential, credentials...)
	return p
}

// WithCredentialJWTs adds JWT-encoded credentials to the presentation.
func (p *Presentation) WithCredentialJWTs(jwts ...string) *Presentation {
	p.VerifiableCredentialJWT = append(p.VerifiableCredentialJWT, jwts...)
	return p
}

// WithType adds a presentation type.
func (p *Presentation) WithType(t PresentationType) *Presentation {
	p.Type = append(p.Type, t)
	return p
}

// WithContext adds a context URI.
func (p *Presentation) WithContext(ctx string) *Presentation {
	p.Context = append(p.Context, ctx)
	return p
}

// WithProof sets the proof.
func (p *Presentation) WithProof(proof *Proof) *Presentation {
	p.Proof = proof
	return p
}

// ============================================================================
// Presentation Validation
// ============================================================================

// Validate checks if the presentation structure is valid.
func (p *Presentation) Validate() error {
	const op = "Presentation.Validate"

	if len(p.Context) == 0 {
		return ErrInvalidPresentation(op, "context is required")
	}

	if len(p.Type) == 0 {
		return ErrInvalidPresentation(op, "type is required")
	}

	// Must include VerifiablePresentation type
	hasVP := false
	for _, t := range p.Type {
		if t == TypeVerifiablePresentation {
			hasVP = true
			break
		}
	}
	if !hasVP {
		return ErrInvalidPresentation(op, "type must include 'VerifiablePresentation'")
	}

	if p.Holder.IsZero() {
		return ErrInvalidHolder(op, "holder is required")
	}

	// Must have at least one credential
	if len(p.VerifiableCredential) == 0 && len(p.VerifiableCredentialJWT) == 0 {
		return ErrInvalidPresentation(op, "at least one credential is required")
	}

	return nil
}

// ============================================================================
// Presentation Queries
// ============================================================================

// CredentialCount returns the total number of credentials.
func (p *Presentation) CredentialCount() int {
	return len(p.VerifiableCredential) + len(p.VerifiableCredentialJWT)
}

// HasCredentials returns true if the presentation contains any credentials.
func (p *Presentation) HasCredentials() bool {
	return p.CredentialCount() > 0
}

// HasType returns true if the presentation has the specified type.
func (p *Presentation) HasType(t PresentationType) bool {
	for _, pt := range p.Type {
		if pt == t {
			return true
		}
	}
	return false
}

// GetCredentialByID finds a credential by ID.
func (p *Presentation) GetCredentialByID(id string) *vc.Credential {
	for i := range p.VerifiableCredential {
		if p.VerifiableCredential[i].ID == id {
			return &p.VerifiableCredential[i]
		}
	}
	return nil
}

// GetCredentialsByType returns credentials matching the type.
func (p *Presentation) GetCredentialsByType(credType vc.CredentialType) []vc.Credential {
	var result []vc.Credential
	for _, cred := range p.VerifiableCredential {
		if cred.HasType(credType) {
			result = append(result, cred)
		}
	}
	return result
}

// HolderDID returns the DID of the holder.
func (p *Presentation) HolderDID() did.DID {
	return p.Holder
}

// ============================================================================
// Proof Builder
// ============================================================================

// NewProof creates a new presentation proof.
func NewProof(proofType vc.ProofType, verificationMethod string) *Proof {
	return &Proof{
		Type:               proofType,
		Created:            time.Now().UTC(),
		VerificationMethod: verificationMethod,
		ProofPurpose:       ProofPurposeAuthentication,
	}
}

// WithChallenge sets the challenge.
func (p *Proof) WithChallenge(challenge string) *Proof {
	p.Challenge = challenge
	return p
}

// WithDomain sets the domain.
func (p *Proof) WithDomain(domain string) *Proof {
	p.Domain = domain
	return p
}

// WithNonce sets the nonce.
func (p *Proof) WithNonce(nonce string) *Proof {
	p.Nonce = nonce
	return p
}

// WithProofValue sets the proof value.
func (p *Proof) WithProofValue(value string) *Proof {
	p.ProofValue = value
	return p
}

// WithJWS sets the JWS signature.
func (p *Proof) WithJWS(jws string) *Proof {
	p.JWS = jws
	return p
}

// Validate checks if the proof is valid.
func (p *Proof) Validate() error {
	const op = "Proof.Validate"

	if p.Type == "" {
		return ErrInvalidPresentation(op, "proof type is required")
	}

	if p.Created.IsZero() {
		return ErrInvalidPresentation(op, "proof created is required")
	}

	if p.VerificationMethod == "" {
		return ErrInvalidPresentation(op, "proof verificationMethod is required")
	}

	if p.ProofPurpose == "" {
		return ErrInvalidPresentation(op, "proof proofPurpose is required")
	}

	if p.ProofValue == "" && p.JWS == "" {
		return ErrInvalidPresentation(op, "proof value or JWS is required")
	}

	return nil
}

// ============================================================================
// ID Generation
// ============================================================================

// GeneratePresentationID generates a unique presentation ID.
func GeneratePresentationID() string {
	id := types.NewID()
	return "urn:uuid:" + id.String()
}
