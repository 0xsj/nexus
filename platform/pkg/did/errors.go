package did

import "github.com/0xsj/nexus/platform/pkg/errors"

// DID-specific error codes.
const (
	CodeInvalidDID        errors.Code = "INVALID_DID"
	CodeUnsupportedMethod errors.Code = "UNSUPPORTED_DID_METHOD"
	CodeResolutionFailed  errors.Code = "DID_RESOLUTION_FAILED"
	CodeDIDNotFound       errors.Code = "DID_NOT_FOUND"
	CodeInvalidDocument   errors.Code = "INVALID_DID_DOCUMENT"
	CodeDeactivated       errors.Code = "DID_DEACTIVATED"
)

// ErrInvalidDID creates an error for invalid DID format.
func ErrInvalidDID(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidDID)
}

// ErrUnsupportedMethod creates an error for unsupported DID methods.
func ErrUnsupportedMethod(operation string, method string) *errors.Error {
	return errors.Validation(operation, "unsupported DID method: "+method).
		WithCode(CodeUnsupportedMethod).
		WithMeta("method", method)
}

// ErrResolutionFailed creates an error for DID resolution failures.
func ErrResolutionFailed(operation string, did string, err error) *errors.Error {
	e := errors.Infrastructure(operation, err).
		WithCode(CodeResolutionFailed).
		WithMeta("did", did)
	if err != nil {
		e = e.WithMeta("cause", err.Error())
	}
	return e
}

// ErrDIDNotFound creates an error when a DID cannot be resolved.
func ErrDIDNotFound(operation string, did string) *errors.Error {
	return errors.NotFound(operation, "DID").
		WithCode(CodeDIDNotFound).
		WithMeta("did", did)
}

// ErrInvalidDocument creates an error for invalid DID documents.
func ErrInvalidDocument(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidDocument)
}

// ErrDeactivated creates an error when a DID has been deactivated.
func ErrDeactivated(operation string, did string) *errors.Error {
	return errors.Domain(operation, "DID has been deactivated").
		WithCode(CodeDeactivated).
		WithMeta("did", did)
}
