package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/integration/app/command"
	"github.com/0xsj/nexus/platform/internal/integration/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Integration context.
type Handler struct {
	commands *command.Handlers
	queries  *query.Handlers
}

// NewHandler creates a new Handler.
func NewHandler(commands *command.Handlers, queries *query.Handlers) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

// ============================================================================
// Connect Provider
// ============================================================================

// ConnectProvider handles POST /api/v1/integrations/connect
func (h *Handler) ConnectProvider(w http.ResponseWriter, r *http.Request) {
	var req ConnectProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.UserID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("user_id", "user_id is required"))
		return
	}
	if req.ProviderType == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("provider_type", "provider_type is required"))
		return
	}
	if req.ProviderUserID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("provider_user_id", "provider_user_id is required"))
		return
	}

	// Build command
	cmd := command.ConnectProvider{
		UserID:           req.UserID,
		ProviderType:     req.ProviderType,
		ProviderUserID:   req.ProviderUserID,
		ProviderUsername: req.ProviderUsername,
		Scopes:           req.Scopes,
	}

	// Execute command
	result, err := h.commands.HandleConnectProvider(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.ConnectProviderResult)
	h.writeJSON(w, http.StatusCreated, ConnectProviderResponse{
		IntegrationID: data.IntegrationID,
		ProviderType:  data.ProviderType,
		Status:        data.Status,
		ConnectedAt:   time.Now().UTC(),
	})
}

// ============================================================================
// List Supported Providers
// ============================================================================

// ListSupportedProviders handles GET /api/v1/integrations/providers
func (h *Handler) ListSupportedProviders(w http.ResponseWriter, r *http.Request) {
	providers := SupportedProvidersResponse{
		Providers: []SupportedProviderResponse{
			{Type: "github", DisplayName: "GitHub", Description: "Connect your GitHub account to verify contributions and repositories"},
			{Type: "linkedin", DisplayName: "LinkedIn", Description: "Connect your LinkedIn account to verify professional experience"},
			{Type: "twitter", DisplayName: "Twitter/X", Description: "Connect your Twitter/X account to verify social presence"},
			{Type: "coursera", DisplayName: "Coursera", Description: "Connect your Coursera account to verify course completions"},
			{Type: "aws", DisplayName: "AWS", Description: "Connect your AWS account to verify cloud certifications"},
			{Type: "google", DisplayName: "Google", Description: "Connect your Google account to verify cloud certifications"},
		},
	}

	h.writeJSON(w, http.StatusOK, providers)
}

// ============================================================================
// Get Integration
// ============================================================================

// GetIntegration handles GET /api/v1/integrations/{integrationId}
func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "integrationId")
	if integrationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("integration_id is required"))
		return
	}

	id, err := types.ParseID(integrationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("integration_id", "invalid integration_id format"))
		return
	}

	result, err := h.queries.HandleGetIntegration(r.Context(), query.GetIntegration{
		IntegrationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toIntegrationResponse(result))
}

// ============================================================================
// List Integrations By User
// ============================================================================

// ListIntegrationsByUser handles GET /api/v1/integrations/user/{userId}
func (h *Handler) ListIntegrationsByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	result, err := h.queries.HandleListIntegrationsByUser(r.Context(), query.ListIntegrationsByUser{
		UserID: userID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toIntegrationListResponse(result))
}

// ============================================================================
// Disconnect Provider
// ============================================================================

// DisconnectProvider handles POST /api/v1/integrations/{integrationId}/disconnect
func (h *Handler) DisconnectProvider(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "integrationId")
	if integrationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("integration_id is required"))
		return
	}

	id, err := types.ParseID(integrationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("integration_id", "invalid integration_id format"))
		return
	}

	result, err := h.commands.HandleDisconnectProvider(r.Context(), command.DisconnectProvider{
		IntegrationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.DisconnectProviderResult)
	h.writeJSON(w, http.StatusOK, DisconnectProviderResponse{
		IntegrationID:  data.IntegrationID,
		Status:         data.Status,
		DisconnectedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Refresh Credentials
// ============================================================================

// RefreshCredentials handles POST /api/v1/integrations/{integrationId}/refresh
func (h *Handler) RefreshCredentials(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "integrationId")
	if integrationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("integration_id is required"))
		return
	}

	id, err := types.ParseID(integrationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("integration_id", "invalid integration_id format"))
		return
	}

	var req RefreshCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional for refresh; default to empty scopes
		req = RefreshCredentialsRequest{}
	}

	result, err := h.commands.HandleRefreshCredentials(r.Context(), command.RefreshCredentials{
		IntegrationID: id,
		Scopes:        req.Scopes,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.RefreshCredentialsResult)
	h.writeJSON(w, http.StatusOK, RefreshCredentialsResponse{
		IntegrationID: data.IntegrationID,
		Status:        data.Status,
		RefreshedAt:   time.Now().UTC(),
	})
}

// ============================================================================
// Suspend Integration
// ============================================================================

// SuspendIntegration handles POST /api/v1/integrations/{integrationId}/suspend
func (h *Handler) SuspendIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "integrationId")
	if integrationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("integration_id is required"))
		return
	}

	id, err := types.ParseID(integrationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("integration_id", "invalid integration_id format"))
		return
	}

	var req SuspendIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Reason == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("reason", "reason is required"))
		return
	}
	if len(req.Reason) > 500 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("reason", "reason must be 500 characters or less"))
		return
	}

	result, err := h.commands.HandleSuspendIntegration(r.Context(), command.SuspendIntegration{
		IntegrationID: id,
		Reason:        req.Reason,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.SuspendIntegrationResult)
	h.writeJSON(w, http.StatusOK, SuspendIntegrationResponse{
		IntegrationID: data.IntegrationID,
		Status:        data.Status,
		SuspendedAt:   time.Now().UTC(),
	})
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func toIntegrationResponse(v *query.IntegrationView) IntegrationResponse {
	return IntegrationResponse{
		IntegrationID:    v.IntegrationID,
		UserID:           v.UserID,
		ProviderType:     v.ProviderType,
		Status:           v.Status,
		ProviderUserID:   v.ProviderUserID,
		ProviderUsername: v.ProviderUsername,
		Scopes:           v.Scopes,
		LastFetchAt:      v.LastFetchAt,
		FetchCount:       v.FetchCount,
		Metadata:         v.Metadata,
		ConnectedAt:      v.ConnectedAt,
		DisconnectedAt:   v.DisconnectedAt,
		SuspendedAt:      v.SuspendedAt,
		SuspensionReason: v.SuspensionReason,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
}

func toIntegrationListResponse(v *query.IntegrationListView) IntegrationListResponse {
	summaries := make([]IntegrationSummaryResponse, len(v.Integrations))
	for i, integ := range v.Integrations {
		summaries[i] = IntegrationSummaryResponse{
			IntegrationID:    integ.IntegrationID,
			UserID:           integ.UserID,
			ProviderType:     integ.ProviderType,
			Status:           integ.Status,
			ProviderUsername: integ.ProviderUsername,
			ConnectedAt:      integ.ConnectedAt,
		}
	}

	return IntegrationListResponse{
		Integrations: summaries,
		TotalCount:   v.TotalCount,
	}
}

// ============================================================================
// Response Writing Helpers
// ============================================================================

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errResp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}
