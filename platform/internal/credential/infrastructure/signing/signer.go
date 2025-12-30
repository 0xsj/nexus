package signing

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
	vcjwt "github.com/0xsj/nexus/platform/pkg/vc/jwt"
)

// ============================================================================
// Signer
// ============================================================================

// Signer implements domain.SigningService using pkg/vc.
type Signer struct {
	issuerDID did.DID
	jwtSigner *vcjwt.Signer
}

// NewSigner creates a new Signer.
func NewSigner(issuerDID did.DID, keyPair crypto.KeyPair, cryptoSigner crypto.Signer) (*Signer, error) {
	if issuerDID.IsZero() {
		return nil, fmt.Errorf("issuer DID is required")
	}
	if keyPair == nil {
		return nil, fmt.Errorf("key pair is required")
	}
	if cryptoSigner == nil {
		return nil, fmt.Errorf("crypto signer is required")
	}

	// Create VC signer config
	config := vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   keyPair,
		Signer:    cryptoSigner,
		Format:    vc.FormatJWT,
	}

	// Create JWT signer
	jwtSigner, err := vcjwt.NewSigner(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT signer: %w", err)
	}

	return &Signer{
		issuerDID: issuerDID,
		jwtSigner: jwtSigner,
	}, nil
}

// ============================================================================
// SigningService Implementation
// ============================================================================

// SignCredential signs a credential and returns the JWT-VC string.
func (s *Signer) SignCredential(ctx context.Context, params domain.SigningParams) (string, error) {
	// Parse holder DID
	holderDID, err := did.Parse(params.HolderDID)
	if err != nil {
		return "", fmt.Errorf("invalid holder DID: %w", err)
	}

	// Build credential ID as URI
	credentialID := params.CredentialID
	if !isURI(credentialID) {
		credentialID = "urn:uuid:" + credentialID
	}

	// Build subject
	subject := vc.Subject{
		ID:     holderDID,
		Claims: params.Claims,
	}

	// Build credential using the builder
	credential := vc.NewCredential(credentialID, s.issuerDID, subject).
		WithIssuanceDate(params.IssuedAt)

	// Add credential type
	if params.CredentialType != "" {
		credential.WithType(vc.CredentialType(params.CredentialType))
	}

	// Set expiration if provided
	if params.ExpiresAt != nil {
		credential.WithExpirationDate(*params.ExpiresAt)
	}

	// Set schema if provided
	if params.SchemaID != "" {
		credential.WithSchema(&vc.CredentialSchema{
			ID:   params.SchemaID,
			Type: "JsonSchema",
		})
	}

	// Sign the credential
	jwtBytes, err := s.jwtSigner.Sign(credential)
	if err != nil {
		return "", fmt.Errorf("failed to sign credential: %w", err)
	}

	return string(jwtBytes), nil
}

// IssuerDID returns the issuer DID.
func (s *Signer) IssuerDID() did.DID {
	return s.issuerDID
}

// ============================================================================
// Factory Functions
// ============================================================================

// Config holds configuration for creating a Signer.
type Config struct {
	// IssuerDID is the DID of the issuer (as string).
	IssuerDID string

	// KeyPair is the signing key pair.
	KeyPair crypto.KeyPair

	// Signer is the cryptographic signer.
	Signer crypto.Signer
}

// NewSignerFromConfig creates a Signer from configuration.
func NewSignerFromConfig(cfg Config) (*Signer, error) {
	if cfg.IssuerDID == "" {
		return nil, fmt.Errorf("issuer DID is required")
	}

	// Parse issuer DID
	issuerDID, err := did.Parse(cfg.IssuerDID)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer DID: %w", err)
	}

	return NewSigner(issuerDID, cfg.KeyPair, cfg.Signer)
}

// ============================================================================
// Helpers
// ============================================================================

// isURI checks if a string is a valid URI.
func isURI(s string) bool {
	if len(s) < 4 {
		return false
	}
	return s[:4] == "urn:" || s[:4] == "did:" || s[:4] == "http" || (len(s) >= 5 && s[:5] == "https")
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.SigningService = (*Signer)(nil)
