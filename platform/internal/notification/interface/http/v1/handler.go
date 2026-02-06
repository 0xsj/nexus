package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/notification/app/command"
	"github.com/0xsj/nexus/platform/internal/notification/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Notification context.
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
// Create Notification
// ============================================================================

// CreateNotification handles POST /api/v1/notifications
func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.RecipientID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("recipient_id", "recipient_id is required"))
		return
	}
	if req.Category == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("category", "category is required"))
		return
	}
	if req.Channel == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("channel", "channel is required"))
		return
	}
	if req.Subject == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("subject", "subject is required"))
		return
	}
	if req.Body == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("body", "body is required"))
		return
	}

	// Build command
	cmd := command.CreateNotification{
		RecipientID: req.RecipientID,
		Category:    req.Category,
		Channel:     req.Channel,
		TemplateID:  req.TemplateID,
		Subject:     req.Subject,
		Body:        req.Body,
		ActionURL:   req.ActionURL,
	}

	// Execute command
	result, err := h.commands.HandleCreateNotification(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.CreateNotificationResult)
	h.writeJSON(w, http.StatusCreated, NotificationCreatedResponse{
		NotificationID: data.NotificationID,
		Status:         data.Status,
		CreatedAt:      time.Now().UTC(),
	})
}

// ============================================================================
// Get Notification
// ============================================================================

// GetNotification handles GET /api/v1/notifications/{notificationId}
func (h *Handler) GetNotification(w http.ResponseWriter, r *http.Request) {
	notificationID := chi.URLParam(r, "notificationId")
	if notificationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("notification_id is required"))
		return
	}

	id, err := types.ParseID(notificationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("notification_id", "invalid notification_id format"))
		return
	}

	result, err := h.queries.HandleGetNotification(r.Context(), query.GetNotification{
		NotificationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toNotificationResponse(result))
}

// ============================================================================
// List Inbox
// ============================================================================

// ListInbox handles GET /api/v1/notifications/inbox/{recipientId}
func (h *Handler) ListInbox(w http.ResponseWriter, r *http.Request) {
	recipientID := chi.URLParam(r, "recipientId")
	if recipientID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("recipient_id is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListNotificationsByRecipient(r.Context(), query.ListNotificationsByRecipient{
		RecipientID: recipientID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toNotificationListResponse(result))
}

// ============================================================================
// Mark Read
// ============================================================================

// MarkRead handles POST /api/v1/notifications/{notificationId}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	notificationID := chi.URLParam(r, "notificationId")
	if notificationID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("notification_id is required"))
		return
	}

	id, err := types.ParseID(notificationID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("notification_id", "invalid notification_id format"))
		return
	}

	result, err := h.commands.HandleMarkNotificationRead(r.Context(), command.MarkNotificationRead{
		NotificationID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.MarkNotificationReadResult)
	h.writeJSON(w, http.StatusOK, NotificationReadResponse{
		NotificationID: data.NotificationID,
		Status:         data.Status,
		ReadAt:         time.Now().UTC(),
	})
}

// ============================================================================
// Get Preferences
// ============================================================================

// GetPreferences handles GET /api/v1/notifications/preferences/{userId}
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	result, err := h.queries.HandleGetPreferences(r.Context(), query.GetPreferences{
		UserID: userID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toPreferencesResponse(result))
}

// ============================================================================
// Update Preferences
// ============================================================================

// UpdatePreferences handles PUT /api/v1/notifications/preferences/{userId}
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Build command
	cmd := command.UpdatePreferences{
		UserID:          userID,
		GlobalEnabled:   req.GlobalEnabled,
		ChannelUpdates:  req.ChannelUpdates,
		CategoryUpdates: req.CategoryUpdates,
		DigestEnabled:   req.DigestEnabled,
		DigestFrequency: req.DigestFrequency,
		Timezone:        req.Timezone,
	}

	// Execute command
	_, err := h.commands.HandleUpdatePreferences(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Return updated preferences
	result, err := h.queries.HandleGetPreferences(r.Context(), query.GetPreferences{
		UserID: userID,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toPreferencesResponse(result))
}

// ============================================================================
// Request Parsing Helpers
// ============================================================================

func (h *Handler) parsePagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func toNotificationResponse(v *query.NotificationView) NotificationResponse {
	return NotificationResponse{
		NotificationID: v.NotificationID,
		RecipientID:    v.RecipientID,
		Category:       v.Category,
		Channel:        v.Channel,
		TemplateID:     v.TemplateID,
		Subject:        v.Subject,
		Body:           v.Body,
		ActionURL:      v.ActionURL,
		Status:         v.Status,
		ErrorMessage:   v.ErrorMessage,
		CreatedAt:      v.CreatedAt,
		SentAt:         v.SentAt,
		ReadAt:         v.ReadAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func toNotificationListResponse(v *query.NotificationListView) NotificationListResponse {
	summaries := make([]NotificationSummaryResponse, len(v.Notifications))
	for i, n := range v.Notifications {
		summaries[i] = NotificationSummaryResponse{
			NotificationID: n.NotificationID,
			Category:       n.Category,
			Channel:        n.Channel,
			Subject:        n.Subject,
			Status:         n.Status,
			CreatedAt:      n.CreatedAt,
			ReadAt:         n.ReadAt,
		}
	}

	return NotificationListResponse{
		Notifications: summaries,
		TotalCount:    v.TotalCount,
		UnreadCount:   v.UnreadCount,
		Limit:         v.Limit,
		Offset:        v.Offset,
		HasMore:       v.HasMore,
	}
}

func toPreferencesResponse(v *query.PreferencesView) PreferencesResponse {
	return PreferencesResponse{
		UserID:           v.UserID,
		GlobalEnabled:    v.GlobalEnabled,
		ChannelEnabled:   v.ChannelEnabled,
		CategoryChannels: v.CategoryChannels,
		DigestEnabled:    v.DigestEnabled,
		DigestFrequency:  v.DigestFrequency,
		Timezone:         v.Timezone,
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
