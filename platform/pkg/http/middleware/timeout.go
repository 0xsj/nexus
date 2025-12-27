package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/pkg/http/response"
)

// ============================================================================
// Timeout Middleware
// ============================================================================

// TimeoutConfig configures the timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration for a request.
	// Default: 30 seconds
	Timeout time.Duration

	// ErrorHandler is called when a timeout occurs.
	// Default: returns 503 Service Unavailable
	ErrorHandler func(w http.ResponseWriter, r *http.Request)

	// SkipPaths are paths to skip timeout for (e.g., long-polling, SSE).
	SkipPaths []string
}

// DefaultTimeoutConfig returns the default timeout configuration.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout:      30 * time.Second,
		ErrorHandler: defaultTimeoutHandler,
		SkipPaths:    nil,
	}
}

// Timeout returns a timeout middleware with the specified duration.
func Timeout(timeout time.Duration) Middleware {
	config := DefaultTimeoutConfig()
	config.Timeout = timeout
	return TimeoutWithConfig(config)
}

// TimeoutWithConfig returns a timeout middleware with custom configuration.
func TimeoutWithConfig(config TimeoutConfig) Middleware {
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultTimeoutHandler
	}

	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip timeout for certain paths
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), config.Timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Create a wrapped response writer that tracks writes
			tw := &timeoutWriter{
				ResponseWriter: w,
				done:           done,
			}

			// Run the handler in a goroutine
			go func() {
				defer close(done)
				next.ServeHTTP(tw, r.WithContext(ctx))
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Handler completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				tw.mu.Lock()
				defer tw.mu.Unlock()

				if !tw.written {
					config.ErrorHandler(w, r)
				}
				// If already written, we can't do much - the response is partial
			}
		})
	}
}

// defaultTimeoutHandler sends a 503 Service Unavailable response.
func defaultTimeoutHandler(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusServiceUnavailable, &response.ErrorResponse{
		Code:    "REQUEST_TIMEOUT",
		Message: "Request timed out",
		Details: "The server took too long to process your request",
	})
}

// ============================================================================
// Timeout Writer
// ============================================================================

// timeoutWriter wraps http.ResponseWriter to track if a response has been written.
type timeoutWriter struct {
	http.ResponseWriter
	mu      sync.Mutex
	written bool
	done    chan struct{}
}

// WriteHeader marks the response as written.
func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.written {
		return
	}

	// Check if we've timed out
	select {
	case <-tw.done:
		// Already done, don't write
		return
	default:
	}

	tw.written = true
	tw.ResponseWriter.WriteHeader(code)
}

// Write marks the response as written.
func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.written {
		// Already started writing, continue
		return tw.ResponseWriter.Write(b)
	}

	// Check if we've timed out
	select {
	case <-tw.done:
		// Already done, don't write
		return 0, context.DeadlineExceeded
	default:
	}

	tw.written = true
	return tw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter.
func (tw *timeoutWriter) Unwrap() http.ResponseWriter {
	return tw.ResponseWriter
}
