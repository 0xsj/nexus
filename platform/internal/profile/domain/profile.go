package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Profile is the aggregate root for public user profiles.
// It manages the public representation of a user's verified identity,
// including badges derived from credentials, sections with visibility
// controls, and vanity URL slugs.
type Profile struct {
	eventsourcing.AggregateRoot

	id          ProfileID
	userID      string
	displayName string
	headline    string
	bio         string
	vanitySlug  string
	badges      map[string]Badge
	sections    []Section
	createdAt   time.Time
	updatedAt   time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// CreateProfile creates a new Profile aggregate.
func CreateProfile(
	id ProfileID,
	userID string,
	displayName string,
	headline string,
	vanitySlug string,
) (*Profile, error) {
	if id.IsZero() {
		return nil, ProfileInvalid("Profile.Create", "profile ID is required")
	}

	if userID == "" {
		return nil, ProfileInvalid("Profile.Create", "user ID is required")
	}

	if displayName == "" {
		return nil, ProfileInvalid("Profile.Create", "display name is required")
	}

	if vanitySlug == "" {
		return nil, ProfileInvalid("Profile.Create", "vanity slug is required")
	}

	p := &Profile{
		badges: make(map[string]Badge),
	}
	p.InitAggregate(AggregateTypeProfile, id.String())

	now := time.Now().UTC()

	p.Raise(p, &ProfileCreatedEvent{
		BaseEvent:   newProfileBaseEvent(id),
		ProfileID:   id.String(),
		UserID:      userID,
		DisplayName: displayName,
		Headline:    headline,
		VanitySlug:  vanitySlug,
		CreatedAt:   now,
	})

	return p, nil
}

// NewProfileFromEvents reconstructs a Profile from events (for hydration).
func NewProfileFromEvents(id string) *Profile {
	p := &Profile{
		badges: make(map[string]Badge),
	}
	p.InitAggregate(AggregateTypeProfile, id)
	return p
}

// ProfileFactory creates a factory for Profile aggregates.
func ProfileFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewProfileFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the profile's ID.
func (p *Profile) ID() ProfileID {
	return p.id
}

// UserID returns the ID of the user who owns this profile.
func (p *Profile) UserID() string {
	return p.userID
}

// DisplayName returns the profile's display name.
func (p *Profile) DisplayName() string {
	return p.displayName
}

// Headline returns the profile's headline.
func (p *Profile) Headline() string {
	return p.headline
}

// Bio returns the profile's bio.
func (p *Profile) Bio() string {
	return p.bio
}

// VanitySlug returns the profile's vanity URL slug.
func (p *Profile) VanitySlug() string {
	return p.vanitySlug
}

// Badges returns a copy of the profile's badges.
func (p *Profile) Badges() map[string]Badge {
	copy := make(map[string]Badge, len(p.badges))
	for k, v := range p.badges {
		copy[k] = v
	}
	return copy
}

// Badge returns a specific badge by ID, or false if not found.
func (p *Profile) Badge(badgeID BadgeID) (Badge, bool) {
	b, ok := p.badges[badgeID.String()]
	return b, ok
}

// Sections returns a copy of the profile's sections.
func (p *Profile) Sections() []Section {
	copy := make([]Section, len(p.sections))
	for i, s := range p.sections {
		copy[i] = s
	}
	return copy
}

// CreatedAt returns when the profile was created.
func (p *Profile) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns when the profile was last updated.
func (p *Profile) UpdatedAt() time.Time {
	return p.updatedAt
}

// ============================================================================
// Command Methods
// ============================================================================

// UpdateMetadata updates the profile's display name, headline, and bio.
func (p *Profile) UpdateMetadata(displayName string, headline string, bio string) error {
	if displayName == "" {
		return ProfileInvalid("Profile.UpdateMetadata", "display name is required")
	}

	p.Raise(p, &ProfileUpdatedEvent{
		BaseEvent:   newProfileBaseEvent(p.id),
		ProfileID:   p.id.String(),
		DisplayName: displayName,
		Headline:    headline,
		Bio:         bio,
		UpdatedAt:   time.Now().UTC(),
	})

	return nil
}

// AddBadge adds a badge to the profile.
func (p *Profile) AddBadge(badge Badge) error {
	if badge.ID().IsZero() {
		return ProfileInvalid("Profile.AddBadge", "badge ID is required")
	}

	if _, exists := p.badges[badge.ID().String()]; exists {
		return ProfileInvalid("Profile.AddBadge", "badge already exists on profile")
	}

	now := time.Now().UTC()

	p.Raise(p, &BadgeAddedEvent{
		BaseEvent:      newProfileBaseEvent(p.id),
		ProfileID:      p.id.String(),
		BadgeID:        badge.ID().String(),
		CredentialID:   badge.CredentialID(),
		BadgeType:      badge.BadgeType(),
		DisplayName:    badge.DisplayName(),
		PrimaryValue:   badge.PrimaryValue(),
		SecondaryValue: badge.SecondaryValue(),
		Icon:           badge.Icon(),
		VerifiedAt:     badge.VerifiedAt(),
		Visibility:     badge.Visibility().String(),
		AddedAt:        now,
	})

	return nil
}

// RemoveBadge removes a badge from the profile.
func (p *Profile) RemoveBadge(badgeID BadgeID) error {
	if _, exists := p.badges[badgeID.String()]; !exists {
		return BadgeNotFound("Profile.RemoveBadge", badgeID.String())
	}

	p.Raise(p, &BadgeRemovedEvent{
		BaseEvent: newProfileBaseEvent(p.id),
		ProfileID: p.id.String(),
		BadgeID:   badgeID.String(),
		RemovedAt: time.Now().UTC(),
	})

	return nil
}

// ChangeBadgeVisibility changes the visibility of a badge on the profile.
func (p *Profile) ChangeBadgeVisibility(badgeID BadgeID, visibility Visibility) error {
	if _, exists := p.badges[badgeID.String()]; !exists {
		return BadgeNotFound("Profile.ChangeBadgeVisibility", badgeID.String())
	}

	if !visibility.IsValid() {
		return ProfileInvalid("Profile.ChangeBadgeVisibility", "invalid visibility")
	}

	p.Raise(p, &BadgeVisibilityChangedEvent{
		BaseEvent:  newProfileBaseEvent(p.id),
		ProfileID:  p.id.String(),
		BadgeID:    badgeID.String(),
		Visibility: visibility.String(),
		ChangedAt:  time.Now().UTC(),
	})

	return nil
}

// ClaimVanityURL claims a new vanity URL slug for the profile.
func (p *Profile) ClaimVanityURL(slug string) error {
	if slug == "" {
		return VanityURLInvalid("Profile.ClaimVanityURL", "slug is required")
	}

	p.Raise(p, &VanityURLClaimedEvent{
		BaseEvent: newProfileBaseEvent(p.id),
		ProfileID: p.id.String(),
		Slug:      slug,
		ClaimedAt: time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (p *Profile) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *ProfileCreatedEvent:
		p.onProfileCreated(e)
	case *ProfileUpdatedEvent:
		p.onProfileUpdated(e)
	case *BadgeAddedEvent:
		p.onBadgeAdded(e)
	case *BadgeRemovedEvent:
		p.onBadgeRemoved(e)
	case *BadgeVisibilityChangedEvent:
		p.onBadgeVisibilityChanged(e)
	case *VanityURLClaimedEvent:
		p.onVanityURLClaimed(e)
	}
}

func (p *Profile) onProfileCreated(e *ProfileCreatedEvent) {
	var err error
	p.id, err = ParseProfileID(e.ProfileID)
	if err != nil {
		panic("corrupt event store: ProfileCreated has invalid ProfileID: " + e.ProfileID)
	}
	p.userID = e.UserID
	p.displayName = e.DisplayName
	p.headline = e.Headline
	p.vanitySlug = e.VanitySlug
	p.createdAt = e.CreatedAt
	p.updatedAt = e.CreatedAt
	if p.badges == nil {
		p.badges = make(map[string]Badge)
	}
}

func (p *Profile) onProfileUpdated(e *ProfileUpdatedEvent) {
	p.displayName = e.DisplayName
	p.headline = e.Headline
	p.bio = e.Bio
	p.updatedAt = e.UpdatedAt
}

func (p *Profile) onBadgeAdded(e *BadgeAddedEvent) {
	badgeID, err := ParseBadgeID(e.BadgeID)
	if err != nil {
		panic("corrupt event store: BadgeAdded has invalid BadgeID: " + e.BadgeID)
	}

	visibility, err := ParseVisibility(e.Visibility)
	if err != nil {
		panic("corrupt event store: BadgeAdded has invalid Visibility: " + e.Visibility)
	}

	badge := NewBadge(
		badgeID,
		e.CredentialID,
		e.BadgeType,
		e.DisplayName,
		e.PrimaryValue,
		e.VerifiedAt,
	).WithSecondaryValue(e.SecondaryValue).
		WithIcon(e.Icon).
		WithVisibility(visibility)

	if p.badges == nil {
		p.badges = make(map[string]Badge)
	}
	p.badges[badgeID.String()] = badge
	p.updatedAt = e.AddedAt
}

func (p *Profile) onBadgeRemoved(e *BadgeRemovedEvent) {
	delete(p.badges, e.BadgeID)
	p.updatedAt = e.RemovedAt
}

func (p *Profile) onBadgeVisibilityChanged(e *BadgeVisibilityChangedEvent) {
	visibility, err := ParseVisibility(e.Visibility)
	if err != nil {
		panic("corrupt event store: BadgeVisibilityChanged has invalid Visibility: " + e.Visibility)
	}

	badge, exists := p.badges[e.BadgeID]
	if !exists {
		panic("corrupt event store: BadgeVisibilityChanged references unknown BadgeID: " + e.BadgeID)
	}

	p.badges[e.BadgeID] = badge.WithVisibility(visibility)
	p.updatedAt = e.ChangedAt
}

func (p *Profile) onVanityURLClaimed(e *VanityURLClaimedEvent) {
	p.vanitySlug = e.Slug
	p.updatedAt = e.ClaimedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (p *Profile) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &p.AggregateRoot
}
