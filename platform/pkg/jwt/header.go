package jwt

import (
	"encoding/base64"
	"encoding/json"
)

// ============================================================================
// Header
// ============================================================================

// Header represents a JWT header (JOSE Header).
type Header struct {
	// Algorithm is the signing algorithm (e.g., "ES256", "EdDSA").
	Algorithm string `json:"alg"`

	// Type is the token type (typically "JWT").
	Type string `json:"typ,omitempty"`

	// KeyID is the key identifier (optional).
	KeyID string `json:"kid,omitempty"`

	// ContentType is used for nested JWTs (optional).
	ContentType string `json:"cty,omitempty"`
}

// NewHeader creates a new JWT header with the specified algorithm.
func NewHeader(algorithm string) Header {
	return Header{
		Algorithm: algorithm,
		Type:      "JWT",
	}
}

// WithKeyID sets the key ID.
func (h Header) WithKeyID(kid string) Header {
	h.KeyID = kid
	return h
}

// WithType sets the token type.
func (h Header) WithType(typ string) Header {
	h.Type = typ
	return h
}

// WithContentType sets the content type.
func (h Header) WithContentType(cty string) Header {
	h.ContentType = cty
	return h
}

// Validate validates the header.
func (h Header) Validate() error {
	const op = "jwt.Header.Validate"

	if h.Algorithm == "" {
		return ErrInvalidHeader(op, "algorithm is required")
	}

	return nil
}

// Encode encodes the header to a base64url string.
func (h Header) Encode() (string, error) {
	const op = "jwt.Header.Encode"

	data, err := json.Marshal(h)
	if err != nil {
		return "", ErrEncodingFailed(op, err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

// ============================================================================
// Decoding
// ============================================================================

// DecodeHeader decodes a base64url encoded header.
func DecodeHeader(encoded string) (Header, error) {
	const op = "jwt.DecodeHeader"

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Header{}, ErrInvalidHeader(op, "invalid base64 encoding")
	}

	var header Header
	if err := json.Unmarshal(data, &header); err != nil {
		return Header{}, ErrInvalidHeader(op, "invalid JSON")
	}

	if err := header.Validate(); err != nil {
		return Header{}, err
	}

	return header, nil
}

// ============================================================================
// Algorithm Constants
// ============================================================================

// Common JWT algorithms.
const (
	// AlgHS256 is HMAC using SHA-256.
	AlgHS256 = "HS256"

	// AlgHS384 is HMAC using SHA-384.
	AlgHS384 = "HS384"

	// AlgHS512 is HMAC using SHA-512.
	AlgHS512 = "HS512"

	// AlgRS256 is RSASSA-PKCS1-v1_5 using SHA-256.
	AlgRS256 = "RS256"

	// AlgRS384 is RSASSA-PKCS1-v1_5 using SHA-384.
	AlgRS384 = "RS384"

	// AlgRS512 is RSASSA-PKCS1-v1_5 using SHA-512.
	AlgRS512 = "RS512"

	// AlgES256 is ECDSA using P-256 and SHA-256.
	AlgES256 = "ES256"

	// AlgES384 is ECDSA using P-384 and SHA-384.
	AlgES384 = "ES384"

	// AlgES512 is ECDSA using P-521 and SHA-512.
	AlgES512 = "ES512"

	// AlgES256K is ECDSA using secp256k1 and SHA-256.
	AlgES256K = "ES256K"

	// AlgEdDSA is EdDSA using Ed25519.
	AlgEdDSA = "EdDSA"

	// AlgNone indicates no signature (UNSAFE - use only for testing).
	AlgNone = "none"
)

// IsSupportedAlgorithm checks if an algorithm is supported.
func IsSupportedAlgorithm(alg string) bool {
	switch alg {
	case AlgHS256, AlgHS384, AlgHS512,
		AlgRS256, AlgRS384, AlgRS512,
		AlgES256, AlgES384, AlgES512, AlgES256K,
		AlgEdDSA:
		return true
	default:
		return false
	}
}

// IsSymmetricAlgorithm returns true if the algorithm uses symmetric keys.
func IsSymmetricAlgorithm(alg string) bool {
	switch alg {
	case AlgHS256, AlgHS384, AlgHS512:
		return true
	default:
		return false
	}
}

// IsAsymmetricAlgorithm returns true if the algorithm uses asymmetric keys.
func IsAsymmetricAlgorithm(alg string) bool {
	switch alg {
	case AlgRS256, AlgRS384, AlgRS512,
		AlgES256, AlgES384, AlgES512, AlgES256K,
		AlgEdDSA:
		return true
	default:
		return false
	}
}
