package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/identity/application/command"
	"github.com/0xsj/nexus/platform/internal/identity/application/query"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// API Key Handlers
// ============================================================================

// ListAPIKeys returns all API keys for the authenticated user.
// GET /v1/api-keys
func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	params := parseAPIKeyListParams(r)

	result, err := dispatchAPIKeyListQuery(ctx, h.queryBus, &query.ListUserAPIKeys{
		UserID:     userID,
		ActiveOnly: params.ActiveOnly,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		h.logger.Error("failed to list api keys",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromAPIKeyListView(result))
}

// GetAPIKey returns a specific API key.
// GET /v1/api-keys/{id}
func (h *Handler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		WriteBadRequest(w, "api key id is required")
		return
	}

	apiKey, err := dispatchAPIKeyQuery(ctx, h.queryBus, &query.GetAPIKey{
		UserID: userID,
		KeyID:  keyID,
	})
	if err != nil {
		h.logger.Error("failed to get api key",
			log.Err(err),
			log.String("user_id", userID),
			log.String("key_id", keyID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromAPIKeyView(apiKey))
}

// CreateAPIKey creates a new API key.
// POST /v1/api-keys
func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := validateCreateAPIKeyRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	var expiresIn time.Duration
	if req.ExpiresIn != nil {
		expiresIn = time.Duration(*req.ExpiresIn) * time.Second
	}

	result, err := h.dispatchCommand(ctx, &command.CreateAPIKey{
		UserID:      userID,
		Name:        req.Name,
		Scopes:      req.ToScopes(),
		ExpiresIn:   expiresIn,
		Description: req.Description,
	})
	if err != nil {
		h.logger.Error("failed to create api key",
			log.Err(err),
			log.String("user_id", userID),
			log.String("name", req.Name),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.CreateAPIKeyResult)

	httpresponse.JSON(w, http.StatusCreated, &APIKeyCreatedResponse{
		ID:        data.KeyID,
		Name:      data.Name,
		Key:       data.RawKey,
		Prefix:    data.Prefix,
		Scopes:    data.Scopes,
		ExpiresAt: data.ExpiresAt,
		CreatedAt: data.CreatedAt,
	})
}

// RevokeAPIKey revokes an API key.
// DELETE /v1/api-keys/{id}
func (h *Handler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		WriteBadRequest(w, "api key id is required")
		return
	}

	var req RevokeAPIKeyRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, "invalid request body")
			return
		}
	}

	result, err := h.dispatchCommand(ctx, &command.RevokeAPIKey{
		UserID: userID,
		KeyID:  keyID,
		Reason: req.Reason,
	})
	if err != nil {
		h.logger.Error("failed to revoke api key",
			log.Err(err),
			log.String("user_id", userID),
			log.String("key_id", keyID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.RevokeAPIKeyResult)

	httpresponse.JSON(w, http.StatusOK, &RevokeAPIKeyResponse{
		KeyID:     data.KeyID,
		Revoked:   data.Revoked,
		RevokedAt: data.RevokedAt,
	})
}

// ============================================================================
// Helpers
// ============================================================================

func parseAPIKeyListParams(r *http.Request) APIKeyListParams {
	q := r.URL.Query()

	params := APIKeyListParams{}
	params.Limit, _ = strconv.Atoi(q.Get("limit"))
	params.Offset, _ = strconv.Atoi(q.Get("offset"))
	params.ActiveOnly = q.Get("active_only") == "true"

	params.WithDefaults()

	return params
}

func validateCreateAPIKeyRequest(req *CreateAPIKeyRequest) error {
	if req.Name == "" {
		return validationError("name is required")
	}
	if len(req.Name) > 100 {
		return validationError("name must be 100 characters or less")
	}
	if len(req.Scopes) == 0 {
		return validationError("at least one scope is required")
	}
	if len(req.Description) > 500 {
		return validationError("description must be 500 characters or less")
	}
	return nil
}
