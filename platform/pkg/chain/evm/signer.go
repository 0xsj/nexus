package evm

import (
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/crypto/secp256k1"
)

// ============================================================================
// SignatureVerifier
// ============================================================================

// SignatureVerifier verifies EVM signatures and recovers addresses.
type SignatureVerifier struct {
	addressValidator *AddressValidator
}

// NewSignatureVerifier creates a new EVM signature verifier.
func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{
		addressValidator: NewAddressValidator(),
	}
}

// ============================================================================
// Signature Verification
// ============================================================================

// VerifySignature verifies that a signature was produced by the owner of the given address.
// The message should be the raw message bytes (not hashed).
// The signature should be 65 bytes: [R (32) | S (32) | V (1)].
func (v *SignatureVerifier) VerifySignature(address string, message []byte, signature []byte) error {
	const op = "evm.SignatureVerifier.VerifySignature"

	// Validate the expected address
	normalizedAddress, err := v.addressValidator.Validate(address)
	if err != nil {
		return err
	}

	// Recover the address from signature
	recoveredAddress, err := v.RecoverAddress(message, signature)
	if err != nil {
		return err
	}

	// Compare addresses (case-insensitive)
	if normalizedAddress != recoveredAddress {
		return chain.ErrAddressMismatchFor(op, normalizedAddress, recoveredAddress)
	}

	return nil
}

// RecoverAddress recovers the signer's address from a message and signature.
// The message should be the raw message bytes (not hashed).
// The signature should be 65 bytes: [R (32) | S (32) | V (1)].
func (v *SignatureVerifier) RecoverAddress(message []byte, signature []byte) (string, error) {
	const op = "evm.SignatureVerifier.RecoverAddress"

	// Validate signature length
	if len(signature) != 65 {
		return "", chain.ErrInvalidSignatureFor(op, fmt.Sprintf("invalid length: expected 65, got %d", len(signature)))
	}

	// Hash the message using Ethereum's message prefix
	hash := secp256k1.HashEthereumMessage(message)

	// Use secp256k1 package to recover address
	address, err := secp256k1.RecoverAddress(hash, signature)
	if err != nil {
		return "", chain.ErrSignatureVerificationFailed(op, err)
	}

	// Normalize to checksummed address
	return v.addressValidator.Normalize(address), nil
}

// VerifyPersonalSign verifies an Ethereum personal_sign signature.
// This is the most common signature type used by wallets.
func (v *SignatureVerifier) VerifyPersonalSign(address string, message []byte, signature []byte) error {
	return v.VerifySignature(address, message, signature)
}

// VerifyHash verifies a signature against a pre-hashed message.
// Use this when the message has already been hashed (e.g., EIP-712 typed data).
func (v *SignatureVerifier) VerifyHash(address string, hash []byte, signature []byte) error {
	const op = "evm.SignatureVerifier.VerifyHash"

	if len(hash) != 32 {
		return chain.ErrInvalidSignatureFor(op, fmt.Sprintf("invalid hash length: expected 32, got %d", len(hash)))
	}

	// Validate the expected address
	normalizedAddress, err := v.addressValidator.Validate(address)
	if err != nil {
		return err
	}

	// Recover address from hash and signature
	recoveredAddress, err := secp256k1.RecoverAddress(hash, signature)
	if err != nil {
		return chain.ErrSignatureVerificationFailed(op, err)
	}

	// Normalize recovered address
	recoveredNormalized := v.addressValidator.Normalize(recoveredAddress)

	// Compare addresses
	if normalizedAddress != recoveredNormalized {
		return chain.ErrAddressMismatchFor(op, normalizedAddress, recoveredNormalized)
	}

	return nil
}

// RecoverAddressFromHash recovers the signer's address from a hash and signature.
// Use this when the message has already been hashed.
func (v *SignatureVerifier) RecoverAddressFromHash(hash []byte, signature []byte) (string, error) {
	const op = "evm.SignatureVerifier.RecoverAddressFromHash"

	if len(hash) != 32 {
		return "", chain.ErrInvalidSignatureFor(op, fmt.Sprintf("invalid hash length: expected 32, got %d", len(hash)))
	}

	if len(signature) != 65 {
		return "", chain.ErrInvalidSignatureFor(op, fmt.Sprintf("invalid signature length: expected 65, got %d", len(signature)))
	}

	address, err := secp256k1.RecoverAddress(hash, signature)
	if err != nil {
		return "", chain.ErrSignatureVerificationFailed(op, err)
	}

	return v.addressValidator.Normalize(address), nil
}
