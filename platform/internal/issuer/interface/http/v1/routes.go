package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Issuer HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/issuers", func(r chi.Router) {
		// Issuer registration
		r.Post("/", h.RegisterIssuer)

		// Issuer listing
		r.Get("/", h.ListIssuers)

		// Issuer by organization
		r.Get("/organization/{orgId}", h.GetIssuerByOrganization)

		// Single issuer operations
		r.Route("/{issuerId}", func(r chi.Router) {
			r.Get("/", h.GetIssuer)
			r.Post("/activate", h.ActivateIssuer)
			r.Post("/suspend", h.SuspendIssuer)

			// Template operations
			r.Post("/templates", h.CreateTemplate)
			r.Get("/templates", h.ListTemplatesByIssuer)

			r.Route("/templates/{templateId}", func(r chi.Router) {
				r.Get("/", h.GetTemplate)
				r.Put("/", h.UpdateTemplate)
				r.Post("/archive", h.ArchiveTemplate)
			})
		})
	})
}
