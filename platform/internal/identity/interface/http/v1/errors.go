package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
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
// Validation Error
// ============================================================================

// ValidationError represents a request validation error.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
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

	// Check for validation errors
	if ve, ok := err.(*ValidationError); ok {
		return http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: ve.Message,
		}
	}

	// Check for domain errors - User
	if pkgerrors.Is(err, domain.ErrUserNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeUserNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrUserAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeUserAlreadyExists),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrUserSuspended) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodeUserSuspended),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrUserNotActive) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodeUserNotActive),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidUserID) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeInvalidUserID),
			Message: err.Error(),
		}
	}

	// Check for domain errors - Email
	if pkgerrors.Is(err, domain.ErrEmailAlreadyRegistered) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeEmailAlreadyRegistered),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrEmailNotVerified) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodeEmailNotVerified),
			Message: err.Error(),
		}
	}

	// Check for domain errors - Session
	if pkgerrors.Is(err, domain.ErrSessionNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeSessionNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSessionExpired) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeSessionExpired),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrSessionRevoked) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeSessionRevoked),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidToken) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeInvalidToken),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrInvalidSessionID) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeInvalidSessionID),
			Message: err.Error(),
		}
	}

	// Check for domain errors - Magic Link
	if pkgerrors.Is(err, domain.ErrMagicLinkExpired) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeMagicLinkExpired),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrMagicLinkUsed) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeMagicLinkUsed),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrMagicLinkInvalid) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeMagicLinkInvalid),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrMagicLinkNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeMagicLinkNotFound),
			Message: err.Error(),
		}
	}

	// Check for domain errors - OAuth
	if pkgerrors.Is(err, domain.ErrOAuthProviderNotSupported) {
		return http.StatusBadRequest, ErrorResponse{
			Code:    string(domain.CodeOAuthProviderNotSupported),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOAuthStateMismatch) {
		return http.StatusUnauthorized, ErrorResponse{
			Code:    string(domain.CodeOAuthStateMismatch),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOAuthCodeExchangeFailed) {
		return http.StatusBadGateway, ErrorResponse{
			Code:    string(domain.CodeOAuthCodeExchangeFailed),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrOAuthAccountLinked) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeOAuthAccountLinked),
			Message: err.Error(),
		}
	}

	// Check for domain errors - Auth Method
	if pkgerrors.Is(err, domain.ErrAuthMethodNotEnabled) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodeAuthMethodNotEnabled),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrAuthMethodExists) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeAuthMethodExists),
			Message: err.Error(),
		}
	}

	// Check for domain errors - DID
	if pkgerrors.Is(err, domain.ErrDIDAlreadyLinked) {
		return http.StatusConflict, ErrorResponse{
			Code:    string(domain.CodeDIDAlreadyLinked),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrDIDNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Code:    string(domain.CodeDIDNotFound),
			Message: err.Error(),
		}
	}

	if pkgerrors.Is(err, domain.ErrPrimaryDIDChange) {
		return http.StatusForbidden, ErrorResponse{
			Code:    string(domain.CodePrimaryDIDChange),
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

// UnauthorizedResponse creates an unauthorized error response.
func UnauthorizedResponse(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// ForbiddenResponse creates a forbidden error response.
func ForbiddenResponse(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// ConflictResponse creates a conflict error response.
func ConflictResponse(message string) ErrorResponse {
	return ErrorResponse{
		Code:    "CONFLICT",
		Message: message,
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
