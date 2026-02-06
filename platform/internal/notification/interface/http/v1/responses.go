package v1

import "time"

// ============================================================================
// Notification Responses
// ============================================================================

// NotificationResponse represents a full notification in API responses.
type NotificationResponse struct {
	NotificationID string     `json:"notification_id"`
	RecipientID    string     `json:"recipient_id"`
	Category       string     `json:"category"`
	Channel        string     `json:"channel"`
	TemplateID     string     `json:"template_id,omitempty"`
	Subject        string     `json:"subject"`
	Body           string     `json:"body"`
	ActionURL      string     `json:"action_url,omitempty"`
	Status         string     `json:"status"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NotificationSummaryResponse represents a lightweight notification in list responses.
type NotificationSummaryResponse struct {
	NotificationID string     `json:"notification_id"`
	Category       string     `json:"category"`
	Channel        string     `json:"channel"`
	Subject        string     `json:"subject"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
}

// NotificationListResponse represents a paginated list of notifications.
type NotificationListResponse struct {
	Notifications []NotificationSummaryResponse `json:"notifications"`
	TotalCount    int                           `json:"total_count"`
	UnreadCount   int                           `json:"unread_count"`
	Limit         int                           `json:"limit"`
	Offset        int                           `json:"offset"`
	HasMore       bool                          `json:"has_more"`
}

// ============================================================================
// Command Result Responses
// ============================================================================

// NotificationCreatedResponse represents the result of creating a notification.
type NotificationCreatedResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// NotificationReadResponse represents the result of marking notifications as read.
type NotificationReadResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	ReadAt         time.Time `json:"read_at"`
}

// ============================================================================
// Preferences Responses
// ============================================================================

// PreferencesResponse represents notification preferences in API responses.
type PreferencesResponse struct {
	UserID           string                     `json:"user_id"`
	GlobalEnabled    bool                       `json:"global_enabled"`
	ChannelEnabled   map[string]bool            `json:"channel_enabled"`
	CategoryChannels map[string]map[string]bool `json:"category_channels"`
	DigestEnabled    bool                       `json:"digest_enabled"`
	DigestFrequency  string                     `json:"digest_frequency"`
	Timezone         string                     `json:"timezone"`
}
