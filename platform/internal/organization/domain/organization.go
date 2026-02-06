package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Organization is the aggregate root for organizational entities.
// It manages teams, membership, roles, verification status, and organizational DIDs.
type Organization struct {
	eventsourcing.AggregateRoot

	id                 OrganizationID
	name               string
	slug               string
	orgType            OrganizationType
	description        string
	verificationStatus VerificationStatus
	did                string
	ownerMemberID      MemberID
	members            map[string]Member // keyed by MemberID string
	invitations        map[string]Invitation
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          time.Time
	deleted            bool
}

// ============================================================================
// Constructors
// ============================================================================

// CreateOrganization creates a new Organization aggregate.
func CreateOrganization(
	id OrganizationID,
	name string,
	slug string,
	orgType OrganizationType,
	ownerUserID string,
	ownerMemberID MemberID,
) (*Organization, error) {
	if id.IsZero() {
		return nil, OrganizationInvalid("Organization.Create", "organization ID is required")
	}
	if name == "" {
		return nil, OrganizationInvalid("Organization.Create", "name is required")
	}
	if slug == "" {
		return nil, OrganizationInvalid("Organization.Create", "slug is required")
	}
	if !orgType.IsValid() {
		return nil, OrganizationInvalid("Organization.Create", "invalid organization type")
	}
	if ownerUserID == "" {
		return nil, OrganizationInvalid("Organization.Create", "owner user ID is required")
	}
	if ownerMemberID.IsZero() {
		return nil, OrganizationInvalid("Organization.Create", "owner member ID is required")
	}

	o := &Organization{
		members:     make(map[string]Member),
		invitations: make(map[string]Invitation),
	}
	o.InitAggregate(AggregateTypeOrganization, id.String())

	now := time.Now().UTC()

	o.Raise(o, &OrganizationCreatedEvent{
		BaseEvent:      newOrganizationBaseEvent(id),
		OrganizationID: id.String(),
		Name:           name,
		Slug:           slug,
		OrgType:        orgType.String(),
		OwnerUserID:    ownerUserID,
		OwnerMemberID:  ownerMemberID.String(),
		CreatedAt:      now,
	})

	return o, nil
}

// NewOrganizationFromEvents reconstructs an Organization from events (for hydration).
func NewOrganizationFromEvents(id string) *Organization {
	o := &Organization{
		members:     make(map[string]Member),
		invitations: make(map[string]Invitation),
	}
	o.InitAggregate(AggregateTypeOrganization, id)
	return o
}

// OrganizationFactory creates a factory for Organization aggregates.
func OrganizationFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewOrganizationFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

func (o *Organization) ID() OrganizationID             { return o.id }
func (o *Organization) Name() string                    { return o.name }
func (o *Organization) Slug() string                    { return o.slug }
func (o *Organization) OrgType() OrganizationType       { return o.orgType }
func (o *Organization) Description() string             { return o.description }
func (o *Organization) VerificationStatus() VerificationStatus {
	return o.verificationStatus
}
func (o *Organization) DID() string                     { return o.did }
func (o *Organization) OwnerMemberID() MemberID         { return o.ownerMemberID }
func (o *Organization) CreatedAt() time.Time            { return o.createdAt }
func (o *Organization) UpdatedAt() time.Time            { return o.updatedAt }
func (o *Organization) IsDeleted() bool                 { return o.deleted }
func (o *Organization) IsVerified() bool                { return o.verificationStatus.IsVerified() }

// MemberCount returns the number of active members.
func (o *Organization) MemberCount() int { return len(o.members) }

// Members returns a copy of all members.
func (o *Organization) Members() []Member {
	members := make([]Member, 0, len(o.members))
	for _, m := range o.members {
		members = append(members, m)
	}
	return members
}

// GetMember returns a member by ID, or false if not found.
func (o *Organization) GetMember(id MemberID) (Member, bool) {
	m, ok := o.members[id.String()]
	return m, ok
}

// GetMemberByUserID returns the member for a given user ID, or false if not found.
func (o *Organization) GetMemberByUserID(userID string) (Member, bool) {
	for _, m := range o.members {
		if m.UserID() == userID {
			return m, true
		}
	}
	return Member{}, false
}

// HasMemberWithUserID returns true if a user is already a member.
func (o *Organization) HasMemberWithUserID(userID string) bool {
	_, ok := o.GetMemberByUserID(userID)
	return ok
}

// ============================================================================
// Command Methods
// ============================================================================

// UpdateMetadata updates the organization's name and/or description.
func (o *Organization) UpdateMetadata(name string, description string) error {
	if o.deleted {
		return OrganizationInvalid("Organization.UpdateMetadata", "organization is deleted")
	}

	o.Raise(o, &OrganizationUpdatedEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		Name:           name,
		Description:    description,
		UpdatedAt:      time.Now().UTC(),
	})

	return nil
}

// AddMember adds a new member to the organization.
func (o *Organization) AddMember(memberID MemberID, userID string, role Role) error {
	if o.deleted {
		return OrganizationInvalid("Organization.AddMember", "organization is deleted")
	}
	if memberID.IsZero() {
		return OrganizationInvalid("Organization.AddMember", "member ID is required")
	}
	if userID == "" {
		return OrganizationInvalid("Organization.AddMember", "user ID is required")
	}
	if o.HasMemberWithUserID(userID) {
		return MemberAlreadyExists("Organization.AddMember", userID)
	}

	o.Raise(o, &MemberAddedEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		MemberID:       memberID.String(),
		UserID:         userID,
		Role:           role.String(),
		JoinedAt:       time.Now().UTC(),
	})

	return nil
}

// RemoveMember removes a member from the organization.
func (o *Organization) RemoveMember(memberID MemberID, reason string) error {
	if o.deleted {
		return OrganizationInvalid("Organization.RemoveMember", "organization is deleted")
	}

	member, ok := o.GetMember(memberID)
	if !ok {
		return MemberNotFound("Organization.RemoveMember", memberID.String())
	}

	if member.IsOwner() {
		return OwnerCannotLeave("Organization.RemoveMember", o.id.String())
	}

	o.Raise(o, &MemberRemovedEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		MemberID:       memberID.String(),
		UserID:         member.UserID(),
		Reason:         reason,
		RemovedAt:      time.Now().UTC(),
	})

	return nil
}

// ChangeMemberRole changes a member's role.
func (o *Organization) ChangeMemberRole(memberID MemberID, newRole Role) error {
	if o.deleted {
		return OrganizationInvalid("Organization.ChangeMemberRole", "organization is deleted")
	}

	member, ok := o.GetMember(memberID)
	if !ok {
		return MemberNotFound("Organization.ChangeMemberRole", memberID.String())
	}

	if member.IsOwner() {
		return OrganizationInvalid("Organization.ChangeMemberRole", "cannot change owner role directly; use TransferOwnership")
	}

	if newRole == RoleOwner {
		return OrganizationInvalid("Organization.ChangeMemberRole", "cannot promote to owner; use TransferOwnership")
	}

	o.Raise(o, &MemberRoleChangedEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		MemberID:       memberID.String(),
		OldRole:        member.Role().String(),
		NewRole:        newRole.String(),
		ChangedAt:      time.Now().UTC(),
	})

	return nil
}

// TransferOwnership transfers ownership to another member.
func (o *Organization) TransferOwnership(toMemberID MemberID) error {
	if o.deleted {
		return OrganizationInvalid("Organization.TransferOwnership", "organization is deleted")
	}

	_, ok := o.GetMember(toMemberID)
	if !ok {
		return MemberNotFound("Organization.TransferOwnership", toMemberID.String())
	}

	if o.ownerMemberID.Equals(toMemberID) {
		return OrganizationInvalid("Organization.TransferOwnership", "member is already the owner")
	}

	o.Raise(o, &OwnershipTransferredEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		FromMemberID:   o.ownerMemberID.String(),
		ToMemberID:     toMemberID.String(),
		TransferredAt:  time.Now().UTC(),
	})

	return nil
}

// CompleteVerification marks the organization as verified and assigns a DID.
func (o *Organization) CompleteVerification(did string) error {
	if o.deleted {
		return OrganizationInvalid("Organization.CompleteVerification", "organization is deleted")
	}
	if !o.verificationStatus.CanTransitionTo(VerificationStatusVerified) {
		return OrganizationInvalid("Organization.CompleteVerification",
			"cannot verify from current status: "+o.verificationStatus.String())
	}

	o.Raise(o, &OrganizationVerifiedEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		DID:            did,
		VerifiedAt:     time.Now().UTC(),
	})

	return nil
}

// Delete soft-deletes the organization.
func (o *Organization) Delete() error {
	if o.deleted {
		return OrganizationInvalid("Organization.Delete", "organization is already deleted")
	}

	o.Raise(o, &OrganizationDeletedEvent{
		BaseEvent:      newOrganizationBaseEvent(o.id),
		OrganizationID: o.id.String(),
		DeletedAt:      time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (o *Organization) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *OrganizationCreatedEvent:
		o.onOrganizationCreated(e)
	case *OrganizationUpdatedEvent:
		o.onOrganizationUpdated(e)
	case *OrganizationVerifiedEvent:
		o.onOrganizationVerified(e)
	case *OrganizationDeletedEvent:
		o.onOrganizationDeleted(e)
	case *MemberAddedEvent:
		o.onMemberAdded(e)
	case *MemberRemovedEvent:
		o.onMemberRemoved(e)
	case *MemberRoleChangedEvent:
		o.onMemberRoleChanged(e)
	case *OwnershipTransferredEvent:
		o.onOwnershipTransferred(e)
	}
}

func (o *Organization) onOrganizationCreated(e *OrganizationCreatedEvent) {
	var err error
	o.id, err = ParseOrganizationID(e.OrganizationID)
	if err != nil {
		panic("corrupt event store: OrganizationCreated has invalid OrganizationID: " + e.OrganizationID)
	}
	o.orgType, err = ParseOrganizationType(e.OrgType)
	if err != nil {
		panic("corrupt event store: OrganizationCreated has invalid OrgType: " + e.OrgType)
	}
	ownerMemberID, err := ParseMemberID(e.OwnerMemberID)
	if err != nil {
		panic("corrupt event store: OrganizationCreated has invalid OwnerMemberID: " + e.OwnerMemberID)
	}

	o.name = e.Name
	o.slug = e.Slug
	o.verificationStatus = VerificationStatusUnverified
	o.ownerMemberID = ownerMemberID
	o.createdAt = e.CreatedAt
	o.updatedAt = e.CreatedAt

	// Auto-add owner as first member.
	o.members[ownerMemberID.String()] = NewMember(ownerMemberID, e.OwnerUserID, RoleOwner, e.CreatedAt)
}

func (o *Organization) onOrganizationUpdated(e *OrganizationUpdatedEvent) {
	if e.Name != "" {
		o.name = e.Name
	}
	if e.Description != "" {
		o.description = e.Description
	}
	o.updatedAt = e.UpdatedAt
}

func (o *Organization) onOrganizationVerified(e *OrganizationVerifiedEvent) {
	o.verificationStatus = VerificationStatusVerified
	o.did = e.DID
	o.updatedAt = e.VerifiedAt
}

func (o *Organization) onOrganizationDeleted(e *OrganizationDeletedEvent) {
	o.deleted = true
	o.deletedAt = e.DeletedAt
	o.updatedAt = e.DeletedAt
}

func (o *Organization) onMemberAdded(e *MemberAddedEvent) {
	memberID, err := ParseMemberID(e.MemberID)
	if err != nil {
		panic("corrupt event store: MemberAdded has invalid MemberID: " + e.MemberID)
	}
	role, err := ParseRole(e.Role)
	if err != nil {
		panic("corrupt event store: MemberAdded has invalid Role: " + e.Role)
	}

	o.members[e.MemberID] = NewMember(memberID, e.UserID, role, e.JoinedAt)
	o.updatedAt = e.JoinedAt
}

func (o *Organization) onMemberRemoved(e *MemberRemovedEvent) {
	delete(o.members, e.MemberID)
	o.updatedAt = e.RemovedAt
}

func (o *Organization) onMemberRoleChanged(e *MemberRoleChangedEvent) {
	newRole, err := ParseRole(e.NewRole)
	if err != nil {
		panic("corrupt event store: MemberRoleChanged has invalid NewRole: " + e.NewRole)
	}

	if existing, ok := o.members[e.MemberID]; ok {
		o.members[e.MemberID] = NewMember(existing.ID(), existing.UserID(), newRole, existing.JoinedAt())
	}
	o.updatedAt = e.ChangedAt
}

func (o *Organization) onOwnershipTransferred(e *OwnershipTransferredEvent) {
	toMemberID, err := ParseMemberID(e.ToMemberID)
	if err != nil {
		panic("corrupt event store: OwnershipTransferred has invalid ToMemberID: " + e.ToMemberID)
	}

	// Demote old owner to admin.
	if old, ok := o.members[e.FromMemberID]; ok {
		o.members[e.FromMemberID] = NewMember(old.ID(), old.UserID(), RoleAdmin, old.JoinedAt())
	}

	// Promote new owner.
	if newOwner, ok := o.members[e.ToMemberID]; ok {
		o.members[e.ToMemberID] = NewMember(newOwner.ID(), newOwner.UserID(), RoleOwner, newOwner.JoinedAt())
	}

	o.ownerMemberID = toMemberID
	o.updatedAt = e.TransferredAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (o *Organization) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &o.AggregateRoot
}
