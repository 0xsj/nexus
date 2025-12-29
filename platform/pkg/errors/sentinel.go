package errors

import "errors"

// Sentinel errors for use with errors.Is().
// These provide fast type checking without inspecting the full error chain.
//
// Usage:
//
//	if errors.Is(err, errors.ErrNotFound) {
//	    // handle not found
//	}
//
// Domains can define their own sentinels:
//
//	// in internal/credential/domain/errors.go
//	var ErrCredentialExpired = errors.New("credential expired")
var (
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrConflict indicates a resource conflict (already exists, version mismatch).
	ErrConflict = errors.New("conflict")

	// ErrValidation indicates a validation failure.
	ErrValidation = errors.New("validation error")

	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates insufficient permissions.
	ErrForbidden = errors.New("forbidden")

	// ErrDomain indicates a business rule violation.
	ErrDomain = errors.New("domain error")

	// ErrInfrastructure indicates an infrastructure/external system failure.
	ErrInfrastructure = errors.New("infrastructure error")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("timeout")

	// ErrRateLimit indicates rate limit exceeded.
	ErrRateLimit = errors.New("rate limit exceeded")

	// ErrInternal indicates an internal/unexpected error.
	ErrInternal = errors.New("internal error")
)

// sentinelForKind returns the appropriate sentinel error for a given kind.
// Used internally to set the wrapped error for errors.Is() support.
func sentinelForKind(k Kind) error {
	switch k {
	case KindNotFound:
		return ErrNotFound
	case KindConflict:
		return ErrConflict
	case KindValidation:
		return ErrValidation
	case KindUnauthorized:
		return ErrUnauthorized
	case KindForbidden:
		return ErrForbidden
	case KindDomain:
		return ErrDomain
	case KindInfrastructure:
		return ErrInfrastructure
	case KindTimeout:
		return ErrTimeout
	case KindRateLimit:
		return ErrRateLimit
	case KindInternal:
		return ErrInternal
	default:
		return ErrInternal
	}
}
