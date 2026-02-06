package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeNotification = "Notification"
)

// Event type constants
const (
	EventTypeNotificationCreated    = "Notification.Created"
	EventTypeNotificationSent       = "Notification.Sent"
	EventTypeNotificationFailed     = "Notification.Failed"
	EventTypeNotificationRead       = "Notification.Read"
	EventTypeNotificationSuppressed = "Notification.Suppressed"
	EventTypePreferencesUpdated     = "Notification.PreferencesUpdated"
)

// ============================================================================
// Notification Events
// ============================================================================

// NotificationCreatedEvent is emitted when a new notification is created.
type NotificationCreatedEvent struct {
	eventsourcing.BaseEvent

	NotificationID string    `json:"notification_id"`
	RecipientID    string    `json:"recipient_id"`
	Category       string    `json:"category"`
	Channel        string    `json:"channel"`
	TemplateID     string    `json:"template_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// EventType returns the event type.
func (e NotificationCreatedEvent) EventType() string {
	return EventTypeNotificationCreated
}

// NotificationSentEvent is emitted when a notification is successfully delivered.
type NotificationSentEvent struct {
	eventsourcing.BaseEvent

	NotificationID string    `json:"notification_id"`
	Channel        string    `json:"channel"`
	SentAt         time.Time `json:"sent_at"`
}

// EventType returns the event type.
func (e NotificationSentEvent) EventType() string {
	return EventTypeNotificationSent
}

// NotificationFailedEvent is emitted when a notification delivery fails.
type NotificationFailedEvent struct {
	eventsourcing.BaseEvent

	NotificationID string    `json:"notification_id"`
	Channel        string    `json:"channel"`
	ErrorMessage   string    `json:"error_message"`
	FailedAt       time.Time `json:"failed_at"`
}

// EventType returns the event type.
func (e NotificationFailedEvent) EventType() string {
	return EventTypeNotificationFailed
}

// NotificationReadEvent is emitted when a user reads a notification.
type NotificationReadEvent struct {
	eventsourcing.BaseEvent

	NotificationID string    `json:"notification_id"`
	ReadAt         time.Time `json:"read_at"`
}

// EventType returns the event type.
func (e NotificationReadEvent) EventType() string {
	return EventTypeNotificationRead
}

// NotificationSuppressedEvent is emitted when a notification is suppressed.
type NotificationSuppressedEvent struct {
	eventsourcing.BaseEvent

	NotificationID string    `json:"notification_id"`
	Reason         string    `json:"reason"`
	SuppressedAt   time.Time `json:"suppressed_at"`
}

// EventType returns the event type.
func (e NotificationSuppressedEvent) EventType() string {
	return EventTypeNotificationSuppressed
}

// PreferencesUpdatedEvent is emitted when a user updates their notification preferences.
type PreferencesUpdatedEvent struct {
	eventsourcing.BaseEvent

	UserID    string    `json:"user_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventType returns the event type.
func (e PreferencesUpdatedEvent) EventType() string {
	return EventTypePreferencesUpdated
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterNotificationEvents registers all Notification domain events with the event registry.
func RegisterNotificationEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeNotificationCreated, func() eventsourcing.Event { return &NotificationCreatedEvent{} })
	registry.Register(EventTypeNotificationSent, func() eventsourcing.Event { return &NotificationSentEvent{} })
	registry.Register(EventTypeNotificationFailed, func() eventsourcing.Event { return &NotificationFailedEvent{} })
	registry.Register(EventTypeNotificationRead, func() eventsourcing.Event { return &NotificationReadEvent{} })
	registry.Register(EventTypeNotificationSuppressed, func() eventsourcing.Event { return &NotificationSuppressedEvent{} })
	registry.Register(EventTypePreferencesUpdated, func() eventsourcing.Event { return &PreferencesUpdatedEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterNotificationEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newNotificationBaseEvent creates a base event for the notification aggregate.
func newNotificationBaseEvent(id NotificationID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeNotification, id.String())
}
