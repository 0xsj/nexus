package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Identity Reader Port
// ============================================================================

// IdentityReader retrieves recipient information from the Identity context.
type IdentityReader interface {
	// GetUserEmail returns the email address for the given user ID.
	GetUserEmail(ctx context.Context, userID string) (string, error)

	// UserExists checks if a user with the given ID exists.
	UserExists(ctx context.Context, userID string) (bool, error)
}

// ============================================================================
// Email Sender Port
// ============================================================================

// EmailSender sends email notifications.
type EmailSender interface {
	// SendEmail sends an email to the given recipient.
	SendEmail(ctx context.Context, to string, subject string, body string) error
}

// ============================================================================
// Push Sender Port
// ============================================================================

// PushSender sends push notifications.
type PushSender interface {
	// SendPush sends a push notification to the given user.
	SendPush(ctx context.Context, userID string, title string, body string) error
}

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	// Publish publishes one or more domain events synchronously.
	Publish(ctx context.Context, events ...eventsourcing.Event) error
}

// ============================================================================
// Null Implementations (for testing)
// ============================================================================

// NullIdentityReader is a no-op implementation of IdentityReader.
type NullIdentityReader struct{}

// NewNullIdentityReader creates a new NullIdentityReader.
func NewNullIdentityReader() *NullIdentityReader {
	return &NullIdentityReader{}
}

// GetUserEmail always returns an empty string and no error.
func (r *NullIdentityReader) GetUserEmail(ctx context.Context, userID string) (string, error) {
	return "", nil
}

// UserExists always returns false.
func (r *NullIdentityReader) UserExists(ctx context.Context, userID string) (bool, error) {
	return false, nil
}

// Ensure NullIdentityReader implements IdentityReader.
var _ IdentityReader = (*NullIdentityReader)(nil)

// NullEmailSender is a no-op implementation of EmailSender.
type NullEmailSender struct{}

// NewNullEmailSender creates a new NullEmailSender.
func NewNullEmailSender() *NullEmailSender {
	return &NullEmailSender{}
}

// SendEmail does nothing and returns nil.
func (s *NullEmailSender) SendEmail(ctx context.Context, to string, subject string, body string) error {
	return nil
}

// Ensure NullEmailSender implements EmailSender.
var _ EmailSender = (*NullEmailSender)(nil)

// NullPushSender is a no-op implementation of PushSender.
type NullPushSender struct{}

// NewNullPushSender creates a new NullPushSender.
func NewNullPushSender() *NullPushSender {
	return &NullPushSender{}
}

// SendPush does nothing and returns nil.
func (s *NullPushSender) SendPush(ctx context.Context, userID string, title string, body string) error {
	return nil
}

// Ensure NullPushSender implements PushSender.
var _ PushSender = (*NullPushSender)(nil)

// NullEventPublisher is a no-op implementation of EventPublisher.
type NullEventPublisher struct{}

// NewNullEventPublisher creates a new NullEventPublisher.
func NewNullEventPublisher() *NullEventPublisher {
	return &NullEventPublisher{}
}

// Publish does nothing and returns nil.
func (p *NullEventPublisher) Publish(ctx context.Context, events ...eventsourcing.Event) error {
	return nil
}

// Ensure NullEventPublisher implements EventPublisher.
var _ EventPublisher = (*NullEventPublisher)(nil)
