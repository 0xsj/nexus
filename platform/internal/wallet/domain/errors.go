// Package domain contains the core business logic for the Wallet bounded context.
package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Wallet-specific)
// ============================================================================

const (
	// Wallet errors
	CodeWalletNotFound      pkgerrors.Code = "WALLET_NOT_FOUND"
	CodeWalletAlreadyLinked pkgerrors.Code = "WALLET_ALREADY_LINKED"
	CodeWalletNotActive     pkgerrors.Code = "WALLET_NOT_ACTIVE"
	CodeWalletUnverified    pkgerrors.Code = "WALLET_UNVERIFIED"

	// Address errors
	CodeInvalidAddress pkgerrors.Code = "INVALID_ADDRESS"

	// Chain errors
	CodeChainNotSupported pkgerrors.Code = "CHAIN_NOT_SUPPORTED"

	// Signature errors
	CodeSignatureInvalid pkgerrors.Code = "SIGNATURE_INVALID"

	// Challenge errors
	CodeChallengeExpired  pkgerrors.Code = "CHALLENGE_EXPIRED"
	CodeChallengeNotFound pkgerrors.Code = "CHALLENGE_NOT_FOUND"

	// Primary wallet errors
	CodePrimaryWalletRequired pkgerrors.Code = "PRIMARY_WALLET_REQUIRED"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	// Wallet errors
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyLinked = errors.New("wallet already linked")
	ErrWalletNotActive     = errors.New("wallet is not active")
	ErrWalletUnverified    = errors.New("wallet is unverified")

	// Address errors
	ErrInvalidAddress = errors.New("invalid wallet address")

	// Chain errors
	ErrChainNotSupported = errors.New("chain not supported")

	// Signature errors
	ErrSignatureInvalid = errors.New("signature is invalid")

	// Challenge errors
	ErrChallengeExpired  = errors.New("challenge has expired")
	ErrChallengeNotFound = errors.New("challenge not found")

	// Primary wallet errors
	ErrPrimaryWalletRequired = errors.New("primary wallet is required")
)

// ============================================================================
// Error Constructors
// ============================================================================

// WalletNotFound creates a wallet not found error.
func WalletNotFound(operation string, walletID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "wallet").
		WithCode(CodeWalletNotFound).
		WithMeta("wallet_id", walletID)
}

// WalletAlreadyLinked creates a wallet already linked error.
func WalletAlreadyLinked(operation string, address string) *pkgerrors.Error {
	return pkgerrors.Conflict(operation, "wallet").
		WithCode(CodeWalletAlreadyLinked).
		WithMeta("address", address)
}

// WalletNotActive creates a wallet not active error.
func WalletNotActive(operation string, walletID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "wallet is not active").
		WithCode(CodeWalletNotActive).
		WithMeta("wallet_id", walletID)
}

// WalletUnverified creates a wallet unverified error.
func WalletUnverified(operation string, walletID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "wallet is unverified").
		WithCode(CodeWalletUnverified).
		WithMeta("wallet_id", walletID)
}

// InvalidAddress creates an invalid address error.
func InvalidAddress(operation string, address string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid wallet address: "+reason).
		WithCode(CodeInvalidAddress).
		WithMeta("address", address)
}

// ChainNotSupported creates a chain not supported error.
func ChainNotSupported(operation string, chainID int) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "chain not supported").
		WithCode(CodeChainNotSupported).
		WithMeta("chain_id", chainID)
}

// SignatureInvalid creates a signature invalid error.
func SignatureInvalid(operation string, address string) *pkgerrors.Error {
	return pkgerrors.Unauthorized(operation, "signature verification failed").
		WithCode(CodeSignatureInvalid).
		WithMeta("address", address)
}

// ChallengeExpired creates a challenge expired error.
func ChallengeExpired(operation string, address string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "authentication challenge has expired").
		WithCode(CodeChallengeExpired).
		WithMeta("address", address)
}

// ChallengeNotFound creates a challenge not found error.
func ChallengeNotFound(operation string, address string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "challenge").
		WithCode(CodeChallengeNotFound).
		WithMeta("address", address)
}

// PrimaryWalletRequired creates a primary wallet required error.
func PrimaryWalletRequired(operation string, walletID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "a primary wallet is required").
		WithCode(CodePrimaryWalletRequired).
		WithMeta("wallet_id", walletID)
}
