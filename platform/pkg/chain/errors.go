package chain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes
// ============================================================================

// Chain-specific error codes.
const (
	// CodeChainNotSupported indicates the chain is not supported.
	CodeChainNotSupported pkgerrors.Code = "CHAIN_NOT_SUPPORTED"

	// CodeChainNotFound indicates the chain was not found in the registry.
	CodeChainNotFound pkgerrors.Code = "CHAIN_NOT_FOUND"

	// CodeInvalidAddress indicates an invalid blockchain address.
	CodeInvalidAddress pkgerrors.Code = "INVALID_ADDRESS"

	// CodeInvalidSignature indicates an invalid cryptographic signature.
	CodeInvalidSignature pkgerrors.Code = "INVALID_SIGNATURE"

	// CodeSignatureVerificationFailed indicates signature verification failed.
	CodeSignatureVerificationFailed pkgerrors.Code = "SIGNATURE_VERIFICATION_FAILED"

	// CodeInvalidChainID indicates an invalid chain ID.
	CodeInvalidChainID pkgerrors.Code = "INVALID_CHAIN_ID"

	// CodeAddressMismatch indicates the recovered address doesn't match expected.
	CodeAddressMismatch pkgerrors.Code = "ADDRESS_MISMATCH"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// ErrChainNotSupported indicates the chain is not supported.
	ErrChainNotSupported = errors.New("chain not supported")

	// ErrChainNotFound indicates the chain was not found.
	ErrChainNotFound = errors.New("chain not found")

	// ErrInvalidAddress indicates an invalid address.
	ErrInvalidAddress = errors.New("invalid address")

	// ErrInvalidSignature indicates an invalid signature.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrAddressMismatch indicates address mismatch.
	ErrAddressMismatch = errors.New("address mismatch")
)

// ============================================================================
// Error Constructors
// ============================================================================

// ErrChainNotSupportedFor creates a chain not supported error.
func ErrChainNotSupportedFor(operation string, chainID string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "chain not supported: "+chainID).
		WithCode(CodeChainNotSupported).
		WithMeta("chain_id", chainID)
}

// ErrChainNotFoundFor creates a chain not found error.
func ErrChainNotFoundFor(operation string, chainID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "chain: "+chainID).
		WithCode(CodeChainNotFound).
		WithMeta("chain_id", chainID)
}

// ErrInvalidAddressFor creates an invalid address error.
func ErrInvalidAddressFor(operation string, address string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid address: "+reason).
		WithCode(CodeInvalidAddress).
		WithMeta("address", address).
		WithMeta("reason", reason)
}

// ErrInvalidSignatureFor creates an invalid signature error.
func ErrInvalidSignatureFor(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid signature: "+reason).
		WithCode(CodeInvalidSignature).
		WithMeta("reason", reason)
}

// ErrSignatureVerificationFailed creates a signature verification failed error.
func ErrSignatureVerificationFailed(operation string, err error) *pkgerrors.Error {
	msg := "signature verification failed"
	if err != nil {
		msg = msg + ": " + err.Error()
	}
	return pkgerrors.Validation(operation, msg).
		WithCode(CodeSignatureVerificationFailed)
}

// ErrInvalidChainIDFor creates an invalid chain ID error.
func ErrInvalidChainIDFor(operation string, chainID string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid chain ID: "+chainID).
		WithCode(CodeInvalidChainID).
		WithMeta("chain_id", chainID)
}

// ErrAddressMismatchFor creates an address mismatch error.
func ErrAddressMismatchFor(operation string, expected, actual string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "address mismatch").
		WithCode(CodeAddressMismatch).
		WithMeta("expected", expected).
		WithMeta("actual", actual)
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsChainNotSupported returns true if the error is a chain not supported error.
func IsChainNotSupported(err error) bool {
	return errors.Is(err, ErrChainNotSupported) || pkgerrors.HasCode(err, CodeChainNotSupported)
}

// IsChainNotFound returns true if the error is a chain not found error.
func IsChainNotFound(err error) bool {
	return errors.Is(err, ErrChainNotFound) || pkgerrors.HasCode(err, CodeChainNotFound)
}

// IsInvalidAddress returns true if the error is an invalid address error.
func IsInvalidAddress(err error) bool {
	return errors.Is(err, ErrInvalidAddress) || pkgerrors.HasCode(err, CodeInvalidAddress)
}

// IsInvalidSignature returns true if the error is an invalid signature error.
func IsInvalidSignature(err error) bool {
	return errors.Is(err, ErrInvalidSignature) || pkgerrors.HasCode(err, CodeInvalidSignature)
}

// IsAddressMismatch returns true if the error is an address mismatch error.
func IsAddressMismatch(err error) bool {
	return errors.Is(err, ErrAddressMismatch) || pkgerrors.HasCode(err, CodeAddressMismatch)
}
