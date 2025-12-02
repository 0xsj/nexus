// pkg/graphql/context.go

package graphql

import (
	"context"

	"github.com/0xsj/nexus-go/pkg/graphql/gqlctx"
	"github.com/0xsj/nexus-go/pkg/security/jwt"
)

// ============================================================================
// Re-export from gqlctx for convenience
// ============================================================================

// Context setters
var (
	WithRequestID     = gqlctx.WithRequestID
	WithClaims        = gqlctx.WithClaims
	WithUserID        = gqlctx.WithUserID
	WithTenantID      = gqlctx.WithTenantID
	WithSessionID     = gqlctx.WithSessionID
	WithEmail         = gqlctx.WithEmail
	WithRole          = gqlctx.WithRole
	WithOperationName = gqlctx.WithOperationName
	WithComplexity    = gqlctx.WithComplexity
)

// Context getters
var (
	GetRequestID     = gqlctx.GetRequestID
	GetClaims        = gqlctx.GetClaims
	GetUserID        = gqlctx.GetUserID
	GetTenantID      = gqlctx.GetTenantID
	GetSessionID     = gqlctx.GetSessionID
	GetEmail         = gqlctx.GetEmail
	GetRole          = gqlctx.GetRole
	GetOperationName = gqlctx.GetOperationName
	GetComplexity    = gqlctx.GetComplexity
	IsAuthenticated  = gqlctx.IsAuthenticated
	GetRequestInfo   = gqlctx.GetRequestInfo
)

// Types
type RequestInfo = gqlctx.RequestInfo

// ============================================================================
// Authentication Helpers (depend on errors.go)
// ============================================================================

// RequireAuth returns an error if the context is not authenticated.
func RequireAuth(ctx context.Context) error {
	if !gqlctx.IsAuthenticated(ctx) {
		return Unauthenticated("authentication required")
	}
	return nil
}

// RequireRole returns an error if the user doesn't have the required role.
func RequireRole(ctx context.Context, requiredRole string) error {
	if err := RequireAuth(ctx); err != nil {
		return err
	}

	role := gqlctx.GetRole(ctx)
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

	userRole := gqlctx.GetRole(ctx)
	for _, role := range roles {
		if userRole == role {
			return nil
		}
	}

	return Forbidden("insufficient permissions")
}

// ============================================================================
// Claims Helpers
// ============================================================================

// GetClaimsValue extracts a specific claim value from context.
func GetClaimsValue(ctx context.Context, key string) (string, bool) {
	claims, ok := gqlctx.GetClaims(ctx)
	if !ok {
		return "", false
	}
	val := claims.GetString(key)
	return val, val != ""
}

// MustGetUserID returns the user ID or panics.
// Use only when authentication is guaranteed.
func MustGetUserID(ctx context.Context) string {
	id := gqlctx.GetUserID(ctx)
	if id == "" {
		panic("user ID not found in context")
	}
	return id
}

// MustGetTenantID returns the tenant ID or panics.
// Use only when tenancy is guaranteed.
func MustGetTenantID(ctx context.Context) string {
	id := gqlctx.GetTenantID(ctx)
	if id == "" {
		panic("tenant ID not found in context")
	}
	return id
}

// Ensure jwt.Claims is available for type references
var _ jwt.Claims
