package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/errors"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
)

// ============================================================================
// Error Response
// ============================================================================

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ============================================================================
// Error Mapper
// ============================================================================

// MapError maps domain/application errors to HTTP responses.
func MapError(err error) (int, *ErrorResponse) {
	if err == nil {
		return http.StatusInternalServerError, &ErrorResponse{
			Error: "unknown error",
			Code:  "INTERNAL_ERROR",
		}
	}

	// Check for domain errors first
	if domainErr := mapDomainError(err); domainErr != nil {
		return domainErr.status, domainErr.response
	}

	// Check for pkg/errors types
	if pkgErr := mapPkgError(err); pkgErr != nil {
		return pkgErr.status, pkgErr.response
	}

	// Default to internal error
	return http.StatusInternalServerError, &ErrorResponse{
		Error:   "internal server error",
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	}
}

type mappedError struct {
	status   int
	response *ErrorResponse
}

func mapDomainError(err error) *mappedError {
	// User errors
	if domain.IsUserNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &ErrorResponse{
				Error: "user not found",
				Code:  "USER_NOT_FOUND",
			},
		}
	}

	if domain.IsUserAlreadyExists(err) {
		return &mappedError{
			status: http.StatusConflict,
			response: &ErrorResponse{
				Error: "user already exists",
				Code:  "USER_ALREADY_EXISTS",
			},
		}
	}

	if domain.IsUserDisabled(err) {
		return &mappedError{
			status: http.StatusForbidden,
			response: &ErrorResponse{
				Error: "user account is disabled",
				Code:  "USER_DISABLED",
			},
		}
	}

	// Session errors
	if domain.IsSessionNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &ErrorResponse{
				Error: "session not found",
				Code:  "SESSION_NOT_FOUND",
			},
		}
	}

	if domain.IsSessionExpired(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "session expired",
				Code:  "SESSION_EXPIRED",
			},
		}
	}

	if domain.IsSessionRevoked(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "session revoked",
				Code:  "SESSION_REVOKED",
			},
		}
	}

	// Token errors
	if domain.IsTokenInvalid(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "invalid token",
				Code:  "TOKEN_INVALID",
			},
		}
	}

	if domain.IsTokenExpired(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "token expired",
				Code:  "TOKEN_EXPIRED",
			},
		}
	}

	// Challenge errors
	if domain.IsChallengeNotFound(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "challenge not found or expired",
				Code:  "CHALLENGE_NOT_FOUND",
			},
		}
	}

	if domain.IsChallengeExpired(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "challenge expired",
				Code:  "CHALLENGE_EXPIRED",
			},
		}
	}

	// Credential errors
	if domain.IsInvalidCredentials(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "invalid credentials",
				Code:  "INVALID_CREDENTIALS",
			},
		}
	}

	// API Key errors
	if domain.IsAPIKeyNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &ErrorResponse{
				Error: "api key not found",
				Code:  "API_KEY_NOT_FOUND",
			},
		}
	}

	if domain.IsAPIKeyExpired(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "api key expired",
				Code:  "API_KEY_EXPIRED",
			},
		}
	}

	if domain.IsAPIKeyRevoked(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "api key revoked",
				Code:  "API_KEY_REVOKED",
			},
		}
	}

	// Connection errors
	if domain.IsConnectionNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &ErrorResponse{
				Error: "connection not found",
				Code:  "CONNECTION_NOT_FOUND",
			},
		}
	}

	if domain.IsConnectionAlreadyExists(err) {
		return &mappedError{
			status: http.StatusConflict,
			response: &ErrorResponse{
				Error: "connection already exists",
				Code:  "CONNECTION_ALREADY_EXISTS",
			},
		}
	}

	// Wallet errors
	if domain.IsWalletNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &ErrorResponse{
				Error: "wallet not found",
				Code:  "WALLET_NOT_FOUND",
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

	// Use the Kind field and its HTTPStatus method
	status := e.Kind.HTTPStatus()
	message := e.Message

	var code string
	var errorText string

	switch e.Kind {
	case errors.KindValidation:
		code = "VALIDATION_ERROR"
		errorText = "validation error"
	case errors.KindNotFound:
		code = "NOT_FOUND"
		errorText = "not found"
	case errors.KindConflict:
		code = "CONFLICT"
		errorText = "conflict"
	case errors.KindUnauthorized:
		code = "UNAUTHORIZED"
		errorText = "unauthorized"
	case errors.KindForbidden:
		code = "FORBIDDEN"
		errorText = "forbidden"
	case errors.KindRateLimit:
		code = "RATE_LIMIT_EXCEEDED"
		errorText = "rate limit exceeded"
	case errors.KindDomain:
		code = "DOMAIN_ERROR"
		errorText = "domain error"
	case errors.KindInfrastructure:
		code = "SERVICE_UNAVAILABLE"
		errorText = "service unavailable"
	case errors.KindTimeout:
		code = "TIMEOUT"
		errorText = "request timeout"
	case errors.KindInternal:
		code = "INTERNAL_ERROR"
		errorText = "internal server error"
		message = "" // Don't expose internal error details
	default:
		code = "INTERNAL_ERROR"
		errorText = "internal server error"
		message = ""
	}

	return &mappedError{
		status: status,
		response: &ErrorResponse{
			Error:   errorText,
			Code:    code,
			Message: message,
		},
	}
}

// ============================================================================
// Response Helpers
// ============================================================================

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, err error) {
	status, errResp := MapError(err)
	httpresponse.JSON(w, status, errResp)
}

// WriteValidationError writes a validation error response.
func WriteValidationError(w http.ResponseWriter, message string, details map[string]string) {
	httpresponse.JSON(w, http.StatusBadRequest, &ErrorResponse{
		Error:   "validation error",
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
	})
}

// WriteBadRequest writes a bad request error.
func WriteBadRequest(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusBadRequest, &ErrorResponse{
		Error:   "bad request",
		Code:    "BAD_REQUEST",
		Message: message,
	})
}

// WriteUnauthorized writes an unauthorized error.
func WriteUnauthorized(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusUnauthorized, &ErrorResponse{
		Error:   "unauthorized",
		Code:    "UNAUTHORIZED",
		Message: message,
	})
}

// WriteForbidden writes a forbidden error.
func WriteForbidden(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusForbidden, &ErrorResponse{
		Error:   "forbidden",
		Code:    "FORBIDDEN",
		Message: message,
	})
}

// WriteNotFound writes a not found error.
func WriteNotFound(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusNotFound, &ErrorResponse{
		Error:   "not found",
		Code:    "NOT_FOUND",
		Message: message,
	})
}

// WriteInternalError writes an internal server error.
func WriteInternalError(w http.ResponseWriter) {
	httpresponse.JSON(w, http.StatusInternalServerError, &ErrorResponse{
		Error: "internal server error",
		Code:  "INTERNAL_ERROR",
	})
}
