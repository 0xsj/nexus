// pkg/graphql/middleware/auth.go

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/0xsj/nexus-go/pkg/graphql/gqlctx"
	"github.com/0xsj/nexus-go/pkg/observability/logger"
	"github.com/0xsj/nexus-go/pkg/security/jwt"
)

// ============================================================================
// Auth Middleware (HTTP level)
// ============================================================================

// AuthMiddleware extracts JWT tokens and populates context.
// This runs at the HTTP level before GraphQL execution.
type AuthMiddleware struct {
	jwtService jwt.Service
	logger     logger.Logger
	opts       authOptions
}

// authOptions holds configuration for the auth middleware.
type authOptions struct {
	tenancyEnabled  bool
	defaultTenantID string
	optional        bool
}

// AuthOption configures the auth middleware.
type AuthOption func(*authOptions)

// WithTenancy enables or disables tenancy support.
func WithTenancy(enabled bool, defaultTenantID string) AuthOption {
	return func(o *authOptions) {
		o.tenancyEnabled = enabled
		o.defaultTenantID = defaultTenantID
	}
}

// WithOptionalAuth allows unauthenticated requests to proceed.
func WithOptionalAuth(optional bool) AuthOption {
	return func(o *authOptions) {
		o.optional = optional
	}
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) *AuthMiddleware {
	options := authOptions{
		tenancyEnabled:  true,
		defaultTenantID: "default",
		optional:        true,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     log,
		opts:       options,
	}
}

// Handler returns an HTTP middleware that extracts auth from requests.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token := extractBearerToken(r)

		if token == "" {
			if !m.opts.optional {
				http.Error(w, `{"errors":[{"message":"unauthorized","extensions":{"code":"UNAUTHENTICATED"}}]}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwtService.Validate(ctx, token)
		if err != nil {
			m.logger.Warn("invalid JWT token",
				logger.String("path", r.URL.Path),
				logger.Err(err),
			)

			if !m.opts.optional {
				http.Error(w, `{"errors":[{"message":"invalid or expired token","extensions":{"code":"UNAUTHENTICATED"}}]}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		ctx = m.addClaimsToContext(ctx, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// addClaimsToContext adds JWT claims to the context.
func (m *AuthMiddleware) addClaimsToContext(ctx context.Context, claims jwt.Claims) context.Context {
	ctx = gqlctx.WithClaims(ctx, claims)

	if userID := claims.GetString("user_id"); userID != "" {
		ctx = gqlctx.WithUserID(ctx, userID)
	} else if subject := claims.Subject(); subject != "" {
		ctx = gqlctx.WithUserID(ctx, subject)
	}

	if m.opts.tenancyEnabled {
		if tenantID := claims.GetString("tenant_id"); tenantID != "" {
			ctx = gqlctx.WithTenantID(ctx, tenantID)
		}
	} else {
		ctx = gqlctx.WithTenantID(ctx, m.opts.defaultTenantID)
	}

	if sessionID := claims.GetString("session_id"); sessionID != "" {
		ctx = gqlctx.WithSessionID(ctx, sessionID)
	}

	if email := claims.GetString("email"); email != "" {
		ctx = gqlctx.WithEmail(ctx, email)
	}

	if role := claims.GetString("role"); role != "" {
		ctx = gqlctx.WithRole(ctx, role)
	}

	return ctx
}

// extractBearerToken extracts the JWT token from Authorization header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return parts[1]
}

// ============================================================================
// Functional API
// ============================================================================

// Auth returns an HTTP middleware function for authentication.
func Auth(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) func(http.Handler) http.Handler {
	m := NewAuthMiddleware(jwtService, log, opts...)
	return m.Handler
}

// ============================================================================
// Websocket Auth (for subscriptions)
// ============================================================================

// WebsocketInitFunc returns a function for authenticating websocket connections.
func WebsocketInitFunc(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) func(ctx context.Context, initPayload map[string]interface{}) (context.Context, *error) {
	options := authOptions{
		tenancyEnabled:  true,
		defaultTenantID: "default",
		optional:        false,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return func(ctx context.Context, initPayload map[string]interface{}) (context.Context, *error) {
		token := extractTokenFromPayload(initPayload)

		if token == "" {
			if options.optional {
				return ctx, nil
			}
			err := error(authError("missing authentication token"))
			return nil, &err
		}

		claims, err := jwtService.Validate(ctx, token)
		if err != nil {
			log.Warn("invalid websocket JWT token", logger.Err(err))

			if options.optional {
				return ctx, nil
			}
			authErr := error(authError("invalid or expired token"))
			return nil, &authErr
		}

		ctx = gqlctx.WithClaims(ctx, claims)

		if userID := claims.GetString("user_id"); userID != "" {
			ctx = gqlctx.WithUserID(ctx, userID)
		} else if subject := claims.Subject(); subject != "" {
			ctx = gqlctx.WithUserID(ctx, subject)
		}

		if options.tenancyEnabled {
			if tenantID := claims.GetString("tenant_id"); tenantID != "" {
				ctx = gqlctx.WithTenantID(ctx, tenantID)
			}
		} else {
			ctx = gqlctx.WithTenantID(ctx, options.defaultTenantID)
		}

		if sessionID := claims.GetString("session_id"); sessionID != "" {
			ctx = gqlctx.WithSessionID(ctx, sessionID)
		}

		if email := claims.GetString("email"); email != "" {
			ctx = gqlctx.WithEmail(ctx, email)
		}

		if role := claims.GetString("role"); role != "" {
			ctx = gqlctx.WithRole(ctx, role)
		}

		return ctx, nil
	}
}

// authError is a simple error type for auth failures.
type authError string

func (e authError) Error() string {
	return string(e)
}

// extractTokenFromPayload extracts the token from websocket init payload.
func extractTokenFromPayload(payload map[string]interface{}) string {
	if auth, ok := payload["Authorization"].(string); ok {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
		return auth
	}

	if auth, ok := payload["authorization"].(string); ok {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
		return auth
	}

	if token, ok := payload["authToken"].(string); ok {
		return token
	}

	if token, ok := payload["token"].(string); ok {
		return token
	}

	return ""
}
