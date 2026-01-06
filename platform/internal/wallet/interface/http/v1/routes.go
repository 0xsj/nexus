package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Router
// ============================================================================

// RouterConfig contains configuration for the router.
type RouterConfig struct {
	CommandBus       cqrs.CommandBus
	QueryBus         cqrs.QueryBus
	ChallengeService domain.ChallengeService
	AuthMiddleware   func(next http.Handler) http.Handler
	Logger           log.Logger
}

// NewRouter creates a new chi router with all wallet routes.
func NewRouter(cfg RouterConfig) chi.Router {
	r := chi.NewRouter()

	// Create handler
	handler := NewHandler(cfg.CommandBus, cfg.QueryBus, cfg.ChallengeService, cfg.Logger)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		// Challenge endpoints
		r.Post("/wallet/challenge", handler.CreateChallenge)
		r.Post("/wallet/challenge/verify", handler.VerifyChallenge)

		// Public verification
		r.Post("/wallet/verify", handler.VerifySignature)

		// Wallet existence check
		r.Get("/wallet/exists", handler.CheckWalletExists)

		// Supported chains
		r.Get("/wallet/chains", handler.GetSupportedChains)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		if cfg.AuthMiddleware != nil {
			r.Use(cfg.AuthMiddleware)
		}

		// Wallet list and stats
		r.Get("/wallets", handler.ListWallets)
		r.Get("/wallets/stats", handler.GetWalletStats)
		r.Get("/wallets/primary", handler.GetPrimaryWallet)

		// Link wallet
		r.Post("/wallets/link", handler.LinkWallet)

		// Wallet by address lookup
		r.Get("/wallets/address/{address}", handler.GetWalletByAddress)

		// Individual wallet operations
		r.Get("/wallets/{id}", handler.GetWallet)
		r.Patch("/wallets/{id}", handler.UpdateWallet)
		r.Delete("/wallets/{id}", handler.DeleteWallet)

		// Primary wallet
		r.Patch("/wallets/{id}/primary", handler.SetPrimaryWallet)

		// Status changes
		r.Post("/wallets/{id}/activate", handler.ActivateWallet)
		r.Post("/wallets/{id}/deactivate", handler.DeactivateWallet)

		// Re-verification
		r.Post("/wallets/{id}/verify", handler.ReverifyWallet)
	})

	return r
}

// MountRouter mounts the wallet router under a path prefix.
func MountRouter(parent chi.Router, prefix string, cfg RouterConfig) {
	parent.Mount(prefix, NewRouter(cfg))
}
