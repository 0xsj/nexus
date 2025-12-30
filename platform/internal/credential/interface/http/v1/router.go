package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/http/middleware"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Router
// ============================================================================

// Router encapsulates the credential HTTP routing.
type Router struct {
	handler *Handler
	logger  log.Logger
}

// NewRouter creates a new credential router.
func NewRouter(commandBus cqrs.CommandBus, queryBus cqrs.QueryBus, logger log.Logger) *Router {
	return &Router{
		handler: NewHandler(commandBus, queryBus, logger),
		logger:  logger,
	}
}

// Routes returns the credential routes as a mountable chi.Router.
func (rt *Router) Routes() chi.Router {
	r := chi.NewRouter()

	// Credential routes
	r.Route("/credentials", func(r chi.Router) {
		// List credentials
		r.Get("/", rt.handler.ListCredentials)

		// Issue a new credential
		r.Post("/", rt.handler.IssueCredential)

		// Request a credential (holder-initiated)
		r.Post("/request", rt.handler.RequestCredential)

		// Verify a credential (JWT-VC)
		r.Post("/verify", rt.handler.VerifyCredential)

		// Single credential operations
		r.Route("/{id}", func(r chi.Router) {
			// Get credential by ID
			r.Get("/", rt.handler.GetCredential)

			// Get signed verifiable credential (JWT-VC)
			r.Get("/vc", rt.handler.GetCredentialVC)

			// Lifecycle operations
			r.Post("/revoke", rt.handler.RevokeCredential)
			r.Post("/suspend", rt.handler.SuspendCredential)
			r.Post("/reinstate", rt.handler.ReinstateCredential)
		})
	})

	// Holder-centric routes
	r.Route("/holders/{holder_did}", func(r chi.Router) {
		r.Get("/credentials", rt.handler.ListCredentialsByHolder)
	})

	// Issuer-centric routes
	r.Route("/issuers/{issuer_did}", func(r chi.Router) {
		r.Get("/credentials", rt.handler.ListCredentialsByIssuer)
	})

	return r
}

// ============================================================================
// Router Options
// ============================================================================

// RouterOption configures the router.
type RouterOption func(*Router)

// WithMiddleware adds middleware to the router.
func WithMiddleware(mw ...middleware.Middleware) RouterOption {
	return func(rt *Router) {
		// Store middleware for later use in Routes()
		// This would require adding a middlewares field to Router
	}
}

// ============================================================================
// Factory
// ============================================================================

// Config holds configuration for creating a credential router.
type Config struct {
	CommandBus cqrs.CommandBus
	QueryBus   cqrs.QueryBus
	Logger     log.Logger
}

// NewRouterFromConfig creates a new router from configuration.
func NewRouterFromConfig(cfg Config) *Router {
	return NewRouter(cfg.CommandBus, cfg.QueryBus, cfg.Logger)
}
