package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Wallet HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/wallets", func(r chi.Router) {
		// Wallet linking
		r.Post("/", h.LinkWallet)

		// List endpoints
		r.Get("/user/{userId}", h.ListWalletsByUser)

		// Single wallet operations
		r.Route("/{walletId}", func(r chi.Router) {
			r.Get("/", h.GetWallet)
			r.Post("/verify", h.VerifyWallet)
			r.Post("/unlink", h.UnlinkWallet)
			r.Post("/primary", h.SetPrimaryWallet)
			r.Patch("/label", h.UpdateWalletLabel)
		})
	})
}
