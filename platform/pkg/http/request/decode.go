package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// DefaultMaxBodySize is the default maximum request body size (1MB).
	DefaultMaxBodySize = 1 << 20 // 1MB

	// MaxBodySize is the absolute maximum request body size (10MB).
	MaxBodySize = 10 << 20 // 10MB
)

// ============================================================================
// JSON Decoding
// ============================================================================

// DecodeJSON decodes the request body as JSON into the given destination.
func DecodeJSON(r *http.Request, dst any) error {
	return DecodeJSONWithLimit(r, dst, DefaultMaxBodySize)
}

// DecodeJSONWithLimit decodes the request body as JSON with a size limit.
func DecodeJSONWithLimit(r *http.Request, dst any, maxSize int64) error {
	if r.Body == nil {
		return pkghttp.NewDecodeError("", "request body is empty", nil)
	}

	// Limit the request body size
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict parsing

	if err := decoder.Decode(dst); err != nil {
		return handleJSONError(err)
	}

	// Check for extra data after JSON object
	if decoder.More() {
		return pkghttp.NewDecodeError("", "request body contains multiple JSON values", nil)
	}

	return nil
}

// DecodeJSONLenient decodes JSON without strict field checking.
func DecodeJSONLenient(r *http.Request, dst any) error {
	if r.Body == nil {
		return pkghttp.NewDecodeError("", "request body is empty", nil)
	}

	r.Body = http.MaxBytesReader(nil, r.Body, DefaultMaxBodySize)

	decoder := json.NewDecoder(r.Body)
	// Don't call DisallowUnknownFields - allow extra fields

	if err := decoder.Decode(dst); err != nil {
		return handleJSONError(err)
	}

	return nil
}

// handleJSONError converts JSON errors to DecodeError.
func handleJSONError(err error) error {
	var syntaxErr *json.SyntaxError
	var unmarshalTypeErr *json.UnmarshalTypeError
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &syntaxErr):
		return pkghttp.NewDecodeError(
			"",
			fmt.Sprintf("malformed JSON at position %d", syntaxErr.Offset),
			err,
		)

	case errors.As(err, &unmarshalTypeErr):
		return pkghttp.NewDecodeError(
			unmarshalTypeErr.Field,
			fmt.Sprintf("invalid type for field '%s': expected %s", unmarshalTypeErr.Field, unmarshalTypeErr.Type),
			err,
		)

	case errors.As(err, &maxBytesErr):
		return pkghttp.NewDecodeError(
			"",
			"request body too large",
			pkghttp.ErrRequestBodyTooLarge,
		)

	case errors.Is(err, io.EOF):
		return pkghttp.NewDecodeError("", "request body is empty", err)

	case errors.Is(err, io.ErrUnexpectedEOF):
		return pkghttp.NewDecodeError("", "request body contains incomplete JSON", err)

	case strings.HasPrefix(err.Error(), "json: unknown field"):
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		field = strings.Trim(field, "\"")
		return pkghttp.NewDecodeError(field, "unknown field", err)

	default:
		return pkghttp.NewDecodeError("", err.Error(), err)
	}
}

// ============================================================================
// Path Parameters (chi)
// ============================================================================

// PathParam returns a path parameter value.
func PathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// PathParamRequired returns a path parameter value or an error if missing.
func PathParamRequired(r *http.Request, name string) (string, error) {
	value := chi.URLParam(r, name)
	if value == "" {
		return "", pkghttp.NewDecodeError(name, "path parameter is required", pkghttp.ErrMissingPathParam)
	}
	return value, nil
}

// PathParamInt returns a path parameter as an integer.
func PathParamInt(r *http.Request, name string) (int, error) {
	value := chi.URLParam(r, name)
	if value == "" {
		return 0, pkghttp.NewDecodeError(name, "path parameter is required", pkghttp.ErrMissingPathParam)
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, pkghttp.NewDecodeError(name, "must be an integer", pkghttp.ErrInvalidQueryParam)
	}

	return i, nil
}

// PathParamInt64 returns a path parameter as an int64.
func PathParamInt64(r *http.Request, name string) (int64, error) {
	value := chi.URLParam(r, name)
	if value == "" {
		return 0, pkghttp.NewDecodeError(name, "path parameter is required", pkghttp.ErrMissingPathParam)
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, pkghttp.NewDecodeError(name, "must be an integer", pkghttp.ErrInvalidQueryParam)
	}

	return i, nil
}

// ============================================================================
// Query Parameters
// ============================================================================

// QueryParam returns a query parameter value.
func QueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// QueryParamDefault returns a query parameter value or a default.
func QueryParamDefault(r *http.Request, name, defaultValue string) string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// QueryParamRequired returns a query parameter value or an error if missing.
func QueryParamRequired(r *http.Request, name string) (string, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return "", pkghttp.NewDecodeError(name, "query parameter is required", pkghttp.ErrMissingQueryParam)
	}
	return value, nil
}

// QueryParamInt returns a query parameter as an integer.
func QueryParamInt(r *http.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, pkghttp.NewDecodeError(name, "must be an integer", pkghttp.ErrInvalidQueryParam)
	}

	return i, nil
}

// QueryParamIntRequired returns a required query parameter as an integer.
func QueryParamIntRequired(r *http.Request, name string) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, pkghttp.NewDecodeError(name, "query parameter is required", pkghttp.ErrMissingQueryParam)
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, pkghttp.NewDecodeError(name, "must be an integer", pkghttp.ErrInvalidQueryParam)
	}

	return i, nil
}

// QueryParamInt64 returns a query parameter as an int64.
func QueryParamInt64(r *http.Request, name string, defaultValue int64) (int64, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, pkghttp.NewDecodeError(name, "must be an integer", pkghttp.ErrInvalidQueryParam)
	}

	return i, nil
}

// QueryParamBool returns a query parameter as a boolean.
func QueryParamBool(r *http.Request, name string, defaultValue bool) (bool, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, pkghttp.NewDecodeError(name, "must be a boolean (true/false)", pkghttp.ErrInvalidQueryParam)
	}

	return b, nil
}

// QueryParamTime returns a query parameter as a time.Time (RFC3339 format).
func QueryParamTime(r *http.Request, name string) (*time.Time, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, pkghttp.NewDecodeError(name, "must be a valid RFC3339 timestamp", pkghttp.ErrInvalidQueryParam)
	}

	return &t, nil
}

// QueryParamSlice returns a query parameter as a string slice.
// Supports both repeated params (?a=1&a=2) and comma-separated (?a=1,2).
func QueryParamSlice(r *http.Request, name string) []string {
	values := r.URL.Query()[name]
	if len(values) == 0 {
		return nil
	}

	// If single value with commas, split it
	if len(values) == 1 && strings.Contains(values[0], ",") {
		parts := strings.Split(values[0], ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}

	return values
}

// ============================================================================
// Pagination Helpers
// ============================================================================

// PaginationParams holds common pagination parameters.
type PaginationParams struct {
	Limit  int
	Offset int
	Cursor string
}

// DefaultPaginationParams returns default pagination parameters.
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Limit:  20,
		Offset: 0,
	}
}

// ParsePagination extracts pagination parameters from the request.
func ParsePagination(r *http.Request, maxLimit int) (PaginationParams, error) {
	params := DefaultPaginationParams()

	if maxLimit <= 0 {
		maxLimit = 100
	}

	// Parse limit
	limit, err := QueryParamInt(r, "limit", params.Limit)
	if err != nil {
		return params, err
	}
	if limit < 1 {
		limit = 1
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	params.Limit = limit

	// Parse offset
	offset, err := QueryParamInt(r, "offset", 0)
	if err != nil {
		return params, err
	}
	if offset < 0 {
		offset = 0
	}
	params.Offset = offset

	// Parse cursor (for cursor-based pagination)
	params.Cursor = QueryParam(r, "cursor")

	return params, nil
}

// ============================================================================
// Sorting Helpers
// ============================================================================

// SortParams holds sorting parameters.
type SortParams struct {
	SortBy    string
	SortOrder string // "asc" or "desc"
}

// ParseSort extracts sorting parameters from the request.
func ParseSort(r *http.Request, allowedFields []string, defaultField, defaultOrder string) SortParams {
	params := SortParams{
		SortBy:    defaultField,
		SortOrder: defaultOrder,
	}

	sortBy := QueryParam(r, "sort_by")
	if sortBy != "" && isAllowedField(sortBy, allowedFields) {
		params.SortBy = sortBy
	}

	sortOrder := strings.ToLower(QueryParam(r, "sort_order"))
	if sortOrder == "asc" || sortOrder == "desc" {
		params.SortOrder = sortOrder
	}

	return params
}

// isAllowedField checks if a field is in the allowed list.
func isAllowedField(field string, allowed []string) bool {
	for _, f := range allowed {
		if f == field {
			return true
		}
	}
	return false
}
