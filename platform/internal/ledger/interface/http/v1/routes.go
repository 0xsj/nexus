package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Ledger HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/ledger", func(r chi.Router) {
		// Entry endpoints
		r.Get("/entries/{id}", h.GetEntry)

		// Activity endpoints
		r.Get("/activity", h.GetActivity)
		r.Get("/stats", h.GetActivityStats)

		// User-specific endpoints
		r.Get("/users/{userId}/activity", h.GetUserActivity)
		r.Get("/users/{userId}/verifications", h.GetVerificationLog)

		// Credential history endpoint
		r.Get("/credentials/{credentialId}/history", h.GetCredentialHistory)

		// Generic subject history endpoint
		r.Get("/subjects/{subjectType}/{subjectId}/history", h.GetSubjectHistory)
	})
}
