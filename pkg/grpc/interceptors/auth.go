// pkg/grpc/interceptors/auth.go

package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/0xsj/nexus-go/pkg/observability/logger"
	"github.com/0xsj/nexus-go/pkg/security/jwt"
)

// ============================================================================
// Auth Interceptor
// ============================================================================

// AuthInterceptor provides JWT authentication for gRPC handlers.
type AuthInterceptor struct {
	jwtService jwt.Service
	logger     logger.Logger
	opts       authOptions
}

// authOptions holds configuration for the auth interceptor.
type authOptions struct {
	// skipMethods is a list of methods that don't require authentication.
	skipMethods map[string]bool

	// tenancyEnabled enables multi-tenancy support.
	tenancyEnabled bool

	// defaultTenantID is used when tenancy is disabled.
	defaultTenantID string
}

// AuthOption configures the auth interceptor.
type AuthOption func(*authOptions)

// WithSkipAuth sets methods that don't require authentication.
func WithSkipAuth(methods ...string) AuthOption {
	return func(o *authOptions) {
		if o.skipMethods == nil {
			o.skipMethods = make(map[string]bool)
		}
		for _, m := range methods {
			o.skipMethods[m] = true
		}
	}
}

// WithTenancy enables or disables tenancy support.
func WithTenancy(enabled bool, defaultTenantID string) AuthOption {
	return func(o *authOptions) {
		o.tenancyEnabled = enabled
		o.defaultTenantID = defaultTenantID
	}
}

// NewAuthInterceptor creates a new auth interceptor.
func NewAuthInterceptor(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) *AuthInterceptor {
	options := authOptions{
		skipMethods:     make(map[string]bool),
		tenancyEnabled:  true,
		defaultTenantID: "default",
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &AuthInterceptor{
		jwtService: jwtService,
		logger:     log,
		opts:       options,
	}
}

// Unary returns a unary server interceptor that validates JWT tokens.
//
// Usage:
//
//	auth := interceptors.NewAuthInterceptor(jwtService, log)
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(auth.Unary()),
//	)
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for excluded methods
		if a.opts.skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Authenticate
		newCtx, err := a.authenticate(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// Stream returns a stream server interceptor that validates JWT tokens.
//
// Usage:
//
//	auth := interceptors.NewAuthInterceptor(jwtService, log)
//	server := grpc.NewServer(
//	    grpc.StreamInterceptor(auth.Stream()),
//	)
func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip auth for excluded methods
		if a.opts.skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// Authenticate
		newCtx, err := a.authenticate(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		// Wrap stream with new context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// authenticate validates the JWT token and returns an enriched context.
func (a *AuthInterceptor) authenticate(ctx context.Context, method string) (context.Context, error) {
	// Extract token from metadata
	token, err := extractTokenFromMetadata(ctx)
	if err != nil {
		a.logger.Warn("missing authorization token",
			logger.String("method", method),
		)
		return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization token")
	}

	// Validate token
	claims, err := a.jwtService.Validate(ctx, token)
	if err != nil {
		a.logger.Warn("invalid JWT token",
			logger.String("method", method),
			logger.Err(err),
		)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	// Add claims to context
	ctx = a.addClaimsToContext(ctx, claims)

	return ctx, nil
}

// addClaimsToContext adds JWT claims to the context.
func (a *AuthInterceptor) addClaimsToContext(ctx context.Context, claims jwt.Claims) context.Context {
	// Store full claims
	ctx = WithClaims(ctx, claims)

	// Add individual claims for convenience
	if userID := claims.GetString("user_id"); userID != "" {
		ctx = WithUserID(ctx, userID)
	} else if subject := claims.Subject(); subject != "" {
		ctx = WithUserID(ctx, subject)
	}

	// Handle tenant ID based on tenancy configuration
	if a.opts.tenancyEnabled {
		if tenantID := claims.GetString("tenant_id"); tenantID != "" {
			ctx = WithTenantID(ctx, tenantID)
		}
	} else {
		ctx = WithTenantID(ctx, a.opts.defaultTenantID)
	}

	if sessionID := claims.GetString("session_id"); sessionID != "" {
		ctx = WithSessionID(ctx, sessionID)
	}

	if email := claims.GetString("email"); email != "" {
		ctx = WithEmail(ctx, email)
	}

	if role := claims.GetString("role"); role != "" {
		ctx = WithRole(ctx, role)
	}

	return ctx
}

// ============================================================================
// Metadata Extraction
// ============================================================================

const (
	// authorizationHeader is the metadata key for the authorization token.
	authorizationHeader = "authorization"

	// bearerPrefix is the prefix for bearer tokens.
	bearerPrefix = "bearer "
)

// extractTokenFromMetadata extracts the JWT token from gRPC metadata.
// Expected format: "authorization: Bearer <token>"
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	if !strings.HasPrefix(strings.ToLower(authHeader), bearerPrefix) {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// ============================================================================
// Wrapped Server Stream
// ============================================================================

// wrappedServerStream wraps a grpc.ServerStream with a custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// ============================================================================
// Functional API (Alternative)
// ============================================================================

// UnaryAuthInterceptor returns a unary interceptor with functional options.
func UnaryAuthInterceptor(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) grpc.UnaryServerInterceptor {
	interceptor := NewAuthInterceptor(jwtService, log, opts...)
	return interceptor.Unary()
}

// StreamAuthInterceptor returns a stream interceptor with functional options.
func StreamAuthInterceptor(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) grpc.StreamServerInterceptor {
	interceptor := NewAuthInterceptor(jwtService, log, opts...)
	return interceptor.Stream()
}

// ============================================================================
// Optional Auth Interceptor
// ============================================================================

// OptionalAuthInterceptor is like AuthInterceptor but doesn't fail if no token is present.
// Useful for endpoints that behave differently for authenticated vs anonymous users.
type OptionalAuthInterceptor struct {
	*AuthInterceptor
}

// NewOptionalAuthInterceptor creates a new optional auth interceptor.
func NewOptionalAuthInterceptor(jwtService jwt.Service, log logger.Logger, opts ...AuthOption) *OptionalAuthInterceptor {
	return &OptionalAuthInterceptor{
		AuthInterceptor: NewAuthInterceptor(jwtService, log, opts...),
	}
}

// Unary returns a unary server interceptor that optionally validates JWT tokens.
func (a *OptionalAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Try to authenticate, but don't fail if no token
		token, err := extractTokenFromMetadata(ctx)
		if err == nil && token != "" {
			claims, err := a.jwtService.Validate(ctx, token)
			if err == nil {
				ctx = a.addClaimsToContext(ctx, claims)
			}
		}

		return handler(ctx, req)
	}
}

// Stream returns a stream server interceptor that optionally validates JWT tokens.
func (a *OptionalAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Try to authenticate, but don't fail if no token
		token, err := extractTokenFromMetadata(ctx)
		if err == nil && token != "" {
			claims, err := a.jwtService.Validate(ctx, token)
			if err == nil {
				ctx = a.addClaimsToContext(ctx, claims)
				ss = &wrappedServerStream{ServerStream: ss, ctx: ctx}
			}
		}

		return handler(srv, ss)
	}
}
