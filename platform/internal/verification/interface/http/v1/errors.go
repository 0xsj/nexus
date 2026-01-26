package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/response"
)

// ============================================================================
// Error Codes
// ============================================================================

const (
	ErrCodeVerificationNotFound     = "VERIFICATION_NOT_FOUND"
	ErrCodeVerificationExists       = "VERIFICATION_ALREADY_EXISTS"
	ErrCodeVerificationInvalidState = "VERIFICATION_INVALID_STATE"
	ErrCodeVerificationExpired      = "VERIFICATION_EXPIRED"
	ErrCodeVerificationFailed       = "VERIFICATION_FAILED"
	ErrCodeProviderNotSupported     = "PROVIDER_NOT_SUPPORTED"
	ErrCodeProviderAuthFailed       = "PROVIDER_AUTH_FAILED"
	ErrCodeProviderUnavailable      = "PROVIDER_UNAVAILABLE"
	ErrCodeOAuthStateMismatch       = "OAUTH_STATE_MISMATCH"
	ErrCodeOAuthCodeInvalid         = "OAUTH_CODE_INVALID"
	ErrCodeOAuthCodeExpired         = "OAUTH_CODE_EXPIRED"
	ErrCodeCredentialIssuanceFailed = "CREDENTIAL_ISSUANCE_FAILED"
	ErrCodeValidation               = "VALIDATION_ERROR"
	ErrCodeConflict                 = "CONCURRENCY_CONFLICT"
	ErrCodeInternal                 = "INTERNAL_ERROR"
)

// ============================================================================
// Error Mapping
// ============================================================================

// MapError maps domain/application errors to HTTP status and response.
func MapError(err error) (int, *response.ErrorResponse) {
	if err == nil {
		return http.StatusOK, nil
	}

	// Check for domain errors first
	if domainErr := mapDomainError(err); domainErr != nil {
		return domainErr.status, domainErr.response
	}

	// Aggregate not found
	if eventsourcing.IsAggregateNotFound(err) {
		return http.StatusNotFound, &response.ErrorResponse{
			Code:    ErrCodeVerificationNotFound,
			Message: "Verification not found",
			Details: err.Error(),
		}
	}

	// Concurrency conflict
	if eventsourcing.IsConcurrencyConflict(err) {
		return http.StatusConflict, &response.ErrorResponse{
			Code:    ErrCodeConflict,
			Message: "Verification was modified by another request",
			Details: err.Error(),
		}
	}

	// Aggregate validation (domain validation)
	if eventsourcing.IsAggregateValidation(err) {
		return http.StatusBadRequest, &response.ErrorResponse{
			Code:    ErrCodeVerificationInvalidState,
			Message: "Invalid verification operation",
			Details: err.Error(),
		}
	}

	// Command validation
	if cqrs.IsCommandValidation(err) {
		return http.StatusBadRequest, &response.ErrorResponse{
			Code:    ErrCodeValidation,
			Message: "Validation error",
			Details: err.Error(),
		}
	}

	// Request validation errors
	if validationErrs, ok := pkghttp.AsValidationErrors(err); ok {
		fields := make([]response.FieldError, len(validationErrs.Errors))
		for i, e := range validationErrs.Errors {
			fields[i] = response.FieldError{
				Field:   e.Field,
				Message: e.Message,
			}
		}
		return http.StatusBadRequest, &response.ErrorResponse{
			Code:    ErrCodeValidation,
			Message: "Validation failed",
			Fields:  fields,
		}
	}

	// Decode errors
	if pkghttp.IsDecodeError(err) {
		return http.StatusBadRequest, &response.ErrorResponse{
			Code:    ErrCodeValidation,
			Message: "Invalid request body",
			Details: err.Error(),
		}
	}

	// Check for pkg/errors types
	if pkgErr := mapPkgError(err); pkgErr != nil {
		return pkgErr.status, pkgErr.response
	}

	// Default to internal error
	return http.StatusInternalServerError, &response.ErrorResponse{
		Code:    ErrCodeInternal,
		Message: "An internal error occurred",
		Details: err.Error(),
	}
}

type mappedError struct {
	status   int
	response *response.ErrorResponse
}

func mapDomainError(err error) *mappedError {
	// Verification not found
	if domain.IsVerificationNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &response.ErrorResponse{
				Code:    ErrCodeVerificationNotFound,
				Message: "Verification not found",
			},
		}
	}

	// Verification already exists
	if domain.IsVerificationAlreadyExists(err) {
		return &mappedError{
			status: http.StatusConflict,
			response: &response.ErrorResponse{
				Code:    ErrCodeVerificationExists,
				Message: "Verification already exists for this provider",
			},
		}
	}

	// Verification expired
	if domain.IsVerificationExpired(err) {
		return &mappedError{
			status: http.StatusGone,
			response: &response.ErrorResponse{
				Code:    ErrCodeVerificationExpired,
				Message: "Verification has expired",
			},
		}
	}

	// Verification failed
	if domain.IsVerificationFailed(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    ErrCodeVerificationFailed,
				Message: "Verification failed",
			},
		}
	}

	// Invalid verification state
	if domain.IsVerificationInvalidState(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    ErrCodeVerificationInvalidState,
				Message: "Invalid verification state for this operation",
			},
		}
	}

	// Provider not supported
	if domain.IsProviderNotSupported(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    ErrCodeProviderNotSupported,
				Message: "Provider is not supported",
			},
		}
	}

	// Provider unavailable
	if domain.IsProviderUnavailable(err) {
		return &mappedError{
			status: http.StatusServiceUnavailable,
			response: &response.ErrorResponse{
				Code:    ErrCodeProviderUnavailable,
				Message: "Provider is temporarily unavailable",
			},
		}
	}

	// Provider auth failed
	if domain.IsProviderAuthFailed(err) {
		return &mappedError{
			status: http.StatusBadGateway,
			response: &response.ErrorResponse{
				Code:    ErrCodeProviderAuthFailed,
				Message: "Authentication with provider failed",
			},
		}
	}

	// Provider rate limited
	if domain.IsProviderRateLimited(err) {
		return &mappedError{
			status: http.StatusTooManyRequests,
			response: &response.ErrorResponse{
				Code:    "PROVIDER_RATE_LIMITED",
				Message: "Rate limited by provider",
			},
		}
	}

	// OAuth state mismatch
	if domain.IsOAuthStateMismatch(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    ErrCodeOAuthStateMismatch,
				Message: "OAuth state mismatch or invalid",
			},
		}
	}

	// OAuth code invalid
	if domain.IsOAuthCodeInvalid(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    ErrCodeOAuthCodeInvalid,
				Message: "OAuth authorization code is invalid",
			},
		}
	}

	// OAuth code expired
	if domain.IsOAuthCodeExpired(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    ErrCodeOAuthCodeExpired,
				Message: "OAuth authorization code has expired",
			},
		}
	}

	// Credential issuance failed
	if domain.IsCredentialIssuanceFailed(err) {
		return &mappedError{
			status: http.StatusInternalServerError,
			response: &response.ErrorResponse{
				Code:    ErrCodeCredentialIssuanceFailed,
				Message: "Failed to issue credential",
			},
		}
	}

	// Insufficient data
	if domain.IsInsufficientData(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &response.ErrorResponse{
				Code:    "INSUFFICIENT_DATA",
				Message: "Insufficient data from provider to issue credential",
			},
		}
	}

	return nil
}

func mapPkgError(err error) *mappedError {
	var e *errors.Error
	if !errors.As(err, &e) {
		return nil
	}

	status := e.Kind.HTTPStatus()
	message := e.Message

	var code string
	var errorText string

	switch e.Kind {
	case errors.KindValidation:
		code = ErrCodeValidation
		errorText = "Validation error"
	case errors.KindNotFound:
		code = ErrCodeVerificationNotFound
		errorText = "Not found"
	case errors.KindConflict:
		code = ErrCodeConflict
		errorText = "Conflict"
	case errors.KindUnauthorized:
		code = "UNAUTHORIZED"
		errorText = "Unauthorized"
	case errors.KindForbidden:
		code = "FORBIDDEN"
		errorText = "Forbidden"
	case errors.KindRateLimit:
		code = "RATE_LIMIT_EXCEEDED"
		errorText = "Rate limit exceeded"
	case errors.KindDomain:
		code = "DOMAIN_ERROR"
		errorText = "Domain error"
	case errors.KindInfrastructure:
		code = "SERVICE_UNAVAILABLE"
		errorText = "Service unavailable"
	case errors.KindTimeout:
		code = "TIMEOUT"
		errorText = "Request timeout"
	case errors.KindInternal:
		code = ErrCodeInternal
		errorText = "Internal server error"
		message = "" // Don't expose internal error details
	default:
		code = ErrCodeInternal
		errorText = "Internal server error"
		message = ""
	}

	return &mappedError{
		status: status,
		response: &response.ErrorResponse{
			Code:    code,
			Message: errorText,
			Details: message,
		},
	}
}

// ============================================================================
// Error Response Helpers
// ============================================================================

// WriteError writes an error response based on the error type.
func WriteError(w http.ResponseWriter, err error) {
	status, errResp := MapError(err)
	response.Error(w, status, errResp)
}

// NotFound writes a 404 not found response.
func NotFound(w http.ResponseWriter, id string) {
	response.NotFound(w, &response.ErrorResponse{
		Code:    ErrCodeVerificationNotFound,
		Message: "Verification not found",
		Details: "verification with ID '" + id + "' not found",
	})
}

// ValidationFailed writes a 400 validation error response.
func ValidationFailed(w http.ResponseWriter, fields []response.FieldError) {
	response.BadRequest(w, &response.ErrorResponse{
		Code:    ErrCodeValidation,
		Message: "Validation failed",
		Fields:  fields,
	})
}

// InvalidState writes a 400 invalid state error response.
func InvalidState(w http.ResponseWriter, message string) {
	response.BadRequest(w, &response.ErrorResponse{
		Code:    ErrCodeVerificationInvalidState,
		Message: message,
	})
}

// ProviderError writes an error response for provider-related errors.
func ProviderError(w http.ResponseWriter, provider string, message string) {
	response.Error(w, http.StatusBadGateway, &response.ErrorResponse{
		Code:    ErrCodeProviderAuthFailed,
		Message: message,
		Details: "provider: " + provider,
	})
}

// OAuthError writes an error response for OAuth-related errors.
func OAuthError(w http.ResponseWriter, code string, message string) {
	response.BadRequest(w, &response.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
