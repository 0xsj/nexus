// pkg/graphql/context.go

package graphql

import (
	"context"

	"github.com/0xsj/nexus-go/pkg/security/jwt"
)

// ============================================================================
// Context Keys
// ============================================================================

// Context keys for GraphQL request data.
type (
	requestIDKey  struct{}
	claimsKey     struct{}
	userIDKey     struct{}
	tenantIDKey   struct{}
	sessionIDKey  struct{}
	emailKey      struct{}
	roleKey       struct{}
	operationKey  struct{}
	complexityKey struct{}
)

// ============================================================================
// Context Setters
// ============================================================================

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// WithClaims adds JWT claims to the context.
func WithClaims(ctx context.Context, claims jwt.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// WithUserID adds a user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// WithTenantID adds a tenant ID to the context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

// WithSessionID adds a session ID to the context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, sessionID)
}

// WithEmail adds an email to the context.
func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailKey{}, email)
}

// WithRole adds a role to the context.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey{}, role)
}

// WithOperationName adds the GraphQL operation name to the context.
func WithOperationName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, operationKey{}, name)
}

// WithComplexity adds the query complexity to the context.
func WithComplexity(ctx context.Context, complexity int) context.Context {
	return context.WithValue(ctx, complexityKey{}, complexity)
}

// ============================================================================
// Context Getters
// ============================================================================

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetClaims extracts JWT claims from context.
func GetClaims(ctx context.Context) (jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(jwt.Claims)
	return claims, ok
}

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetTenantID extracts tenant ID from context.
func GetTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(tenantIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetSessionID extracts session ID from context.
func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(sessionIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GetEmail extracts email from context.
func GetEmail(ctx context.Context) string {
	if email, ok := ctx.Value(emailKey{}).(string); ok {
		return email
	}
	return ""
}

// GetRole extracts role from context.
func GetRole(ctx context.Context) string {
	if role, ok := ctx.Value(roleKey{}).(string); ok {
		return role
	}
	return ""
}

// GetOperationName extracts the GraphQL operation name from context.
func GetOperationName(ctx context.Context) string {
	if name, ok := ctx.Value(operationKey{}).(string); ok {
		return name
	}
	return ""
}

// GetComplexity extracts the query complexity from context.
func GetComplexity(ctx context.Context) int {
	if complexity, ok := ctx.Value(complexityKey{}).(int); ok {
		return complexity
	}
	return 0
}

// ============================================================================
// Authentication Helpers
// ============================================================================

// IsAuthenticated returns true if the context has valid authentication.
func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != ""
}

// RequireAuth returns an error if the context is not authenticated.
func RequireAuth(ctx context.Context) error {
	if !IsAuthenticated(ctx) {
		return Unauthenticated("authentication required")
	}
	return nil
}

// RequireRole returns an error if the user doesn't have the required role.
func RequireRole(ctx context.Context, requiredRole string) error {
	if err := RequireAuth(ctx); err != nil {
		return err
	}

	role := GetRole(ctx)
	if role != requiredRole {
		return Forbidden("insufficient permissions")
	}
	return nil
}

// RequireAnyRole returns an error if the user doesn't have any of the required roles.
func RequireAnyRole(ctx context.Context, roles ...string) error {
	if err := RequireAuth(ctx); err != nil {
		return err
	}

	userRole := GetRole(ctx)
	for _, role := range roles {
		if userRole == role {
			return nil
		}
	}

	return Forbidden("insufficient permissions")
}

// ============================================================================
// Request Info
// ============================================================================

// RequestInfo contains information about the current request.
type RequestInfo struct {
	RequestID     string
	UserID        string
	TenantID      string
	SessionID     string
	Email         string
	Role          string
	OperationName string
	Complexity    int
	Authenticated bool
}

// GetRequestInfo extracts all request information from context.
func GetRequestInfo(ctx context.Context) RequestInfo {
	return RequestInfo{
		RequestID:     GetRequestID(ctx),
		UserID:        GetUserID(ctx),
		TenantID:      GetTenantID(ctx),
		SessionID:     GetSessionID(ctx),
		Email:         GetEmail(ctx),
		Role:          GetRole(ctx),
		OperationName: GetOperationName(ctx),
		Complexity:    GetComplexity(ctx),
		Authenticated: IsAuthenticated(ctx),
	}
}
