package crypto

import "github.com/0xsj/nexus/platform/pkg/errors"

// Crypto-specific error codes.
// Extend the generic codes from pkg/errors.
const (
	CodeInvalidKey           errors.Code = "INVALID_KEY"
	CodeInvalidSignature     errors.Code = "INVALID_SIGNATURE"
	CodeKeyGenerationFailed  errors.Code = "KEY_GENERATION_FAILED"
	CodeSigningFailed        errors.Code = "SIGNING_FAILED"
	CodeVerificationFailed   errors.Code = "VERIFICATION_FAILED"
	CodeUnsupportedAlgorithm errors.Code = "UNSUPPORTED_ALGORITHM"
	CodeInvalidKeyFormat     errors.Code = "INVALID_KEY_FORMAT"
	CodeKeyMismatch          errors.Code = "KEY_MISMATCH"
)

// ErrInvalidKey creates an error for invalid key data.
func ErrInvalidKey(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidKey)
}

// ErrInvalidSignature creates an error for invalid signature data.
func ErrInvalidSignature(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidSignature)
}

// ErrKeyGenerationFailed creates an error for key generation failures.
func ErrKeyGenerationFailed(operation string, err error) *errors.Error {
	return errors.Internal(operation, err).
		WithCode(CodeKeyGenerationFailed)
}

// ErrSigningFailed creates an error for signing failures.
func ErrSigningFailed(operation string, err error) *errors.Error {
	return errors.Internal(operation, err).
		WithCode(CodeSigningFailed)
}

// ErrVerificationFailed creates an error for signature verification failures.
func ErrVerificationFailed(operation string) *errors.Error {
	return errors.Validation(operation, "signature verification failed").
		WithCode(CodeVerificationFailed)
}

// ErrUnsupportedAlgorithm creates an error for unsupported algorithms.
func ErrUnsupportedAlgorithm(operation string, algorithm Algorithm) *errors.Error {
	return errors.Validation(operation, "unsupported algorithm: "+algorithm.String()).
		WithCode(CodeUnsupportedAlgorithm).
		WithMeta("algorithm", algorithm.String())
}

// ErrInvalidKeyFormat creates an error for invalid key encoding/format.
func ErrInvalidKeyFormat(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidKeyFormat)
}

// ErrKeyMismatch creates an error when public/private keys don't match.
func ErrKeyMismatch(operation string) *errors.Error {
	return errors.Validation(operation, "public key does not match private key").
		WithCode(CodeKeyMismatch)
}
