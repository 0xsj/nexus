package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Credential HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/credentials", func(r chi.Router) {
		// Credential issuance
		r.Post("/", h.IssueCredential)

		// List endpoints
		r.Get("/subject/{subjectDid}", h.ListCredentialsBySubject)
		r.Get("/issuer/{issuerDid}", h.ListCredentialsByIssuer)

		// Single credential operations
		r.Route("/{credentialId}", func(r chi.Router) {
			r.Get("/", h.GetCredential)
			r.Post("/revoke", h.RevokeCredential)
			r.Post("/expire", h.ExpireCredential)
		})
	})
}
