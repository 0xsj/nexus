package query

import (
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Query name constants
const (
	QueryGetNotification              = "notification.GetNotification"
	QueryListNotificationsByRecipient = "notification.ListNotificationsByRecipient"
	QueryGetPreferences               = "notification.GetPreferences"
)

// ============================================================================
// GetNotification
// ============================================================================

// GetNotification retrieves a single notification by ID.
type GetNotification struct {
	NotificationID types.ID `json:"notification_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetNotification) QueryName() string {
	return QueryGetNotification
}

// Validate implements cqrs.Validatable.
func (q GetNotification) Validate() error {
	if q.NotificationID.IsZero() {
		return cqrs.ErrQueryValidation("GetNotification.Validate", "notification_id is required")
	}
	return nil
}

// ============================================================================
// ListNotificationsByRecipient
// ============================================================================

// ListNotificationsByRecipient lists notifications for a recipient (inbox).
type ListNotificationsByRecipient struct {
	RecipientID string `json:"recipient_id" validate:"required"`
	Limit       int    `json:"limit" validate:"omitempty"`
	Offset      int    `json:"offset" validate:"omitempty"`
}

// QueryName implements cqrs.Query.
func (q ListNotificationsByRecipient) QueryName() string {
	return QueryListNotificationsByRecipient
}

// Validate implements cqrs.Validatable.
func (q ListNotificationsByRecipient) Validate() error {
	if q.RecipientID == "" {
		return cqrs.ErrQueryValidation("ListNotificationsByRecipient.Validate", "recipient_id is required")
	}
	return nil
}

// ============================================================================
// GetPreferences
// ============================================================================

// GetPreferences retrieves notification preferences for a user.
type GetPreferences struct {
	UserID string `json:"user_id" validate:"required"`
}

// QueryName implements cqrs.Query.
func (q GetPreferences) QueryName() string {
	return QueryGetPreferences
}

// Validate implements cqrs.Validatable.
func (q GetPreferences) Validate() error {
	if q.UserID == "" {
		return cqrs.ErrQueryValidation("GetPreferences.Validate", "user_id is required")
	}
	return nil
}
