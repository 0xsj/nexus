package errors

// Code is a machine-readable error identifier.
// Clients use codes for programmatic error handling (e.g., showing specific UI messages).
//
// Codes are strings to allow domains to define their own:
//
//	// in pkg/errors/codes.go (generic)
//	const CodeNotFound Code = "NOT_FOUND"
//
//	// in internal/credential/domain/errors.go (domain-specific)
//	const CodeCredentialExpired errors.Code = "CREDENTIAL_EXPIRED"
type Code string

// Generic error codes.
// These cover common cases across all domains.
// Domains should define their own specific codes for domain-specific errors.
const (
	// CodeUnknown indicates an unknown or unclassified error.
	CodeUnknown Code = "UNKNOWN"

	// CodeInternal indicates an internal server error.
	CodeInternal Code = "INTERNAL_ERROR"

	// CodeNotFound indicates a resource was not found.
	CodeNotFound Code = "NOT_FOUND"

	// CodeAlreadyExists indicates a resource already exists.
	CodeAlreadyExists Code = "ALREADY_EXISTS"

	// CodeConflict indicates a conflict (e.g., version mismatch).
	CodeConflict Code = "CONFLICT"

	// CodeInvalidInput indicates invalid input from the client.
	CodeInvalidInput Code = "INVALID_INPUT"

	// CodeValidationFailed indicates validation rules were not satisfied.
	CodeValidationFailed Code = "VALIDATION_FAILED"

	// CodeUnauthenticated indicates missing or invalid authentication.
	CodeUnauthenticated Code = "UNAUTHENTICATED"

	// CodeUnauthorized indicates insufficient permissions.
	CodeUnauthorized Code = "UNAUTHORIZED"

	// CodeForbidden indicates the operation is forbidden.
	CodeForbidden Code = "FORBIDDEN"

	// CodeTimeout indicates an operation timed out.
	CodeTimeout Code = "TIMEOUT"

	// CodeUnavailable indicates a service is unavailable.
	CodeUnavailable Code = "UNAVAILABLE"

	// CodeRateLimitExceeded indicates too many requests.
	CodeRateLimitExceeded Code = "RATE_LIMIT_EXCEEDED"

	// CodeVersionMismatch indicates an optimistic locking failure.
	CodeVersionMismatch Code = "VERSION_MISMATCH"
)

// String returns the string representation of the code.
func (c Code) String() string {
	return string(c)
}

// IsEmpty returns true if the code is empty.
func (c Code) IsEmpty() bool {
	return c == ""
}