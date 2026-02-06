package domain

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Organization Reader Port
// ============================================================================

// OrganizationReader reads organization data from the Organization context.
type OrganizationReader interface {
	// GetOrganization verifies that an organization exists.
	GetOrganization(ctx context.Context, orgID string) error

	// IsVerified checks if an organization is verified.
	IsVerified(ctx context.Context, orgID string) (bool, error)
}

// ============================================================================
// Schema Reader Port
// ============================================================================

// SchemaReader reads schema data from the Schema context.
type SchemaReader interface {
	// GetSchema verifies that a schema type exists.
	GetSchema(ctx context.Context, schemaType string) error

	// SchemaExists checks if a schema type exists.
	SchemaExists(ctx context.Context, schemaType string) (bool, error)
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

// NullOrganizationReader is a no-op implementation of OrganizationReader.
type NullOrganizationReader struct{}

var _ OrganizationReader = (*NullOrganizationReader)(nil)

// GetOrganization returns an error indicating organization reader is not configured.
func (n *NullOrganizationReader) GetOrganization(_ context.Context, _ string) error {
	return fmt.Errorf("organization reader not configured")
}

// IsVerified returns an error indicating organization reader is not configured.
func (n *NullOrganizationReader) IsVerified(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("organization reader not configured")
}

// NullSchemaReader is a no-op implementation of SchemaReader.
type NullSchemaReader struct{}

var _ SchemaReader = (*NullSchemaReader)(nil)

// GetSchema returns an error indicating schema reader is not configured.
func (n *NullSchemaReader) GetSchema(_ context.Context, _ string) error {
	return fmt.Errorf("schema reader not configured")
}

// SchemaExists returns an error indicating schema reader is not configured.
func (n *NullSchemaReader) SchemaExists(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("schema reader not configured")
}

// NullEventPublisher is a no-op implementation of EventPublisher.
// Silently discards all events.
type NullEventPublisher struct{}

var _ EventPublisher = (*NullEventPublisher)(nil)

// Publish silently discards all events.
func (n *NullEventPublisher) Publish(_ context.Context, _ ...eventsourcing.Event) error {
	return nil
}
