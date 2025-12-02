// pkg/graphql/errors.go

package graphql

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus-go/pkg/errors"
)

// ============================================================================
// GraphQL Error Extension Codes
// ============================================================================

// ExtensionCode represents a GraphQL error extension code.
// These follow common GraphQL conventions (Apollo, etc.).
type ExtensionCode string

const (
	// Client errors
	CodeBadUserInput     ExtensionCode = "BAD_USER_INPUT"
	CodeNotFound         ExtensionCode = "NOT_FOUND"
	CodeConflict         ExtensionCode = "CONFLICT"
	CodeUnauthenticated  ExtensionCode = "UNAUTHENTICATED"
	CodeForbidden        ExtensionCode = "FORBIDDEN"
	CodeValidationFailed ExtensionCode = "VALIDATION_FAILED"

	// Business logic errors
	CodeBusinessRuleViolation ExtensionCode = "BUSINESS_RULE_VIOLATION"

	// Server errors
	CodeInternalError      ExtensionCode = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable ExtensionCode = "SERVICE_UNAVAILABLE"
	CodeTimeout            ExtensionCode = "TIMEOUT"
	CodeRateLimited        ExtensionCode = "RATE_LIMITED"

	// Unknown
	CodeUnknown ExtensionCode = "UNKNOWN_ERROR"
)

// ============================================================================
// Kind to Extension Code Mapping
// ============================================================================

// KindToExtensionCode maps a domain error Kind to a GraphQL extension code.
//
// Mapping follows GraphQL/Apollo conventions:
//   - KindNotFound       → NOT_FOUND
//   - KindValidation     → BAD_USER_INPUT
//   - KindConflict       → CONFLICT
//   - KindUnauthorized   → UNAUTHENTICATED
//   - KindForbidden      → FORBIDDEN
//   - KindDomain         → BUSINESS_RULE_VIOLATION
//   - KindInfrastructure → SERVICE_UNAVAILABLE
//   - KindTimeout        → TIMEOUT
//   - KindRateLimit      → RATE_LIMITED
//   - KindInternal       → INTERNAL_SERVER_ERROR
//   - KindOther          → UNKNOWN_ERROR
func KindToExtensionCode(kind errors.Kind) ExtensionCode {
	switch kind {
	case errors.KindNotFound:
		return CodeNotFound
	case errors.KindValidation:
		return CodeBadUserInput
	case errors.KindConflict:
		return CodeConflict
	case errors.KindUnauthorized:
		return CodeUnauthenticated
	case errors.KindForbidden:
		return CodeForbidden
	case errors.KindDomain:
		return CodeBusinessRuleViolation
	case errors.KindInfrastructure:
		return CodeServiceUnavailable
	case errors.KindTimeout:
		return CodeTimeout
	case errors.KindRateLimit:
		return CodeRateLimited
	case errors.KindInternal:
		return CodeInternalError
	default:
		return CodeUnknown
	}
}

// ExtensionCodeToKind maps a GraphQL extension code to a domain error Kind.
// Used when receiving errors from GraphQL services.
func ExtensionCodeToKind(code ExtensionCode) errors.Kind {
	switch code {
	case CodeNotFound:
		return errors.KindNotFound
	case CodeBadUserInput, CodeValidationFailed:
		return errors.KindValidation
	case CodeConflict:
		return errors.KindConflict
	case CodeUnauthenticated:
		return errors.KindUnauthorized
	case CodeForbidden:
		return errors.KindForbidden
	case CodeBusinessRuleViolation:
		return errors.KindDomain
	case CodeServiceUnavailable:
		return errors.KindInfrastructure
	case CodeTimeout:
		return errors.KindTimeout
	case CodeRateLimited:
		return errors.KindRateLimit
	case CodeInternalError:
		return errors.KindInternal
	default:
		return errors.KindOther
	}
}

// ============================================================================
// GraphQL Error Type
// ============================================================================

// Error represents a GraphQL error with extensions.
// Implements the error interface and provides extensions for clients.
type Error struct {
	// Message is the human-readable error message.
	Message string `json:"message"`

	// Extensions contains machine-readable error metadata.
	Extensions Extensions `json:"extensions,omitempty"`

	// Path is the GraphQL query path where the error occurred.
	Path []interface{} `json:"path,omitempty"`
}

// Extensions contains GraphQL error extension fields.
type Extensions struct {
	// Code is the primary error classification (e.g., NOT_FOUND).
	Code ExtensionCode `json:"code"`

	// ErrorCode is the domain-specific error code (e.g., USER_NOT_FOUND).
	ErrorCode string `json:"errorCode,omitempty"`

	// Operation identifies where the error occurred.
	Operation string `json:"operation,omitempty"`

	// Retryable indicates if the operation can be safely retried.
	Retryable bool `json:"retryable,omitempty"`

	// Field is the specific field that caused the error (for validation).
	Field string `json:"field,omitempty"`

	// Fields contains multiple field-level errors (for validation).
	Fields map[string]string `json:"fields,omitempty"`

	// Metadata contains additional context.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// ============================================================================
// Domain Error to GraphQL Error Conversion
// ============================================================================

// ToGraphQLError converts a domain error to a GraphQL error.
//
// Example:
//
//	err := errors.NotFound("UserRepository.FindByID", "user not found")
//	gqlErr := graphql.ToGraphQLError(err)
func ToGraphQLError(err error) *Error {
	if err == nil {
		return nil
	}

	// Check if already a GraphQL error
	if gqlErr, ok := err.(*Error); ok {
		return gqlErr
	}

	// Extract domain error
	domainErr := errors.AsError(err)
	if domainErr == nil {
		// Wrap non-domain error as internal
		return &Error{
			Message: err.Error(),
			Extensions: Extensions{
				Code: CodeInternalError,
			},
		}
	}

	// Build GraphQL error
	gqlErr := &Error{
		Message: domainErr.Message,
		Extensions: Extensions{
			Code:      KindToExtensionCode(domainErr.Kind),
			ErrorCode: string(domainErr.Code),
			Operation: domainErr.Operation,
			Retryable: domainErr.Retryable,
		},
	}

	// Extract field information for validation errors
	if domainErr.Kind == errors.KindValidation {
		if field, ok := domainErr.Metadata["field"].(string); ok {
			gqlErr.Extensions.Field = field
		}
		if fields, ok := domainErr.Metadata["fields"].(map[string]string); ok {
			gqlErr.Extensions.Fields = fields
		}
	}

	// Copy relevant metadata (excluding already extracted fields)
	if len(domainErr.Metadata) > 0 {
		metadata := make(map[string]interface{})
		for k, v := range domainErr.Metadata {
			if k != "field" && k != "fields" {
				metadata[k] = v
			}
		}
		if len(metadata) > 0 {
			gqlErr.Extensions.Metadata = metadata
		}
	}

	return gqlErr
}

// ============================================================================
// GraphQL Error to Domain Error Conversion
// ============================================================================

// FromGraphQLError converts a GraphQL error to a domain error.
// Used when receiving errors from GraphQL services.
func FromGraphQLError(gqlErr *Error) *errors.Error {
	if gqlErr == nil {
		return nil
	}

	kind := ExtensionCodeToKind(gqlErr.Extensions.Code)

	domainErr := &errors.Error{
		Kind:      kind,
		Code:      errors.Code(gqlErr.Extensions.ErrorCode),
		Operation: gqlErr.Extensions.Operation,
		Message:   gqlErr.Message,
		Retryable: gqlErr.Extensions.Retryable,
		Metadata:  make(map[string]any),
	}

	// Restore field information
	if gqlErr.Extensions.Field != "" {
		domainErr.Metadata["field"] = gqlErr.Extensions.Field
	}
	if len(gqlErr.Extensions.Fields) > 0 {
		domainErr.Metadata["fields"] = gqlErr.Extensions.Fields
	}

	// Copy metadata
	for k, v := range gqlErr.Extensions.Metadata {
		domainErr.Metadata[k] = v
	}

	return domainErr
}

// ============================================================================
// gqlgen Integration
// ============================================================================

// ErrorPresenterFunc is the signature for gqlgen's error presenter.
// Use this with gqlgen's server configuration.
type ErrorPresenterFunc func(ctx context.Context, err error) *Error

// NewErrorPresenter creates an error presenter for gqlgen.
//
// Usage with gqlgen:
//
//	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
//	srv.SetErrorPresenter(graphql.NewErrorPresenter(logger))
func NewErrorPresenter(logFn func(ctx context.Context, err error)) ErrorPresenterFunc {
	return func(ctx context.Context, err error) *Error {
		// Log the error if logger provided
		if logFn != nil {
			logFn(ctx, err)
		}

		return ToGraphQLError(err)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// Errorf creates a new GraphQL error with formatted message.
func Errorf(code ExtensionCode, format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Extensions: Extensions{
			Code: code,
		},
	}
}

// NotFound creates a NOT_FOUND GraphQL error.
func NotFound(message string) *Error {
	return &Error{
		Message: message,
		Extensions: Extensions{
			Code: CodeNotFound,
		},
	}
}

// BadUserInput creates a BAD_USER_INPUT GraphQL error.
func BadUserInput(message string) *Error {
	return &Error{
		Message: message,
		Extensions: Extensions{
			Code: CodeBadUserInput,
		},
	}
}

// ValidationError creates a BAD_USER_INPUT error with field details.
func ValidationError(field, message string) *Error {
	return &Error{
		Message: message,
		Extensions: Extensions{
			Code:  CodeBadUserInput,
			Field: field,
		},
	}
}

// ValidationErrors creates a BAD_USER_INPUT error with multiple field errors.
func ValidationErrors(fields map[string]string) *Error {
	return &Error{
		Message: "Validation failed",
		Extensions: Extensions{
			Code:   CodeBadUserInput,
			Fields: fields,
		},
	}
}

// Unauthenticated creates an UNAUTHENTICATED GraphQL error.
func Unauthenticated(message string) *Error {
	if message == "" {
		message = "Authentication required"
	}
	return &Error{
		Message: message,
		Extensions: Extensions{
			Code: CodeUnauthenticated,
		},
	}
}

// Forbidden creates a FORBIDDEN GraphQL error.
func Forbidden(message string) *Error {
	if message == "" {
		message = "Access denied"
	}
	return &Error{
		Message: message,
		Extensions: Extensions{
			Code: CodeForbidden,
		},
	}
}

// InternalError creates an INTERNAL_SERVER_ERROR GraphQL error.
// Message is sanitized to avoid leaking internal details.
func InternalError() *Error {
	return &Error{
		Message: "An internal error occurred",
		Extensions: Extensions{
			Code: CodeInternalError,
		},
	}
}

// ============================================================================
// Error Checking Helpers
// ============================================================================

// IsNotFound checks if the error is a NOT_FOUND error.
func IsNotFound(err error) bool {
	return hasExtensionCode(err, CodeNotFound)
}

// IsUnauthenticated checks if the error is an UNAUTHENTICATED error.
func IsUnauthenticated(err error) bool {
	return hasExtensionCode(err, CodeUnauthenticated)
}

// IsForbidden checks if the error is a FORBIDDEN error.
func IsForbidden(err error) bool {
	return hasExtensionCode(err, CodeForbidden)
}

// IsValidationError checks if the error is a BAD_USER_INPUT error.
func IsValidationError(err error) bool {
	return hasExtensionCode(err, CodeBadUserInput)
}

// IsRetryable checks if the error indicates a retryable condition.
func IsRetryable(err error) bool {
	if gqlErr, ok := err.(*Error); ok {
		if gqlErr.Extensions.Retryable {
			return true
		}
		// Fall back to code-based check
		switch gqlErr.Extensions.Code {
		case CodeServiceUnavailable, CodeTimeout, CodeRateLimited:
			return true
		}
	}
	return false
}

// hasExtensionCode checks if an error has a specific extension code.
func hasExtensionCode(err error, code ExtensionCode) bool {
	if gqlErr, ok := err.(*Error); ok {
		return gqlErr.Extensions.Code == code
	}
	return false
}

// ExtensionCode extracts the extension code from an error.
func GetExtensionCode(err error) ExtensionCode {
	if gqlErr, ok := err.(*Error); ok {
		return gqlErr.Extensions.Code
	}
	return CodeUnknown
}
