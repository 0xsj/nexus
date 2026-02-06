package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Identity Reader Port
// ============================================================================

// IdentityReader retrieves user information for membership validation.
type IdentityReader interface {
	// UserExists checks if a user with the given ID exists in the Identity context.
	UserExists(ctx context.Context, userID string) (bool, error)

	// GetUserEmail retrieves a user's email by their ID.
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

// ============================================================================
// DID Service Port
// ============================================================================

// DIDService manages organizational DIDs.
type DIDService interface {
	// CreateOrganizationDID creates a did:web for the given organization slug.
	CreateOrganizationDID(ctx context.Context, orgSlug string) (string, error)
}

// ============================================================================
// Notification Service Port
// ============================================================================

// NotificationService sends membership-related notifications.
type NotificationService interface {
	// SendInvitation sends an invitation notification to the invitee.
	SendInvitation(ctx context.Context, email string, orgName string, role Role) error

	// NotifyMemberRemoved notifies a member that they have been removed.
	NotifyMemberRemoved(ctx context.Context, userID string, orgName string, reason string) error
}

// ============================================================================
// Slug Lookup Port
// ============================================================================

// SlugLookup checks slug uniqueness across organizations.
type SlugLookup interface {
	// SlugExists checks if the given slug is already taken.
	SlugExists(ctx context.Context, slug string) (bool, error)
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
type NullIdentityReader struct{}

func NewNullIdentityReader() *NullIdentityReader { return &NullIdentityReader{} }

func (r *NullIdentityReader) UserExists(ctx context.Context, userID string) (bool, error) {
	return true, nil
}

func (r *NullIdentityReader) GetUserEmail(ctx context.Context, userID string) (string, error) {
	return "", nil
}

var _ IdentityReader = (*NullIdentityReader)(nil)

// NullDIDService is a no-op implementation of DIDService.
type NullDIDService struct{}

func NewNullDIDService() *NullDIDService { return &NullDIDService{} }

func (s *NullDIDService) CreateOrganizationDID(ctx context.Context, orgSlug string) (string, error) {
	return "", nil
}

var _ DIDService = (*NullDIDService)(nil)

// NullNotificationService is a no-op implementation of NotificationService.
type NullNotificationService struct{}

func NewNullNotificationService() *NullNotificationService { return &NullNotificationService{} }

func (s *NullNotificationService) SendInvitation(ctx context.Context, email string, orgName string, role Role) error {
	return nil
}

func (s *NullNotificationService) NotifyMemberRemoved(ctx context.Context, userID string, orgName string, reason string) error {
	return nil
}

var _ NotificationService = (*NullNotificationService)(nil)

// NullSlugLookup is a no-op implementation of SlugLookup.
type NullSlugLookup struct{}

func NewNullSlugLookup() *NullSlugLookup { return &NullSlugLookup{} }

func (l *NullSlugLookup) SlugExists(ctx context.Context, slug string) (bool, error) {
	return false, nil
}

var _ SlugLookup = (*NullSlugLookup)(nil)

// NullEventPublisher is a no-op implementation of EventPublisher.
type NullEventPublisher struct{}

func NewNullEventPublisher() *NullEventPublisher { return &NullEventPublisher{} }

func (p *NullEventPublisher) Publish(ctx context.Context, events ...eventsourcing.Event) error {
	return nil
}

var _ EventPublisher = (*NullEventPublisher)(nil)
