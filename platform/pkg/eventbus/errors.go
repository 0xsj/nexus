package eventbus

import (
	"github.com/0xsj/nexus/platform/pkg/errors"
)

// Event bus specific error codes.
const (
	CodePublishFailed       errors.Code = "EVENTBUS_PUBLISH_FAILED"
	CodeSubscribeFailed     errors.Code = "EVENTBUS_SUBSCRIBE_FAILED"
	CodeConnectionFailed    errors.Code = "EVENTBUS_CONNECTION_FAILED"
	CodeSerializationFailed errors.Code = "EVENTBUS_SERIALIZATION_FAILED"
	CodeHandlerFailed       errors.Code = "EVENTBUS_HANDLER_FAILED"
	CodeBusTimeout          errors.Code = "EVENTBUS_TIMEOUT"
	CodeBusClosed           errors.Code = "EVENTBUS_CLOSED"
)

// ErrPublishFailed creates an error when publishing events fails.
func ErrPublishFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodePublishFailed)
}

// ErrSubscribeFailed creates an error when subscribing fails.
func ErrSubscribeFailed(operation string, pattern string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeSubscribeFailed).
		WithMeta("pattern", pattern)
}

// ErrConnectionFailed creates an error when connecting to the bus backend fails.
func ErrConnectionFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeConnectionFailed)
}

// ErrSerializationFailed creates an error when event serialization/deserialization fails.
func ErrSerializationFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeSerializationFailed)
}

// ErrHandlerFailed creates an error when an event handler fails.
func ErrHandlerFailed(operation string, eventType string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeHandlerFailed).
		WithMeta("event_type", eventType)
}

// ErrBusTimeout creates an error when an event bus operation times out.
func ErrBusTimeout(operation string, detail string) *errors.Error {
	return errors.Timeout(operation, detail).
		WithCode(CodeBusTimeout)
}

// ErrBusClosed creates an error when the bus is used after closing.
func ErrBusClosed(operation string) *errors.Error {
	return errors.Domain(operation, "event bus is closed").
		WithCode(CodeBusClosed)
}

// IsPublishFailed returns true if the error is a publish failure.
func IsPublishFailed(err error) bool {
	return errors.HasCode(err, CodePublishFailed)
}

// IsSubscribeFailed returns true if the error is a subscribe failure.
func IsSubscribeFailed(err error) bool {
	return errors.HasCode(err, CodeSubscribeFailed)
}

// IsConnectionFailed returns true if the error is a connection failure.
func IsConnectionFailed(err error) bool {
	return errors.HasCode(err, CodeConnectionFailed)
}

// IsHandlerFailed returns true if the error is a handler failure.
func IsHandlerFailed(err error) bool {
	return errors.HasCode(err, CodeHandlerFailed)
}

// IsBusClosed returns true if the error indicates the bus is closed.
func IsBusClosed(err error) bool {
	return errors.HasCode(err, CodeBusClosed)
}
