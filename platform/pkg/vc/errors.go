package vc

import "github.com/0xsj/nexus/platform/pkg/errors"

// VC-specific error codes.
const (
	CodeInvalidCredential   errors.Code = "INVALID_CREDENTIAL"
	CodeCredentialExpired   errors.Code = "CREDENTIAL_EXPIRED"
	CodeCredentialRevoked   errors.Code = "CREDENTIAL_REVOKED"
	CodeCredentialNotActive errors.Code = "CREDENTIAL_NOT_ACTIVE"
	CodeInvalidProof        errors.Code = "INVALID_PROOF"
	CodeProofExpired        errors.Code = "PROOF_EXPIRED"
	CodeInvalidIssuer       errors.Code = "INVALID_ISSUER"
	CodeInvalidSubject      errors.Code = "INVALID_SUBJECT"
	CodeInvalidSchema       errors.Code = "INVALID_SCHEMA"
	CodeSigningFailed       errors.Code = "VC_SIGNING_FAILED"
	CodeVerificationFailed  errors.Code = "VC_VERIFICATION_FAILED"
	CodeUnsupportedFormat   errors.Code = "UNSUPPORTED_VC_FORMAT"
)

// ErrInvalidCredential creates an error for invalid credential structure.
func ErrInvalidCredential(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidCredential)
}

// ErrCredentialExpired creates an error for expired credentials.
func ErrCredentialExpired(operation string, credentialID string) *errors.Error {
	return errors.Domain(operation, "credential has expired").
		WithCode(CodeCredentialExpired).
		WithMeta("credential_id", credentialID)
}

// ErrCredentialRevoked creates an error for revoked credentials.
func ErrCredentialRevoked(operation string, credentialID string) *errors.Error {
	return errors.Domain(operation, "credential has been revoked").
		WithCode(CodeCredentialRevoked).
		WithMeta("credential_id", credentialID)
}

// ErrCredentialNotActive creates an error for credentials not yet active.
func ErrCredentialNotActive(operation string, credentialID string) *errors.Error {
	return errors.Domain(operation, "credential is not yet active").
		WithCode(CodeCredentialNotActive).
		WithMeta("credential_id", credentialID)
}

// ErrInvalidProof creates an error for invalid proof structure or signature.
func ErrInvalidProof(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidProof)
}

// ErrProofExpired creates an error for expired proofs.
func ErrProofExpired(operation string) *errors.Error {
	return errors.Domain(operation, "proof has expired").
		WithCode(CodeProofExpired)
}

// ErrInvalidIssuer creates an error for invalid issuer.
func ErrInvalidIssuer(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidIssuer)
}

// ErrInvalidSubject creates an error for invalid credential subject.
func ErrInvalidSubject(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidSubject)
}

// ErrInvalidSchema creates an error for schema validation failures.
func ErrInvalidSchema(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidSchema)
}

// ErrSigningFailed creates an error for credential signing failures.
func ErrSigningFailed(operation string, err error) *errors.Error {
	return errors.Internal(operation, err).
		WithCode(CodeSigningFailed)
}

// ErrVerificationFailed creates an error for credential verification failures.
func ErrVerificationFailed(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeVerificationFailed)
}

// ErrUnsupportedFormat creates an error for unsupported credential formats.
func ErrUnsupportedFormat(operation string, format string) *errors.Error {
	return errors.Validation(operation, "unsupported credential format: "+format).
		WithCode(CodeUnsupportedFormat).
		WithMeta("format", format)
}
