package query

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all query handlers for the Notification context.
type Handlers struct {
	lookup          *postgres.NotificationLookup
	preferencesRepo domain.PreferencesRepository
	logger          log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	lookup *postgres.NotificationLookup,
	preferencesRepo domain.PreferencesRepository,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		lookup:          lookup,
		preferencesRepo: preferencesRepo,
		logger:          logger,
	}
}

// ============================================================================
// GetNotification Handler
// ============================================================================

// HandleGetNotification handles the GetNotification query.
func (h *Handlers) HandleGetNotification(ctx context.Context, q GetNotification) (*NotificationView, error) {
	const op = "Handlers.HandleGetNotification"

	proj, err := h.lookup.GetByID(ctx, q.NotificationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapProjectionToView(proj), nil
}

// ============================================================================
// ListNotificationsByRecipient Handler
// ============================================================================

// HandleListNotificationsByRecipient handles the ListNotificationsByRecipient query.
func (h *Handlers) HandleListNotificationsByRecipient(ctx context.Context, q ListNotificationsByRecipient) (*NotificationListView, error) {
	const op = "Handlers.HandleListNotificationsByRecipient"

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}

	notifications, err := h.lookup.ListByRecipient(ctx, q.RecipientID, limit, q.Offset)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	totalCount, err := h.lookup.CountByRecipient(ctx, q.RecipientID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	unreadCount, err := h.lookup.CountUnreadByRecipient(ctx, q.RecipientID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	summaries := make([]NotificationSummaryView, len(notifications))
	for i, notif := range notifications {
		summaries[i] = mapProjectionToSummary(&notif)
	}

	return &NotificationListView{
		Notifications: summaries,
		TotalCount:    totalCount,
		UnreadCount:   unreadCount,
		Limit:         limit,
		Offset:        q.Offset,
		HasMore:       q.Offset+limit < totalCount,
	}, nil
}

// ============================================================================
// GetPreferences Handler
// ============================================================================

// HandleGetPreferences handles the GetPreferences query.
func (h *Handlers) HandleGetPreferences(ctx context.Context, q GetPreferences) (*PreferencesView, error) {
	const op = "Handlers.HandleGetPreferences"

	prefs, err := h.preferencesRepo.GetPreferences(ctx, q.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return mapPreferencesToView(prefs), nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// mapProjectionToView maps a NotificationProjection to a NotificationView.
func mapProjectionToView(proj *postgres.NotificationProjection) *NotificationView {
	return &NotificationView{
		NotificationID: proj.ID,
		RecipientID:    proj.RecipientID,
		Category:       proj.Category,
		Channel:        proj.Channel,
		TemplateID:     proj.TemplateID,
		Subject:        proj.Subject,
		Body:           proj.Body,
		ActionURL:      proj.ActionURL,
		Status:         proj.Status,
		ErrorMessage:   proj.ErrorMessage,
		CreatedAt:      proj.CreatedAt,
		SentAt:         proj.SentAt,
		ReadAt:         proj.ReadAt,
		UpdatedAt:      proj.UpdatedAt,
	}
}

// mapProjectionToSummary maps a NotificationProjection to a NotificationSummaryView.
func mapProjectionToSummary(proj *postgres.NotificationProjection) NotificationSummaryView {
	return NotificationSummaryView{
		NotificationID: proj.ID,
		Category:       proj.Category,
		Channel:        proj.Channel,
		Subject:        proj.Subject,
		Status:         proj.Status,
		CreatedAt:      proj.CreatedAt,
		ReadAt:         proj.ReadAt,
	}
}

// mapPreferencesToView maps a domain NotificationPreferences to a PreferencesView.
func mapPreferencesToView(prefs domain.NotificationPreferences) *PreferencesView {
	// Convert channel map
	channelEnabled := make(map[string]bool)
	for ch, enabled := range prefs.ChannelEnabled() {
		channelEnabled[ch.String()] = enabled
	}

	// Convert category map
	categoryChannels := make(map[string]map[string]bool)
	for cat, channels := range prefs.CategoryChannels() {
		chMap := make(map[string]bool)
		for ch, enabled := range channels {
			chMap[ch.String()] = enabled
		}
		categoryChannels[cat.String()] = chMap
	}

	return &PreferencesView{
		UserID:           prefs.UserID(),
		GlobalEnabled:    prefs.GlobalEnabled(),
		ChannelEnabled:   channelEnabled,
		CategoryChannels: categoryChannels,
		DigestEnabled:    prefs.DigestEnabled(),
		DigestFrequency:  prefs.DigestFrequency().String(),
		Timezone:         prefs.Timezone(),
	}
}
