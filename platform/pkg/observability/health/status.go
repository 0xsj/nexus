package health

import (
	"encoding/json"
	"time"
)

// Status represents the health status of a component or system.
type Status string

const (
	// StatusUp indicates the component is healthy.
	StatusUp Status = "up"

	// StatusDown indicates the component is unhealthy.
	StatusDown Status = "down"

	// StatusDegraded indicates the component is partially healthy.
	StatusDegraded Status = "degraded"

	// StatusUnknown indicates the health status is unknown.
	StatusUnknown Status = "unknown"
)

// String returns the string representation.
func (s Status) String() string {
	return string(s)
}

// IsHealthy returns true if the status indicates healthy operation.
func (s Status) IsHealthy() bool {
	return s == StatusUp
}

// IsUnhealthy returns true if the status indicates unhealthy operation.
func (s Status) IsUnhealthy() bool {
	return s == StatusDown
}

// ============================================================================
// Health Result
// ============================================================================

// Result represents the result of a health check.
type Result struct {
	// Status is the health status.
	Status Status `json:"status"`

	// Message is an optional human-readable message.
	Message string `json:"message,omitempty"`

	// Error is the error message if unhealthy.
	Error string `json:"error,omitempty"`

	// Duration is how long the check took.
	Duration time.Duration `json:"duration,omitempty"`

	// Timestamp is when the check was performed.
	Timestamp time.Time `json:"timestamp"`

	// Details contains additional check-specific information.
	Details map[string]any `json:"details,omitempty"`
}

// NewResult creates a new health result.
func NewResult(status Status) *Result {
	return &Result{
		Status:    status,
		Timestamp: time.Now().UTC(),
	}
}

// Up creates a healthy result.
func Up() *Result {
	return NewResult(StatusUp)
}

// Down creates an unhealthy result.
func Down(err error) *Result {
	r := NewResult(StatusDown)
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

// Degraded creates a degraded result.
func Degraded(message string) *Result {
	r := NewResult(StatusDegraded)
	r.Message = message
	return r
}

// Unknown creates an unknown status result.
func Unknown() *Result {
	return NewResult(StatusUnknown)
}

// WithMessage sets the message.
func (r *Result) WithMessage(msg string) *Result {
	r.Message = msg
	return r
}

// WithDuration sets the duration.
func (r *Result) WithDuration(d time.Duration) *Result {
	r.Duration = d
	return r
}

// WithDetail adds a detail.
func (r *Result) WithDetail(key string, value any) *Result {
	if r.Details == nil {
		r.Details = make(map[string]any)
	}
	r.Details[key] = value
	return r
}

// WithDetails sets multiple details.
func (r *Result) WithDetails(details map[string]any) *Result {
	r.Details = details
	return r
}

// ============================================================================
// System Health Response
// ============================================================================

// Response represents the overall system health response.
type Response struct {
	// Status is the overall system status.
	Status Status `json:"status"`

	// Version is the application version.
	Version string `json:"version,omitempty"`

	// Uptime is how long the system has been running.
	Uptime time.Duration `json:"uptime,omitempty"`

	// Timestamp is when the health check was performed.
	Timestamp time.Time `json:"timestamp"`

	// Checks contains individual component check results.
	Checks map[string]*Result `json:"checks,omitempty"`
}

// NewResponse creates a new health response.
func NewResponse(status Status) *Response {
	return &Response{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]*Result),
	}
}

// WithVersion sets the version.
func (r *Response) WithVersion(version string) *Response {
	r.Version = version
	return r
}

// WithUptime sets the uptime.
func (r *Response) WithUptime(uptime time.Duration) *Response {
	r.Uptime = uptime
	return r
}

// WithCheck adds a check result.
func (r *Response) WithCheck(name string, result *Result) *Response {
	r.Checks[name] = result
	return r
}

// HTTPStatus returns the appropriate HTTP status code.
func (r *Response) HTTPStatus() int {
	switch r.Status {
	case StatusUp:
		return 200
	case StatusDegraded:
		return 200 // Still serve traffic, but indicate degradation
	case StatusDown:
		return 503
	default:
		return 503
	}
}

// MarshalJSON implements custom JSON marshaling.
func (r *Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	return json.Marshal(&struct {
		*Alias
		Uptime string `json:"uptime,omitempty"`
	}{
		Alias:  (*Alias)(r),
		Uptime: r.Uptime.String(),
	})
}

// ============================================================================
// Liveness / Readiness / Startup Responses
// ============================================================================

// LivenessResponse is a minimal response for liveness probes.
type LivenessResponse struct {
	Status Status `json:"status"`
}

// ReadinessResponse is a response for readiness probes.
type ReadinessResponse struct {
	Status Status            `json:"status"`
	Checks map[string]Status `json:"checks,omitempty"`
}

// StartupResponse is a response for startup probes.
type StartupResponse struct {
	Status  Status `json:"status"`
	Started bool   `json:"started"`
}
