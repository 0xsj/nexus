package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/organization/domain"
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

	// Check for organization domain errors
	if pkgerrors.Is(err, domain.ErrOrganizationNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeOrganizationNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOrganizationAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeOrganizationAlreadyExists),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOrganizationInvalid) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeOrganizationInvalid),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOrganizationNotVerified) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeOrganizationNotVerified),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrMemberNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeMemberNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrMemberAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeMemberAlreadyExists),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSlugAlreadyTaken) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeSlugAlreadyTaken),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOwnerCannotLeave) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeOwnerCannotLeave),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInsufficientPermissions) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodeInsufficientPermissions),
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
