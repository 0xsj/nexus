package vc

import (
	"time"
)

// Proof represents a cryptographic proof for a Verifiable Credential.
// Used in JSON-LD format credentials. For JWT format, the proof is the JWT signature.
type Proof struct {
	// Type specifies the proof type.
	Type ProofType `json:"type"`

	// Created is when the proof was created.
	Created time.Time `json:"created"`

	// VerificationMethod is the ID of the key used to create the proof.
	// Format: did:key:z6Mk...#z6Mk...
	VerificationMethod string `json:"verificationMethod"`

	// ProofPurpose indicates the purpose of the proof.
	// Common values: "assertionMethod", "authentication"
	ProofPurpose string `json:"proofPurpose"`

	// ProofValue is the signature value (base64 or multibase encoded).
	ProofValue string `json:"proofValue,omitempty"`

	// JWS is the JSON Web Signature (for JwtProof2020).
	JWS string `json:"jws,omitempty"`

	// Nonce is an optional nonce for replay protection.
	Nonce string `json:"nonce,omitempty"`

	// Domain is an optional domain binding.
	Domain string `json:"domain,omitempty"`

	// Challenge is an optional challenge for authentication proofs.
	Challenge string `json:"challenge,omitempty"`
}

// Proof purposes.
const (
	ProofPurposeAssertionMethod      = "assertionMethod"
	ProofPurposeAuthentication       = "authentication"
	ProofPurposeKeyAgreement         = "keyAgreement"
	ProofPurposeCapabilityInvocation = "capabilityInvocation"
	ProofPurposeCapabilityDelegation = "capabilityDelegation"
)

// NewProof creates a new proof with required fields.
func NewProof(proofType ProofType, verificationMethod string) *Proof {
	return &Proof{
		Type:               proofType,
		Created:            time.Now().UTC(),
		VerificationMethod: verificationMethod,
		ProofPurpose:       ProofPurposeAssertionMethod,
	}
}

// WithCreated sets the creation time.
func (p *Proof) WithCreated(t time.Time) *Proof {
	p.Created = t.UTC()
	return p
}

// WithProofPurpose sets the proof purpose.
func (p *Proof) WithProofPurpose(purpose string) *Proof {
	p.ProofPurpose = purpose
	return p
}

// WithProofValue sets the proof value (signature).
func (p *Proof) WithProofValue(value string) *Proof {
	p.ProofValue = value
	return p
}

// WithJWS sets the JWS signature.
func (p *Proof) WithJWS(jws string) *Proof {
	p.JWS = jws
	return p
}

// WithNonce sets the nonce.
func (p *Proof) WithNonce(nonce string) *Proof {
	p.Nonce = nonce
	return p
}

// WithDomain sets the domain.
func (p *Proof) WithDomain(domain string) *Proof {
	p.Domain = domain
	return p
}

// WithChallenge sets the challenge.
func (p *Proof) WithChallenge(challenge string) *Proof {
	p.Challenge = challenge
	return p
}

// Validate checks if the proof structure is valid.
func (p *Proof) Validate() error {
	const op = "Proof.Validate"

	if err := p.Type.Validate(); err != nil {
		return ErrInvalidProof(op, "invalid proof type: "+err.Error())
	}

	if p.Created.IsZero() {
		return ErrInvalidProof(op, "created is required")
	}

	if p.VerificationMethod == "" {
		return ErrInvalidProof(op, "verificationMethod is required")
	}

	if p.ProofPurpose == "" {
		return ErrInvalidProof(op, "proofPurpose is required")
	}

	// Must have either ProofValue or JWS
	if p.ProofValue == "" && p.JWS == "" {
		return ErrInvalidProof(op, "proofValue or jws is required")
	}

	return nil
}

// IsExpired checks if the proof has expired based on a maximum age.
func (p *Proof) IsExpired(maxAge time.Duration) bool {
	return time.Since(p.Created) > maxAge
}

// ============================================================================
// Presentation Proof
// ============================================================================

// PresentationProof represents a proof for a Verifiable Presentation.
// Similar to Proof but may include additional presentation-specific fields.
type PresentationProof struct {
	Proof

	// Holder is the DID of the presentation holder.
	Holder string `json:"holder,omitempty"`
}

// NewPresentationProof creates a new presentation proof.
func NewPresentationProof(proofType ProofType, verificationMethod string, holder string) *PresentationProof {
	return &PresentationProof{
		Proof: Proof{
			Type:               proofType,
			Created:            time.Now().UTC(),
			VerificationMethod: verificationMethod,
			ProofPurpose:       ProofPurposeAuthentication,
		},
		Holder: holder,
	}
}

// ============================================================================
// Derived Proof (for BBS+ selective disclosure)
// ============================================================================

// DerivedProof represents a derived proof for selective disclosure.
// Used with BBS+ signatures to reveal only selected claims.
type DerivedProof struct {
	// Type is always BbsBlsSignatureProof2020 for derived proofs.
	Type ProofType `json:"type"`

	// Created is when the derived proof was created.
	Created time.Time `json:"created"`

	// VerificationMethod is the ID of the original signing key.
	VerificationMethod string `json:"verificationMethod"`

	// ProofPurpose is the purpose of the proof.
	ProofPurpose string `json:"proofPurpose"`

	// ProofValue is the derived proof value.
	ProofValue string `json:"proofValue"`

	// Nonce is the verifier-provided nonce.
	Nonce string `json:"nonce"`

	// RevealedIndexes indicates which claims are revealed.
	RevealedIndexes []int `json:"-"`
}

// NewDerivedProof creates a new derived proof for selective disclosure.
func NewDerivedProof(verificationMethod string, nonce string) *DerivedProof {
	return &DerivedProof{
		Type:               ProofTypeBbsBlsSignatureProof2020,
		Created:            time.Now().UTC(),
		VerificationMethod: verificationMethod,
		ProofPurpose:       ProofPurposeAssertionMethod,
		Nonce:              nonce,
	}
}

// WithProofValue sets the derived proof value.
func (p *DerivedProof) WithProofValue(value string) *DerivedProof {
	p.ProofValue = value
	return p
}

// WithRevealedIndexes sets the revealed claim indexes.
func (p *DerivedProof) WithRevealedIndexes(indexes []int) *DerivedProof {
	p.RevealedIndexes = indexes
	return p
}

// Validate checks if the derived proof is valid.
func (p *DerivedProof) Validate() error {
	const op = "DerivedProof.Validate"

	if p.Type != ProofTypeBbsBlsSignatureProof2020 {
		return ErrInvalidProof(op, "derived proof must use BbsBlsSignatureProof2020")
	}

	if p.Created.IsZero() {
		return ErrInvalidProof(op, "created is required")
	}

	if p.VerificationMethod == "" {
		return ErrInvalidProof(op, "verificationMethod is required")
	}

	if p.ProofValue == "" {
		return ErrInvalidProof(op, "proofValue is required")
	}

	if p.Nonce == "" {
		return ErrInvalidProof(op, "nonce is required for derived proofs")
	}

	return nil
}
