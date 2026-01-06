package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Router Configuration
// ============================================================================

// RouterConfig contains configuration for the router.
type RouterConfig struct {
	CommandBus       cqrs.CommandBus
	QueryBus         cqrs.QueryBus
	ChallengeService domain.ChallengeService
	AuthMiddleware   func(http.Handler) http.Handler
	Logger           log.Logger
}

// ============================================================================
// Routers
// ============================================================================

// NewPublicRouter creates routes for public wallet endpoints (no auth required).
// Mount at: /api/v1/wallet
func NewPublicRouter(cfg RouterConfig) chi.Router {
	r := chi.NewRouter()

	handler := NewHandler(cfg.CommandBus, cfg.QueryBus, cfg.ChallengeService, cfg.Logger)

	// Challenge endpoints
	r.Post("/challenge", handler.CreateChallenge)
	r.Post("/challenge/verify", handler.VerifyChallenge)

	// Public verification
	r.Post("/verify", handler.VerifySignature)

	// Wallet existence check
	r.Get("/exists", handler.CheckWalletExists)

	// Supported chains
	r.Get("/chains", handler.GetSupportedChains)

	return r
}

// NewProtectedRouter creates routes for protected wallet endpoints (auth required).
// Mount at: /api/v1/wallets
func NewProtectedRouter(cfg RouterConfig) chi.Router {
	r := chi.NewRouter()

	handler := NewHandler(cfg.CommandBus, cfg.QueryBus, cfg.ChallengeService, cfg.Logger)

	// Apply auth middleware
	if cfg.AuthMiddleware != nil {
		r.Use(cfg.AuthMiddleware)
	}

	// Wallet list and stats
	r.Get("/", handler.ListWallets)
	r.Get("/stats", handler.GetWalletStats)
	r.Get("/primary", handler.GetPrimaryWallet)

	// Link wallet
	r.Post("/link", handler.LinkWallet)

	// Wallet by address lookup
	r.Get("/address/{address}", handler.GetWalletByAddress)

	// Individual wallet operations
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", handler.GetWallet)
		r.Patch("/", handler.UpdateWallet)
		r.Delete("/", handler.DeleteWallet)

		// Primary wallet
		r.Patch("/primary", handler.SetPrimaryWallet)

		// Status changes
		r.Post("/activate", handler.ActivateWallet)
		r.Post("/deactivate", handler.DeactivateWallet)

		// Re-verification
		r.Post("/verify", handler.ReverifyWallet)
	})

	return r
}

// ============================================================================
// Module Routes Container
// ============================================================================

// Routes contains all wallet routers for mounting.
type Routes struct {
	Public    chi.Router // Mount at /api/v1/wallet
	Protected chi.Router // Mount at /api/v1/wallets
}

// NewRoutes creates all wallet routes.
func NewRoutes(cfg RouterConfig) *Routes {
	return &Routes{
		Public:    NewPublicRouter(cfg),
		Protected: NewProtectedRouter(cfg),
	}
}
