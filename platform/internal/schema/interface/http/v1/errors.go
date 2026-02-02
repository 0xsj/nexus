package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/schema/domain"
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

	// Check for domain errors
	if pkgerrors.Is(err, domain.ErrSchemaNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeSchemaNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSchemaAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeSchemaAlreadyExists),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSchemaDeprecated) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeSchemaDeprecated),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSchemaInvalid) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeSchemaInvalid),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrVersionNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeVersionNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrVersionAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeVersionAlreadyExists),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidVersionFormat) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeInvalidVersionFormat),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrClaimNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeClaimNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrClaimAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeClaimAlreadyExists),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidClaimDefinition) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeInvalidClaimDefinition),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidDataType) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeInvalidDataType),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrIssuerNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeIssuerNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrIssuerInactive) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeIssuerInactive),
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

// NotFoundResponse creates a not found error response.
func NotFoundResponse(resource string) ErrorResponse {
	return ErrorResponse{
		Code:    "NOT_FOUND",
		Message: resource + " not found",
	}
}

// InternalErrorResponse creates an internal error response.
func InternalErrorResponse() ErrorResponse {
	return ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "an unexpected error occurred",
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

// ConflictErrorResponse creates a conflict error response.
func ConflictErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "CONFLICT",
		Message: message,
	}
}
