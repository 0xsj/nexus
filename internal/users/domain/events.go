package domain

import (
	"github.com/0xsj/nexus/pkg/events"
)

// UserCreatedEvent represents a user creation event.
type UserCreatedEvent struct {
	events.BaseEvent
	UserID   string
	Email    string
	Username string
}

// NewUserCreatedEvent creates a new UserCreated event.
func NewUserCreatedEvent(userID, email, username string) *UserCreatedEvent {
	payload := map[string]interface{}{
		"user_id":  userID,
		"email":    email,
		"username": username,
	}

	base := events.NewBaseEvent(
		"user.created",
		"user",
		userID,
		payload,
	)

	return &UserCreatedEvent{
		BaseEvent: base,
		UserID:    userID,
		Email:     email,
		Username:  username,
	}
}

// UserUpdatedEvent represents a user update event.
type UserUpdatedEvent struct {
	events.BaseEvent
	UserID string
}

// NewUserUpdatedEvent creates a new UserUpdated event.
func NewUserUpdatedEvent(userID string) *UserUpdatedEvent {
	payload := map[string]interface{}{
		"user_id": userID,
	}

	base := events.NewBaseEvent(
		"user.updated",
		"user",
		userID,
		payload,
	)

	return &UserUpdatedEvent{
		BaseEvent: base,
		UserID:    userID,
	}
}

// UserDeletedEvent represents a user deletion event.
type UserDeletedEvent struct {
	events.BaseEvent
	UserID string
	Reason string
}

// NewUserDeletedEvent creates a new UserDeleted event.
func NewUserDeletedEvent(userID, reason string) *UserDeletedEvent {
	payload := map[string]interface{}{
		"user_id": userID,
		"reason":  reason,
	}

	base := events.NewBaseEvent(
		"user.deleted",
		"user",
		userID,
		payload,
	)

	return &UserDeletedEvent{
		BaseEvent: base,
		UserID:    userID,
		Reason:    reason,
	}
}
