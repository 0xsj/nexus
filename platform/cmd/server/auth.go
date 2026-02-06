package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// newAuthMiddleware creates a Chi middleware that validates session tokens.
// It extracts the Bearer token from the Authorization header, hashes it,
// looks up the session, and injects "user_id" and "session_id" into the
// request context if valid. Returns 401 for invalid/expired sessions.
func newAuthMiddleware(sessionLookup domain.SessionLookup, logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			token, err := domain.ParseToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid token format"}`, http.StatusUnauthorized)
				return
			}

			session, err := sessionLookup.GetSessionByTokenHash(r.Context(), token.HashString())
			if err != nil {
				logger.Debug("session lookup failed", log.String("error", err.Error()))
				http.Error(w, `{"error":"invalid session"}`, http.StatusUnauthorized)
				return
			}

			if !session.IsActive() {
				http.Error(w, `{"error":"session expired or revoked"}`, http.StatusUnauthorized)
				return
			}

			// Inject user_id and session_id into context.
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", session.UserID().String())
			ctx = context.WithValue(ctx, "session_id", session.ID().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken extracts the token from the "Authorization: Bearer <token>" header.
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
