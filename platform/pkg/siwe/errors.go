package siwe

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Error codes for SIWE operations.
const (
	ErrCodeInvalidMessage     errors.Code = "SIWE_INVALID_MESSAGE"
	ErrCodeInvalidSignature   errors.Code = "SIWE_INVALID_SIGNATURE"
	ErrCodeInvalidAddress     errors.Code = "SIWE_INVALID_ADDRESS"
	ErrCodeInvalidDomain      errors.Code = "SIWE_INVALID_DOMAIN"
	ErrCodeInvalidNonce       errors.Code = "SIWE_INVALID_NONCE"
	ErrCodeInvalidURI         errors.Code = "SIWE_INVALID_URI"
	ErrCodeInvalidChainID     errors.Code = "SIWE_INVALID_CHAIN_ID"
	ErrCodeMessageExpired     errors.Code = "SIWE_MESSAGE_EXPIRED"
	ErrCodeMessageNotYetValid errors.Code = "SIWE_MESSAGE_NOT_YET_VALID"
	ErrCodeAddressMismatch    errors.Code = "SIWE_ADDRESS_MISMATCH"
	ErrCodeDomainMismatch     errors.Code = "SIWE_DOMAIN_MISMATCH"
	ErrCodeNonceMismatch      errors.Code = "SIWE_NONCE_MISMATCH"
)

// ErrInvalidMessage indicates the SIWE message format is invalid.
func ErrInvalidMessage(op string, detail string) *errors.Error {
	return errors.Validation(op, "invalid SIWE message: "+detail).
		WithCode(ErrCodeInvalidMessage)
}

// ErrInvalidSignature indicates the signature is invalid.
func ErrInvalidSignature(op string) *errors.Error {
	return errors.Unauthorized(op, "invalid signature").
		WithCode(ErrCodeInvalidSignature)
}

// ErrInvalidAddress indicates the Ethereum address is invalid.
func ErrInvalidAddress(op string, address string) *errors.Error {
	return errors.Validation(op, "invalid Ethereum address: "+address).
		WithCode(ErrCodeInvalidAddress)
}

// ErrInvalidDomain indicates the domain is invalid.
func ErrInvalidDomain(op string, domain string) *errors.Error {
	return errors.Validation(op, "invalid domain: "+domain).
		WithCode(ErrCodeInvalidDomain)
}

// ErrInvalidNonce indicates the nonce is invalid.
func ErrInvalidNonce(op string) *errors.Error {
	return errors.Validation(op, "invalid nonce").
		WithCode(ErrCodeInvalidNonce)
}

// ErrInvalidURI indicates the URI is invalid.
func ErrInvalidURI(op string, uri string) *errors.Error {
	return errors.Validation(op, "invalid URI: "+uri).
		WithCode(ErrCodeInvalidURI)
}

// ErrInvalidChainID indicates the chain ID is invalid.
func ErrInvalidChainID(op string, chainID string) *errors.Error {
	return errors.Validation(op, "invalid chain ID: "+chainID).
		WithCode(ErrCodeInvalidChainID)
}

// ErrMessageExpired indicates the SIWE message has expired.
func ErrMessageExpired(op string) *errors.Error {
	return errors.Unauthorized(op, "SIWE message has expired").
		WithCode(ErrCodeMessageExpired)
}

// ErrMessageNotYetValid indicates the SIWE message is not yet valid.
func ErrMessageNotYetValid(op string) *errors.Error {
	return errors.Unauthorized(op, "SIWE message is not yet valid").
		WithCode(ErrCodeMessageNotYetValid)
}

// ErrAddressMismatch indicates the recovered address doesn't match.
func ErrAddressMismatch(op string) *errors.Error {
	return errors.Unauthorized(op, "recovered address does not match").
		WithCode(ErrCodeAddressMismatch)
}

// ErrDomainMismatch indicates the domain doesn't match expected.
func ErrDomainMismatch(op string, expected, actual string) *errors.Error {
	return errors.Validation(op, "domain mismatch: expected "+expected+", got "+actual).
		WithCode(ErrCodeDomainMismatch)
}

// ErrNonceMismatch indicates the nonce doesn't match expected.
func ErrNonceMismatch(op string) *errors.Error {
	return errors.Unauthorized(op, "nonce mismatch").
		WithCode(ErrCodeNonceMismatch)
}

// ============================================================================
// Error Checks
// ============================================================================

// IsInvalidMessage checks if the error is an invalid message error.
func IsInvalidMessage(err error) bool {
	return errors.HasCode(err, ErrCodeInvalidMessage)
}

// IsInvalidSignature checks if the error is an invalid signature error.
func IsInvalidSignature(err error) bool {
	return errors.HasCode(err, ErrCodeInvalidSignature)
}

// IsMessageExpired checks if the error is a message expired error.
func IsMessageExpired(err error) bool {
	return errors.HasCode(err, ErrCodeMessageExpired)
}

// IsAddressMismatch checks if the error is an address mismatch error.
func IsAddressMismatch(err error) bool {
	return errors.HasCode(err, ErrCodeAddressMismatch)
}
