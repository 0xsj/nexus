package config

import "github.com/0xsj/nexus/platform/pkg/errors"

// Config-specific error codes.
const (
	CodeConfigNotFound   errors.Code = "CONFIG_NOT_FOUND"
	CodeConfigInvalid    errors.Code = "CONFIG_INVALID"
	CodeConfigParseError errors.Code = "CONFIG_PARSE_ERROR"
	CodeConfigRequired   errors.Code = "CONFIG_REQUIRED"
	CodeConfigTypeError  errors.Code = "CONFIG_TYPE_ERROR"
	CodeConfigFileError  errors.Code = "CONFIG_FILE_ERROR"
)

// ErrConfigNotFound creates an error when a config source is not found.
func ErrConfigNotFound(operation string, source string) *errors.Error {
	return errors.NotFound(operation, "config source: "+source).
		WithCode(CodeConfigNotFound).
		WithMeta("source", source)
}

// ErrConfigInvalid creates an error for invalid configuration.
func ErrConfigInvalid(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeConfigInvalid)
}

// ErrConfigParseError creates an error for parsing failures.
func ErrConfigParseError(operation string, field string, err error) *errors.Error {
	msg := "failed to parse field: " + field
	if err != nil {
		msg += ": " + err.Error()
	}
	return errors.Validation(operation, msg).
		WithCode(CodeConfigParseError).
		WithMeta("field", field)
}

// ErrConfigRequired creates an error for missing required fields.
func ErrConfigRequired(operation string, field string) *errors.Error {
	return errors.Validation(operation, "required field missing: "+field).
		WithCode(CodeConfigRequired).
		WithMeta("field", field)
}

// ErrConfigTypeError creates an error for type conversion failures.
func ErrConfigTypeError(operation string, field string, expected string, got string) *errors.Error {
	return errors.Validation(operation, "type error for field: "+field).
		WithCode(CodeConfigTypeError).
		WithMeta("field", field).
		WithMeta("expected", expected).
		WithMeta("got", got)
}

// ErrConfigFileError creates an error for file read failures.
func ErrConfigFileError(operation string, path string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeConfigFileError).
		WithMeta("path", path)
}

// IsConfigNotFound returns true if the error is a config not found error.
func IsConfigNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeConfigNotFound
	}
	return false
}

// IsConfigRequired returns true if the error is a required field error.
func IsConfigRequired(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeConfigRequired
	}
	return false
}

// IsConfigParseError returns true if the error is a parse error.
func IsConfigParseError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeConfigParseError
	}
	return false
}
