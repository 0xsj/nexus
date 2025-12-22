package errors

import "fmt"

// Severity indicates the error severity level.
// Used for log level mapping, alerting decisions, and operational response.
//
// Severity levels follow standard operational practices:
//   - Warning: Recoverable issues, may need attention
//   - Error: Failed operations, requires investigation
//   - Critical: System-level failures, requires immediate attention
type Severity int

const (
	// SeverityWarning indicates a recoverable issue.
	// The operation may have partially succeeded or used a fallback.
	// Examples: deprecated API usage, fallback to default config, retry succeeded
	SeverityWarning Severity = iota

	// SeverityError indicates a failed operation.
	// The operation did not complete successfully.
	// Examples: validation failed, resource not found, unauthorized access
	SeverityError

	// SeverityCritical indicates a system-level failure.
	// Requires immediate attention and likely triggers alerts.
	// Examples: database connection lost, out of memory, data corruption
	SeverityCritical
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// Validate checks if the severity is a valid value.
func (s Severity) Validate() error {
	switch s {
	case SeverityWarning, SeverityError, SeverityCritical:
		return nil
	default:
		return fmt.Errorf("invalid severity: %d", s)
	}
}

// IsHigherThan returns true if s is more severe than other.
func (s Severity) IsHigherThan(other Severity) bool {
	return s > other
}

// ShouldAlert returns true if the severity level should trigger alerts.
func (s Severity) ShouldAlert() bool {
	return s >= SeverityCritical
}
