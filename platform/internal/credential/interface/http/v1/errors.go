package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/response"
)

// ============================================================================
// Error Codes
// ============================================================================

const (
	ErrCodeCredentialNotFound     = "CREDENTIAL_NOT_FOUND"
	ErrCodeCredentialExists       = "CREDENTIAL_ALREADY_EXISTS"
	ErrCodeCredentialInvalidState = "CREDENTIAL_INVALID_STATE"
	ErrCodeValidation             = "VALIDATION_ERROR"
	ErrCodeConflict               = "CONCURRENCY_CONFLICT"
	ErrCodeInternal               = "INTERNAL_ERROR"
)

// ============================================================================
// Error Mapping
// ============================================================================

// MapError maps domain/application errors to HTTP status and response.
func MapError(err error) (int, *response.ErrorResponse) {
	if err == nil {
		return http.StatusOK, nil
	}

	// Aggregate not found
	if eventsourcing.IsAggregateNotFound(err) {
		return http.StatusNotFound, &response.ErrorResponse{
			Code:    ErrCodeCredentialNotFound,
			Message: "Credential not found",
			Details: err.Error(),
		}
	}

	// Concurrency conflict
	if eventsourcing.IsConcurrencyConflict(err) {
		return http.StatusConflict, &response.ErrorResponse{
			Code:    ErrCodeConflict,
			Message: "Credential was modified by another request",
			Details: err.Error(),
		}
	}

	// Aggregate validation (domain validation)
	if eventsourcing.IsAggregateValidation(err) {
		return http.StatusBadRequest, &response.ErrorResponse{
			Code:    ErrCodeCredentialInvalidState,
			Message: "Invalid credential operation",
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

	// Default to internal error
	return http.StatusInternalServerError, &response.ErrorResponse{
		Code:    ErrCodeInternal,
		Message: "An internal error occurred",
		Details: err.Error(),
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
		Code:    ErrCodeCredentialNotFound,
		Message: "Credential not found",
		Details: "credential with ID '" + id + "' not found",
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
		Code:    ErrCodeCredentialInvalidState,
		Message: message,
	})
}
