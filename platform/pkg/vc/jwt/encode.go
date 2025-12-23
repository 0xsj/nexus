package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Signer signs Verifiable Credentials as JWTs.
type Signer struct {
	issuerDID            did.DID
	signer               crypto.Signer
	verificationMethodID string
	issuerName           string
	issuerURL            string
}

// Ensure Signer implements vc.Signer.
var _ vc.Signer = (*Signer)(nil)

// NewSigner creates a new JWT-VC signer.
func NewSigner(config vc.SignerConfig) (*Signer, error) {
	const op = "jwt.NewSigner"

	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.Format != vc.FormatJWT {
		return nil, vc.ErrUnsupportedFormat(op, config.Format.String())
	}

	return &Signer{
		issuerDID:            config.IssuerDID,
		signer:               config.Signer,
		verificationMethodID: config.VerificationMethod(),
		issuerName:           config.IssuerName,
		issuerURL:            config.IssuerURL,
	}, nil
}

// Sign signs a credential and returns the JWT string.
func (s *Signer) Sign(credential *vc.Credential) ([]byte, error) {
	const op = "jwt.Signer.Sign"

	// Validate credential
	if err := credential.Validate(); err != nil {
		return nil, err
	}

	// Set issuer if not set
	if credential.Issuer.ID.IsZero() {
		credential.Issuer.ID = s.issuerDID
	}
	if credential.Issuer.Name == "" && s.issuerName != "" {
		credential.Issuer.Name = s.issuerName
	}
	if credential.Issuer.URL == "" && s.issuerURL != "" {
		credential.Issuer.URL = s.issuerURL
	}

	// Build JWT claims
	claims := s.buildClaims(credential)

	// Build header
	header := s.buildHeader()

	// Encode header and claims
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return nil, vc.ErrSigningFailed(op, fmt.Errorf("failed to marshal header: %w", err))
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, vc.ErrSigningFailed(op, fmt.Errorf("failed to marshal claims: %w", err))
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Sign
	signingInput := headerB64 + "." + claimsB64
	signature, err := s.signer.Sign([]byte(signingInput))
	if err != nil {
		return nil, vc.ErrSigningFailed(op, err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	// Build JWT
	jwt := signingInput + "." + signatureB64

	return []byte(jwt), nil
}

// Format returns the output format.
func (s *Signer) Format() vc.Format {
	return vc.FormatJWT
}

// IssuerDID returns the issuer DID.
func (s *Signer) IssuerDID() did.DID {
	return s.issuerDID
}

// Algorithm returns the signing algorithm.
func (s *Signer) Algorithm() crypto.Algorithm {
	return s.signer.Algorithm()
}

// buildHeader builds the JWT header.
func (s *Signer) buildHeader() Header {
	return Header{
		Alg: s.signer.Algorithm().JWAName(),
		Typ: "JWT",
		Kid: s.verificationMethodID,
	}
}

// buildClaims builds the JWT claims from a credential.
func (s *Signer) buildClaims(credential *vc.Credential) Claims {
	claims := Claims{
		Issuer:    credential.Issuer.ID.String(),
		Subject:   credential.CredentialSubject.ID.String(),
		IssuedAt:  credential.IssuanceDate.Unix(),
		NotBefore: credential.IssuanceDate.Unix(),
		JWTID:     credential.ID,
		VC:        credentialToVC(credential),
	}

	if credential.ExpirationDate != nil {
		exp := credential.ExpirationDate.Unix()
		claims.ExpiresAt = &exp
	}

	return claims
}

// credentialToVC converts a credential to the VC claim format.
func credentialToVC(credential *vc.Credential) VCClaim {
	return VCClaim{
		Context:           credential.Context,
		Type:              credentialTypesToStrings(credential.Type),
		CredentialSubject: credential.CredentialSubject,
		CredentialStatus:  credential.CredentialStatus,
		CredentialSchema:  credential.CredentialSchema,
		Evidence:          credential.Evidence,
	}
}

// credentialTypesToStrings converts credential types to strings.
func credentialTypesToStrings(types []vc.CredentialType) []string {
	strs := make([]string, len(types))
	for i, t := range types {
		strs[i] = t.String()
	}
	return strs
}

// ============================================================================
// JWT Structures
// ============================================================================

// Header represents a JWT header.
type Header struct {
	// Alg is the signing algorithm.
	Alg string `json:"alg"`

	// Typ is the token type (always "JWT").
	Typ string `json:"typ"`

	// Kid is the key ID (verification method ID).
	Kid string `json:"kid,omitempty"`
}

// Claims represents JWT claims for a Verifiable Credential.
type Claims struct {
	// Standard JWT claims
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	ExpiresAt *int64 `json:"exp,omitempty"`
	JWTID     string `json:"jti,omitempty"`

	// VC claim contains the credential
	VC VCClaim `json:"vc"`
}

// VCClaim represents the "vc" claim in a JWT-VC.
type VCClaim struct {
	Context           []string             `json:"@context"`
	Type              []string             `json:"type"`
	CredentialSubject vc.Subject           `json:"credentialSubject"`
	CredentialStatus  *vc.CredentialStatus `json:"credentialStatus,omitempty"`
	CredentialSchema  *vc.CredentialSchema `json:"credentialSchema,omitempty"`
	Evidence          []vc.Evidence        `json:"evidence,omitempty"`
}

// ============================================================================
// Helper Functions
// ============================================================================

// EncodeCredential encodes a credential as a JWT without signing.
// Useful for testing or when signing is done externally.
func EncodeCredential(credential *vc.Credential, algorithm string, keyID string) (string, error) {
	const op = "jwt.EncodeCredential"

	header := Header{
		Alg: algorithm,
		Typ: "JWT",
		Kid: keyID,
	}

	claims := Claims{
		Issuer:    credential.Issuer.ID.String(),
		Subject:   credential.CredentialSubject.ID.String(),
		IssuedAt:  credential.IssuanceDate.Unix(),
		NotBefore: credential.IssuanceDate.Unix(),
		JWTID:     credential.ID,
		VC:        credentialToVC(credential),
	}

	if credential.ExpirationDate != nil {
		exp := credential.ExpirationDate.Unix()
		claims.ExpiresAt = &exp
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", vc.ErrSigningFailed(op, fmt.Errorf("failed to marshal header: %w", err))
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", vc.ErrSigningFailed(op, fmt.Errorf("failed to marshal claims: %w", err))
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return headerB64 + "." + claimsB64, nil
}

// CreateJWT creates a complete JWT from header, claims, and signature.
func CreateJWT(header Header, claims Claims, signature []byte) string {
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return headerB64 + "." + claimsB64 + "." + signatureB64
}

// TimestampToTime converts a Unix timestamp to time.Time.
func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).UTC()
}
