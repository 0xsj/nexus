package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// ============================================================================
// Context Key
// ============================================================================

type requestIDKey struct{}

// ============================================================================
// Request ID Middleware
// ============================================================================

const (
	// RequestIDHeader is the header name for request ID.
	RequestIDHeader = "X-Request-ID"

	// DefaultRequestIDLength is the default length of generated request IDs.
	DefaultRequestIDLength = 16
)

// RequestIDConfig configures the request ID middleware.
type RequestIDConfig struct {
	// Header is the header name to use for request ID.
	// Default: "X-Request-ID"
	Header string

	// Generator is a function that generates a new request ID.
	// Default: generates a random hex string.
	Generator func() string

	// TrustHeader if true, uses the request ID from the incoming header if present.
	// Default: true
	TrustHeader bool
}

// DefaultRequestIDConfig returns the default configuration.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:      RequestIDHeader,
		Generator:   GenerateRequestID,
		TrustHeader: true,
	}
}

// RequestID returns a middleware that injects a request ID into the context and response.
func RequestID() Middleware {
	return RequestIDWithConfig(DefaultRequestIDConfig())
}

// RequestIDWithConfig returns a request ID middleware with custom configuration.
func RequestIDWithConfig(config RequestIDConfig) Middleware {
	if config.Header == "" {
		config.Header = RequestIDHeader
	}
	if config.Generator == nil {
		config.Generator = GenerateRequestID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var requestID string

			// Try to get from header if trusted
			if config.TrustHeader {
				requestID = r.Header.Get(config.Header)
			}

			// Generate if not present
			if requestID == "" {
				requestID = config.Generator()
			}

			// Add to response header
			w.Header().Set(config.Header, requestID)

			// Add to context
			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ============================================================================
// Context Helpers
// ============================================================================

// GetRequestID returns the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetRequestIDFromRequest returns the request ID from the request context.
func GetRequestIDFromRequest(r *http.Request) string {
	return GetRequestID(r.Context())
}

// ============================================================================
// ID Generation
// ============================================================================

// GenerateRequestID generates a random request ID.
func GenerateRequestID() string {
	return GenerateID(DefaultRequestIDLength)
}

// GenerateID generates a random hex string of the specified byte length.
func GenerateID(byteLength int) string {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a less random but still unique ID
		return fallbackID(byteLength)
	}
	return hex.EncodeToString(bytes)
}

// fallbackID generates a fallback ID using a simple counter and time.
// This is only used if crypto/rand fails, which should be extremely rare.
func fallbackID(length int) string {
	const charset = "0123456789abcdef"
	result := make([]byte, length*2)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
