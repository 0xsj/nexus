package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Schema HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/schemas", func(r chi.Router) {
		// Schema CRUD
		r.Post("/", h.RegisterSchema)
		r.Get("/", h.ListSchemas)

		// Lookup endpoints
		r.Get("/exists", h.CheckSchemaExists)
		r.Get("/resolve/{schemaType}", h.ResolveSchemaType)
		r.Get("/type/{schemaType}", h.GetSchemaByType)

		// Single schema operations
		r.Route("/{schemaId}", func(r chi.Router) {
			r.Get("/", h.GetSchema)
			r.Patch("/", h.UpdateSchemaMetadata)

			// Status management
			r.Post("/deprecate", h.DeprecateSchema)
			r.Post("/activate", h.ActivateSchema)

			// Version management
			r.Route("/versions", func(r chi.Router) {
				r.Post("/", h.AddSchemaVersion)
				r.Get("/", h.ListSchemaVersions)
				r.Get("/{version}", h.GetSchemaVersion)
			})
		})
	})
}
