package middleware

import (
	"net/http"
	"strings"

	"github.com/0xsj/nexus/pkg/security/token/jwt"
)

// AuthenticateConfig holds configuration for the authentication middleware.
type AuthenticateConfig struct {
	// TokenExtractor is a custom function to extract token from request
	// If nil, defaults to extracting from Authorization header
	TokenExtractor func(*http.Request) string

	// ErrorHandler is a custom function to handle authentication errors
	// If nil, defaults to returning JSON error response
	ErrorHandler func(http.ResponseWriter, *http.Request, error)

	// Optional indicates whether authentication is optional
	// If true, continues to next handler even if token is invalid/missing
	// but still adds claims to context if token is valid
	Optional bool
}

// DefaultAuthenticateConfig returns configuration with sensible defaults.
func DefaultAuthenticateConfig() AuthenticateConfig {
	return AuthenticateConfig{
		TokenExtractor: extractBearerToken,
		ErrorHandler:   defaultErrorHandler,
		Optional:       false,
	}
}

// Authenticate creates an authentication middleware using JWT.
func Authenticate(jwtManager jwt.Manager) func(http.Handler) http.Handler {
	return AuthenticateWithConfig(jwtManager, DefaultAuthenticateConfig())
}

// AuthenticateWithConfig creates an authentication middleware with custom config.
func AuthenticateWithConfig(jwtManager jwt.Manager, config AuthenticateConfig) func(http.Handler) http.Handler {
	// Set defaults
	if config.TokenExtractor == nil {
		config.TokenExtractor = extractBearerToken
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token
			token := config.TokenExtractor(r)
			if token == "" {
				if config.Optional {
					// Continue without authentication
					next.ServeHTTP(w, r)
					return
				}
				config.ErrorHandler(w, r, ErrMissingBearerToken)
				return
			}

			// Validate token
			claimsResult := jwtManager.ValidateAccessToken(token)
			if claimsResult.IsErr() {
				err := claimsResult.UnwrapErr()

				if config.Optional {
					// Continue without authentication
					next.ServeHTTP(w, r)
					return
				}

				config.ErrorHandler(w, r, err)
				return
			}

			// Get claims
			claims := claimsResult.Unwrap()

			// Add claims and user ID to context
			ctx := WithClaims(r.Context(), claims)
			ctx = WithUserID(ctx, claims.UserID)

			// Continue with authenticated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthenticate creates an authentication middleware that doesn't fail on missing/invalid tokens.
// Useful for endpoints that work for both authenticated and anonymous users.
func OptionalAuthenticate(jwtManager jwt.Manager) func(http.Handler) http.Handler {
	config := DefaultAuthenticateConfig()
	config.Optional = true
	return AuthenticateWithConfig(jwtManager, config)
}

// extractBearerToken extracts the JWT token from the Authorization header.
// Expected format: "Authorization: Bearer <token>"
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Split "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	scheme := strings.ToLower(parts[0])
	if scheme != "bearer" {
		return ""
	}

	return parts[1]
}

// extractTokenFromQuery extracts token from query parameter.
// Useful for WebSocket connections or download links.
func extractTokenFromQuery(paramName string) func(*http.Request) string {
	return func(r *http.Request) string {
		return r.URL.Query().Get(paramName)
	}
}

// extractTokenFromCookie extracts token from cookie.
func extractTokenFromCookie(cookieName string) func(*http.Request) string {
	return func(r *http.Request) string {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return ""
		}
		return cookie.Value
	}
}

// ChainTokenExtractors tries multiple token extractors in order.
// Returns the first non-empty token found.
func ChainTokenExtractors(extractors ...func(*http.Request) string) func(*http.Request) string {
	return func(r *http.Request) string {
		for _, extractor := range extractors {
			if token := extractor(r); token != "" {
				return token
			}
		}
		return ""
	}
}

// defaultErrorHandler returns a JSON error response.
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Determine status code based on error
	statusCode := http.StatusUnauthorized
	message := "unauthorized"

	switch err {
	case ErrMissingAuthHeader, ErrMissingBearerToken:
		statusCode = http.StatusUnauthorized
		message = "missing authentication token"
	case ErrInvalidAuthHeader, ErrInvalidToken:
		statusCode = http.StatusUnauthorized
		message = "invalid authentication token"
	case ErrTokenExpired:
		statusCode = http.StatusUnauthorized
		message = "authentication token expired"
	default:
		statusCode = http.StatusUnauthorized
		message = "unauthorized"
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
