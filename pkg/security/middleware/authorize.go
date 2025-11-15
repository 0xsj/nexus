package middleware

import (
	"net/http"
)

// AuthorizeConfig holds configuration for the authorization middleware.
type AuthorizeConfig struct {
	// ErrorHandler is a custom function to handle authorization errors
	// If nil, defaults to returning JSON error response
	ErrorHandler func(http.ResponseWriter, *http.Request, error)
}

// DefaultAuthorizeConfig returns configuration with sensible defaults.
func DefaultAuthorizeConfig() AuthorizeConfig {
	return AuthorizeConfig{
		ErrorHandler: defaultAuthzErrorHandler,
	}
}

// RequireRole creates middleware that requires the user to have a specific role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return RequireRoleWithConfig(role, DefaultAuthorizeConfig())
}

// RequireRoleWithConfig creates middleware with custom config.
func RequireRoleWithConfig(role string, config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context (should be set by Authenticate middleware)
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				config.ErrorHandler(w, r, ErrMissingUserInCtx)
				return
			}

			// Check if user has the required role
			if !claims.HasRole(role) {
				err := ErrMissingRole{
					Required: []string{role},
					Actual:   claims.Roles,
				}
				config.ErrorHandler(w, r, err)
				return
			}

			// User has required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole creates middleware that requires the user to have ANY of the specified roles.
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return RequireAnyRoleWithConfig(roles, DefaultAuthorizeConfig())
}

// RequireAnyRoleWithConfig creates middleware with custom config.
func RequireAnyRoleWithConfig(roles []string, config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				config.ErrorHandler(w, r, ErrMissingUserInCtx)
				return
			}

			// Check if user has any of the required roles
			if !claims.HasAnyRole(roles...) {
				err := ErrMissingRole{
					Required: roles,
					Actual:   claims.Roles,
				}
				config.ErrorHandler(w, r, err)
				return
			}

			// User has at least one required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllRoles creates middleware that requires the user to have ALL of the specified roles.
func RequireAllRoles(roles ...string) func(http.Handler) http.Handler {
	return RequireAllRolesWithConfig(roles, DefaultAuthorizeConfig())
}

// RequireAllRolesWithConfig creates middleware with custom config.
func RequireAllRolesWithConfig(roles []string, config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				config.ErrorHandler(w, r, ErrMissingUserInCtx)
				return
			}

			// Check if user has all required roles
			if !claims.HasAllRoles(roles...) {
				err := ErrMissingRole{
					Required: roles,
					Actual:   claims.Roles,
				}
				config.ErrorHandler(w, r, err)
				return
			}

			// User has all required roles, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthenticated creates middleware that simply requires authentication.
// Use this when you don't care about specific roles, just that the user is logged in.
func RequireAuthenticated() func(http.Handler) http.Handler {
	return RequireAuthenticatedWithConfig(DefaultAuthorizeConfig())
}

// RequireAuthenticatedWithConfig creates middleware with custom config.
func RequireAuthenticatedWithConfig(config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated
			if !IsAuthenticated(r.Context()) {
				config.ErrorHandler(w, r, ErrUnauthorized)
				return
			}

			// User is authenticated, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnership creates middleware that checks if the authenticated user
// is the owner of a resource. The ownerExtractor function should return the
// owner's user ID from the request (e.g., from path params).
func RequireOwnership(ownerExtractor func(*http.Request) string) func(http.Handler) http.Handler {
	return RequireOwnershipWithConfig(ownerExtractor, DefaultAuthorizeConfig())
}

// RequireOwnershipWithConfig creates middleware with custom config.
func RequireOwnershipWithConfig(ownerExtractor func(*http.Request) string, config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authenticated user ID
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				config.ErrorHandler(w, r, ErrMissingUserInCtx)
				return
			}

			// Get resource owner ID
			ownerID := ownerExtractor(r)
			if ownerID == "" {
				config.ErrorHandler(w, r, ErrForbidden)
				return
			}

			// Check if user is the owner
			if userID != ownerID {
				config.ErrorHandler(w, r, ErrForbidden)
				return
			}

			// User is the owner, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwnershipOrRole combines ownership check with role check.
// Allows access if user is either the owner OR has one of the specified roles (e.g., admin).
func RequireOwnershipOrRole(ownerExtractor func(*http.Request) string, roles ...string) func(http.Handler) http.Handler {
	return RequireOwnershipOrRoleWithConfig(ownerExtractor, roles, DefaultAuthorizeConfig())
}

// RequireOwnershipOrRoleWithConfig creates middleware with custom config.
func RequireOwnershipOrRoleWithConfig(ownerExtractor func(*http.Request) string, roles []string, config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				config.ErrorHandler(w, r, ErrMissingUserInCtx)
				return
			}

			// Check if user has any of the privileged roles
			if claims.HasAnyRole(roles...) {
				// User has privileged role, allow access
				next.ServeHTTP(w, r)
				return
			}

			// Check ownership
			ownerID := ownerExtractor(r)
			if ownerID == "" {
				config.ErrorHandler(w, r, ErrForbidden)
				return
			}

			if claims.UserID != ownerID {
				config.ErrorHandler(w, r, ErrForbidden)
				return
			}

			// User is the owner, continue
			next.ServeHTTP(w, r)
		})
	}
}

// CustomAuthorize creates middleware with a custom authorization function.
// The authorizer function should return true if access is allowed.
func CustomAuthorize(authorizer func(*http.Request) bool) func(http.Handler) http.Handler {
	return CustomAuthorizeWithConfig(authorizer, DefaultAuthorizeConfig())
}

// CustomAuthorizeWithConfig creates middleware with custom config.
func CustomAuthorizeWithConfig(authorizer func(*http.Request) bool, config AuthorizeConfig) func(http.Handler) http.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultAuthzErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authorizer(r) {
				config.ErrorHandler(w, r, ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// defaultAuthzErrorHandler returns a JSON error response for authorization errors.
func defaultAuthzErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusForbidden
	message := "forbidden"

	switch err {
	case ErrMissingUserInCtx, ErrUnauthorized:
		statusCode = http.StatusUnauthorized
		message = "unauthorized"
	case ErrForbidden:
		statusCode = http.StatusForbidden
		message = "forbidden: insufficient permissions"
	default:
		if _, ok := err.(ErrMissingRole); ok {
			statusCode = http.StatusForbidden
			message = "forbidden: insufficient role"
		} else {
			statusCode = http.StatusForbidden
			message = "forbidden"
		}
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
