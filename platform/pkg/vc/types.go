package vc

import "fmt"

// Format represents the credential encoding format.
type Format string

const (
	// FormatJWT is JWT-encoded Verifiable Credential.
	// Compact, widely supported, good for transport.
	FormatJWT Format = "jwt_vc"

	// FormatJSONLD is JSON-LD encoded Verifiable Credential.
	// Richer semantics, supports linked data proofs.
	FormatJSONLD Format = "ldp_vc"
)

// String returns the string representation.
func (f Format) String() string {
	return string(f)
}

// Validate checks if the format is supported.
func (f Format) Validate() error {
	switch f {
	case FormatJWT, FormatJSONLD:
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", f)
	}
}

// IsSupported returns true if the format is currently implemented.
func (f Format) IsSupported() bool {
	switch f {
	case FormatJWT:
		return true
	case FormatJSONLD:
		// Phase 2
		return false
	default:
		return false
	}
}

// ProofType represents the type of cryptographic proof.
type ProofType string

const (
	// ProofTypeJWT indicates a JWT-based proof (signature in JWT).
	ProofTypeJWT ProofType = "JwtProof2020"

	// ProofTypeEd25519Signature2020 is an Ed25519 linked data proof.
	ProofTypeEd25519Signature2020 ProofType = "Ed25519Signature2020"

	// ProofTypeEcdsaSecp256k1Signature2019 is a secp256k1 linked data proof.
	ProofTypeEcdsaSecp256k1Signature2019 ProofType = "EcdsaSecp256k1Signature2019"

	// ProofTypeBbsBlsSignature2020 is a BBS+ signature proof.
	ProofTypeBbsBlsSignature2020 ProofType = "BbsBlsSignature2020"

	// ProofTypeBbsBlsSignatureProof2020 is a BBS+ derived proof (selective disclosure).
	ProofTypeBbsBlsSignatureProof2020 ProofType = "BbsBlsSignatureProof2020"
)

// String returns the string representation.
func (p ProofType) String() string {
	return string(p)
}

// Validate checks if the proof type is recognized.
func (p ProofType) Validate() error {
	switch p {
	case ProofTypeJWT,
		ProofTypeEd25519Signature2020,
		ProofTypeEcdsaSecp256k1Signature2019,
		ProofTypeBbsBlsSignature2020,
		ProofTypeBbsBlsSignatureProof2020:
		return nil
	default:
		return fmt.Errorf("unsupported proof type: %s", p)
	}
}

// IsSupported returns true if the proof type is currently implemented.
func (p ProofType) IsSupported() bool {
	switch p {
	case ProofTypeJWT, ProofTypeEd25519Signature2020:
		return true
	case ProofTypeEcdsaSecp256k1Signature2019,
		ProofTypeBbsBlsSignature2020,
		ProofTypeBbsBlsSignatureProof2020:
		// Phase 2
		return false
	default:
		return false
	}
}

// SupportsSelectiveDisclosure returns true if the proof type supports selective disclosure.
func (p ProofType) SupportsSelectiveDisclosure() bool {
	switch p {
	case ProofTypeBbsBlsSignature2020, ProofTypeBbsBlsSignatureProof2020:
		return true
	default:
		return false
	}
}

// Status represents the status of a credential.
type Status string

const (
	// StatusActive indicates the credential is valid and active.
	StatusActive Status = "active"

	// StatusRevoked indicates the credential has been revoked.
	StatusRevoked Status = "revoked"

	// StatusExpired indicates the credential has expired.
	StatusExpired Status = "expired"

	// StatusSuspended indicates the credential is temporarily suspended.
	StatusSuspended Status = "suspended"
)

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// Validate checks if the status is valid.
func (s Status) Validate() error {
	switch s {
	case StatusActive, StatusRevoked, StatusExpired, StatusSuspended:
		return nil
	default:
		return fmt.Errorf("invalid status: %s", s)
	}
}

// IsValid returns true if the credential is usable (active).
func (s Status) IsValid() bool {
	return s == StatusActive
}

// CredentialType represents standard credential types.
type CredentialType string

const (
	// TypeVerifiableCredential is the base type for all VCs.
	TypeVerifiableCredential CredentialType = "VerifiableCredential"

	// TypeVerifiablePresentation is the base type for all VPs.
	TypeVerifiablePresentation CredentialType = "VerifiablePresentation"

	// TypeGitHubContributorCredential is for GitHub contribution verification.
	TypeGitHubContributorCredential CredentialType = "GitHubContributorCredential"

	// TypeLinkedInEmploymentCredential is for LinkedIn employment verification.
	TypeLinkedInEmploymentCredential CredentialType = "LinkedInEmploymentCredential"

	// TypeEducationCredential is for education/course completion.
	TypeEducationCredential CredentialType = "EducationCredential"

	// TypeCertificationCredential is for professional certifications.
	TypeCertificationCredential CredentialType = "CertificationCredential"

	// TypeIdentityCredential is for identity verification.
	TypeIdentityCredential CredentialType = "IdentityCredential"

	// TypeSkillCredential is for verified skills.
	TypeSkillCredential CredentialType = "SkillCredential"
)

// String returns the string representation.
func (c CredentialType) String() string {
	return string(c)
}

// Context URIs for JSON-LD credentials.
const (
	ContextCredentialsV1 = "https://www.w3.org/2018/credentials/v1"
	ContextCredentialsV2 = "https://www.w3.org/ns/credentials/v2"
	ContextEd25519       = "https://w3id.org/security/suites/ed25519-2020/v1"
	ContextBBS           = "https://w3id.org/security/bbs/v1"
)

// DefaultContext returns the default JSON-LD context for credentials.
func DefaultContext() []string {
	return []string{
		ContextCredentialsV1,
		ContextEd25519,
	}
}
