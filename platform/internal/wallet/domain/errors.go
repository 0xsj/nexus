package domain

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Wallet domain error codes.
const (
	CodeInvalidAddress    errors.Code = "INVALID_ADDRESS"
	CodeInvalidSignature  errors.Code = "INVALID_SIGNATURE"
	CodeInvalidChain      errors.Code = "INVALID_CHAIN"
	CodeChainNotSupported errors.Code = "CHAIN_NOT_SUPPORTED"
	CodeAddressMismatch   errors.Code = "ADDRESS_MISMATCH"
	CodeChainMismatch     errors.Code = "CHAIN_MISMATCH"

	CodeWalletNotFound      errors.Code = "WALLET_NOT_FOUND"
	CodeWalletAlreadyExists errors.Code = "WALLET_ALREADY_EXISTS"

	CodeSignatureExpired   errors.Code = "SIGNATURE_EXPIRED"
	CodeSignatureUsed      errors.Code = "SIGNATURE_USED"
	CodeInvalidNonce       errors.Code = "INVALID_NONCE"
	CodeVerificationFailed errors.Code = "VERIFICATION_FAILED"
)

// ============================================================================
// Address Errors
// ============================================================================

// ErrInvalidAddress creates an invalid address error.
func ErrInvalidAddress(operation string, address string, reason string) *errors.Error {
	return errors.Validation(operation, "invalid address: "+reason).
		WithCode(CodeInvalidAddress).
		WithMeta("address", address)
}

// ErrAddressMismatch creates an address mismatch error.
func ErrAddressMismatch(operation string, expected, actual string) *errors.Error {
	return errors.Validation(operation, "address mismatch").
		WithCode(CodeAddressMismatch).
		WithMeta("expected", expected).
		WithMeta("actual", actual)
}

// ============================================================================
// Chain Errors
// ============================================================================

// ErrInvalidChain creates an invalid chain error.
func ErrInvalidChain(operation string, chainID string) *errors.Error {
	return errors.Validation(operation, "invalid chain: "+chainID).
		WithCode(CodeInvalidChain).
		WithMeta("chain_id", chainID)
}

// ErrChainNotSupported creates a chain not supported error.
func ErrChainNotSupported(operation string, chainID string) *errors.Error {
	return errors.Validation(operation, "chain not supported: "+chainID).
		WithCode(CodeChainNotSupported).
		WithMeta("chain_id", chainID)
}

// ErrChainMismatch creates a chain mismatch error.
func ErrChainMismatch(operation string, expected, actual string) *errors.Error {
	return errors.Validation(operation, "chain mismatch").
		WithCode(CodeChainMismatch).
		WithMeta("expected", expected).
		WithMeta("actual", actual)
}

// ============================================================================
// Wallet Errors
// ============================================================================

// ErrWalletNotFound creates a wallet not found error.
func ErrWalletNotFound(operation string, identifier string) *errors.Error {
	return errors.NotFound(operation, "wallet: "+identifier).
		WithCode(CodeWalletNotFound).
		WithMeta("identifier", identifier)
}

// ErrWalletAlreadyExists creates a wallet already exists error.
func ErrWalletAlreadyExists(operation string, address string, chainID string) *errors.Error {
	return errors.Conflict(operation, "wallet already exists").
		WithCode(CodeWalletAlreadyExists).
		WithMeta("address", address).
		WithMeta("chain_id", chainID)
}

// ============================================================================
// Signature Errors
// ============================================================================

// ErrInvalidSignature creates an invalid signature error.
func ErrInvalidSignature(operation string, reason string) *errors.Error {
	return errors.Validation(operation, "invalid signature: "+reason).
		WithCode(CodeInvalidSignature)
}

// ErrSignatureExpired creates a signature expired error.
func ErrSignatureExpired(operation string) *errors.Error {
	return errors.Unauthorized(operation, "signature or challenge has expired").
		WithCode(CodeSignatureExpired)
}

// ErrSignatureUsed creates a signature already used error.
func ErrSignatureUsed(operation string, nonce string) *errors.Error {
	return errors.Unauthorized(operation, "signature or nonce already used").
		WithCode(CodeSignatureUsed).
		WithMeta("nonce", nonce)
}

// ErrInvalidNonce creates an invalid nonce error.
func ErrInvalidNonce(operation string, nonce string) *errors.Error {
	return errors.Validation(operation, "invalid nonce").
		WithCode(CodeInvalidNonce).
		WithMeta("nonce", nonce)
}

// ErrVerificationFailed creates a verification failed error.
func ErrVerificationFailed(operation string, reason string) *errors.Error {
	return errors.Unauthorized(operation, "signature verification failed: "+reason).
		WithCode(CodeVerificationFailed)
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsInvalidAddress checks if error is invalid address.
func IsInvalidAddress(err error) bool {
	return errors.GetCode(err) == CodeInvalidAddress
}

// IsInvalidSignature checks if error is invalid signature.
func IsInvalidSignature(err error) bool {
	return errors.GetCode(err) == CodeInvalidSignature
}

// IsInvalidChain checks if error is invalid chain.
func IsInvalidChain(err error) bool {
	return errors.GetCode(err) == CodeInvalidChain
}

// IsChainNotSupported checks if error is chain not supported.
func IsChainNotSupported(err error) bool {
	return errors.GetCode(err) == CodeChainNotSupported
}

// IsAddressMismatch checks if error is address mismatch.
func IsAddressMismatch(err error) bool {
	return errors.GetCode(err) == CodeAddressMismatch
}

// IsChainMismatch checks if error is chain mismatch.
func IsChainMismatch(err error) bool {
	return errors.GetCode(err) == CodeChainMismatch
}

// IsWalletNotFound checks if error is wallet not found.
func IsWalletNotFound(err error) bool {
	return errors.GetCode(err) == CodeWalletNotFound
}

// IsWalletAlreadyExists checks if error is wallet already exists.
func IsWalletAlreadyExists(err error) bool {
	return errors.GetCode(err) == CodeWalletAlreadyExists
}

// IsSignatureExpired checks if error is signature expired.
func IsSignatureExpired(err error) bool {
	return errors.GetCode(err) == CodeSignatureExpired
}

// IsSignatureUsed checks if error is signature used.
func IsSignatureUsed(err error) bool {
	return errors.GetCode(err) == CodeSignatureUsed
}

// IsInvalidNonce checks if error is invalid nonce.
func IsInvalidNonce(err error) bool {
	return errors.GetCode(err) == CodeInvalidNonce
}

// IsVerificationFailed checks if error is verification failed.
func IsVerificationFailed(err error) bool {
	return errors.GetCode(err) == CodeVerificationFailed
}
