package domain

import (
	"fmt"
	"time"
)

// ============================================================================
// DeliveryStatus
// ============================================================================

// DeliveryStatus represents the delivery state of a notification.
type DeliveryStatus int

const (
	// DeliveryStatusQueued indicates the notification is waiting for delivery.
	DeliveryStatusQueued DeliveryStatus = 1

	// DeliveryStatusSending indicates delivery is in progress.
	DeliveryStatusSending DeliveryStatus = 2

	// DeliveryStatusDelivered indicates the notification was successfully delivered.
	DeliveryStatusDelivered DeliveryStatus = 3

	// DeliveryStatusFailed indicates the delivery failed.
	DeliveryStatusFailed DeliveryStatus = 4

	// DeliveryStatusSuppressed indicates the notification was suppressed.
	DeliveryStatusSuppressed DeliveryStatus = 5

	// DeliveryStatusRead indicates the user has read the notification.
	DeliveryStatusRead DeliveryStatus = 6
)

// String returns the string representation of the delivery status.
func (s DeliveryStatus) String() string {
	switch s {
	case DeliveryStatusQueued:
		return "queued"
	case DeliveryStatusSending:
		return "sending"
	case DeliveryStatusDelivered:
		return "delivered"
	case DeliveryStatusFailed:
		return "failed"
	case DeliveryStatusSuppressed:
		return "suppressed"
	case DeliveryStatusRead:
		return "read"
	default:
		return "unknown"
	}
}

// ParseDeliveryStatus parses a string into a DeliveryStatus.
func ParseDeliveryStatus(s string) (DeliveryStatus, error) {
	switch s {
	case "queued":
		return DeliveryStatusQueued, nil
	case "sending":
		return DeliveryStatusSending, nil
	case "delivered":
		return DeliveryStatusDelivered, nil
	case "failed":
		return DeliveryStatusFailed, nil
	case "suppressed":
		return DeliveryStatusSuppressed, nil
	case "read":
		return DeliveryStatusRead, nil
	default:
		return 0, fmt.Errorf("invalid delivery status: %s", s)
	}
}

// IsTerminal returns true if the delivery status is a terminal state.
func (s DeliveryStatus) IsTerminal() bool {
	return s == DeliveryStatusFailed || s == DeliveryStatusSuppressed
}

// IsDelivered returns true if the notification has been delivered or read.
func (s DeliveryStatus) IsDelivered() bool {
	return s == DeliveryStatusDelivered || s == DeliveryStatusRead
}

// ============================================================================
// DeliveryAttempt
// ============================================================================

// DeliveryAttempt is an immutable value object recording a single delivery attempt.
type DeliveryAttempt struct {
	channel      Channel
	attemptedAt  time.Time
	status       DeliveryStatus
	errorMessage string
}

// NewDeliveryAttempt creates a new DeliveryAttempt with the given channel and status.
func NewDeliveryAttempt(channel Channel, status DeliveryStatus) DeliveryAttempt {
	return DeliveryAttempt{
		channel:     channel,
		attemptedAt: time.Now().UTC(),
		status:      status,
	}
}

// WithError returns a new DeliveryAttempt with the error message set (immutable).
func (d DeliveryAttempt) WithError(msg string) DeliveryAttempt {
	return DeliveryAttempt{
		channel:      d.channel,
		attemptedAt:  d.attemptedAt,
		status:       d.status,
		errorMessage: msg,
	}
}

// Channel returns the delivery channel.
func (d DeliveryAttempt) Channel() Channel {
	return d.channel
}

// AttemptedAt returns the time the delivery was attempted.
func (d DeliveryAttempt) AttemptedAt() time.Time {
	return d.attemptedAt
}

// Status returns the delivery status.
func (d DeliveryAttempt) Status() DeliveryStatus {
	return d.status
}

// ErrorMessage returns the error message, if any.
func (d DeliveryAttempt) ErrorMessage() string {
	return d.errorMessage
}
