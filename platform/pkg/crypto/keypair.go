package crypto

// KeyPair represents a public/private key pair.
// Implementations provide algorithm-specific key generation and serialization.
type KeyPair interface {
	// Algorithm returns the cryptographic algorithm.
	Algorithm() Algorithm

	// PublicKey returns the raw public key bytes.
	PublicKey() []byte

	// PrivateKey returns the raw private key bytes.
	// Handle with care — this is sensitive material.
	PrivateKey() []byte

	// PublicKeyBase58 returns the public key as base58 encoded string.
	// Used in DID documents and verification methods.
	PublicKeyBase58() string

	// PublicKeyMultibase returns the public key as multibase encoded string.
	// Used in did:key generation.
	PublicKeyMultibase() string
}

// KeyGenerator generates new key pairs.
type KeyGenerator interface {
	// Generate creates a new random key pair.
	Generate() (KeyPair, error)

	// FromPrivateKey reconstructs a key pair from a private key.
	FromPrivateKey(privateKey []byte) (KeyPair, error)

	// FromSeed generates a deterministic key pair from a seed.
	// Seed must be at least 32 bytes.
	FromSeed(seed []byte) (KeyPair, error)
}

// KeyEncoder handles key serialization formats.
type KeyEncoder interface {
	// EncodePublicKey encodes a public key to the specified format.
	EncodePublicKey(publicKey []byte, format KeyFormat) ([]byte, error)

	// DecodePublicKey decodes a public key from the specified format.
	DecodePublicKey(encoded []byte, format KeyFormat) ([]byte, error)

	// EncodePrivateKey encodes a private key to the specified format.
	EncodePrivateKey(privateKey []byte, format KeyFormat) ([]byte, error)

	// DecodePrivateKey decodes a private key from the specified format.
	DecodePrivateKey(encoded []byte, format KeyFormat) ([]byte, error)
}

// KeyFormat specifies key encoding format.
type KeyFormat string

const (
	// KeyFormatRaw is raw bytes (no encoding).
	KeyFormatRaw KeyFormat = "raw"

	// KeyFormatBase58 is base58 encoding (Bitcoin style).
	KeyFormatBase58 KeyFormat = "base58"

	// KeyFormatBase64 is standard base64 encoding.
	KeyFormatBase64 KeyFormat = "base64"

	// KeyFormatBase64URL is URL-safe base64 encoding.
	KeyFormatBase64URL KeyFormat = "base64url"

	// KeyFormatHex is hexadecimal encoding.
	KeyFormatHex KeyFormat = "hex"

	// KeyFormatMultibase is multibase encoding (self-describing).
	KeyFormatMultibase KeyFormat = "multibase"

	// KeyFormatJWK is JSON Web Key format.
	KeyFormatJWK KeyFormat = "jwk"

	// KeyFormatPEM is PEM encoding.
	KeyFormatPEM KeyFormat = "pem"
)

// String returns the string representation.
func (f KeyFormat) String() string {
	return string(f)
}

// Validate checks if the format is supported.
func (f KeyFormat) Validate() error {
	switch f {
	case KeyFormatRaw, KeyFormatBase58, KeyFormatBase64, KeyFormatBase64URL,
		KeyFormatHex, KeyFormatMultibase, KeyFormatJWK, KeyFormatPEM:
		return nil
	default:
		return ErrInvalidKeyFormat("KeyFormat.Validate", "unsupported format: "+string(f))
	}
}
