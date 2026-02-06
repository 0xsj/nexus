package query

import "time"

// ============================================================================
// Notification Views
// ============================================================================

// NotificationView is the full notification read model.
type NotificationView struct {
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

// NotificationSummaryView is a lightweight notification representation for lists.
type NotificationSummaryView struct {
	NotificationID string     `json:"notification_id"`
	Category       string     `json:"category"`
	Channel        string     `json:"channel"`
	Subject        string     `json:"subject"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
}

// NotificationListView is a paginated list of notifications with unread count.
type NotificationListView struct {
	Notifications []NotificationSummaryView `json:"notifications"`
	TotalCount    int                       `json:"total_count"`
	UnreadCount   int                       `json:"unread_count"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
	HasMore       bool                      `json:"has_more"`
}

// ============================================================================
// Preferences Views
// ============================================================================

// PreferencesView is the notification preferences read model.
type PreferencesView struct {
	UserID           string                     `json:"user_id"`
	GlobalEnabled    bool                       `json:"global_enabled"`
	ChannelEnabled   map[string]bool            `json:"channel_enabled"`
	CategoryChannels map[string]map[string]bool `json:"category_channels"`
	DigestEnabled    bool                       `json:"digest_enabled"`
	DigestFrequency  string                     `json:"digest_frequency"`
	Timezone         string                     `json:"timezone"`
}
