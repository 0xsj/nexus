package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Integration HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/integrations", func(r chi.Router) {
		// Provider connection
		r.Post("/connect", h.ConnectProvider)

		// Supported providers (static list)
		r.Get("/providers", h.ListSupportedProviders)

		// User integrations
		r.Get("/user/{userId}", h.ListIntegrationsByUser)

		// Single integration operations
		r.Route("/{integrationId}", func(r chi.Router) {
			r.Get("/", h.GetIntegration)
			r.Post("/disconnect", h.DisconnectProvider)
			r.Post("/refresh", h.RefreshCredentials)
			r.Post("/suspend", h.SuspendIntegration)
		})
	})
}
