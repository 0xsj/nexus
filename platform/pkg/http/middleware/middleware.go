package middleware

import (
	"net/http"
)

// ============================================================================
// Middleware Type
// ============================================================================

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// ============================================================================
// Chain
// ============================================================================

// Chain creates a middleware chain that executes in order.
// The first middleware is the outermost (executes first on request, last on response).
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// ============================================================================
// Handler Helpers
// ============================================================================

// Then applies a middleware chain to a handler.
func Then(h http.Handler, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(h)
}

// ThenFunc applies a middleware chain to a handler function.
func ThenFunc(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(h)
}

// ============================================================================
// Adapt
// ============================================================================

// Adapt converts a function that takes specific arguments into a Middleware.
// Useful for adapting third-party middleware.
func Adapt(fn func(http.Handler) http.Handler) Middleware {
	return Middleware(fn)
}

// AdaptFunc converts a handler function to an http.Handler.
func AdaptFunc(fn http.HandlerFunc) http.Handler {
	return fn
}

// ============================================================================
// Conditional Middleware
// ============================================================================

// If applies middleware only if the condition is true.
func If(condition bool, m Middleware) Middleware {
	if condition {
		return m
	}
	return func(next http.Handler) http.Handler {
		return next
	}
}

// IfFunc applies middleware only if the condition function returns true.
// The condition is evaluated per-request.
func IfFunc(condition func(*http.Request) bool, m Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if condition(r) {
				m(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// Unless applies middleware only if the condition is false.
func Unless(condition bool, m Middleware) Middleware {
	return If(!condition, m)
}

// ============================================================================
// Skip Middleware
// ============================================================================

// SkipPaths returns a middleware that skips the given middleware for specific paths.
func SkipPaths(m Middleware, paths ...string) Middleware {
	skipSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		skipSet[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipSet[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			m(next).ServeHTTP(w, r)
		})
	}
}

// SkipPathPrefixes returns a middleware that skips for paths with given prefixes.
func SkipPathPrefixes(m Middleware, prefixes ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, prefix := range prefixes {
				if len(r.URL.Path) >= len(prefix) && r.URL.Path[:len(prefix)] == prefix {
					next.ServeHTTP(w, r)
					return
				}
			}
			m(next).ServeHTTP(w, r)
		})
	}
}

// OnlyPaths returns a middleware that only applies for specific paths.
func OnlyPaths(m Middleware, paths ...string) Middleware {
	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pathSet[r.URL.Path] {
				m(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// No-Op Middleware
// ============================================================================

// Noop is a middleware that does nothing.
func Noop() Middleware {
	return func(next http.Handler) http.Handler {
		return next
	}
}
