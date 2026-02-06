package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Presentation HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/presentations", func(r chi.Router) {
		// Presentation creation
		r.Post("/", h.CreatePresentation)

		// List by holder
		r.Get("/holder/{holderDid}", h.ListPresentationsByHolder)

		// Single presentation operations
		r.Route("/{presentationId}", func(r chi.Router) {
			r.Get("/", h.GetPresentation)
			r.Post("/revoke", h.RevokePresentation)

			// Share links under a presentation
			r.Post("/share-links", h.CreateShareLink)
			r.Get("/share-links", h.ListShareLinks)
		})
	})

	r.Route("/api/v1/share-links", func(r chi.Router) {
		r.Route("/{shareLinkId}", func(r chi.Router) {
			r.Get("/", h.GetShareLink)
			r.Post("/access", h.AccessShareLink)
			r.Post("/revoke", h.RevokeShareLink)
			r.Get("/access-log", h.GetAccessLog)
		})
	})
}
