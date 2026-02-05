package domain

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Credential Signer Port
// ============================================================================

// CredentialSigner signs a credential and returns its JWT representation.
type CredentialSigner interface {
	// Sign signs a credential and returns the JWT string.
	Sign(ctx context.Context, credential *Credential) (string, error)
}

// ============================================================================
// Schema Resolver Port
// ============================================================================

// SchemaResolver resolves credential schemas and validates claims.
type SchemaResolver interface {
	// GetSchema returns the schema ID for a given credential type.
	GetSchema(ctx context.Context, credType CredentialType) (schemaID string, err error)

	// ValidateClaims validates claims against the schema for a given credential type.
	ValidateClaims(ctx context.Context, credType CredentialType, claims Claims) error
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

// NullCredentialSigner is a no-op implementation of CredentialSigner.
// Returns an error indicating signing is not configured.
type NullCredentialSigner struct{}

var _ CredentialSigner = (*NullCredentialSigner)(nil)

// Sign returns an error indicating signing is not configured.
func (n *NullCredentialSigner) Sign(_ context.Context, _ *Credential) (string, error) {
	return "", fmt.Errorf("credential signer not configured")
}

// NullSchemaResolver is a no-op implementation of SchemaResolver.
// Returns an error indicating schema resolution is not configured.
type NullSchemaResolver struct{}

var _ SchemaResolver = (*NullSchemaResolver)(nil)

// GetSchema returns an error indicating schema resolution is not configured.
func (n *NullSchemaResolver) GetSchema(_ context.Context, _ CredentialType) (string, error) {
	return "", fmt.Errorf("schema resolver not configured")
}

// ValidateClaims returns an error indicating schema resolution is not configured.
func (n *NullSchemaResolver) ValidateClaims(_ context.Context, _ CredentialType, _ Claims) error {
	return fmt.Errorf("schema resolver not configured")
}

// NullEventPublisher is a no-op implementation of EventPublisher.
// Silently discards all events.
type NullEventPublisher struct{}

var _ EventPublisher = (*NullEventPublisher)(nil)

// Publish silently discards all events.
func (n *NullEventPublisher) Publish(_ context.Context, _ ...eventsourcing.Event) error {
	return nil
}
