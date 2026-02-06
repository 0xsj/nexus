package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeProfile = "Profile"
)

// Event type constants
const (
	EventTypeProfileCreated          = "Profile.Created"
	EventTypeProfileUpdated          = "Profile.Updated"
	EventTypeBadgeAdded              = "Profile.BadgeAdded"
	EventTypeBadgeRemoved            = "Profile.BadgeRemoved"
	EventTypeBadgeVisibilityChanged  = "Profile.BadgeVisibilityChanged"
	EventTypeVanityURLClaimed        = "Profile.VanityURLClaimed"
)

// ============================================================================
// Profile Events
// ============================================================================

// ProfileCreatedEvent is emitted when a new profile is created.
type ProfileCreatedEvent struct {
	eventsourcing.BaseEvent

	ProfileID   string    `json:"profile_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Headline    string    `json:"headline"`
	VanitySlug  string    `json:"vanity_slug"`
	CreatedAt   time.Time `json:"created_at"`
}

// EventType returns the event type.
func (e ProfileCreatedEvent) EventType() string {
	return EventTypeProfileCreated
}

// ProfileUpdatedEvent is emitted when profile metadata is updated.
type ProfileUpdatedEvent struct {
	eventsourcing.BaseEvent

	ProfileID   string    `json:"profile_id"`
	DisplayName string    `json:"display_name"`
	Headline    string    `json:"headline"`
	Bio         string    `json:"bio"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EventType returns the event type.
func (e ProfileUpdatedEvent) EventType() string {
	return EventTypeProfileUpdated
}

// BadgeAddedEvent is emitted when a badge is added to a profile.
type BadgeAddedEvent struct {
	eventsourcing.BaseEvent

	ProfileID      string    `json:"profile_id"`
	BadgeID        string    `json:"badge_id"`
	CredentialID   string    `json:"credential_id"`
	BadgeType      string    `json:"badge_type"`
	DisplayName    string    `json:"display_name"`
	PrimaryValue   string    `json:"primary_value"`
	SecondaryValue string    `json:"secondary_value,omitempty"`
	Icon           string    `json:"icon,omitempty"`
	VerifiedAt     time.Time `json:"verified_at"`
	Visibility     string    `json:"visibility"`
	AddedAt        time.Time `json:"added_at"`
}

// EventType returns the event type.
func (e BadgeAddedEvent) EventType() string {
	return EventTypeBadgeAdded
}

// BadgeRemovedEvent is emitted when a badge is removed from a profile.
type BadgeRemovedEvent struct {
	eventsourcing.BaseEvent

	ProfileID string    `json:"profile_id"`
	BadgeID   string    `json:"badge_id"`
	RemovedAt time.Time `json:"removed_at"`
}

// EventType returns the event type.
func (e BadgeRemovedEvent) EventType() string {
	return EventTypeBadgeRemoved
}

// BadgeVisibilityChangedEvent is emitted when a badge's visibility is changed.
type BadgeVisibilityChangedEvent struct {
	eventsourcing.BaseEvent

	ProfileID  string    `json:"profile_id"`
	BadgeID    string    `json:"badge_id"`
	Visibility string    `json:"visibility"`
	ChangedAt  time.Time `json:"changed_at"`
}

// EventType returns the event type.
func (e BadgeVisibilityChangedEvent) EventType() string {
	return EventTypeBadgeVisibilityChanged
}

// VanityURLClaimedEvent is emitted when a user claims a vanity URL slug.
type VanityURLClaimedEvent struct {
	eventsourcing.BaseEvent

	ProfileID string    `json:"profile_id"`
	Slug      string    `json:"slug"`
	ClaimedAt time.Time `json:"claimed_at"`
}

// EventType returns the event type.
func (e VanityURLClaimedEvent) EventType() string {
	return EventTypeVanityURLClaimed
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterProfileEvents registers all Profile domain events with the event registry.
func RegisterProfileEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeProfileCreated, func() eventsourcing.Event { return &ProfileCreatedEvent{} })
	registry.Register(EventTypeProfileUpdated, func() eventsourcing.Event { return &ProfileUpdatedEvent{} })
	registry.Register(EventTypeBadgeAdded, func() eventsourcing.Event { return &BadgeAddedEvent{} })
	registry.Register(EventTypeBadgeRemoved, func() eventsourcing.Event { return &BadgeRemovedEvent{} })
	registry.Register(EventTypeBadgeVisibilityChanged, func() eventsourcing.Event { return &BadgeVisibilityChangedEvent{} })
	registry.Register(EventTypeVanityURLClaimed, func() eventsourcing.Event { return &VanityURLClaimedEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterProfileEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newProfileBaseEvent creates a base event for profile aggregate.
func newProfileBaseEvent(id ProfileID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeProfile, id.String())
}
