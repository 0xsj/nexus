package websocket

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// WebSocket specific error codes.
const (
	CodeUpgradeFailed    errors.Code = "WS_UPGRADE_FAILED"
	CodeConnectionClosed errors.Code = "WS_CONNECTION_CLOSED"
	CodeWriteFailed      errors.Code = "WS_WRITE_FAILED"
	CodeReadFailed       errors.Code = "WS_READ_FAILED"
	CodeInvalidMessage   errors.Code = "WS_INVALID_MESSAGE"
)

// ErrUpgradeFailed creates an error when WebSocket upgrade fails.
func ErrUpgradeFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeUpgradeFailed)
}

// ErrConnectionClosed creates an error when writing to a closed connection.
func ErrConnectionClosed(operation string, connID string) *errors.Error {
	return errors.Domain(operation, "websocket connection is closed").
		WithCode(CodeConnectionClosed).
		WithMeta("connection_id", connID)
}

// ErrWriteFailed creates an error when a write operation fails.
func ErrWriteFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeWriteFailed)
}

// ErrReadFailed creates an error when a read operation fails.
func ErrReadFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeReadFailed)
}

// ErrInvalidMessage creates an error for malformed messages.
func ErrInvalidMessage(operation string, reason string) *errors.Error {
	return errors.Validation(operation, reason).
		WithCode(CodeInvalidMessage)
}

// IsUpgradeFailed returns true if the error is an upgrade failure.
func IsUpgradeFailed(err error) bool {
	return errors.HasCode(err, CodeUpgradeFailed)
}

// IsConnectionClosed returns true if the error indicates a closed connection.
func IsConnectionClosed(err error) bool {
	return errors.HasCode(err, CodeConnectionClosed)
}

// IsWriteFailed returns true if the error is a write failure.
func IsWriteFailed(err error) bool {
	return errors.HasCode(err, CodeWriteFailed)
}

// IsInvalidMessage returns true if the error is an invalid message error.
func IsInvalidMessage(err error) bool {
	return errors.HasCode(err, CodeInvalidMessage)
}
