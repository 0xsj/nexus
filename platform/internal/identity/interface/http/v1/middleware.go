package v1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Context Keys
// ============================================================================

type contextKey string

const (
	contextKeyUserID    contextKey = "user_id"
	contextKeyDID       contextKey = "did"
	contextKeySessionID contextKey = "session_id"
	contextKeyAPIKeyID  contextKey = "api_key_id"
	contextKeyScopes    contextKey = "scopes"
)

// ============================================================================
// Context Helpers
// ============================================================================

// UserIDFromContext extracts the user ID from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(string)
	return userID, ok && userID != ""
}

// DIDFromContext extracts the DID from context.
func DIDFromContext(ctx context.Context) (string, bool) {
	did, ok := ctx.Value(contextKeyDID).(string)
	return did, ok && did != ""
}

// SessionIDFromContext extracts the session ID from context.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(contextKeySessionID).(string)
	return sessionID, ok && sessionID != ""
}

// APIKeyIDFromContext extracts the API key ID from context.
func APIKeyIDFromContext(ctx context.Context) (string, bool) {
	keyID, ok := ctx.Value(contextKeyAPIKeyID).(string)
	return keyID, ok && keyID != ""
}

// ScopesFromContext extracts the scopes from context.
func ScopesFromContext(ctx context.Context) []string {
	scopes, ok := ctx.Value(contextKeyScopes).([]string)
	if !ok {
		return nil
	}
	return scopes
}

// ============================================================================
// Auth Context
// ============================================================================

// AuthContext contains authenticated user information.
type AuthContext struct {
	UserID    string
	DID       string
	SessionID string
	APIKeyID  string
	Scopes    []string
}

// AuthContextFromContext extracts the full auth context.
func AuthContextFromContext(ctx context.Context) *AuthContext {
	userID, _ := UserIDFromContext(ctx)
	did, _ := DIDFromContext(ctx)
	sessionID, _ := SessionIDFromContext(ctx)
	apiKeyID, _ := APIKeyIDFromContext(ctx)
	scopes := ScopesFromContext(ctx)

	if userID == "" {
		return nil
	}

	return &AuthContext{
		UserID:    userID,
		DID:       did,
		SessionID: sessionID,
		APIKeyID:  apiKeyID,
		Scopes:    scopes,
	}
}

// ============================================================================
// Middleware
// ============================================================================

// Middleware provides authentication middleware.
type Middleware struct {
	tokenService domain.TokenService
	logger       log.Logger
}

// NewMiddleware creates a new middleware instance.
func NewMiddleware(tokenService domain.TokenService, logger log.Logger) *Middleware {
	return &Middleware{
		tokenService: tokenService,
		logger:       logger,
	}
}

// AuthRequired validates JWT and requires authentication.
func (m *Middleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] AuthRequired: path=%s\n", r.URL.Path)
		fmt.Printf("[DEBUG] Authorization header: %s\n", r.Header.Get("Authorization"))

		ctx, err := m.authenticate(r)
		if err != nil {
			fmt.Printf("[DEBUG] Auth failed: %v\n", err)
			WriteUnauthorized(w, "authentication required")
			return
		}

		fmt.Printf("[DEBUG] Auth success\n")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthOptional extracts auth info if present but doesn't require it.
func (m *Middleware) AuthOptional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := m.authenticate(r)
		if err != nil {
			// Continue without auth context
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authenticate extracts and validates auth credentials.
func (m *Middleware) authenticate(r *http.Request) (context.Context, error) {
	// Try Bearer token first
	token := extractBearerToken(r)
	if token != "" {
		return m.authenticateWithToken(r.Context(), token)
	}

	// No valid authentication found
	return nil, domain.ErrTokenInvalid("middleware.authenticate", "no token provided")
}

// authenticateWithToken validates a JWT access token.
func (m *Middleware) authenticateWithToken(ctx context.Context, token string) (context.Context, error) {
	claims, err := m.tokenService.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Add claims to context
	ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
	ctx = context.WithValue(ctx, contextKeyDID, claims.DID)
	ctx = context.WithValue(ctx, contextKeySessionID, claims.SessionID)

	return ctx, nil
}

// ============================================================================
// API Key Middleware
// ============================================================================

// APIKeyMiddleware provides API key authentication.
type APIKeyMiddleware struct {
	apiKeyValidator APIKeyValidator
	logger          log.Logger
}

// APIKeyValidator validates API keys.
type APIKeyValidator interface {
	ValidateAPIKey(ctx context.Context, rawKey string) (*APIKeyValidationResult, error)
}

// APIKeyValidationResult contains the result of API key validation.
type APIKeyValidationResult struct {
	Valid   bool
	UserID  string
	DID     string
	KeyID   string
	KeyName string
	Scopes  []string
}

// NewAPIKeyMiddleware creates a new API key middleware.
func NewAPIKeyMiddleware(validator APIKeyValidator, logger log.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeyValidator: validator,
		logger:          logger,
	}
}

// APIKeyAuth validates API key from header.
func (m *APIKeyMiddleware) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := extractAPIKey(r)
		if apiKey == "" {
			WriteUnauthorized(w, "api key required")
			return
		}

		result, err := m.apiKeyValidator.ValidateAPIKey(r.Context(), apiKey)
		if err != nil || !result.Valid {
			WriteUnauthorized(w, "invalid api key")
			return
		}

		// Add to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, contextKeyUserID, result.UserID)
		ctx = context.WithValue(ctx, contextKeyDID, result.DID)
		ctx = context.WithValue(ctx, contextKeyAPIKeyID, result.KeyID)
		ctx = context.WithValue(ctx, contextKeyScopes, result.Scopes)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APIKeyOrBearerAuth allows either API key or Bearer token authentication.
func (m *APIKeyMiddleware) APIKeyOrBearerAuth(tokenMiddleware *Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try API key first
			apiKey := extractAPIKey(r)
			if apiKey != "" {
				result, err := m.apiKeyValidator.ValidateAPIKey(r.Context(), apiKey)
				if err == nil && result.Valid {
					ctx := r.Context()
					ctx = context.WithValue(ctx, contextKeyUserID, result.UserID)
					ctx = context.WithValue(ctx, contextKeyDID, result.DID)
					ctx = context.WithValue(ctx, contextKeyAPIKeyID, result.KeyID)
					ctx = context.WithValue(ctx, contextKeyScopes, result.Scopes)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Try Bearer token
			token := extractBearerToken(r)
			if token != "" {
				ctx, err := tokenMiddleware.authenticateWithToken(r.Context(), token)
				if err == nil {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			WriteUnauthorized(w, "authentication required")
		})
	}
}

// ============================================================================
// Scope Middleware
// ============================================================================

// RequireScope creates middleware that requires specific scopes.
func RequireScope(required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes := ScopesFromContext(r.Context())

			// If no scopes in context, allow (JWT auth doesn't use scopes)
			if scopes == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has any of the required scopes
			hasScope := false
			for _, req := range required {
				for _, s := range scopes {
					if s == req || s == "admin" {
						hasScope = true
						break
					}
				}
				if hasScope {
					break
				}
			}

			if !hasScope {
				WriteForbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Token Extraction
// ============================================================================

// extractBearerToken extracts Bearer token from Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

// extractAPIKey extracts API key from X-API-Key header.
func extractAPIKey(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}

// ============================================================================
// Request Helpers
// ============================================================================

// GetClientIP extracts the client IP address from the request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in the form "IP:port"
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}

	return addr
}

// GetUserAgent extracts the User-Agent from the request.
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}
