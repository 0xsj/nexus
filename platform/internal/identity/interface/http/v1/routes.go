package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all Identity HTTP routes.
func RegisterRoutes(r chi.Router, h *Handler, authMiddleware func(next http.Handler) http.Handler) {
	// Health check (public)
	r.Get("/health", h.Health)

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		// Magic link
		r.Post("/magic-link", h.SendMagicLink)
		r.Post("/magic-link/verify", h.VerifyMagicLink)

		// Wallet (SIWE)
		r.Post("/wallet/verify", h.VerifyWallet)

		// OAuth
		r.Get("/oauth/{provider}", h.InitiateOAuth)
		r.Post("/oauth/{provider}/callback", h.OAuthCallback)

		// Session management (authenticated)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/refresh", h.RefreshSession)
			r.Post("/revoke", h.RevokeSession)
			r.Post("/logout", h.Logout)
		})
	})

	// Current user routes (authenticated)
	r.Route("/me", func(r chi.Router) {
		r.Use(authMiddleware)

		// Profile
		r.Get("/", h.GetCurrentUser)
		r.Delete("/", h.DeleteAccount)
		r.Patch("/display-name", h.UpdateDisplayName)
		r.Patch("/email", h.UpdateEmail)

		// DIDs
		r.Get("/dids", h.GetUserDIDs)
		r.Post("/dids", h.AddDID)
		r.Delete("/dids/{did}", h.RemoveDID)

		// OAuth account linking
		r.Get("/oauth", h.GetLinkedOAuthAccounts)
		r.Get("/oauth/{provider}", h.InitiateLinkOAuth)
		r.Post("/oauth/{provider}/callback", h.LinkOAuthCallback)
		r.Delete("/oauth/{provider}", h.UnlinkOAuth)
	})

	// Sessions (authenticated)
	r.Route("/sessions", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", h.ListSessions)
	})

	// Users (public read)
	r.Route("/users", func(r chi.Router) {
		r.Get("/{userId}", h.GetUser)
		r.Get("/{userId}/profile", h.GetUserProfile)
	})

	// DID resolution (public)
	r.Get("/dids/{did}", h.ResolveDID)
}
