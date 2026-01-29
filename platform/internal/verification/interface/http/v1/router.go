package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/id"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Router
// ============================================================================

// Router encapsulates the verification HTTP routing.
type Router struct {
	handler        *Handler
	authMiddleware func(http.Handler) http.Handler
	logger         log.Logger
}

// NewRouter creates a new verification router.
func NewRouter(
	commandBus cqrs.CommandBus,
	queryBus cqrs.QueryBus,
	idGenerator id.Generator,
	logger log.Logger,
) *Router {
	return &Router{
		handler: NewHandler(commandBus, queryBus, idGenerator, logger),
		logger:  logger,
	}
}

// WithAuthMiddleware sets the auth middleware.
func (rt *Router) WithAuthMiddleware(mw func(http.Handler) http.Handler) *Router {
	rt.authMiddleware = mw
	return rt
}

// Routes returns the verification routes as a mountable chi.Router.
func (rt *Router) Routes() chi.Router {
	r := chi.NewRouter()

	// OAuth callback (public - no auth required)
	r.Get("/callback", rt.handler.HandleOAuthCallback)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		// Apply auth middleware if set
		if rt.authMiddleware != nil {
			r.Use(rt.authMiddleware)
		}

		// List user's verifications
		r.Get("/", rt.handler.ListVerifications)

		// Initiate a new verification
		r.Post("/", rt.handler.InitiateVerification)

		// Provider connections
		r.Get("/connections", rt.handler.GetProviderConnections)
		r.Get("/connections/{provider}", rt.handler.CheckProviderConnected)

		// Provider-specific convenience endpoints
		r.Post("/github", rt.handler.InitiateGitHubVerification)
		r.Post("/linkedin", rt.handler.InitiateLinkedInVerification)

		// Single verification operations
		r.Route("/{id}", func(r chi.Router) {
			// Get verification details
			r.Get("/", rt.handler.GetVerification)

			// Cancel verification
			r.Post("/cancel", rt.handler.CancelVerification)
		})
	})

	return r
}

// ============================================================================
// Router Configuration
// ============================================================================

// Config holds configuration for creating a verification router.
type Config struct {
	CommandBus     cqrs.CommandBus
	QueryBus       cqrs.QueryBus
	IDGenerator    id.Generator
	AuthMiddleware func(http.Handler) http.Handler
	Logger         log.Logger
}

// NewRouterFromConfig creates a new router from configuration.
func NewRouterFromConfig(cfg Config) *Router {
	rt := NewRouter(cfg.CommandBus, cfg.QueryBus, cfg.IDGenerator, cfg.Logger)
	if cfg.AuthMiddleware != nil {
		rt.WithAuthMiddleware(cfg.AuthMiddleware)
	}
	return rt
}
