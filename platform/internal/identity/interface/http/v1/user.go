package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/identity/application/query"
	"github.com/0xsj/nexus/platform/internal/identity/domain"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// User Handlers
// ============================================================================

// GetCurrentUser returns the authenticated user's profile.
// GET /v1/users/me
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	user, err := dispatchUserQuery(ctx, h.queryBus, &query.GetUser{
		UserID: userID,
	})
	if err != nil {
		h.logger.Error("failed to get current user",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromUserView(user))
}

// GetUser returns a user by ID.
// GET /v1/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := chi.URLParam(r, "id")
	if userID == "" {
		WriteBadRequest(w, "user id is required")
		return
	}

	user, err := dispatchUserQuery(ctx, h.queryBus, &query.GetUser{
		UserID: userID,
	})
	if err != nil {
		h.logger.Error("failed to get user",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromUserView(user))
}

// GetUserStats returns statistics for the authenticated user.
// GET /v1/users/me/stats
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	stats, err := dispatchUserStatsQuery(ctx, h.queryBus, &query.GetUserStats{
		UserID: userID,
	})
	if err != nil {
		h.logger.Error("failed to get user stats",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromUserStatsView(stats))
}

// GetUserByDID returns a user by DID.
// GET /v1/users/did/{did}
func (h *Handler) GetUserByDID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	did := chi.URLParam(r, "did")
	if did == "" {
		WriteBadRequest(w, "did is required")
		return
	}

	user, err := dispatchUserQuery(ctx, h.queryBus, &query.GetUserByDID{
		DID: did,
	})
	if err != nil {
		h.logger.Error("failed to get user by DID",
			log.Err(err),
			log.String("did", did),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromUserView(user))
}

// GetPublicProfile returns the public profile for a DID.
// GET /v1/profiles/{did}
func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	did := chi.URLParam(r, "did")
	if did == "" {
		WriteBadRequest(w, "did is required")
		return
	}

	profile, err := dispatchPublicProfileQuery(ctx, h.queryBus, &query.GetPublicProfile{
		DID: did,
	})
	if err != nil {
		h.logger.Error("failed to get public profile",
			log.Err(err),
			log.String("did", did),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, profile)
}

// CheckUserExists checks if a user exists by various identifiers.
// GET /v1/users/exists
func (h *Handler) CheckUserExists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	q := r.URL.Query()

	checkQuery := &query.CheckUserExists{}

	if email := q.Get("email"); email != "" {
		checkQuery.Email = &email
	}
	if did := q.Get("did"); did != "" {
		checkQuery.DID = &did
	}
	if address := q.Get("address"); address != "" {
		chain := q.Get("chain")
		if chain == "" {
			chain = "ethereum"
		}
		checkQuery.Wallet = &query.WalletIdentifier{
			Address: address,
			Chain:   domain.Chain(chain),
		}
	}

	if checkQuery.Email == nil && checkQuery.DID == nil && checkQuery.Wallet == nil {
		WriteValidationError(w, "at least one identifier (email, did, address) is required", nil)
		return
	}

	exists, err := dispatchBoolQuery(ctx, h.queryBus, checkQuery)
	if err != nil {
		h.logger.Error("failed to check user exists", log.Err(err))
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, map[string]bool{"exists": exists})
}
