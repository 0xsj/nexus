package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// ============================================================================
// CORS Middleware
// ============================================================================

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed.
	// Use "*" to allow all origins.
	// Default: []
	AllowedOrigins []string

	// AllowedMethods is a list of methods that are allowed.
	// Default: GET, POST, PUT, PATCH, DELETE, OPTIONS
	AllowedMethods []string

	// AllowedHeaders is a list of headers that are allowed.
	// Default: Accept, Authorization, Content-Type, X-Request-ID
	AllowedHeaders []string

	// ExposedHeaders is a list of headers that are exposed to the browser.
	// Default: []
	ExposedHeaders []string

	// AllowCredentials indicates whether credentials are allowed.
	// Default: false
	AllowCredentials bool

	// MaxAge is the maximum age (in seconds) of the preflight cache.
	// Default: 86400 (24 hours)
	MaxAge int
}

// DefaultCORSConfig returns the default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a CORS middleware with the provided allowed origins.
func CORS(allowedOrigins ...string) Middleware {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	return CORSWithConfig(config)
}

// CORSAllowAll returns a CORS middleware that allows all origins.
// Use with caution - only for development or public APIs.
func CORSAllowAll() Middleware {
	config := DefaultCORSConfig()
	config.AllowedOrigins = []string{"*"}
	return CORSWithConfig(config)
}

// CORSWithConfig returns a CORS middleware with custom configuration.
func CORSWithConfig(config CORSConfig) Middleware {
	// Pre-compute joined strings
	allowedMethods := strings.Join(config.AllowedMethods, ", ")
	allowedHeaders := strings.Join(config.AllowedHeaders, ", ")
	exposedHeaders := strings.Join(config.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(config.MaxAge)

	// Build origin lookup map
	allowAllOrigins := false
	originMap := make(map[string]bool, len(config.AllowedOrigins))
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			allowAllOrigins = true
			break
		}
		originMap[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If no origin header, this isn't a CORS request
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if origin is allowed
			allowed := allowAllOrigins || originMap[origin]
			if !allowed {
				// Origin not allowed - don't set CORS headers
				next.ServeHTTP(w, r)
				return
			}

			// Set CORS headers
			if allowAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if exposedHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				w.Header().Set("Access-Control-Max-Age", maxAge)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// CORS Builder
// ============================================================================

// CORSBuilder provides a fluent API for configuring CORS.
type CORSBuilder struct {
	config CORSConfig
}

// NewCORSBuilder creates a new CORS builder with default configuration.
func NewCORSBuilder() *CORSBuilder {
	return &CORSBuilder{
		config: DefaultCORSConfig(),
	}
}

// AllowOrigins sets the allowed origins.
func (b *CORSBuilder) AllowOrigins(origins ...string) *CORSBuilder {
	b.config.AllowedOrigins = origins
	return b
}

// AllowMethods sets the allowed methods.
func (b *CORSBuilder) AllowMethods(methods ...string) *CORSBuilder {
	b.config.AllowedMethods = methods
	return b
}

// AllowHeaders sets the allowed headers.
func (b *CORSBuilder) AllowHeaders(headers ...string) *CORSBuilder {
	b.config.AllowedHeaders = headers
	return b
}

// ExposeHeaders sets the exposed headers.
func (b *CORSBuilder) ExposeHeaders(headers ...string) *CORSBuilder {
	b.config.ExposedHeaders = headers
	return b
}

// AllowCredentials enables credentials.
func (b *CORSBuilder) AllowCredentials(allow bool) *CORSBuilder {
	b.config.AllowCredentials = allow
	return b
}

// MaxAge sets the preflight cache max age in seconds.
func (b *CORSBuilder) MaxAge(seconds int) *CORSBuilder {
	b.config.MaxAge = seconds
	return b
}

// Build creates the CORS middleware.
func (b *CORSBuilder) Build() Middleware {
	return CORSWithConfig(b.config)
}
