package v1

// import (
// 	"errors"
// 	"net/http"

// 	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
// 	"github.com/0xsj/nexus/platform/pkg/http/response"
// )

// // ============================================================================
// // Error Codes
// // ============================================================================

// const (
// 	ErrCodeCredentialNotFound     = "CREDENTIAL_NOT_FOUND"
// 	ErrCodeCredentialExists       = "CREDENTIAL_ALREADY_EXISTS"
// 	ErrCodeCredentialInvalidState = "CREDENTIAL_INVALID_STATE"
// 	ErrCodeCredentialValidation   = "CREDENTIAL_VALIDATION_ERROR"
// 	ErrCodeCredentialConflict     = "CREDENTIAL_CONCURRENCY_CONFLICT"
// 	ErrCodeInternalError          = "INTERNAL_ERROR"
// )

// // ============================================================================
// // Error Mapping
// // ============================================================================

// // MapError maps domain errors to HTTP responses.
// func MapError(err error) (int, *response.ErrorResponse) {
// 	if err == nil {
// 		return http.StatusOK, nil
// 	}

// 	// Check for aggregate not found
// 	if eventsourcing.IsAggregateNotFound(err) {
// 		return http.StatusNotFound, &response.ErrorResponse{
// 			Code:    ErrCodeCredentialNotFound,
// 			Message: "Credential not found",
// 			Details: err.Error(),
// 		}
// 	}

// 	// Check for concurrency conflict
// 	if eventsourcing.IsConcurrencyConflict(err) {
// 		return http.StatusConflict, &response.ErrorResponse{
// 			Code:    ErrCodeCredentialConflict,
// 			Message: "Credential was modified by another request",
// 			Details: err.Error(),
// 		}
// 	}

// 	// Check for validation errors
// 	if eventsourcing.IsAggregateValidation(err) {
// 		return http.StatusBadRequest, &response.ErrorResponse{
// 			Code:    ErrCodeCredentialValidation,
// 			Message: "Validation error",
// 			Details: err.Error(),
// 		}
// 	}

// 	// Check for already exists (validation with "already exists" message)
// 	if containsAlreadyExists(err) {
// 		return http.StatusConflict, &response.ErrorResponse{
// 			Code:    ErrCodeCredentialExists,
// 			Message: "Credential already exists",
// 			Details: err.Error(),
// 		}
// 	}

// 	// Default to internal error
// 	return http.StatusInternalServerError, &response.ErrorResponse{
// 		Code:    ErrCodeInternalError,
// 		Message: "An internal error occurred",
// 		Details: err.Error(),
// 	}
// }

// // containsAlreadyExists checks if the error message contains "already exists".
// func containsAlreadyExists(err error) bool {
// 	if err == nil {
// 		return false
// 	}
// 	return contains(err.Error(), "already exists")
// }

// // contains checks if s contains substr (case-insensitive would be better but keeping simple).
// func contains(s, substr string) bool {
// 	return len(s) >= len(substr) && searchString(s, substr)
// }

// func searchString(s, substr string) bool {
// 	for i := 0; i <= len(s)-len(substr); i++ {
// 		if s[i:i+len(substr)] == substr {
// 			return true
// 		}
// 	}
// 	return false
// }

// // ============================================================================
// // Error Response Helpers
// // ============================================================================

// // ValidationError creates a validation error response.
// func ValidationError(field, message string) *response.ErrorResponse {
// 	return &response.ErrorResponse{
// 		Code:    ErrCodeCredentialValidation,
// 		Message: "Validation error",
// 		Details: field + ": " + message,
// 	}
// }

// // NotFoundError creates a not found error response.
// func NotFoundError(id string) *response.ErrorResponse {
// 	return &response.ErrorResponse{
// 		Code:    ErrCodeCredentialNotFound,
// 		Message: "Credential not found",
// 		Details: "credential with ID '" + id + "' not found",
// 	}
// }

// // InvalidStateError creates an invalid state error response.
// func InvalidStateError(currentState, requiredState string) *response.ErrorResponse {
// 	return &response.ErrorResponse{
// 		Code:    ErrCodeCredentialInvalidState,
// 		Message: "Invalid credential state",
// 		Details: "credential is in state '" + currentState + "', required: '" + requiredState + "'",
// 	}
// }

// // ============================================================================
// // Ensure we don't have unused import
// // ============================================================================

// var _ = errors.New // Keep errors import for future use
