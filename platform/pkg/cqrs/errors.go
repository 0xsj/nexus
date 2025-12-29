package cqrs

import (
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/errors"
)

// CQRS-specific error codes.
const (
	CodeCommandValidation   errors.Code = "COMMAND_VALIDATION"
	CodeCommandNotFound     errors.Code = "COMMAND_NOT_FOUND"
	CodeCommandFailed       errors.Code = "COMMAND_FAILED"
	CodeCommandUnauthorized errors.Code = "COMMAND_UNAUTHORIZED"

	CodeQueryNotFound errors.Code = "QUERY_NOT_FOUND"
	CodeQueryFailed   errors.Code = "QUERY_FAILED"
	CodeQueryTimeout  errors.Code = "QUERY_TIMEOUT"

	CodeHandlerNotFound errors.Code = "HANDLER_NOT_FOUND"
	CodeHandlerPanic    errors.Code = "HANDLER_PANIC"
)

// ============================================================================
// Command Errors
// ============================================================================

// ErrCommandValidation creates a command validation error.
func ErrCommandValidation(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeCommandValidation)
}

// ErrCommandNotFound creates an error when a command handler is not found.
func ErrCommandNotFound(operation string, commandType string) *errors.Error {
	return errors.NotFound(operation, "command handler: "+commandType).
		WithCode(CodeCommandNotFound).
		WithMeta("command_type", commandType)
}

// ErrCommandFailed creates an error when a command execution fails.
func ErrCommandFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeCommandFailed)
}

// ErrCommandUnauthorized creates an error when a command is not authorized.
func ErrCommandUnauthorized(operation string, reason string) *errors.Error {
	return errors.Forbidden(operation, reason).
		WithCode(CodeCommandUnauthorized)
}

// ============================================================================
// Query Errors
// ============================================================================

// ErrQueryNotFound creates an error when a query handler is not found.
func ErrQueryNotFound(operation string, queryType string) *errors.Error {
	return errors.NotFound(operation, "query handler: "+queryType).
		WithCode(CodeQueryNotFound).
		WithMeta("query_type", queryType)
}

// ErrQueryFailed creates an error when a query execution fails.
func ErrQueryFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeQueryFailed)
}

// ErrQueryTimeout creates an error when a query times out.
func ErrQueryTimeout(operation string, queryType string) *errors.Error {
	return errors.Timeout(operation, "query timed out: "+queryType).
		WithCode(CodeQueryTimeout).
		WithMeta("query_type", queryType)
}

// ============================================================================
// Handler Errors
// ============================================================================

// ErrHandlerNotFound creates an error when a handler is not registered.
func ErrHandlerNotFound(operation string, handlerType string) *errors.Error {
	return errors.NotFound(operation, "handler: "+handlerType).
		WithCode(CodeHandlerNotFound).
		WithMeta("handler_type", handlerType)
}

// ErrHandlerPanic creates an error when a handler panics.
func ErrHandlerPanic(operation string, recovered any) *errors.Error {
	return errors.Internalf(operation, "handler panic: %v", recovered).
		WithCode(CodeHandlerPanic).
		WithMeta("panic", fmt.Sprintf("%v", recovered))
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsCommandValidation returns true if the error is a command validation error.
func IsCommandValidation(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeCommandValidation
	}
	return false
}

// IsCommandNotFound returns true if the error is a command not found error.
func IsCommandNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeCommandNotFound
	}
	return false
}

// IsQueryNotFound returns true if the error is a query not found error.
func IsQueryNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeQueryNotFound
	}
	return false
}

// IsHandlerNotFound returns true if the error is a handler not found error.
func IsHandlerNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok {
		return e.Code == CodeHandlerNotFound
	}
	return false
}
