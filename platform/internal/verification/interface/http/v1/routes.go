package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Verification HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/verifications", func(r chi.Router) {
		// Start a new verification
		r.Post("/", h.StartVerification)

		// List verifications by user
		r.Get("/user/{userId}", h.ListByUser)

		// Single verification operations
		r.Route("/{verificationId}", func(r chi.Router) {
			r.Get("/", h.GetVerification)
			r.Post("/callback", h.OAuthCallback)
		})
	})
}
