package errors

import "net/http"

// Kind categorizes errors for protocol mapping.
// Each kind maps to appropriate HTTP status codes, gRPC codes, etc.
//
// Kind represents "what class of error" rather than "what specific error".
// Domains define specific error codes; Kind determines how they're communicated.
type Kind int

const (
	// KindUnknown indicates an unknown or unclassified error.
	// Maps to HTTP 500. Use sparingly — prefer specific kinds.
	KindUnknown Kind = iota

	// KindValidation indicates invalid input from the client.
	// Maps to HTTP 400 (Bad Request).
	// Examples: malformed JSON, missing required fields, invalid email format
	KindValidation

	// KindNotFound indicates a requested resource doesn't exist.
	// Maps to HTTP 404 (Not Found).
	// Examples: user not found, credential not found, DID not resolvable
	KindNotFound

	// KindConflict indicates a resource already exists or version conflict.
	// Maps to HTTP 409 (Conflict).
	// Examples: email already registered, duplicate credential, optimistic lock failure
	KindConflict

	// KindUnauthorized indicates missing or invalid authentication.
	// Maps to HTTP 401 (Unauthorized).
	// Examples: missing token, expired token, invalid signature
	KindUnauthorized

	// KindForbidden indicates insufficient permissions.
	// Maps to HTTP 403 (Forbidden).
	// Examples: user lacks role, credential not owned by requester
	KindForbidden

	// KindDomain indicates a business rule violation.
	// Maps to HTTP 422 (Unprocessable Entity).
	// Examples: credential expired, DID revoked, verification already completed
	KindDomain

	// KindInfrastructure indicates an external system failure.
	// Maps to HTTP 503 (Service Unavailable).
	// Examples: database down, GitHub API unreachable, blockchain RPC failed
	KindInfrastructure

	// KindTimeout indicates an operation exceeded its deadline.
	// Maps to HTTP 504 (Gateway Timeout).
	// Examples: database query timeout, external API timeout
	KindTimeout

	// KindRateLimit indicates too many requests.
	// Maps to HTTP 429 (Too Many Requests).
	// Examples: API rate limit exceeded, verification attempts exceeded
	KindRateLimit

	// KindInternal indicates an unexpected internal error.
	// Maps to HTTP 500 (Internal Server Error).
	// Examples: nil pointer, unhandled case, programming error
	KindInternal
)

// String returns the string representation of the kind.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "unknown"
	case KindValidation:
		return "validation"
	case KindNotFound:
		return "not_found"
	case KindConflict:
		return "conflict"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindDomain:
		return "domain"
	case KindInfrastructure:
		return "infrastructure"
	case KindTimeout:
		return "timeout"
	case KindRateLimit:
		return "rate_limit"
	case KindInternal:
		return "internal"
	default:
		return "unknown"
	}
}

// HTTPStatus returns the appropriate HTTP status code for the kind.
func (k Kind) HTTPStatus() int {
	switch k {
	case KindValidation:
		return http.StatusBadRequest // 400
	case KindUnauthorized:
		return http.StatusUnauthorized // 401
	case KindForbidden:
		return http.StatusForbidden // 403
	case KindNotFound:
		return http.StatusNotFound // 404
	case KindConflict:
		return http.StatusConflict // 409
	case KindDomain:
		return http.StatusUnprocessableEntity // 422
	case KindRateLimit:
		return http.StatusTooManyRequests // 429
	case KindInternal:
		return http.StatusInternalServerError // 500
	case KindInfrastructure:
		return http.StatusServiceUnavailable // 503
	case KindTimeout:
		return http.StatusGatewayTimeout // 504
	default:
		return http.StatusInternalServerError // 500
	}
}

// IsRetryableByDefault returns whether errors of this kind are typically retryable.
// Individual errors can override this via WithRetryable().
func (k Kind) IsRetryableByDefault() bool {
	switch k {
	case KindInfrastructure, KindTimeout, KindRateLimit:
		return true
	default:
		return false
	}
}

// IsClientError returns true if the error is due to client input.
func (k Kind) IsClientError() bool {
	switch k {
	case KindValidation, KindNotFound, KindConflict, KindUnauthorized, KindForbidden, KindDomain, KindRateLimit:
		return true
	default:
		return false
	}
}

// IsServerError returns true if the error is due to server/infrastructure issues.
func (k Kind) IsServerError() bool {
	switch k {
	case KindInfrastructure, KindTimeout, KindInternal, KindUnknown:
		return true
	default:
		return false
	}
}
