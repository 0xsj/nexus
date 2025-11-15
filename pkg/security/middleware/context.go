package middleware

import (
	"context"

	"github.com/0xsj/nexus/pkg/security/token/jwt"
)

// Context keys for storing security information in context.
type contextKey string

const (
	claimsContextKey contextKey = "security:claims"
	userIDContextKey contextKey = "security:user_id"
)

// WithClaims adds JWT claims to the context.
func WithClaims(ctx context.Context, claims *jwt.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// ClaimsFromContext retrieves JWT claims from the context.
func ClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*jwt.Claims)
	return claims, ok
}

// MustClaimsFromContext retrieves JWT claims from context or panics.
// Use this only in handlers where authentication is guaranteed by middleware.
func MustClaimsFromContext(ctx context.Context) *jwt.Claims {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		panic("claims not found in context - ensure authentication middleware is applied")
	}
	return claims
}

// WithUserID adds user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// UserIDFromContext retrieves user ID from the context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

// MustUserIDFromContext retrieves user ID from context or panics.
// Use this only in handlers where authentication is guaranteed by middleware.
func MustUserIDFromContext(ctx context.Context) string {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		panic("user ID not found in context - ensure authentication middleware is applied")
	}
	return userID
}

// UserInfo represents authenticated user information extracted from context.
type UserInfo struct {
	UserID   string
	Email    string
	Username string
	Roles    []string
}

// UserInfoFromContext extracts user information from JWT claims in context.
func UserInfoFromContext(ctx context.Context) (*UserInfo, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return nil, false
	}

	return &UserInfo{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Roles:    claims.Roles,
	}, true
}

// MustUserInfoFromContext extracts user information or panics.
func MustUserInfoFromContext(ctx context.Context) *UserInfo {
	info, ok := UserInfoFromContext(ctx)
	if !ok {
		panic("user info not found in context - ensure authentication middleware is applied")
	}
	return info
}

// IsAuthenticated checks if the request is authenticated.
func IsAuthenticated(ctx context.Context) bool {
	_, ok := ClaimsFromContext(ctx)
	return ok
}

// HasRole checks if the authenticated user has a specific role.
func HasRole(ctx context.Context, role string) bool {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.HasRole(role)
}

// HasAnyRole checks if the authenticated user has any of the specified roles.
func HasAnyRole(ctx context.Context, roles ...string) bool {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.HasAnyRole(roles...)
}

// HasAllRoles checks if the authenticated user has all of the specified roles.
func HasAllRoles(ctx context.Context, roles ...string) bool {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.HasAllRoles(roles...)
}
