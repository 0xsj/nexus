package command

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Command name constants
const (
	CommandCreateNotification     = "notification.CreateNotification"
	CommandMarkNotificationSent   = "notification.MarkNotificationSent"
	CommandMarkNotificationFailed = "notification.MarkNotificationFailed"
	CommandMarkNotificationRead   = "notification.MarkNotificationRead"
	CommandSuppressNotification   = "notification.SuppressNotification"
	CommandUpdatePreferences      = "notification.UpdatePreferences"
)

// ============================================================================
// CreateNotification
// ============================================================================

// CreateNotification creates a new notification queued for delivery.
type CreateNotification struct {
	RecipientID string `json:"recipient_id" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Channel     string `json:"channel" validate:"required"`
	TemplateID  string `json:"template_id" validate:"omitempty"`
	Subject     string `json:"subject" validate:"required"`
	Body        string `json:"body" validate:"required"`
	ActionURL   string `json:"action_url" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c CreateNotification) CommandName() string {
	return CommandCreateNotification
}

// Validate implements cqrs.Validatable.
func (c CreateNotification) Validate() error {
	if c.RecipientID == "" {
		return cqrs.ErrCommandValidation("CreateNotification.Validate", "recipient_id is required")
	}
	if c.Category == "" {
		return cqrs.ErrCommandValidation("CreateNotification.Validate", "category is required")
	}
	if c.Channel == "" {
		return cqrs.ErrCommandValidation("CreateNotification.Validate", "channel is required")
	}
	if c.Subject == "" {
		return cqrs.ErrCommandValidation("CreateNotification.Validate", "subject is required")
	}
	if c.Body == "" {
		return cqrs.ErrCommandValidation("CreateNotification.Validate", "body is required")
	}
	return nil
}

// CreateNotificationResult is the result data for CreateNotification.
type CreateNotificationResult struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

// ============================================================================
// MarkNotificationSent
// ============================================================================

// MarkNotificationSent marks a notification as successfully delivered.
type MarkNotificationSent struct {
	NotificationID types.ID `json:"notification_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c MarkNotificationSent) CommandName() string {
	return CommandMarkNotificationSent
}

// Validate implements cqrs.Validatable.
func (c MarkNotificationSent) Validate() error {
	if c.NotificationID.IsZero() {
		return cqrs.ErrCommandValidation("MarkNotificationSent.Validate", "notification_id is required")
	}
	return nil
}

// MarkNotificationSentResult is the result data for MarkNotificationSent.
type MarkNotificationSentResult struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

// ============================================================================
// MarkNotificationFailed
// ============================================================================

// MarkNotificationFailed marks a notification as failed to deliver.
type MarkNotificationFailed struct {
	NotificationID types.ID `json:"notification_id" validate:"required"`
	ErrorMessage   string   `json:"error_message" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c MarkNotificationFailed) CommandName() string {
	return CommandMarkNotificationFailed
}

// Validate implements cqrs.Validatable.
func (c MarkNotificationFailed) Validate() error {
	if c.NotificationID.IsZero() {
		return cqrs.ErrCommandValidation("MarkNotificationFailed.Validate", "notification_id is required")
	}
	if c.ErrorMessage == "" {
		return cqrs.ErrCommandValidation("MarkNotificationFailed.Validate", "error_message is required")
	}
	return nil
}

// MarkNotificationFailedResult is the result data for MarkNotificationFailed.
type MarkNotificationFailedResult struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

// ============================================================================
// MarkNotificationRead
// ============================================================================

// MarkNotificationRead marks a notification as read by the user.
type MarkNotificationRead struct {
	NotificationID types.ID `json:"notification_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c MarkNotificationRead) CommandName() string {
	return CommandMarkNotificationRead
}

// Validate implements cqrs.Validatable.
func (c MarkNotificationRead) Validate() error {
	if c.NotificationID.IsZero() {
		return cqrs.ErrCommandValidation("MarkNotificationRead.Validate", "notification_id is required")
	}
	return nil
}

// MarkNotificationReadResult is the result data for MarkNotificationRead.
type MarkNotificationReadResult struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

// ============================================================================
// SuppressNotification
// ============================================================================

// SuppressNotification suppresses a queued notification.
type SuppressNotification struct {
	NotificationID types.ID `json:"notification_id" validate:"required"`
	Reason         string   `json:"reason" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c SuppressNotification) CommandName() string {
	return CommandSuppressNotification
}

// Validate implements cqrs.Validatable.
func (c SuppressNotification) Validate() error {
	if c.NotificationID.IsZero() {
		return cqrs.ErrCommandValidation("SuppressNotification.Validate", "notification_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("SuppressNotification.Validate", "reason is required")
	}
	return nil
}

// SuppressNotificationResult is the result data for SuppressNotification.
type SuppressNotificationResult struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

// ============================================================================
// UpdatePreferences
// ============================================================================

// UpdatePreferences updates a user's notification preferences.
type UpdatePreferences struct {
	UserID          string                     `json:"user_id" validate:"required"`
	GlobalEnabled   *bool                      `json:"global_enabled" validate:"omitempty"`
	ChannelUpdates  map[string]bool            `json:"channel_updates" validate:"omitempty"`
	CategoryUpdates map[string]map[string]bool `json:"category_updates" validate:"omitempty"`
	DigestEnabled   *bool                      `json:"digest_enabled" validate:"omitempty"`
	DigestFrequency string                     `json:"digest_frequency" validate:"omitempty"`
	Timezone        string                     `json:"timezone" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c UpdatePreferences) CommandName() string {
	return CommandUpdatePreferences
}

// Validate implements cqrs.Validatable.
func (c UpdatePreferences) Validate() error {
	if c.UserID == "" {
		return cqrs.ErrCommandValidation("UpdatePreferences.Validate", "user_id is required")
	}
	return nil
}

// UpdatePreferencesResult is the result data for UpdatePreferences.
type UpdatePreferencesResult struct {
	UserID string `json:"user_id"`
}
