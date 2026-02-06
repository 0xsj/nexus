package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/presentation/domain"
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

	// Check for presentation domain errors
	if pkgerrors.Is(err, domain.ErrPresentationNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodePresentationNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrPresentationRevoked) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodePresentationRevoked),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrPresentationInvalid) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodePresentationInvalid),
			Message: err.Error(),
		}
	}

	// Check for share link domain errors
	if pkgerrors.Is(err, domain.ErrShareLinkNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeShareLinkNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrShareLinkRevoked) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeShareLinkRevoked),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrShareLinkExpired) {
		return http.StatusGone, ErrorResponse{
			Code:    string(domain.CodeShareLinkExpired),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrShareLinkMaxViewsReached) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodeShareLinkMaxViewsReached),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrShareLinkInvalid) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeShareLinkInvalid),
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
