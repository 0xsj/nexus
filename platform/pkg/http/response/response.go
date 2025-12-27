package response

import (
	"encoding/json"
	"net/http"
)

// ============================================================================
// Response Types
// ============================================================================

// Response is a generic API response wrapper.
type Response struct {
	Success bool           `json:"success"`
	Data    any            `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Meta    *Meta          `json:"meta,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details string       `json:"details,omitempty"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// FieldError represents a validation error on a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Meta contains pagination and other metadata.
type Meta struct {
	Total      int    `json:"total,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// ============================================================================
// Response Helpers
// ============================================================================

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but can't do much else at this point
			http.Error(w, `{"success":false,"error":{"code":"ENCODE_ERROR","message":"Failed to encode response"}}`, http.StatusInternalServerError)
		}
	}
}

// OK writes a successful response with data.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// OKWithMeta writes a successful response with data and metadata.
func OKWithMeta(w http.ResponseWriter, data any, meta *Meta) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Error Helpers
// ============================================================================

// Error writes an error response.
func Error(w http.ResponseWriter, status int, err *ErrorResponse) {
	JSON(w, status, Response{
		Success: false,
		Error:   err,
	})
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, err *ErrorResponse) {
	Error(w, http.StatusBadRequest, err)
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, err *ErrorResponse) {
	Error(w, http.StatusUnauthorized, err)
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, err *ErrorResponse) {
	Error(w, http.StatusForbidden, err)
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, err *ErrorResponse) {
	Error(w, http.StatusNotFound, err)
}

// Conflict writes a 409 Conflict response.
func Conflict(w http.ResponseWriter, err *ErrorResponse) {
	Error(w, http.StatusConflict, err)
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, err *ErrorResponse) {
	Error(w, http.StatusInternalServerError, err)
}

// ============================================================================
// Common Errors
// ============================================================================

// ErrBadRequest creates a bad request error.
func ErrBadRequest(message, details string) *ErrorResponse {
	return &ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: message,
		Details: details,
	}
}

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// ErrNotFound creates a not found error.
func ErrNotFound(resource, id string) *ErrorResponse {
	return &ErrorResponse{
		Code:    "NOT_FOUND",
		Message: resource + " not found",
		Details: resource + " with ID '" + id + "' not found",
	}
}

// ErrConflict creates a conflict error.
func ErrConflict(message, details string) *ErrorResponse {
	return &ErrorResponse{
		Code:    "CONFLICT",
		Message: message,
		Details: details,
	}
}

// ErrInternal creates an internal error.
func ErrInternal(details string) *ErrorResponse {
	return &ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "An internal error occurred",
		Details: details,
	}
}

// ErrValidation creates a validation error with field errors.
func ErrValidation(fields []FieldError) *ErrorResponse {
	return &ErrorResponse{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Fields:  fields,
	}
}
