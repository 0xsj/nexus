package v1

import (
	"github.com/go-chi/chi/v5"
)

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Notification HTTP routes on a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/notifications", func(r chi.Router) {
		// Create notification
		r.Post("/", h.CreateNotification)

		// Inbox endpoint
		r.Get("/inbox/{recipientId}", h.ListInbox)

		// Preferences endpoints
		r.Get("/preferences/{userId}", h.GetPreferences)
		r.Put("/preferences/{userId}", h.UpdatePreferences)

		// Single notification operations
		r.Route("/{notificationId}", func(r chi.Router) {
			r.Get("/", h.GetNotification)
			r.Post("/read", h.MarkRead)
		})
	})
}
