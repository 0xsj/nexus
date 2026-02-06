package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Trust HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/trust", func(r chi.Router) {
		// Vouch operations
		r.Route("/vouches", func(r chi.Router) {
			// Create a vouch
			r.Post("/", h.GiveVouch)

			// List endpoints
			r.Get("/received/{userId}", h.ListVouchesByVouchee)
			r.Get("/given/{userId}", h.ListVouchesByVoucher)

			// Single vouch operations
			r.Route("/{vouchId}", func(r chi.Router) {
				r.Get("/", h.GetVouch)
				r.Post("/accept", h.AcceptVouch)
				r.Post("/revoke", h.RevokeVouch)
			})
		})

		// Reputation
		r.Get("/reputation/{userId}", h.GetReputation)
	})
}
