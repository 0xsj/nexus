package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Credential Reader Port
// ============================================================================

// CredentialReader provides read access to credential data for badge generation.
// Used by command handlers to validate credential existence before adding badges.
type CredentialReader interface {
	// GetCredentialByID checks if a credential exists and returns its type.
	GetCredentialByID(ctx context.Context, credentialID string) (exists bool, credType string, err error)
}

// ============================================================================
// Vanity Slug Lookup Port
// ============================================================================

// VanitySlugLookup provides read-optimized queries for vanity URL slug validation.
// Used by command handlers to check slug uniqueness before claiming.
type VanitySlugLookup interface {
	// SlugExists checks if a vanity slug is already taken.
	SlugExists(ctx context.Context, slug string) (bool, error)
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

// NullCredentialReader is a no-op implementation of CredentialReader.
type NullCredentialReader struct{}

// NewNullCredentialReader creates a new NullCredentialReader.
func NewNullCredentialReader() *NullCredentialReader {
	return &NullCredentialReader{}
}

// GetCredentialByID always returns not found.
func (r *NullCredentialReader) GetCredentialByID(ctx context.Context, credentialID string) (bool, string, error) {
	return false, "", nil
}

// Ensure NullCredentialReader implements CredentialReader.
var _ CredentialReader = (*NullCredentialReader)(nil)

// NullVanitySlugLookup is a no-op implementation of VanitySlugLookup.
type NullVanitySlugLookup struct{}

// NewNullVanitySlugLookup creates a new NullVanitySlugLookup.
func NewNullVanitySlugLookup() *NullVanitySlugLookup {
	return &NullVanitySlugLookup{}
}

// SlugExists always returns false.
func (l *NullVanitySlugLookup) SlugExists(ctx context.Context, slug string) (bool, error) {
	return false, nil
}

// Ensure NullVanitySlugLookup implements VanitySlugLookup.
var _ VanitySlugLookup = (*NullVanitySlugLookup)(nil)

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
