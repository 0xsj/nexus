package domain

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Identity Reader Port
// ============================================================================

// IdentityReader retrieves user information from the Identity context.
type IdentityReader interface {
	// GetUser checks if a user exists. Returns an error if the user is not found.
	GetUser(ctx context.Context, userID string) error

	// HasVerifiedCredential checks if a user has at least one verified credential.
	HasVerifiedCredential(ctx context.Context, userID string) (bool, error)
}

// ============================================================================
// Credential Reader Port
// ============================================================================

// CredentialReader retrieves credential information from the Credential context.
type CredentialReader interface {
	// GetCredential checks if a credential exists. Returns an error if not found.
	GetCredential(ctx context.Context, credentialID string) error

	// CountUserCredentials returns the number of credentials for a user.
	CountUserCredentials(ctx context.Context, userID string) (int, error)
}

// ============================================================================
// Organization Reader Port
// ============================================================================

// OrganizationReader retrieves organization information from the Organization context.
type OrganizationReader interface {
	// IsTrustAnchor checks if an organization is a trust anchor.
	IsTrustAnchor(ctx context.Context, orgID string) (bool, error)
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

// NullIdentityReader is a no-op implementation of IdentityReader.
// Returns an error indicating identity reader is not configured.
type NullIdentityReader struct{}

var _ IdentityReader = (*NullIdentityReader)(nil)

// GetUser returns an error indicating identity reader is not configured.
func (n *NullIdentityReader) GetUser(_ context.Context, _ string) error {
	return fmt.Errorf("identity reader not configured")
}

// HasVerifiedCredential returns an error indicating identity reader is not configured.
func (n *NullIdentityReader) HasVerifiedCredential(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("identity reader not configured")
}

// NullCredentialReader is a no-op implementation of CredentialReader.
// Returns an error indicating credential reader is not configured.
type NullCredentialReader struct{}

var _ CredentialReader = (*NullCredentialReader)(nil)

// GetCredential returns an error indicating credential reader is not configured.
func (n *NullCredentialReader) GetCredential(_ context.Context, _ string) error {
	return fmt.Errorf("credential reader not configured")
}

// CountUserCredentials returns an error indicating credential reader is not configured.
func (n *NullCredentialReader) CountUserCredentials(_ context.Context, _ string) (int, error) {
	return 0, fmt.Errorf("credential reader not configured")
}

// NullOrganizationReader is a no-op implementation of OrganizationReader.
// Returns an error indicating organization reader is not configured.
type NullOrganizationReader struct{}

var _ OrganizationReader = (*NullOrganizationReader)(nil)

// IsTrustAnchor returns an error indicating organization reader is not configured.
func (n *NullOrganizationReader) IsTrustAnchor(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("organization reader not configured")
}

// NullEventPublisher is a no-op implementation of EventPublisher.
// Silently discards all events.
type NullEventPublisher struct{}

var _ EventPublisher = (*NullEventPublisher)(nil)

// Publish silently discards all events.
func (n *NullEventPublisher) Publish(_ context.Context, _ ...eventsourcing.Event) error {
	return nil
}
