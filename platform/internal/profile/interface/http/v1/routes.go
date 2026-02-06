package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Profile HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/profiles", func(r chi.Router) {
		// Profile creation
		r.Post("/", h.CreateProfile)

		// Profile listing
		r.Get("/", h.ListProfiles)

		// Profile lookup by user
		r.Get("/user/{userId}", h.GetProfileByUser)

		// Profile lookup by vanity slug
		r.Get("/@{vanitySlug}", h.GetProfileByVanitySlug)

		// Single profile operations
		r.Route("/{profileId}", func(r chi.Router) {
			r.Get("/", h.GetProfile)
			r.Patch("/", h.UpdateProfile)

			// Badge management
			r.Post("/badges", h.AddBadge)
			r.Delete("/badges/{badgeId}", h.RemoveBadge)
			r.Patch("/badges/{badgeId}/visibility", h.ChangeBadgeVisibility)

			// Vanity URL
			r.Post("/vanity-url", h.ClaimVanityURL)
		})
	})
}
