package log

import "github.com/0xsj/nexus/platform/pkg/errors"

// Log-specific error codes.
const (
	CodeLogConfigError errors.Code = "LOG_CONFIG_ERROR"
	CodeLogWriteError  errors.Code = "LOG_WRITE_ERROR"
	CodeLogParseError  errors.Code = "LOG_PARSE_ERROR"
)

// ErrLogConfigError creates an error for logger configuration failures.
func ErrLogConfigError(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeLogConfigError)
}

// ErrLogWriteError creates an error for log write failures.
func ErrLogWriteError(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeLogWriteError)
}

// ErrLogParseError creates an error for log parsing failures.
func ErrLogParseError(operation string, value string) *errors.Error {
	return errors.Validation(operation, "invalid log value: "+value).
		WithCode(CodeLogParseError).
		WithMeta("value", value)
}

// IsLogConfigError returns true if the error is a log config error.
func IsLogConfigError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeLogConfigError
	}
	return false
}
