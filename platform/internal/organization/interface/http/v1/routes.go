package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Organization HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/organizations", func(r chi.Router) {
		// Organization CRUD
		r.Post("/", h.CreateOrganization)
		r.Get("/", h.ListOrganizations)

		// Slug lookup
		r.Get("/slug/{slug}", h.GetOrganizationBySlug)

		// Single organization operations
		r.Route("/{orgId}", func(r chi.Router) {
			r.Get("/", h.GetOrganization)
			r.Patch("/", h.UpdateOrganization)
			r.Delete("/", h.DeleteOrganization)

			// Member management
			r.Post("/members", h.AddMember)
			r.Delete("/members/{memberId}", h.RemoveMember)
			r.Patch("/members/{memberId}/role", h.ChangeMemberRole)

			// Ownership and verification
			r.Post("/transfer-ownership", h.TransferOwnership)
			r.Post("/verify", h.CompleteVerification)
		})
	})
}
