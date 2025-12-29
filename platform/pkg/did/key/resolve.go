package key

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
)

// Resolver resolves did:key DIDs.
// did:key is self-resolving — the DID document is derived from the DID itself.
type Resolver struct{}

// Ensure Resolver implements did.Resolver.
var _ did.Resolver = (*Resolver)(nil)

// NewResolver creates a new did:key resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Method returns the DID method this resolver handles.
func (r *Resolver) Method() did.Method {
	return did.MethodKey
}

// Resolve resolves a did:key DID to a DID Document.
// The document is derived directly from the public key encoded in the DID.
func (r *Resolver) Resolve(ctx context.Context, d did.DID) (*did.Document, error) {
	const op = "did.key.Resolver.Resolve"

	// Validate method
	if d.Method() != did.MethodKey {
		return nil, did.ErrUnsupportedMethod(op, d.Method().String())
	}

	// Parse the multibase-encoded public key
	multibase := d.MethodSpecificID()
	publicKey, algorithm, err := decodeMultibase(multibase)
	if err != nil {
		return nil, did.ErrInvalidDID(op, err.Error())
	}

	// Build the DID document based on key type
	doc, err := buildDocument(d, publicKey, algorithm)
	if err != nil {
		return nil, did.ErrInvalidDocument(op, err.Error())
	}

	return doc, nil
}

// decodeMultibase decodes a multibase string and returns the public key and algorithm.
func decodeMultibase(multibase string) ([]byte, crypto.Algorithm, error) {
	if len(multibase) < 2 {
		return nil, "", fmt.Errorf("multibase string too short")
	}

	// First character is the multibase prefix
	prefix := multibase[0]
	if prefix != 'z' {
		return nil, "", fmt.Errorf("unsupported multibase prefix: %c (expected 'z' for base58btc)", prefix)
	}

	// Decode base58
	encoded := multibase[1:]
	decoded, err := base58Decode(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("base58 decode failed: %w", err)
	}

	// Check multicodec prefix
	if len(decoded) < 2 {
		return nil, "", fmt.Errorf("decoded data too short for multicodec prefix")
	}

	// Identify algorithm from multicodec prefix
	algorithm, prefixLen, err := identifyAlgorithm(decoded)
	if err != nil {
		return nil, "", err
	}

	// Extract public key
	publicKey := decoded[prefixLen:]

	return publicKey, algorithm, nil
}

// identifyAlgorithm identifies the algorithm from multicodec prefix.
func identifyAlgorithm(data []byte) (crypto.Algorithm, int, error) {
	if len(data) < 2 {
		return "", 0, fmt.Errorf("data too short for multicodec prefix")
	}

	// Ed25519: 0xed01
	if data[0] == 0xed && data[1] == 0x01 {
		return crypto.AlgorithmEd25519, 2, nil
	}

	// secp256k1: 0xe701
	if data[0] == 0xe7 && data[1] == 0x01 {
		return crypto.AlgorithmSecp256k1, 2, nil
	}

	// BLS12-381 G2: 0xeb01
	if data[0] == 0xeb && data[1] == 0x01 {
		return crypto.AlgorithmBLS12381, 2, nil
	}

	return "", 0, fmt.Errorf("unsupported multicodec prefix: 0x%x%x", data[0], data[1])
}

// buildDocument builds a DID document for the given public key.
func buildDocument(d did.DID, publicKey []byte, algorithm crypto.Algorithm) (*did.Document, error) {
	switch algorithm {
	case crypto.AlgorithmEd25519:
		return buildEd25519Document(d, publicKey)
	case crypto.AlgorithmSecp256k1:
		return buildSecp256k1Document(d, publicKey)
	case crypto.AlgorithmBLS12381:
		return buildBLS12381Document(d, publicKey)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// buildEd25519Document builds a DID document for an Ed25519 key.
func buildEd25519Document(d did.DID, publicKey []byte) (*did.Document, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid Ed25519 public key size: expected %d, got %d",
			ed25519.PublicKeySize, len(publicKey))
	}

	// Verification method ID is the DID with fragment
	vmID := d.Fragment(d.MethodSpecificID())

	// Create key pair for multibase encoding
	kp, err := ed25519.FromPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	doc := did.NewDocument(d)
	doc.AddVerificationMethod(did.VerificationMethod{
		ID:                 vmID,
		Type:               did.Ed25519VerificationKey2020,
		Controller:         d,
		PublicKeyMultibase: kp.PublicKeyMultibase(),
	})
	doc.AddAuthentication(vmID)
	doc.AddAssertionMethod(vmID)
	doc.AddCapabilityInvocation(vmID)
	doc.AddCapabilityDelegation(vmID)

	return doc, nil
}

// buildSecp256k1Document builds a DID document for a secp256k1 key.
// Phase 2: Stub implementation.
func buildSecp256k1Document(d did.DID, publicKey []byte) (*did.Document, error) {
	vmID := d.Fragment(d.MethodSpecificID())

	doc := did.NewDocument(d)
	doc.AddVerificationMethod(did.VerificationMethod{
		ID:              vmID,
		Type:            did.EcdsaSecp256k1VerificationKey2019,
		Controller:      d,
		PublicKeyBase58: base58Encode(publicKey),
	})
	doc.AddAuthentication(vmID)
	doc.AddAssertionMethod(vmID)

	return doc, nil
}

// buildBLS12381Document builds a DID document for a BLS12-381 key.
// Phase 2: Stub implementation.
func buildBLS12381Document(d did.DID, publicKey []byte) (*did.Document, error) {
	vmID := d.Fragment(d.MethodSpecificID())

	doc := did.NewDocument(d)
	doc.AddVerificationMethod(did.VerificationMethod{
		ID:              vmID,
		Type:            did.Bls12381G2Key2020,
		Controller:      d,
		PublicKeyBase58: base58Encode(publicKey),
	})
	doc.AddAuthentication(vmID)
	doc.AddAssertionMethod(vmID)

	return doc, nil
}

// ============================================================================
// Base58 Encoding (Bitcoin alphabet)
// ============================================================================

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// base58Encode encodes bytes to base58 string.
func base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	// Count leading zeros
	zeros := 0
	for _, b := range input {
		if b != 0 {
			break
		}
		zeros++
	}

	// Allocate enough space
	size := len(input)*138/100 + 1
	buf := make([]byte, size)

	for _, b := range input {
		carry := int(b)
		for i := size - 1; i >= 0; i-- {
			carry += 256 * int(buf[i])
			buf[i] = byte(carry % 58)
			carry /= 58
		}
	}

	// Skip leading zeros in base58 result
	i := 0
	for i < len(buf) && buf[i] == 0 {
		i++
	}

	// Build result
	result := make([]byte, zeros+len(buf)-i)
	for j := 0; j < zeros; j++ {
		result[j] = '1'
	}
	for j := zeros; i < len(buf); i, j = i+1, j+1 {
		result[j] = base58Alphabet[buf[i]]
	}

	return string(result)
}

// base58Decode decodes a base58 string to bytes.
func base58Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}

	// Build alphabet index
	alphabetIdx := make(map[rune]int)
	for i, c := range base58Alphabet {
		alphabetIdx[c] = i
	}

	// Count leading '1's (zeros in result)
	zeros := 0
	for _, c := range input {
		if c != '1' {
			break
		}
		zeros++
	}

	// Allocate enough space
	size := len(input)*733/1000 + 1
	buf := make([]byte, size)

	for _, c := range input {
		idx, ok := alphabetIdx[c]
		if !ok {
			return nil, fmt.Errorf("invalid base58 character: %c", c)
		}

		carry := idx
		for i := size - 1; i >= 0; i-- {
			carry += 58 * int(buf[i])
			buf[i] = byte(carry % 256)
			carry /= 256
		}
	}

	// Skip leading zeros in buf
	i := 0
	for i < len(buf) && buf[i] == 0 {
		i++
	}

	// Build result with leading zeros
	result := make([]byte, zeros+len(buf)-i)
	copy(result[zeros:], buf[i:])

	return result, nil
}

// ============================================================================
// Helpers
// ============================================================================

// ExtractPublicKey extracts the public key from a did:key DID.
func ExtractPublicKey(d did.DID) ([]byte, crypto.Algorithm, error) {
	const op = "did.key.ExtractPublicKey"

	if d.Method() != did.MethodKey {
		return nil, "", did.ErrUnsupportedMethod(op, d.Method().String())
	}

	return decodeMultibase(d.MethodSpecificID())
}

// ToDID creates a did:key from a multibase-encoded public key string.
func ToDID(multibase string) (did.DID, error) {
	// Validate by attempting to decode
	_, _, err := decodeMultibase(multibase)
	if err != nil {
		return did.DID{}, err
	}

	return did.New(did.MethodKey, multibase)
}

// Fingerprint returns the fingerprint (method-specific-id) for a public key.
func Fingerprint(publicKey []byte, algorithm crypto.Algorithm) string {
	prefix := algorithm.MulticodecPrefix()
	if prefix == nil {
		return ""
	}

	prefixed := append(prefix, publicKey...)
	return "z" + base58Encode(prefixed)
}

// IsKeyDID returns true if the string looks like a did:key DID.
func IsKeyDID(s string) bool {
	return strings.HasPrefix(s, "did:key:")
}
