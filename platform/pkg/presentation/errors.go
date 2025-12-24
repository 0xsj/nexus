package presentation

import "github.com/0xsj/nexus/platform/pkg/errors"

// Presentation-specific error codes.
const (
	CodeInvalidPresentation errors.Code = "INVALID_PRESENTATION"
	CodePresentationExpired errors.Code = "PRESENTATION_EXPIRED"
	CodeInvalidHolder       errors.Code = "INVALID_HOLDER"
	CodeInvalidChallenge    errors.Code = "INVALID_CHALLENGE"
	CodeInvalidDomain       errors.Code = "INVALID_DOMAIN"
	CodeCredentialMissing   errors.Code = "CREDENTIAL_MISSING"
	CodeCredentialInvalid   errors.Code = "CREDENTIAL_INVALID"
	CodeSigningFailed       errors.Code = "VP_SIGNING_FAILED"
	CodeVerificationFailed  errors.Code = "VP_VERIFICATION_FAILED"
	CodeUnsupportedFormat   errors.Code = "UNSUPPORTED_VP_FORMAT"
)

// ErrInvalidPresentation creates an error for invalid presentation structure.
func ErrInvalidPresentation(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidPresentation)
}

// ErrPresentationExpired creates an error for expired presentations.
func ErrPresentationExpired(operation string, presentationID string) *errors.Error {
	return errors.Domain(operation, "presentation has expired").
		WithCode(CodePresentationExpired).
		WithMeta("presentation_id", presentationID)
}

// ErrInvalidHolder creates an error for invalid holder.
func ErrInvalidHolder(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidHolder)
}

// ErrInvalidChallenge creates an error for challenge mismatch.
func ErrInvalidChallenge(operation string) *errors.Error {
	return errors.Validation(operation, "challenge does not match").
		WithCode(CodeInvalidChallenge)
}

// ErrInvalidDomain creates an error for domain mismatch.
func ErrInvalidDomain(operation string, expected, actual string) *errors.Error {
	return errors.Validation(operation, "domain does not match").
		WithCode(CodeInvalidDomain).
		WithMeta("expected", expected).
		WithMeta("actual", actual)
}

// ErrCredentialMissing creates an error when required credentials are missing.
func ErrCredentialMissing(operation string, credentialType string) *errors.Error {
	return errors.Validation(operation, "required credential missing: "+credentialType).
		WithCode(CodeCredentialMissing).
		WithMeta("credential_type", credentialType)
}

// ErrCredentialInvalid creates an error when a credential in the presentation is invalid.
func ErrCredentialInvalid(operation string, credentialID string, reason string) *errors.Error {
	return errors.Validation(operation, "credential invalid: "+reason).
		WithCode(CodeCredentialInvalid).
		WithMeta("credential_id", credentialID)
}

// ErrSigningFailed creates an error for presentation signing failures.
func ErrSigningFailed(operation string, err error) *errors.Error {
	return errors.Internal(operation, err).
		WithCode(CodeSigningFailed)
}

// ErrVerificationFailed creates an error for presentation verification failures.
func ErrVerificationFailed(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeVerificationFailed)
}

// ErrUnsupportedFormat creates an error for unsupported presentation formats.
func ErrUnsupportedFormat(operation string, format string) *errors.Error {
	return errors.Validation(operation, "unsupported presentation format: "+format).
		WithCode(CodeUnsupportedFormat).
		WithMeta("format", format)
}
