package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Notification is the aggregate root for outbound notifications.
// It manages the lifecycle of a single notification from creation
// through delivery and read tracking.
type Notification struct {
	eventsourcing.AggregateRoot

	id               NotificationID
	recipientID      string
	category         Category
	channel          Channel
	templateID       string
	subject          string
	body             string
	actionURL        string
	status           DeliveryStatus
	deliveryAttempts []DeliveryAttempt
	createdAt        time.Time
	sentAt           time.Time
	readAt           time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// CreateNotification creates a new Notification aggregate.
func CreateNotification(
	id NotificationID,
	recipientID string,
	category Category,
	channel Channel,
	templateID string,
	subject string,
	body string,
) (*Notification, error) {
	if id.IsZero() {
		return nil, NotificationInvalid("Notification.Create", "notification ID is required")
	}

	if recipientID == "" {
		return nil, NotificationInvalid("Notification.Create", "recipient ID is required")
	}

	if !category.IsValid() {
		return nil, NotificationInvalid("Notification.Create", "invalid category")
	}

	if !channel.IsValid() {
		return nil, NotificationInvalid("Notification.Create", "invalid channel")
	}

	if subject == "" {
		return nil, NotificationInvalid("Notification.Create", "subject is required")
	}

	if body == "" {
		return nil, NotificationInvalid("Notification.Create", "body is required")
	}

	n := &Notification{}
	n.InitAggregate(AggregateTypeNotification, id.String())

	now := time.Now().UTC()

	n.Raise(n, &NotificationCreatedEvent{
		BaseEvent:      newNotificationBaseEvent(id),
		NotificationID: id.String(),
		RecipientID:    recipientID,
		Category:       category.String(),
		Channel:        channel.String(),
		TemplateID:     templateID,
		CreatedAt:      now,
	})

	return n, nil
}

// NewNotificationFromEvents reconstructs a Notification from events (for hydration).
func NewNotificationFromEvents(id string) *Notification {
	n := &Notification{}
	n.InitAggregate(AggregateTypeNotification, id)
	return n
}

// NotificationFactory creates a factory for Notification aggregates.
func NotificationFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewNotificationFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the notification's ID.
func (n *Notification) ID() NotificationID {
	return n.id
}

// RecipientID returns the recipient's user ID.
func (n *Notification) RecipientID() string {
	return n.recipientID
}

// Category returns the notification's category.
func (n *Notification) Category() Category {
	return n.category
}

// Channel returns the notification's delivery channel.
func (n *Notification) Channel() Channel {
	return n.channel
}

// TemplateID returns the template used for this notification.
func (n *Notification) TemplateID() string {
	return n.templateID
}

// Subject returns the notification's subject.
func (n *Notification) Subject() string {
	return n.subject
}

// Body returns the notification's body.
func (n *Notification) Body() string {
	return n.body
}

// ActionURL returns the notification's action URL.
func (n *Notification) ActionURL() string {
	return n.actionURL
}

// Status returns the notification's delivery status.
func (n *Notification) Status() DeliveryStatus {
	return n.status
}

// DeliveryAttempts returns a copy of the delivery attempts.
func (n *Notification) DeliveryAttempts() []DeliveryAttempt {
	attempts := make([]DeliveryAttempt, len(n.deliveryAttempts))
	copy(attempts, n.deliveryAttempts)
	return attempts
}

// CreatedAt returns when the notification was created.
func (n *Notification) CreatedAt() time.Time {
	return n.createdAt
}

// SentAt returns when the notification was sent.
func (n *Notification) SentAt() time.Time {
	return n.sentAt
}

// ReadAt returns when the notification was read.
func (n *Notification) ReadAt() time.Time {
	return n.readAt
}

// ============================================================================
// Command Methods
// ============================================================================

// WithActionURL sets the action URL on the notification before save.
func (n *Notification) WithActionURL(url string) *Notification {
	n.actionURL = url
	return n
}

// MarkAsSent marks the notification as successfully delivered.
func (n *Notification) MarkAsSent() error {
	if n.status != DeliveryStatusQueued {
		return NotificationInvalid("Notification.MarkAsSent", "can only mark queued notifications as sent").
			WithMeta("current_status", n.status.String())
	}

	now := time.Now().UTC()

	n.Raise(n, &NotificationSentEvent{
		BaseEvent:      newNotificationBaseEvent(n.id),
		NotificationID: n.id.String(),
		Channel:        n.channel.String(),
		SentAt:         now,
	})

	return nil
}

// MarkAsFailed marks the notification as failed to deliver.
func (n *Notification) MarkAsFailed(errorMessage string) error {
	if n.status != DeliveryStatusQueued && n.status != DeliveryStatusSending {
		return NotificationInvalid("Notification.MarkAsFailed", "can only mark queued or sending notifications as failed").
			WithMeta("current_status", n.status.String())
	}

	now := time.Now().UTC()

	n.Raise(n, &NotificationFailedEvent{
		BaseEvent:      newNotificationBaseEvent(n.id),
		NotificationID: n.id.String(),
		Channel:        n.channel.String(),
		ErrorMessage:   errorMessage,
		FailedAt:       now,
	})

	return nil
}

// MarkAsRead marks the notification as read by the user.
func (n *Notification) MarkAsRead() error {
	if n.status != DeliveryStatusDelivered {
		return NotificationAlreadyRead("Notification.MarkAsRead", n.id.String()).
			WithMessage("can only mark delivered notifications as read, current status: " + n.status.String())
	}

	now := time.Now().UTC()

	n.Raise(n, &NotificationReadEvent{
		BaseEvent:      newNotificationBaseEvent(n.id),
		NotificationID: n.id.String(),
		ReadAt:         now,
	})

	return nil
}

// Suppress suppresses the notification with the given reason.
func (n *Notification) Suppress(reason string) error {
	if n.status != DeliveryStatusQueued {
		return NotificationInvalid("Notification.Suppress", "can only suppress queued notifications").
			WithMeta("current_status", n.status.String())
	}

	now := time.Now().UTC()

	n.Raise(n, &NotificationSuppressedEvent{
		BaseEvent:      newNotificationBaseEvent(n.id),
		NotificationID: n.id.String(),
		Reason:         reason,
		SuppressedAt:   now,
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (n *Notification) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *NotificationCreatedEvent:
		n.onNotificationCreated(e)
	case *NotificationSentEvent:
		n.onNotificationSent(e)
	case *NotificationFailedEvent:
		n.onNotificationFailed(e)
	case *NotificationReadEvent:
		n.onNotificationRead(e)
	case *NotificationSuppressedEvent:
		n.onNotificationSuppressed(e)
	}
}

func (n *Notification) onNotificationCreated(e *NotificationCreatedEvent) {
	var err error
	n.id, err = ParseNotificationID(e.NotificationID)
	if err != nil {
		panic("corrupt event store: NotificationCreated has invalid NotificationID: " + e.NotificationID)
	}

	n.category, err = ParseCategory(e.Category)
	if err != nil {
		panic("corrupt event store: NotificationCreated has invalid Category: " + e.Category)
	}

	n.channel, err = ParseChannel(e.Channel)
	if err != nil {
		panic("corrupt event store: NotificationCreated has invalid Channel: " + e.Channel)
	}

	n.recipientID = e.RecipientID
	n.templateID = e.TemplateID
	n.status = DeliveryStatusQueued
	n.createdAt = e.CreatedAt
	n.deliveryAttempts = make([]DeliveryAttempt, 0)
}

func (n *Notification) onNotificationSent(e *NotificationSentEvent) {
	channel, err := ParseChannel(e.Channel)
	if err != nil {
		panic("corrupt event store: NotificationSent has invalid Channel: " + e.Channel)
	}

	n.status = DeliveryStatusDelivered
	n.sentAt = e.SentAt
	n.deliveryAttempts = append(n.deliveryAttempts, NewDeliveryAttempt(channel, DeliveryStatusDelivered))
}

func (n *Notification) onNotificationFailed(e *NotificationFailedEvent) {
	channel, err := ParseChannel(e.Channel)
	if err != nil {
		panic("corrupt event store: NotificationFailed has invalid Channel: " + e.Channel)
	}

	n.status = DeliveryStatusFailed
	n.deliveryAttempts = append(n.deliveryAttempts,
		NewDeliveryAttempt(channel, DeliveryStatusFailed).WithError(e.ErrorMessage))
}

func (n *Notification) onNotificationRead(e *NotificationReadEvent) {
	n.status = DeliveryStatusRead
	n.readAt = e.ReadAt
}

func (n *Notification) onNotificationSuppressed(e *NotificationSuppressedEvent) {
	n.status = DeliveryStatusSuppressed
	n.deliveryAttempts = append(n.deliveryAttempts,
		NewDeliveryAttempt(n.channel, DeliveryStatusSuppressed).WithError(e.Reason))
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (n *Notification) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &n.AggregateRoot
}
