package adapters

import (
	"context"
	"fmt"

	credentialdomain "github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/vc"
	vcjwt "github.com/0xsj/nexus/platform/pkg/vc/jwt"
)

// Compile-time interface check.
var _ credentialdomain.CredentialSigner = (*JWTCredentialSigner)(nil)

// JWTCredentialSigner signs credentials as JWTs using Ed25519.
// Generates a fresh issuer keypair on startup. In production this would
// load from Vault/KMS.
type JWTCredentialSigner struct {
	vcSigner  vc.Signer
	issuerDID did.DID
}

// NewJWTCredentialSigner creates a new JWTCredentialSigner with a fresh Ed25519 keypair.
func NewJWTCredentialSigner() (*JWTCredentialSigner, error) {
	kp, err := ed25519.Generate()
	if err != nil {
		return nil, fmt.Errorf("generating issuer keypair: %w", err)
	}

	issuerDID, err := key.FromKeyPair(kp)
	if err != nil {
		return nil, fmt.Errorf("deriving issuer DID: %w", err)
	}

	cryptoSigner, err := ed25519.NewSigner(kp)
	if err != nil {
		return nil, fmt.Errorf("creating crypto signer: %w", err)
	}

	vcSigner, err := vcjwt.NewSigner(vc.SignerConfig{
		IssuerDID: issuerDID,
		KeyPair:   kp,
		Signer:    cryptoSigner,
		Format:    vc.FormatJWT,
	})
	if err != nil {
		return nil, fmt.Errorf("creating VC JWT signer: %w", err)
	}

	return &JWTCredentialSigner{
		vcSigner:  vcSigner,
		issuerDID: issuerDID,
	}, nil
}

// Sign signs a credential and returns the JWT string.
func (s *JWTCredentialSigner) Sign(_ context.Context, cred *credentialdomain.Credential) (string, error) {
	subjectDID, err := did.Parse(cred.SubjectDID())
	if err != nil {
		return "", fmt.Errorf("parsing subject DID: %w", err)
	}

	subject := vc.NewSubject(subjectDID).WithClaims(cred.Claims().ToMap())

	vcCred := vc.NewCredential(cred.ID().String(), s.issuerDID, subject).
		WithType(vc.CredentialType(cred.Type().String()))

	if !cred.ExpiresAt().IsZero() {
		vcCred = vcCred.WithExpirationDate(cred.ExpiresAt())
	}

	jwtBytes, err := s.vcSigner.Sign(vcCred)
	if err != nil {
		return "", fmt.Errorf("signing credential: %w", err)
	}

	return string(jwtBytes), nil
}
