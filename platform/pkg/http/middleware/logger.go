package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Logger Middleware
// ============================================================================

// LoggerConfig configures the logger middleware.
type LoggerConfig struct {
	// Logger is the logger to use.
	// If nil, logs to stdout.
	Logger log.Logger

	// SkipPaths are paths to skip logging for (e.g., health checks).
	SkipPaths []string

	// IncludeHeaders if true, logs request headers.
	IncludeHeaders bool

	// IncludeQuery if true, logs query parameters.
	IncludeQuery bool

	// SlowThreshold logs requests slower than this as warnings.
	// Default: 0 (disabled)
	SlowThreshold time.Duration
}

// DefaultLoggerConfig returns the default configuration.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Logger:         nil,
		SkipPaths:      nil,
		IncludeHeaders: false,
		IncludeQuery:   true,
		SlowThreshold:  0,
	}
}

// Logger returns a request logging middleware with the provided logger.
func Logger(logger log.Logger) Middleware {
	config := DefaultLoggerConfig()
	config.Logger = logger
	return LoggerWithConfig(config)
}

// LoggerWithConfig returns a logger middleware with custom configuration.
func LoggerWithConfig(config LoggerConfig) Middleware {
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for certain paths
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Record start time
			start := time.Now()

			// Wrap response writer to capture status
			recorder := NewStatusRecorder(w)

			// Process request
			next.ServeHTTP(recorder, r)

			// Calculate duration
			duration := time.Since(start)

			// Build log entry
			entry := LogEntry{
				RequestID:    GetRequestID(r.Context()),
				Method:       r.Method,
				Path:         r.URL.Path,
				Query:        "",
				StatusCode:   recorder.StatusCode,
				Duration:     duration,
				BytesWritten: recorder.BytesWritten,
				ClientIP:     getClientIP(r),
				UserAgent:    r.UserAgent(),
			}

			if config.IncludeQuery && r.URL.RawQuery != "" {
				entry.Query = r.URL.RawQuery
			}

			// Log the request
			logRequest(config, entry)
		})
	}
}

// ============================================================================
// Log Entry
// ============================================================================

// LogEntry represents a single log entry for an HTTP request.
type LogEntry struct {
	RequestID    string
	Method       string
	Path         string
	Query        string
	StatusCode   int
	Duration     time.Duration
	BytesWritten int
	ClientIP     string
	UserAgent    string
}

// ============================================================================
// Log Output
// ============================================================================

// logRequest logs the request using the configured logger.
func logRequest(config LoggerConfig, entry LogEntry) {
	fields := []log.Field{
		log.String("request_id", entry.RequestID),
		log.String("method", entry.Method),
		log.String("path", entry.Path),
		log.Int("status", entry.StatusCode),
		log.Int64("duration_ms", entry.Duration.Milliseconds()),
		log.Int("bytes", entry.BytesWritten),
		log.String("client_ip", entry.ClientIP),
	}

	if entry.Query != "" {
		fields = append(fields, log.String("query", entry.Query))
	}

	// Determine log level
	isSlow := config.SlowThreshold > 0 && entry.Duration > config.SlowThreshold
	isError := entry.StatusCode >= 500

	if config.Logger != nil {
		switch {
		case isError:
			config.Logger.Error("http request", fields...)
		case isSlow:
			config.Logger.Warn("http request (slow)", fields...)
		default:
			config.Logger.Info("http request", fields...)
		}
	} else {
		// Default stdout logging
		logToStdout(entry, isSlow, isError)
	}
}

// logToStdout logs to stdout with a simple format.
func logToStdout(entry LogEntry, isSlow, isError bool) {
	level := "INFO"
	if isError {
		level = "ERROR"
	} else if isSlow {
		level = "WARN"
	}

	path := entry.Path
	if entry.Query != "" {
		path = entry.Path + "?" + entry.Query
	}

	fmt.Printf("[%s] %s %s %d %dms %db request_id=%s client_ip=%s\n",
		level,
		entry.Method,
		path,
		entry.StatusCode,
		entry.Duration.Milliseconds(),
		entry.BytesWritten,
		entry.RequestID,
		entry.ClientIP,
	)
}

// ============================================================================
// Client IP Helpers
// ============================================================================

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (common for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header (nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return cfip
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr

	// Remove port if present
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == ':' {
			return ip[:i]
		}
		if ip[i] == ']' {
			// IPv6 with brackets
			break
		}
	}

	return ip
}

// ============================================================================
// Convenience Constructors
// ============================================================================

// LoggerWithSkipPaths returns a logger middleware that skips certain paths.
func LoggerWithSkipPaths(logger log.Logger, paths ...string) Middleware {
	config := DefaultLoggerConfig()
	config.Logger = logger
	config.SkipPaths = paths
	return LoggerWithConfig(config)
}

// LoggerWithSlowThreshold returns a logger middleware that warns on slow requests.
func LoggerWithSlowThreshold(logger log.Logger, threshold time.Duration) Middleware {
	config := DefaultLoggerConfig()
	config.Logger = logger
	config.SlowThreshold = threshold
	return LoggerWithConfig(config)
}
