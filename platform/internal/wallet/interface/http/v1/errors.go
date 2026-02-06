package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Response
// ============================================================================

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ============================================================================
// Error Mapping
// ============================================================================

// MapError maps a domain/application error to an HTTP status code and response.
func MapError(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		}
	}

	// Check for wallet domain errors
	if pkgerrors.Is(err, domain.ErrWalletNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeWalletNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrWalletAlreadyLinked) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeWalletAlreadyLinked),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrWalletNotActive) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeWalletNotActive),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrWalletUnverified) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeWalletUnverified),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidAddress) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeInvalidAddress),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrChainNotSupported) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeChainNotSupported),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSignatureInvalid) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeSignatureInvalid),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrChallengeExpired) {
		return http.StatusGone, ErrorResponse{
			Code:    string(domain.CodeChallengeExpired),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrChallengeNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeChallengeNotFound),
			Message: err.Error(),
		}
	}

	// Check for pkg/errors types
	var appErr *pkgerrors.Error
	if pkgerrors.As(err, &appErr) {
		return appErr.Kind.HTTPStatus(), ErrorResponse{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Metadata,
		}
	}

	// Default to internal server error
	return http.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "an unexpected error occurred",
	}
}

// ============================================================================
// Common Error Responses
// ============================================================================

// BadRequestResponse creates a bad request error response.
func BadRequestResponse(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

// ValidationErrorResponse creates a validation error response.
func ValidationErrorResponse(field, message string) ErrorResponse {
	return ErrorResponse{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: map[string]any{
			"field": field,
		},
	}
}
