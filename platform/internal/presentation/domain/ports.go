package domain

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Credential Reader Port
// ============================================================================

// CredentialReader reads credentials for inclusion in presentations.
type CredentialReader interface {
	// GetCredential checks if a credential exists and is valid.
	GetCredential(ctx context.Context, credentialID string) error

	// GetUserCredentials returns credential IDs for a user.
	GetUserCredentials(ctx context.Context, userID string) ([]string, error)
}

// ============================================================================
// Identity Reader Port
// ============================================================================

// IdentityReader retrieves holder DID for signing.
type IdentityReader interface {
	// GetUserDID returns the DID for a user.
	GetUserDID(ctx context.Context, userID string) (string, error)
}

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	// Publish publishes one or more domain events.
	Publish(ctx context.Context, events ...eventsourcing.Event) error
}

// ============================================================================
// Null Implementations
// ============================================================================

// NullCredentialReader is a no-op implementation of CredentialReader.
// Returns an error indicating credential reading is not configured.
type NullCredentialReader struct{}

var _ CredentialReader = (*NullCredentialReader)(nil)

// GetCredential returns an error indicating credential reading is not configured.
func (n *NullCredentialReader) GetCredential(_ context.Context, _ string) error {
	return fmt.Errorf("credential reader not configured")
}

// GetUserCredentials returns an error indicating credential reading is not configured.
func (n *NullCredentialReader) GetUserCredentials(_ context.Context, _ string) ([]string, error) {
	return nil, fmt.Errorf("credential reader not configured")
}

// NullIdentityReader is a no-op implementation of IdentityReader.
// Returns an error indicating identity reading is not configured.
type NullIdentityReader struct{}

var _ IdentityReader = (*NullIdentityReader)(nil)

// GetUserDID returns an error indicating identity reading is not configured.
func (n *NullIdentityReader) GetUserDID(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("identity reader not configured")
}

// NullEventPublisher is a no-op implementation of EventPublisher.
// Silently discards all events.
type NullEventPublisher struct{}

var _ EventPublisher = (*NullEventPublisher)(nil)

// Publish silently discards all events.
func (n *NullEventPublisher) Publish(_ context.Context, _ ...eventsourcing.Event) error {
	return nil
}
