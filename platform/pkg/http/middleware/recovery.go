package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/0xsj/nexus/platform/pkg/http/response"
)

// ============================================================================
// Recovery Middleware
// ============================================================================

// RecoveryConfig configures the recovery middleware.
type RecoveryConfig struct {
	// EnableStackTrace includes the stack trace in logs.
	// Default: true
	EnableStackTrace bool

	// LogFunc is called when a panic is recovered.
	// Default: prints to stderr
	LogFunc func(r *http.Request, err any, stack []byte)

	// ResponseFunc customizes the error response.
	// Default: returns 500 Internal Server Error
	ResponseFunc func(w http.ResponseWriter, r *http.Request, err any)
}

// DefaultRecoveryConfig returns the default configuration.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		EnableStackTrace: true,
		LogFunc:          defaultRecoveryLog,
		ResponseFunc:     defaultRecoveryResponse,
	}
}

// Recovery returns a middleware that recovers from panics.
func Recovery() Middleware {
	return RecoveryWithConfig(DefaultRecoveryConfig())
}

// RecoveryWithConfig returns a recovery middleware with custom configuration.
func RecoveryWithConfig(config RecoveryConfig) Middleware {
	if config.LogFunc == nil {
		config.LogFunc = defaultRecoveryLog
	}
	if config.ResponseFunc == nil {
		config.ResponseFunc = defaultRecoveryResponse
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					var stack []byte
					if config.EnableStackTrace {
						stack = debug.Stack()
					}

					// Log the panic
					config.LogFunc(r, err, stack)

					// Send error response
					config.ResponseFunc(w, r, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Default Handlers
// ============================================================================

// defaultRecoveryLog logs the panic to stderr.
func defaultRecoveryLog(r *http.Request, err any, stack []byte) {
	requestID := GetRequestID(r.Context())

	fmt.Printf("[PANIC RECOVERY] request_id=%s method=%s path=%s error=%v\n",
		requestID, r.Method, r.URL.Path, err)

	if len(stack) > 0 {
		fmt.Printf("[STACK TRACE]\n%s\n", string(stack))
	}
}

// defaultRecoveryResponse sends a generic 500 error response.
func defaultRecoveryResponse(w http.ResponseWriter, r *http.Request, err any) {
	// Check if headers have already been written
	if isHeaderWritten(w) {
		return
	}

	response.InternalError(w, &response.ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "An unexpected error occurred",
	})
}

// ============================================================================
// Recovery with Logger Integration
// ============================================================================

// Logger interface for recovery middleware.
type RecoveryLogger interface {
	Error(msg string, args ...any)
}

// RecoveryWithLogger returns a recovery middleware that uses the provided logger.
func RecoveryWithLogger(logger RecoveryLogger) Middleware {
	config := DefaultRecoveryConfig()
	config.LogFunc = func(r *http.Request, err any, stack []byte) {
		requestID := GetRequestID(r.Context())

		if len(stack) > 0 {
			logger.Error("panic recovered",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"error", fmt.Sprintf("%v", err),
				"stack", string(stack),
			)
		} else {
			logger.Error("panic recovered",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"error", fmt.Sprintf("%v", err),
			)
		}
	}

	return RecoveryWithConfig(config)
}

// ============================================================================
// Response Writer Wrapper
// ============================================================================

// isHeaderWritten checks if response headers have been written.
// This is a best-effort check since standard ResponseWriter doesn't expose this.
func isHeaderWritten(w http.ResponseWriter) bool {
	// If we can unwrap to our StatusRecorder, check its state
	if sr, ok := w.(*StatusRecorder); ok {
		return sr.Written
	}
	return false
}

// StatusRecorder wraps http.ResponseWriter to record the status code.
type StatusRecorder struct {
	http.ResponseWriter
	StatusCode   int
	Written      bool
	BytesWritten int
}

// NewStatusRecorder creates a new StatusRecorder.
func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// WriteHeader records the status code.
func (r *StatusRecorder) WriteHeader(code int) {
	if !r.Written {
		r.StatusCode = code
		r.Written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

// Write records that the response has been written.
func (r *StatusRecorder) Write(b []byte) (int, error) {
	if !r.Written {
		r.Written = true
	}
	n, err := r.ResponseWriter.Write(b)
	r.BytesWritten += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter.
func (r *StatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// ============================================================================
// Record Status Middleware
// ============================================================================

// RecordStatus wraps the response writer to record status code and bytes written.
func RecordStatus() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := NewStatusRecorder(w)
			next.ServeHTTP(recorder, r)
		})
	}
}
