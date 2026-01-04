package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/identity/application/command"
	"github.com/0xsj/nexus/platform/internal/identity/application/query"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Session Handlers
// ============================================================================

// ListSessions returns all sessions for the authenticated user.
// GET /v1/sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	currentSessionID, _ := SessionIDFromContext(ctx)

	// Parse query params
	params := parseSessionListParams(r)

	result, err := dispatchSessionListQuery(ctx, h.queryBus, &query.ListUserSessions{
		UserID:         userID,
		CurrentSession: currentSessionID,
		ActiveOnly:     params.ActiveOnly,
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		h.logger.Error("failed to list sessions",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromSessionListView(result))
}

// GetSession returns a specific session.
// GET /v1/sessions/{id}
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteBadRequest(w, "session id is required")
		return
	}

	currentSessionID, _ := SessionIDFromContext(ctx)

	session, err := dispatchSessionQuery(ctx, h.queryBus, &query.GetSession{
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		h.logger.Error("failed to get session",
			log.Err(err),
			log.String("user_id", userID),
			log.String("session_id", sessionID),
		)
		WriteError(w, err)
		return
	}

	resp := FromSessionView(session)
	resp.IsCurrent = sessionID == currentSessionID

	httpresponse.JSON(w, http.StatusOK, resp)
}

// RevokeSession revokes a specific session.
// DELETE /v1/sessions/{id}
func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		WriteBadRequest(w, "session id is required")
		return
	}

	var req RevokeSessionRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, "invalid request body")
			return
		}
	}

	result, err := h.dispatchCommand(ctx, &command.RevokeSession{
		UserID:    userID,
		SessionID: sessionID,
		Reason:    req.Reason,
	})
	if err != nil {
		h.logger.Error("failed to revoke session",
			log.Err(err),
			log.String("user_id", userID),
			log.String("session_id", sessionID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.RevokeSessionResult)

	httpresponse.JSON(w, http.StatusOK, &RevokeSessionResponse{
		SessionID: data.SessionID,
		Revoked:   data.Revoked,
	})
}

// RevokeAllSessions revokes all sessions for the authenticated user.
// DELETE /v1/sessions
func (h *Handler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	currentSessionID, _ := SessionIDFromContext(ctx)

	var req RevokeAllSessionsRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, "invalid request body")
			return
		}
	}

	result, err := h.dispatchCommand(ctx, &command.RevokeAllSessions{
		UserID:         userID,
		ExceptCurrent:  req.ExceptCurrent,
		CurrentSession: currentSessionID,
		Reason:         req.Reason,
	})
	if err != nil {
		h.logger.Error("failed to revoke all sessions",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.RevokeAllSessionsResult)

	httpresponse.JSON(w, http.StatusOK, &RevokeAllSessionsResponse{
		RevokedCount: data.RevokedCount,
		SessionIDs:   data.SessionIDs,
	})
}

// ============================================================================
// Helpers
// ============================================================================

func parseSessionListParams(r *http.Request) SessionListParams {
	q := r.URL.Query()

	params := SessionListParams{}
	params.Limit, _ = strconv.Atoi(q.Get("limit"))
	params.Offset, _ = strconv.Atoi(q.Get("offset"))
	params.ActiveOnly = q.Get("active_only") == "true"

	params.WithDefaults()

	return params
}
