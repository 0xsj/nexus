package siwe

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/0xsj/nexus/platform/pkg/crypto/secp256k1"
)

// ============================================================================
// Verification
// ============================================================================

// Verify verifies a SIWE message signature and returns the parsed message.
// It validates the message format, timing, and signature.
func Verify(message string, signature string) (*Message, error) {
	const op = "siwe.Verify"

	// Parse the message
	msg, err := Parse(message)
	if err != nil {
		return nil, err
	}

	// Validate timing
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Verify signature
	recoveredAddress, err := RecoverAddress(message, signature)
	if err != nil {
		return nil, ErrInvalidSignature(op)
	}

	// Compare addresses (case-insensitive)
	if !AddressesEqual(recoveredAddress, msg.Address) {
		return nil, ErrAddressMismatch(op)
	}

	return msg, nil
}

// VerifyWithOptions verifies a SIWE message with additional validation options.
func VerifyWithOptions(message string, signature string, opts VerifyOptions) (*Message, error) {
	const op = "siwe.VerifyWithOptions"

	// Parse the message
	msg, err := Parse(message)
	if err != nil {
		return nil, err
	}

	// Validate timing
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Validate domain if specified
	if opts.ExpectedDomain != "" {
		if err := msg.ValidateDomain(opts.ExpectedDomain); err != nil {
			return nil, err
		}
	}

	// Validate nonce if specified
	if opts.ExpectedNonce != "" {
		if err := msg.ValidateNonce(opts.ExpectedNonce); err != nil {
			return nil, err
		}
	}

	// Validate chain ID if specified
	if opts.ExpectedChainID != 0 && msg.ChainID != opts.ExpectedChainID {
		return nil, ErrInvalidChainID(op, fmt.Sprintf("expected %d, got %d", opts.ExpectedChainID, msg.ChainID))
	}

	// Verify signature
	recoveredAddress, err := RecoverAddress(message, signature)
	if err != nil {
		return nil, ErrInvalidSignature(op)
	}

	// Compare addresses (case-insensitive)
	if !AddressesEqual(recoveredAddress, msg.Address) {
		return nil, ErrAddressMismatch(op)
	}

	return msg, nil
}

// VerifyOptions contains options for verification.
type VerifyOptions struct {
	// ExpectedDomain validates the domain matches.
	ExpectedDomain string

	// ExpectedNonce validates the nonce matches.
	ExpectedNonce string

	// ExpectedChainID validates the chain ID matches.
	ExpectedChainID int
}

// ============================================================================
// Signature Recovery
// ============================================================================

// RecoverAddress recovers the Ethereum address from a message and signature.
func RecoverAddress(message string, signature string) (string, error) {
	const op = "siwe.RecoverAddress"

	// Decode signature from hex
	sigBytes, err := decodeSignature(signature)
	if err != nil {
		return "", ErrInvalidSignature(op)
	}

	// Hash the message with Ethereum prefix
	hash := secp256k1.HashEthereumMessage([]byte(message))

	// Recover address
	address, err := secp256k1.RecoverAddress(hash, sigBytes)
	if err != nil {
		return "", ErrInvalidSignature(op)
	}

	return address, nil
}

// ============================================================================
// Internal Helpers
// ============================================================================

// decodeSignature decodes a hex signature string to bytes.
// Handles both 0x-prefixed and non-prefixed signatures.
// Normalizes V from 27/28 to 0/1.
func decodeSignature(signature string) ([]byte, error) {
	// Remove 0x prefix if present
	sig := strings.TrimPrefix(signature, "0x")

	// Decode hex
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}

	// Signature must be 65 bytes: R (32) + S (32) + V (1)
	if len(sigBytes) != 65 {
		return nil, fmt.Errorf("invalid signature length: %d", len(sigBytes))
	}

	// Normalize V: Ethereum uses 27/28, we need 0/1
	v := sigBytes[64]
	if v >= 27 {
		sigBytes[64] = v - 27
	}

	return sigBytes, nil
}

// ============================================================================
// Verifier Interface
// ============================================================================

// Verifier verifies SIWE messages.
type Verifier interface {
	// Verify verifies a message and signature.
	Verify(message string, signature string) (*Message, error)

	// VerifyWithNonce verifies and validates the nonce.
	VerifyWithNonce(message string, signature string, expectedNonce string) (*Message, error)
}

// DefaultVerifier is the default SIWE verifier with preset domain and chain ID.
type DefaultVerifier struct {
	domain  string
	chainID int
}

// NewVerifier creates a new SIWE verifier.
func NewVerifier(domain string, chainID int) *DefaultVerifier {
	return &DefaultVerifier{
		domain:  domain,
		chainID: chainID,
	}
}

// Verify verifies a message and signature.
func (v *DefaultVerifier) Verify(message string, signature string) (*Message, error) {
	return VerifyWithOptions(message, signature, VerifyOptions{
		ExpectedDomain:  v.domain,
		ExpectedChainID: v.chainID,
	})
}

// VerifyWithNonce verifies and validates the nonce.
func (v *DefaultVerifier) VerifyWithNonce(message string, signature string, expectedNonce string) (*Message, error) {
	return VerifyWithOptions(message, signature, VerifyOptions{
		ExpectedDomain:  v.domain,
		ExpectedNonce:   expectedNonce,
		ExpectedChainID: v.chainID,
	})
}

// ============================================================================
// Signing (for testing)
// ============================================================================

// SignMessage signs a SIWE message with a private key.
// Returns a hex-encoded signature with 0x prefix.
// Primarily for testing purposes.
func SignMessage(message string, privateKey []byte) (string, error) {
	const op = "siwe.SignMessage"

	signer, err := secp256k1.NewSignerFromPrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Sign with Ethereum prefix (returns V as 27/28)
	sig, err := signer.SignEthereum([]byte(message))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return "0x" + hex.EncodeToString(sig), nil
}
