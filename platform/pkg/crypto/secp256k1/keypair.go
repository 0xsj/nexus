package secp256k1

import (
	"crypto/rand"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/0xsj/nexus/platform/pkg/crypto"
)

const (
	// PublicKeySize is the size of a compressed secp256k1 public key in bytes.
	PublicKeySize = 33

	// PublicKeyUncompressedSize is the size of an uncompressed public key.
	PublicKeyUncompressedSize = 65

	// PrivateKeySize is the size of a secp256k1 private key in bytes.
	PrivateKeySize = 32

	// SignatureSize is the size of a secp256k1 signature in bytes (r + s).
	SignatureSize = 64

	// SignatureWithRecoverySize is the size with recovery byte (r + s + v).
	SignatureWithRecoverySize = 65
)

// KeyPair is a secp256k1 key pair.
// Used for Ethereum wallet compatibility (did:pkh, did:ethr).
type KeyPair struct {
	privKey *secp256k1.PrivateKey
	pubKey  *secp256k1.PublicKey
}

// Ensure KeyPair implements crypto.KeyPair.
var _ crypto.KeyPair = (*KeyPair)(nil)

// Generate creates a new random secp256k1 key pair.
func Generate() (*KeyPair, error) {
	const op = "secp256k1.Generate"

	seed := make([]byte, PrivateKeySize)
	if _, err := rand.Read(seed); err != nil {
		return nil, crypto.ErrKeyGenerationFailed(op, err)
	}

	privKey := secp256k1.PrivKeyFromBytes(seed)

	return &KeyPair{
		privKey: privKey,
		pubKey:  privKey.PubKey(),
	}, nil
}

// FromPrivateKey reconstructs a key pair from a private key.
func FromPrivateKey(privateKey []byte) (*KeyPair, error) {
	const op = "secp256k1.FromPrivateKey"

	if len(privateKey) != PrivateKeySize {
		return nil, crypto.ErrInvalidKey(op, "private key must be 32 bytes")
	}

	privKey := secp256k1.PrivKeyFromBytes(privateKey)

	return &KeyPair{
		privKey: privKey,
		pubKey:  privKey.PubKey(),
	}, nil
}

// FromSeed generates a deterministic key pair from a 32-byte seed.
func FromSeed(seed []byte) (*KeyPair, error) {
	const op = "secp256k1.FromSeed"

	if len(seed) != PrivateKeySize {
		return nil, crypto.ErrInvalidKey(op, "seed must be 32 bytes")
	}

	privKey := secp256k1.PrivKeyFromBytes(seed)

	return &KeyPair{
		privKey: privKey,
		pubKey:  privKey.PubKey(),
	}, nil
}

// FromPublicKey creates a key pair with only the public key.
// The resulting KeyPair cannot be used for signing.
func FromPublicKey(publicKey []byte) (*KeyPair, error) {
	const op = "secp256k1.FromPublicKey"

	pubKey, err := secp256k1.ParsePubKey(publicKey)
	if err != nil {
		return nil, crypto.ErrInvalidKey(op, "invalid public key: "+err.Error())
	}

	return &KeyPair{
		privKey: nil,
		pubKey:  pubKey,
	}, nil
}

// Algorithm returns the cryptographic algorithm.
func (kp *KeyPair) Algorithm() crypto.Algorithm {
	return crypto.AlgorithmSecp256k1
}

// PublicKey returns the raw public key bytes (compressed, 33 bytes).
func (kp *KeyPair) PublicKey() []byte {
	if kp.pubKey == nil {
		return nil
	}
	return kp.pubKey.SerializeCompressed()
}

// PublicKeyUncompressed returns the uncompressed public key (65 bytes).
func (kp *KeyPair) PublicKeyUncompressed() []byte {
	if kp.pubKey == nil {
		return nil
	}
	return kp.pubKey.SerializeUncompressed()
}

// PrivateKey returns the raw private key bytes.
func (kp *KeyPair) PrivateKey() []byte {
	if kp.privKey == nil {
		return nil
	}
	return kp.privKey.Serialize()
}

// PublicKeyBase58 returns the public key as base58 encoded string.
func (kp *KeyPair) PublicKeyBase58() string {
	pubKey := kp.PublicKey()
	if pubKey == nil {
		return ""
	}
	return base58Encode(pubKey)
}

// PublicKeyMultibase returns the public key as multibase encoded string.
func (kp *KeyPair) PublicKeyMultibase() string {
	pubKey := kp.PublicKey()
	if pubKey == nil {
		return ""
	}
	// secp256k1-pub multicodec prefix: 0xe7 0x01
	prefixed := append([]byte{0xe7, 0x01}, pubKey...)
	return "z" + base58Encode(prefixed)
}

// HasPrivateKey returns true if the private key is available.
func (kp *KeyPair) HasPrivateKey() bool {
	return kp.privKey != nil
}

// EthereumAddress returns the Ethereum address derived from the public key.
func (kp *KeyPair) EthereumAddress() string {
	if kp.pubKey == nil {
		return ""
	}

	// Get uncompressed public key without the 0x04 prefix
	uncompressed := kp.pubKey.SerializeUncompressed()
	pubKeyBytes := uncompressed[1:] // Remove 0x04 prefix

	// Keccak256 hash
	hash := Keccak256(pubKeyBytes)

	// Take last 20 bytes
	address := hash[12:]

	return "0x" + hexEncode(address)
}

// ============================================================================
// Internal Helpers
// ============================================================================

// hexEncode encodes bytes to lowercase hex string.
func hexEncode(data []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(data)*2)
	for i, b := range data {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}
	return string(result)
}

// base58Encode encodes bytes to base58 string.
func base58Encode(data []byte) string {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	if len(data) == 0 {
		return ""
	}

	// Count leading zeros
	leadingZeros := 0
	for _, b := range data {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	// Allocate enough space
	size := len(data)*138/100 + 1
	buf := make([]byte, size)

	// Process each byte
	for _, b := range data {
		carry := int(b)
		for j := size - 1; j >= 0; j-- {
			carry += 256 * int(buf[j])
			buf[j] = byte(carry % 58)
			carry /= 58
		}
	}

	// Skip leading zeros in buf
	start := 0
	for start < size && buf[start] == 0 {
		start++
	}

	// Build result
	result := make([]byte, leadingZeros+(size-start))
	for i := 0; i < leadingZeros; i++ {
		result[i] = alphabet[0]
	}
	for i := leadingZeros; start < size; i++ {
		result[i] = alphabet[buf[start]]
		start++
	}

	return string(result)
}
