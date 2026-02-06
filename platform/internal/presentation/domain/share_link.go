package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ShareLink is an aggregate that manages the lifecycle of a shareable link
// to a presentation. It tracks access, enforces view limits, and supports
// time-based expiration.
type ShareLink struct {
	eventsourcing.AggregateRoot

	id             ShareLinkID
	presentationID PresentationID
	token          string
	expiresAt      time.Time
	maxViews       int
	currentViews   int
	pinHash        string
	audience       string
	status         ShareLinkStatus
	createdAt      time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// CreateShareLink creates a new ShareLink aggregate.
func CreateShareLink(
	id ShareLinkID,
	presentationID PresentationID,
	token string,
	expiresAt time.Time,
	maxViews int,
	pinHash string,
	audience string,
) (*ShareLink, error) {
	if id.IsZero() {
		return nil, ShareLinkInvalid("ShareLink.Create", "share link ID is required")
	}

	if presentationID.IsZero() {
		return nil, ShareLinkInvalid("ShareLink.Create", "presentation ID is required")
	}

	if token == "" {
		return nil, ShareLinkInvalid("ShareLink.Create", "token is required")
	}

	if maxViews < 0 {
		return nil, ShareLinkInvalid("ShareLink.Create", "max views cannot be negative")
	}

	sl := &ShareLink{}
	sl.InitAggregate(AggregateTypeShareLink, id.String())

	now := time.Now().UTC()

	sl.Raise(sl, &ShareLinkCreatedEvent{
		BaseEvent:      newShareLinkBaseEvent(id),
		ShareLinkID:    id.String(),
		PresentationID: presentationID.String(),
		Token:          token,
		ExpiresAt:      expiresAt,
		MaxViews:       maxViews,
		PinHash:        pinHash,
		Audience:       audience,
		CreatedAt:      now,
	})

	return sl, nil
}

// NewShareLinkFromEvents reconstructs a ShareLink from events (for hydration).
func NewShareLinkFromEvents(id string) *ShareLink {
	sl := &ShareLink{}
	sl.InitAggregate(AggregateTypeShareLink, id)
	return sl
}

// ShareLinkFactory creates a factory for ShareLink aggregates.
func ShareLinkFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewShareLinkFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the share link's ID.
func (sl *ShareLink) ID() ShareLinkID {
	return sl.id
}

// PresentationID returns the associated presentation's ID.
func (sl *ShareLink) PresentationID() PresentationID {
	return sl.presentationID
}

// Token returns the unique URL token.
func (sl *ShareLink) Token() string {
	return sl.token
}

// ExpiresAt returns when the share link expires.
func (sl *ShareLink) ExpiresAt() time.Time {
	return sl.expiresAt
}

// MaxViews returns the maximum number of views allowed. 0 means unlimited.
func (sl *ShareLink) MaxViews() int {
	return sl.maxViews
}

// CurrentViews returns the current number of views.
func (sl *ShareLink) CurrentViews() int {
	return sl.currentViews
}

// PinHash returns the hashed PIN for access protection.
func (sl *ShareLink) PinHash() string {
	return sl.pinHash
}

// Audience returns the intended audience DID.
func (sl *ShareLink) Audience() string {
	return sl.audience
}

// Status returns the share link's status.
func (sl *ShareLink) Status() ShareLinkStatus {
	return sl.status
}

// CreatedAt returns when the share link was created.
func (sl *ShareLink) CreatedAt() time.Time {
	return sl.createdAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Access records an access to the share link.
// If maxViews > 0 and currentViews reaches maxViews after this access,
// the share link is automatically expired.
func (sl *ShareLink) Access(verifierDID, ipAddress string) error {
	if !sl.status.IsActive() {
		if sl.status.IsRevoked() {
			return ShareLinkRevoked("ShareLink.Access", sl.id.String())
		}
		return ShareLinkExpiredErr("ShareLink.Access", sl.id.String())
	}

	// Check if max views would be exceeded
	if sl.maxViews > 0 && sl.currentViews >= sl.maxViews {
		return ShareLinkMaxViewsReached("ShareLink.Access", sl.id.String())
	}

	now := time.Now().UTC()

	sl.Raise(sl, &ShareLinkAccessedEvent{
		BaseEvent:   newShareLinkBaseEvent(sl.id),
		ShareLinkID: sl.id.String(),
		VerifierDID: verifierDID,
		IPAddress:   ipAddress,
		AccessedAt:  now,
	})

	// After access, check if we've reached max views and auto-expire
	if sl.maxViews > 0 && sl.currentViews >= sl.maxViews {
		sl.Raise(sl, &ShareLinkExpiredEvent{
			BaseEvent:   newShareLinkBaseEvent(sl.id),
			ShareLinkID: sl.id.String(),
			ExpiredAt:   now,
		})
	}

	return nil
}

// Revoke revokes an active share link.
func (sl *ShareLink) Revoke() error {
	if !sl.status.CanTransitionTo(ShareLinkStatusRevoked) {
		return ShareLinkRevoked("ShareLink.Revoke", sl.id.String()).
			WithMessage("cannot revoke share link in current status: " + sl.status.String())
	}

	sl.Raise(sl, &ShareLinkRevokedEvent{
		BaseEvent:   newShareLinkBaseEvent(sl.id),
		ShareLinkID: sl.id.String(),
		RevokedAt:   time.Now().UTC(),
	})

	return nil
}

// Expire marks an active share link as expired.
func (sl *ShareLink) Expire() error {
	if !sl.status.CanTransitionTo(ShareLinkStatusExpired) {
		return ShareLinkExpiredErr("ShareLink.Expire", sl.id.String()).
			WithMessage("cannot expire share link in current status: " + sl.status.String())
	}

	sl.Raise(sl, &ShareLinkExpiredEvent{
		BaseEvent:   newShareLinkBaseEvent(sl.id),
		ShareLinkID: sl.id.String(),
		ExpiredAt:   time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (sl *ShareLink) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *ShareLinkCreatedEvent:
		sl.onShareLinkCreated(e)
	case *ShareLinkAccessedEvent:
		sl.onShareLinkAccessed(e)
	case *ShareLinkRevokedEvent:
		sl.onShareLinkRevoked(e)
	case *ShareLinkExpiredEvent:
		sl.onShareLinkExpired(e)
	}
}

func (sl *ShareLink) onShareLinkCreated(e *ShareLinkCreatedEvent) {
	var err error
	sl.id, err = ParseShareLinkID(e.ShareLinkID)
	if err != nil {
		panic("corrupt event store: ShareLinkCreated has invalid ShareLinkID: " + e.ShareLinkID)
	}

	sl.presentationID, err = ParsePresentationID(e.PresentationID)
	if err != nil {
		panic("corrupt event store: ShareLinkCreated has invalid PresentationID: " + e.PresentationID)
	}

	sl.token = e.Token
	sl.expiresAt = e.ExpiresAt
	sl.maxViews = e.MaxViews
	sl.currentViews = 0
	sl.pinHash = e.PinHash
	sl.audience = e.Audience
	sl.status = ShareLinkStatusActive
	sl.createdAt = e.CreatedAt
}

func (sl *ShareLink) onShareLinkAccessed(e *ShareLinkAccessedEvent) {
	_ = e // ensure parameter is used
	sl.currentViews++
}

func (sl *ShareLink) onShareLinkRevoked(e *ShareLinkRevokedEvent) {
	_ = e // ensure parameter is used
	sl.status = ShareLinkStatusRevoked
}

func (sl *ShareLink) onShareLinkExpired(e *ShareLinkExpiredEvent) {
	_ = e // ensure parameter is used
	sl.status = ShareLinkStatusExpired
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (sl *ShareLink) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &sl.AggregateRoot
}
