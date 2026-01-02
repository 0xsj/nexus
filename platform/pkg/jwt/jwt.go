package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/0xsj/nexus/platform/pkg/crypto"
)

// ============================================================================
// Token
// ============================================================================

// Token represents a parsed JWT.
type Token struct {
	Header    Header
	Claims    json.RawMessage
	Signature []byte

	// Raw parts for signature verification
	raw    string // Full token
	signed string // Header.Claims (signed portion)
}

// RawToken returns the original token string.
func (t *Token) RawToken() string {
	return t.raw
}

// SignedPortion returns the header.claims portion that was signed.
func (t *Token) SignedPortion() string {
	return t.signed
}

// StandardClaims parses the claims as standard JWT claims.
func (t *Token) StandardClaims() (Claims, error) {
	var claims Claims
	if err := json.Unmarshal(t.Claims, &claims); err != nil {
		return Claims{}, ErrInvalidClaims("jwt.Token.StandardClaims", "failed to parse claims")
	}
	return claims, nil
}

// ParseClaims parses the claims into a custom struct.
func (t *Token) ParseClaims(v any) error {
	if err := json.Unmarshal(t.Claims, v); err != nil {
		return ErrInvalidClaims("jwt.Token.ParseClaims", "failed to parse claims")
	}
	return nil
}

// ============================================================================
// Encoding
// ============================================================================

// Encode encodes a header and claims to an unsigned JWT (header.claims).
// Use Sign to create a complete signed JWT.
func Encode(header Header, claims any) (string, error) {
	const op = "jwt.Encode"

	// Validate header
	if err := header.Validate(); err != nil {
		return "", err
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", ErrEncodingFailed(op, err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", ErrEncodingFailed(op, err)
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return headerB64 + "." + claimsB64, nil
}

// Sign creates a complete signed JWT.
func Sign(header Header, claims any, signer crypto.Signer) (string, error) {
	const op = "jwt.Sign"

	// Set algorithm from signer if not set
	if header.Algorithm == "" {
		header.Algorithm = signer.Algorithm().JWAName()
	}

	// Encode header and claims
	unsigned, err := Encode(header, claims)
	if err != nil {
		return "", err
	}

	// Sign
	signature, err := signer.Sign([]byte(unsigned))
	if err != nil {
		return "", ErrSigningFailed(op, err)
	}

	// Encode signature
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return unsigned + "." + signatureB64, nil
}

// SignWithKey creates a signed JWT using a key ID.
func SignWithKey(claims any, signer crypto.Signer, keyID string) (string, error) {
	header := NewHeader(signer.Algorithm().JWAName()).WithKeyID(keyID)
	return Sign(header, claims, signer)
}

// ============================================================================
// Decoding
// ============================================================================

// Parse parses a JWT string without verifying the signature.
// Use Verify for signature verification.
func Parse(tokenString string) (*Token, error) {
	const op = "jwt.Parse"

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken(op, "token must have 3 parts")
	}

	// Decode header
	header, err := DecodeHeader(parts[0])
	if err != nil {
		return nil, err
	}

	// Decode claims (keep as raw JSON)
	claimsData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidClaims(op, "invalid base64 encoding")
	}
	if !json.Valid(claimsData) {
		return nil, ErrInvalidClaims(op, "invalid JSON")
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken(op, "invalid signature encoding")
	}

	return &Token{
		Header:    header,
		Claims:    json.RawMessage(claimsData),
		Signature: signature,
		raw:       tokenString,
		signed:    parts[0] + "." + parts[1],
	}, nil
}

// Verify parses and verifies a JWT signature.
func Verify(tokenString string, verifier crypto.Verifier) (*Token, error) {
	const op = "jwt.Verify"

	// Parse token
	token, err := Parse(tokenString)
	if err != nil {
		return nil, err
	}

	// Check algorithm matches
	if token.Header.Algorithm != verifier.Algorithm().JWAName() {
		return nil, ErrUnsupportedAlgorithm(op, token.Header.Algorithm)
	}

	// Verify signature
	if err := verifier.Verify([]byte(token.signed), token.Signature); err != nil {
		return nil, ErrInvalidSignature(op)
	}

	return token, nil
}

// VerifyAndValidate parses, verifies signature, and validates standard claims.
func VerifyAndValidate(tokenString string, verifier crypto.Verifier) (*Token, Claims, error) {
	const op = "jwt.VerifyAndValidate"

	// Verify signature
	token, err := Verify(tokenString, verifier)
	if err != nil {
		return nil, Claims{}, err
	}

	// Parse standard claims
	claims, err := token.StandardClaims()
	if err != nil {
		return nil, Claims{}, err
	}

	// Validate claims
	if err := claims.Validate(); err != nil {
		return nil, Claims{}, err
	}

	return token, claims, nil
}

// ============================================================================
// Utility Functions
// ============================================================================

// ExtractHeader extracts and decodes only the header from a token string.
// Useful for determining the algorithm without parsing the entire token.
func ExtractHeader(tokenString string) (Header, error) {
	const op = "jwt.ExtractHeader"

	parts := strings.Split(tokenString, ".")
	if len(parts) < 1 {
		return Header{}, ErrInvalidToken(op, "empty token")
	}

	return DecodeHeader(parts[0])
}

// ExtractKeyID extracts the key ID from a token's header.
// Returns empty string if no key ID is present.
func ExtractKeyID(tokenString string) (string, error) {
	header, err := ExtractHeader(tokenString)
	if err != nil {
		return "", err
	}
	return header.KeyID, nil
}

// IsExpired checks if a token is expired without full verification.
// Useful for quick checks before attempting verification.
func IsExpired(tokenString string) (bool, error) {
	const op = "jwt.IsExpired"

	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return false, ErrInvalidToken(op, "invalid token format")
	}

	claims, err := DecodeClaims(parts[1])
	if err != nil {
		return false, err
	}

	return claims.IsExpired(), nil
}

// ============================================================================
// Builder Pattern
// ============================================================================

// Builder provides a fluent interface for creating JWTs.
type Builder struct {
	header Header
	claims map[string]any
	err    error
}

// NewBuilder creates a new JWT builder.
func NewBuilder() *Builder {
	return &Builder{
		header: Header{Type: "JWT"},
		claims: make(map[string]any),
	}
}

// WithAlgorithm sets the algorithm.
func (b *Builder) WithAlgorithm(alg string) *Builder {
	b.header.Algorithm = alg
	return b
}

// WithKeyID sets the key ID.
func (b *Builder) WithKeyID(kid string) *Builder {
	b.header.KeyID = kid
	return b
}

// WithIssuer sets the issuer claim.
func (b *Builder) WithIssuer(iss string) *Builder {
	b.claims["iss"] = iss
	return b
}

// WithSubject sets the subject claim.
func (b *Builder) WithSubject(sub string) *Builder {
	b.claims["sub"] = sub
	return b
}

// WithAudience sets the audience claim.
func (b *Builder) WithAudience(aud ...string) *Builder {
	if len(aud) == 1 {
		b.claims["aud"] = aud[0]
	} else {
		b.claims["aud"] = aud
	}
	return b
}

// WithExpiresAt sets the expiration time.
func (b *Builder) WithExpiresAt(exp int64) *Builder {
	b.claims["exp"] = exp
	return b
}

// WithNotBefore sets the not-before time.
func (b *Builder) WithNotBefore(nbf int64) *Builder {
	b.claims["nbf"] = nbf
	return b
}

// WithIssuedAt sets the issued-at time.
func (b *Builder) WithIssuedAt(iat int64) *Builder {
	b.claims["iat"] = iat
	return b
}

// WithID sets the JWT ID.
func (b *Builder) WithID(jti string) *Builder {
	b.claims["jti"] = jti
	return b
}

// WithClaim sets a custom claim.
func (b *Builder) WithClaim(key string, value any) *Builder {
	b.claims[key] = value
	return b
}

// WithClaims merges multiple claims.
func (b *Builder) WithClaims(claims map[string]any) *Builder {
	for k, v := range claims {
		b.claims[k] = v
	}
	return b
}

// Build creates an unsigned JWT string.
func (b *Builder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}
	return Encode(b.header, b.claims)
}

// Sign creates a signed JWT string.
func (b *Builder) Sign(signer crypto.Signer) (string, error) {
	if b.err != nil {
		return "", b.err
	}
	return Sign(b.header, b.claims, signer)
}
