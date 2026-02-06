package domain

import (
	"fmt"
)

// IntegrationStatus represents the lifecycle status of an integration.
type IntegrationStatus int

const (
	// IntegrationStatusConnected indicates the integration is active and connected.
	IntegrationStatusConnected IntegrationStatus = 1

	// IntegrationStatusDisconnected indicates the integration has been disconnected.
	IntegrationStatusDisconnected IntegrationStatus = 2

	// IntegrationStatusError indicates the integration encountered an error.
	IntegrationStatusError IntegrationStatus = 3

	// IntegrationStatusSuspended indicates the integration has been suspended.
	IntegrationStatusSuspended IntegrationStatus = 4
)

// String returns the string representation of the integration status.
func (s IntegrationStatus) String() string {
	switch s {
	case IntegrationStatusConnected:
		return "connected"
	case IntegrationStatusDisconnected:
		return "disconnected"
	case IntegrationStatusError:
		return "error"
	case IntegrationStatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

// IsConnected returns true if the integration is connected.
func (s IntegrationStatus) IsConnected() bool {
	return s == IntegrationStatusConnected
}

// IsDisconnected returns true if the integration has been disconnected.
func (s IntegrationStatus) IsDisconnected() bool {
	return s == IntegrationStatusDisconnected
}

// IsError returns true if the integration is in an error state.
func (s IntegrationStatus) IsError() bool {
	return s == IntegrationStatusError
}

// IsSuspended returns true if the integration has been suspended.
func (s IntegrationStatus) IsSuspended() bool {
	return s == IntegrationStatusSuspended
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
func (s IntegrationStatus) CanTransitionTo(target IntegrationStatus) bool {
	switch s {
	case IntegrationStatusConnected:
		return target == IntegrationStatusDisconnected ||
			target == IntegrationStatusError ||
			target == IntegrationStatusSuspended
	case IntegrationStatusError:
		return target == IntegrationStatusConnected ||
			target == IntegrationStatusDisconnected ||
			target == IntegrationStatusSuspended
	case IntegrationStatusDisconnected:
		return target == IntegrationStatusConnected
	case IntegrationStatusSuspended:
		return false
	default:
		return false
	}
}

// ParseIntegrationStatus parses a string into an IntegrationStatus.
func ParseIntegrationStatus(s string) (IntegrationStatus, error) {
	switch s {
	case "connected":
		return IntegrationStatusConnected, nil
	case "disconnected":
		return IntegrationStatusDisconnected, nil
	case "error":
		return IntegrationStatusError, nil
	case "suspended":
		return IntegrationStatusSuspended, nil
	default:
		return 0, fmt.Errorf("invalid integration status: %s", s)
	}
}
