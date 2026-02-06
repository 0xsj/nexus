package v1

// ============================================================================
// Notification Requests
// ============================================================================

// CreateNotificationRequest represents a request to create a new notification.
type CreateNotificationRequest struct {
	RecipientID string `json:"recipient_id" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Channel     string `json:"channel" validate:"required"`
	TemplateID  string `json:"template_id,omitempty"`
	Subject     string `json:"subject" validate:"required"`
	Body        string `json:"body" validate:"required"`
	ActionURL   string `json:"action_url,omitempty"`
}

// MarkReadRequest represents a request to mark notifications as read.
type MarkReadRequest struct {
	NotificationIDs []string `json:"notification_ids" validate:"required"`
}

// UpdatePreferencesRequest represents a request to update notification preferences.
type UpdatePreferencesRequest struct {
	GlobalEnabled   *bool                      `json:"global_enabled,omitempty"`
	ChannelUpdates  map[string]bool            `json:"channel_updates,omitempty"`
	CategoryUpdates map[string]map[string]bool `json:"category_updates,omitempty"`
	DigestEnabled   *bool                      `json:"digest_enabled,omitempty"`
	DigestFrequency string                     `json:"digest_frequency,omitempty"`
	Timezone        string                     `json:"timezone,omitempty"`
}
