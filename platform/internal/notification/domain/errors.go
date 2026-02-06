package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Error Codes (Notification-specific)
// ============================================================================

const (
	CodeNotificationNotFound    pkgerrors.Code = "NOTIFICATION_NOT_FOUND"
	CodeNotificationAlreadyRead pkgerrors.Code = "NOTIFICATION_ALREADY_READ"
	CodeNotificationInvalid     pkgerrors.Code = "NOTIFICATION_INVALID"
	CodePreferencesNotFound     pkgerrors.Code = "PREFERENCES_NOT_FOUND"
	CodeTemplateNotFound        pkgerrors.Code = "TEMPLATE_NOT_FOUND"
	CodeDeliveryFailed          pkgerrors.Code = "DELIVERY_FAILED"
)

// ============================================================================
// Sentinel Errors
// ============================================================================

var (
	ErrNotificationNotFound    = errors.New("notification not found")
	ErrNotificationAlreadyRead = errors.New("notification already read")
	ErrNotificationInvalid     = errors.New("notification is invalid")
	ErrPreferencesNotFound     = errors.New("preferences not found")
	ErrTemplateNotFound        = errors.New("template not found")
	ErrDeliveryFailed          = errors.New("delivery failed")
)

// ============================================================================
// Error Constructors
// ============================================================================

// NotificationNotFound creates a notification not found error.
func NotificationNotFound(operation string, notificationID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "notification").
		WithCode(CodeNotificationNotFound).
		WithMeta("notification_id", notificationID)
}

// NotificationAlreadyRead creates a notification already read error.
func NotificationAlreadyRead(operation string, notificationID string) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "notification has already been read").
		WithCode(CodeNotificationAlreadyRead).
		WithMeta("notification_id", notificationID)
}

// NotificationInvalid creates a notification invalid error.
func NotificationInvalid(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "notification is invalid: "+reason).
		WithCode(CodeNotificationInvalid)
}

// PreferencesNotFound creates a preferences not found error.
func PreferencesNotFound(operation string, userID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "notification preferences").
		WithCode(CodePreferencesNotFound).
		WithMeta("user_id", userID)
}

// TemplateNotFound creates a template not found error.
func TemplateNotFound(operation string, templateID string) *pkgerrors.Error {
	return pkgerrors.NotFound(operation, "notification template").
		WithCode(CodeTemplateNotFound).
		WithMeta("template_id", templateID)
}

// DeliveryFailed creates a delivery failed error.
func DeliveryFailed(operation string, channel Channel, err error) *pkgerrors.Error {
	return pkgerrors.Domain(operation, "notification delivery failed: "+err.Error()).
		WithCode(CodeDeliveryFailed).
		WithMeta("channel", channel.String())
}
