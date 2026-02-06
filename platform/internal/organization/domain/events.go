package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeOrganization = "Organization"
)

// Event type constants
const (
	EventTypeOrganizationCreated  = "Organization.Created"
	EventTypeOrganizationUpdated  = "Organization.Updated"
	EventTypeOrganizationVerified = "Organization.Verified"
	EventTypeOrganizationDeleted  = "Organization.Deleted"
	EventTypeMemberAdded          = "Organization.MemberAdded"
	EventTypeMemberRemoved        = "Organization.MemberRemoved"
	EventTypeMemberRoleChanged    = "Organization.MemberRoleChanged"
	EventTypeOwnershipTransferred = "Organization.OwnershipTransferred"
)

// ============================================================================
// Organization Events
// ============================================================================

// OrganizationCreatedEvent is emitted when a new organization is created.
type OrganizationCreatedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	OrgType        string `json:"org_type"`
	OwnerUserID    string `json:"owner_user_id"`
	OwnerMemberID  string `json:"owner_member_id"`
	CreatedAt      time.Time `json:"created_at"`
}

func (e OrganizationCreatedEvent) EventType() string { return EventTypeOrganizationCreated }

// OrganizationUpdatedEvent is emitted when organization metadata is changed.
type OrganizationUpdatedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string `json:"organization_id"`
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (e OrganizationUpdatedEvent) EventType() string { return EventTypeOrganizationUpdated }

// OrganizationVerifiedEvent is emitted when KYB verification completes.
type OrganizationVerifiedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string `json:"organization_id"`
	DID            string `json:"did,omitempty"`
	VerifiedAt     time.Time `json:"verified_at"`
}

func (e OrganizationVerifiedEvent) EventType() string { return EventTypeOrganizationVerified }

// OrganizationDeletedEvent is emitted when an organization is deleted.
type OrganizationDeletedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string    `json:"organization_id"`
	DeletedAt      time.Time `json:"deleted_at"`
}

func (e OrganizationDeletedEvent) EventType() string { return EventTypeOrganizationDeleted }

// MemberAddedEvent is emitted when a user joins an organization.
type MemberAddedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string    `json:"organization_id"`
	MemberID       string    `json:"member_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

func (e MemberAddedEvent) EventType() string { return EventTypeMemberAdded }

// MemberRemovedEvent is emitted when a member is removed from an organization.
type MemberRemovedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string    `json:"organization_id"`
	MemberID       string    `json:"member_id"`
	UserID         string    `json:"user_id"`
	Reason         string    `json:"reason,omitempty"`
	RemovedAt      time.Time `json:"removed_at"`
}

func (e MemberRemovedEvent) EventType() string { return EventTypeMemberRemoved }

// MemberRoleChangedEvent is emitted when a member's role changes.
type MemberRoleChangedEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string    `json:"organization_id"`
	MemberID       string    `json:"member_id"`
	OldRole        string    `json:"old_role"`
	NewRole        string    `json:"new_role"`
	ChangedAt      time.Time `json:"changed_at"`
}

func (e MemberRoleChangedEvent) EventType() string { return EventTypeMemberRoleChanged }

// OwnershipTransferredEvent is emitted when ownership is transferred.
type OwnershipTransferredEvent struct {
	eventsourcing.BaseEvent

	OrganizationID string    `json:"organization_id"`
	FromMemberID   string    `json:"from_member_id"`
	ToMemberID     string    `json:"to_member_id"`
	TransferredAt  time.Time `json:"transferred_at"`
}

func (e OwnershipTransferredEvent) EventType() string { return EventTypeOwnershipTransferred }

// ============================================================================
// Event Registration
// ============================================================================

// RegisterOrganizationEvents registers all Organization domain events.
func RegisterOrganizationEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeOrganizationCreated, func() eventsourcing.Event { return &OrganizationCreatedEvent{} })
	registry.Register(EventTypeOrganizationUpdated, func() eventsourcing.Event { return &OrganizationUpdatedEvent{} })
	registry.Register(EventTypeOrganizationVerified, func() eventsourcing.Event { return &OrganizationVerifiedEvent{} })
	registry.Register(EventTypeOrganizationDeleted, func() eventsourcing.Event { return &OrganizationDeletedEvent{} })
	registry.Register(EventTypeMemberAdded, func() eventsourcing.Event { return &MemberAddedEvent{} })
	registry.Register(EventTypeMemberRemoved, func() eventsourcing.Event { return &MemberRemovedEvent{} })
	registry.Register(EventTypeMemberRoleChanged, func() eventsourcing.Event { return &MemberRoleChangedEvent{} })
	registry.Register(EventTypeOwnershipTransferred, func() eventsourcing.Event { return &OwnershipTransferredEvent{} })
}

func init() {
	RegisterOrganizationEvents(eventsourcing.DefaultRegistry)
}

// newOrganizationBaseEvent creates a base event for the organization aggregate.
func newOrganizationBaseEvent(id OrganizationID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeOrganization, id.String())
}
