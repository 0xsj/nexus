package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Router
// ============================================================================

// RouterConfig contains configuration for the router.
type RouterConfig struct {
	CommandBus   cqrs.CommandBus
	QueryBus     cqrs.QueryBus
	TokenService domain.TokenService
	Logger       log.Logger
}

// NewRouter creates a new chi router with all identity routes.
func NewRouter(cfg RouterConfig) chi.Router {
	r := chi.NewRouter()

	// Create handler
	handler := NewHandler(cfg.CommandBus, cfg.QueryBus, cfg.Logger)

	// Create middleware
	authMiddleware := NewMiddleware(cfg.TokenService, cfg.Logger)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Auth endpoints - Wallet
		r.Post("/auth/challenge", handler.RequestChallenge)
		r.Post("/auth/register/wallet", handler.RegisterWithWallet)
		r.Post("/auth/login/wallet", handler.LoginWithWallet)
		r.Post("/auth/refresh", handler.RefreshToken)

		// Auth endpoints - Magic Link
		r.Post("/auth/magic-link/request", handler.RequestMagicLink)
		r.Post("/auth/magic-link/verify", handler.VerifyMagicLink)

		// Public profile
		r.Get("/profiles/{did}", handler.GetPublicProfile)

		// User existence check
		r.Get("/users/exists", handler.CheckUserExists)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthRequired)

		// Auth
		r.Post("/auth/logout", handler.Logout)

		// Current user
		r.Get("/users/me", handler.GetCurrentUser)
		r.Get("/users/me/stats", handler.GetUserStats)

		// User lookup
		r.Get("/users/{id}", handler.GetUser)
		r.Get("/users/did/{did}", handler.GetUserByDID)

		// Sessions
		r.Get("/sessions", handler.ListSessions)
		r.Get("/sessions/{id}", handler.GetSession)
		r.Delete("/sessions/{id}", handler.RevokeSession)
		r.Delete("/sessions", handler.RevokeAllSessions)

		// API Keys
		r.Get("/api-keys", handler.ListAPIKeys)
		r.Get("/api-keys/{id}", handler.GetAPIKey)
		r.Post("/api-keys", handler.CreateAPIKey)
		r.Delete("/api-keys/{id}", handler.RevokeAPIKey)
	})

	return r
}

// MountRouter mounts the identity router under a path prefix.
func MountRouter(parent chi.Router, prefix string, cfg RouterConfig) {
	parent.Mount(prefix, NewRouter(cfg))
}
