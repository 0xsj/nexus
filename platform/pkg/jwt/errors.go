package jwt

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Error codes for JWT operations.
const (
	ErrCodeInvalidToken     errors.Code = "JWT_INVALID_TOKEN"
	ErrCodeTokenExpired     errors.Code = "JWT_TOKEN_EXPIRED"
	ErrCodeTokenNotYetValid errors.Code = "JWT_TOKEN_NOT_YET_VALID"
	ErrCodeInvalidSignature errors.Code = "JWT_INVALID_SIGNATURE"
	ErrCodeInvalidHeader    errors.Code = "JWT_INVALID_HEADER"
	ErrCodeInvalidClaims    errors.Code = "JWT_INVALID_CLAIMS"
	ErrCodeMissingClaim     errors.Code = "JWT_MISSING_CLAIM"
	ErrCodeSigningFailed    errors.Code = "JWT_SIGNING_FAILED"
	ErrCodeEncodingFailed   errors.Code = "JWT_ENCODING_FAILED"
	ErrCodeDecodingFailed   errors.Code = "JWT_DECODING_FAILED"
	ErrCodeUnsupportedAlg   errors.Code = "JWT_UNSUPPORTED_ALGORITHM"
)

// ErrInvalidToken indicates the token format is invalid.
func ErrInvalidToken(op string, detail string) *errors.Error {
	return errors.Validation(op, "invalid token: "+detail).
		WithCode(ErrCodeInvalidToken)
}

// ErrTokenExpired indicates the token has expired.
func ErrTokenExpired(op string) *errors.Error {
	return errors.Unauthorized(op, "token has expired").
		WithCode(ErrCodeTokenExpired)
}

// ErrTokenNotYetValid indicates the token is not yet valid (nbf claim).
func ErrTokenNotYetValid(op string) *errors.Error {
	return errors.Unauthorized(op, "token is not yet valid").
		WithCode(ErrCodeTokenNotYetValid)
}

// ErrInvalidSignature indicates the signature verification failed.
func ErrInvalidSignature(op string) *errors.Error {
	return errors.Unauthorized(op, "invalid token signature").
		WithCode(ErrCodeInvalidSignature)
}

// ErrInvalidHeader indicates the header is malformed.
func ErrInvalidHeader(op string, detail string) *errors.Error {
	return errors.Validation(op, "invalid header: "+detail).
		WithCode(ErrCodeInvalidHeader)
}

// ErrInvalidClaims indicates the claims are malformed.
func ErrInvalidClaims(op string, detail string) *errors.Error {
	return errors.Validation(op, "invalid claims: "+detail).
		WithCode(ErrCodeInvalidClaims)
}

// ErrMissingClaim indicates a required claim is missing.
func ErrMissingClaim(op string, claim string) *errors.Error {
	return errors.Validation(op, "missing required claim: "+claim).
		WithCode(ErrCodeMissingClaim)
}

// ErrSigningFailed indicates signing operation failed.
func ErrSigningFailed(op string, err error) *errors.Error {
	return errors.Internal(op, err).
		WithCode(ErrCodeSigningFailed).
		WithMessage("failed to sign token")
}

// ErrEncodingFailed indicates encoding operation failed.
func ErrEncodingFailed(op string, err error) *errors.Error {
	return errors.Internal(op, err).
		WithCode(ErrCodeEncodingFailed).
		WithMessage("failed to encode token")
}

// ErrDecodingFailed indicates decoding operation failed.
func ErrDecodingFailed(op string, err error) *errors.Error {
	return errors.Validation(op, "failed to decode token").
		WithCode(ErrCodeDecodingFailed)
}

// ErrUnsupportedAlgorithm indicates the algorithm is not supported.
func ErrUnsupportedAlgorithm(op string, alg string) *errors.Error {
	return errors.Validation(op, "unsupported algorithm: "+alg).
		WithCode(ErrCodeUnsupportedAlg)
}

// ============================================================================
// Error Checks
// ============================================================================

// IsTokenExpired checks if the error is a token expired error.
func IsTokenExpired(err error) bool {
	return errors.HasCode(err, ErrCodeTokenExpired)
}

// IsInvalidToken checks if the error is an invalid token error.
func IsInvalidToken(err error) bool {
	return errors.HasCode(err, ErrCodeInvalidToken)
}

// IsInvalidSignature checks if the error is an invalid signature error.
func IsInvalidSignature(err error) bool {
	return errors.HasCode(err, ErrCodeInvalidSignature)
}

// IsTokenNotYetValid checks if the error is a token not yet valid error.
func IsTokenNotYetValid(err error) bool {
	return errors.HasCode(err, ErrCodeTokenNotYetValid)
}
