package jwt

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/presentation"
	"github.com/0xsj/nexus/platform/pkg/vc"
)

// Signer signs Verifiable Presentations as JWTs.
type Signer struct {
	holderDID          did.DID
	keyPair            crypto.KeyPair
	signer             crypto.Signer
	verificationMethod string
}

// Ensure Signer implements presentation.Signer.
var _ presentation.Signer = (*Signer)(nil)

// NewSigner creates a new JWT presentation signer.
func NewSigner(config presentation.SignerConfig) (*Signer, error) {
	const op = "jwt.NewSigner"

	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.Format != vc.FormatJWT && config.Format != "" {
		return nil, presentation.ErrUnsupportedFormat(op, string(config.Format))
	}

	return &Signer{
		holderDID:          config.HolderDID,
		keyPair:            config.KeyPair,
		signer:             config.Signer,
		verificationMethod: config.GetVerificationMethod(),
	}, nil
}

// Sign signs a presentation and returns the JWT bytes.
func (s *Signer) Sign(p *presentation.Presentation, opts presentation.SigningOptions) ([]byte, error) {
	const op = "jwt.Signer.Sign"

	// Set holder if not set
	if p.Holder.IsZero() {
		p.Holder = s.holderDID
	}

	// Validate presentation
	if err := p.Validate(); err != nil {
		return nil, err
	}

	// Build JWT header
	header := Header{
		Alg: algorithmToJWA(s.signer.Algorithm()),
		Typ: "JWT",
		Kid: s.verificationMethod,
	}

	// Determine timestamps
	now := time.Now().Unix()
	if opts.Created != 0 {
		now = opts.Created
	}

	// Build JWT claims
	claims := Claims{
		Issuer:    s.holderDID.String(), // In VP, holder is the "issuer" of the presentation
		Subject:   s.holderDID.String(),
		IssuedAt:  now,
		NotBefore: now,
		JWTID:     p.ID,
		Nonce:     opts.Nonce,
		VP:        buildVPClaim(p),
	}

	// Encode header and claims
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return nil, presentation.ErrSigningFailed(op, err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, presentation.ErrSigningFailed(op, err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signing input
	signingInput := headerB64 + "." + claimsB64

	// Sign
	signature, err := s.signer.Sign([]byte(signingInput))
	if err != nil {
		return nil, presentation.ErrSigningFailed(op, err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	// Build JWT
	jwt := signingInput + "." + signatureB64

	return []byte(jwt), nil
}

// SignWithRequest signs a presentation using parameters from a request.
func (s *Signer) SignWithRequest(p *presentation.Presentation, request *presentation.Request) ([]byte, error) {
	opts := presentation.OptionsFromRequest(request)
	return s.Sign(p, opts)
}

// HolderDID returns the holder DID.
func (s *Signer) HolderDID() did.DID {
	return s.holderDID
}

// Algorithm returns the signing algorithm.
func (s *Signer) Algorithm() crypto.Algorithm {
	return s.signer.Algorithm()
}

// Format returns the output format.
func (s *Signer) Format() vc.Format {
	return vc.FormatJWT
}

// ============================================================================
// JWT Structures
// ============================================================================

// Header represents a JWT header.
type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid,omitempty"`
}

// Claims represents JWT claims for a Verifiable Presentation.
type Claims struct {
	// Standard JWT claims
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud,omitempty"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf,omitempty"`
	ExpiresAt *int64 `json:"exp,omitempty"`
	JWTID     string `json:"jti,omitempty"`

	// VP-specific claims
	Nonce string  `json:"nonce,omitempty"`
	VP    VPClaim `json:"vp"`
}

// VPClaim represents the VP claim in a JWT.
type VPClaim struct {
	Context              []string `json:"@context"`
	Type                 []string `json:"type"`
	VerifiableCredential []string `json:"verifiableCredential,omitempty"` // JWT-encoded VCs
}

// ============================================================================
// Helpers
// ============================================================================

// buildVPClaim creates the VP claim from a presentation.
func buildVPClaim(p *presentation.Presentation) VPClaim {
	// Convert types to strings
	types := make([]string, len(p.Type))
	for i, t := range p.Type {
		types[i] = t.String()
	}

	// Collect credential JWTs
	var credentialJWTs []string

	// Add already-JWT credentials
	credentialJWTs = append(credentialJWTs, p.VerifiableCredentialJWT...)

	// Note: If p.VerifiableCredential contains Credential objects,
	// they would need to be signed as JWTs first.
	// For now, we expect credentials to already be in JWT format.

	return VPClaim{
		Context:              p.Context,
		Type:                 types,
		VerifiableCredential: credentialJWTs,
	}
}

// algorithmToJWA converts crypto algorithm to JWA identifier.
func algorithmToJWA(alg crypto.Algorithm) string {
	switch alg {
	case crypto.AlgorithmEd25519:
		return "EdDSA"
	case crypto.AlgorithmSecp256k1:
		return "ES256K"
	default:
		return string(alg)
	}
}

// ============================================================================
// Encoding Helpers
// ============================================================================

// EncodePresentation encodes a presentation to JWT format without signing.
// Returns header.claims (no signature).
func EncodePresentation(p *presentation.Presentation, alg string, kid string) (string, error) {
	const op = "jwt.EncodePresentation"

	header := Header{
		Alg: alg,
		Typ: "JWT",
		Kid: kid,
	}

	now := time.Now().Unix()

	claims := Claims{
		Issuer:    p.Holder.String(),
		Subject:   p.Holder.String(),
		IssuedAt:  now,
		NotBefore: now,
		JWTID:     p.ID,
		VP:        buildVPClaim(p),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", presentation.ErrSigningFailed(op, err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", presentation.ErrSigningFailed(op, err)
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
