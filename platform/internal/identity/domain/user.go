package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// User is the aggregate root for user identity.
// It manages user profile, authentication methods, DIDs, and linked OAuth accounts.
type User struct {
	eventsourcing.AggregateRoot

	id          UserID
	email       types.Email
	displayName DisplayName
	status      UserStatus
	primaryDID  string
	dids        []string
	oauthLinks  []OAuthSubject
	createdAt   time.Time
	updatedAt   time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// NewUser creates a new User aggregate for registration.
func NewUser(
	id UserID,
	email types.Email,
	displayName DisplayName,
	authMethod AuthMethod,
	primaryDID string,
) (*User, error) {
	if id.IsZero() {
		return nil, InvalidUserID("User.New", "")
	}

	if primaryDID == "" {
		return nil, PrimaryDIDCannotChange("User.New").
			WithMessage("primary DID is required")
	}

	u := &User{}
	u.InitAggregate(AggregateTypeUser, id.String())

	now := time.Now().UTC()

	u.Raise(u, &UserRegisteredEvent{
		BaseEvent:    newUserBaseEvent(id),
		UserID:       id.String(),
		Email:        email.String(),
		DisplayName:  displayName.String(),
		AuthMethod:   authMethod.String(),
		PrimaryDID:   primaryDID,
		RegisteredAt: now,
	})

	return u, nil
}

// NewUserFromEvents reconstructs a User from events (for hydration).
func NewUserFromEvents(id string) *User {
	u := &User{}
	u.InitAggregate(AggregateTypeUser, id)
	return u
}

// UserFactory creates a factory for User aggregates.
func UserFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewUserFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the user's ID.
func (u *User) ID() UserID {
	return u.id
}

// Email returns the user's email.
func (u *User) Email() types.Email {
	return u.email
}

// DisplayName returns the user's display name.
func (u *User) DisplayName() DisplayName {
	return u.displayName
}

// Status returns the user's status.
func (u *User) Status() UserStatus {
	return u.status
}

// PrimaryDID returns the user's primary DID.
func (u *User) PrimaryDID() string {
	return u.primaryDID
}

// DIDs returns all DIDs associated with the user.
func (u *User) DIDs() []string {
	result := make([]string, len(u.dids))
	copy(result, u.dids)
	return result
}

// OAuthLinks returns all linked OAuth accounts.
func (u *User) OAuthLinks() []OAuthSubject {
	result := make([]OAuthSubject, len(u.oauthLinks))
	copy(result, u.oauthLinks)
	return result
}

// CreatedAt returns when the user was created.
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns when the user was last updated.
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// ============================================================================
// Query Methods
// ============================================================================

// IsActive returns true if the user account is active.
func (u *User) IsActive() bool {
	return u.status.IsActive()
}

// IsPending returns true if the user account is pending activation.
func (u *User) IsPending() bool {
	return u.status.IsPending()
}

// IsSuspended returns true if the user account is suspended.
func (u *User) IsSuspended() bool {
	return u.status.IsSuspended()
}

// CanAuthenticate returns true if the user is allowed to authenticate.
func (u *User) CanAuthenticate() bool {
	return u.status.CanAuthenticate()
}

// HasDID returns true if the user has the specified DID.
func (u *User) HasDID(did string) bool {
	for _, d := range u.dids {
		if d == did {
			return true
		}
	}
	return false
}

// HasOAuthLink returns true if the user has linked the specified OAuth provider.
func (u *User) HasOAuthLink(provider OAuthProvider) bool {
	for _, link := range u.oauthLinks {
		if link.Provider() == provider {
			return true
		}
	}
	return false
}

// GetOAuthLink returns the OAuth link for the specified provider, if it exists.
func (u *User) GetOAuthLink(provider OAuthProvider) (OAuthSubject, bool) {
	for _, link := range u.oauthLinks {
		if link.Provider() == provider {
			return link, true
		}
	}
	return OAuthSubject{}, false
}

// ============================================================================
// Command Methods
// ============================================================================

// Activate activates a pending user account.
func (u *User) Activate() error {
	if !u.status.CanTransitionTo(UserStatusActive) {
		return UserNotActive("User.Activate", u.id.String()).
			WithMessage("cannot activate user in current status: " + u.status.String())
	}

	u.Raise(u, &UserActivatedEvent{
		BaseEvent:   newUserBaseEvent(u.id),
		UserID:      u.id.String(),
		ActivatedAt: time.Now().UTC(),
	})

	return nil
}

// Suspend suspends the user account.
func (u *User) Suspend(reason string) error {
	if !u.status.CanTransitionTo(UserStatusSuspended) {
		return UserSuspended("User.Suspend", u.id.String()).
			WithMessage("cannot suspend user in current status: " + u.status.String())
	}

	u.Raise(u, &UserSuspendedEvent{
		BaseEvent:   newUserBaseEvent(u.id),
		UserID:      u.id.String(),
		Reason:      reason,
		SuspendedAt: time.Now().UTC(),
	})

	return nil
}

// Reactivate reactivates a suspended user account.
func (u *User) Reactivate() error {
	if !u.status.CanTransitionTo(UserStatusActive) {
		return UserNotActive("User.Reactivate", u.id.String()).
			WithMessage("cannot reactivate user in current status: " + u.status.String())
	}

	u.Raise(u, &UserReactivatedEvent{
		BaseEvent:     newUserBaseEvent(u.id),
		UserID:        u.id.String(),
		ReactivatedAt: time.Now().UTC(),
	})

	return nil
}

// Delete soft-deletes the user account.
func (u *User) Delete() error {
	if !u.status.CanTransitionTo(UserStatusDeleted) {
		return UserNotActive("User.Delete", u.id.String()).
			WithMessage("cannot delete user in current status: " + u.status.String())
	}

	u.Raise(u, &UserDeletedEvent{
		BaseEvent: newUserBaseEvent(u.id),
		UserID:    u.id.String(),
		DeletedAt: time.Now().UTC(),
	})

	return nil
}

// ChangeDisplayName changes the user's display name.
func (u *User) ChangeDisplayName(newName DisplayName) error {
	if !u.IsActive() {
		return UserNotActive("User.ChangeDisplayName", u.id.String())
	}

	if u.displayName.Equals(newName) {
		return nil // No change needed
	}

	u.Raise(u, &UserDisplayNameChangedEvent{
		BaseEvent:      newUserBaseEvent(u.id),
		UserID:         u.id.String(),
		OldDisplayName: u.displayName.String(),
		NewDisplayName: newName.String(),
		ChangedAt:      time.Now().UTC(),
	})

	return nil
}

// ChangeEmail changes the user's email address.
func (u *User) ChangeEmail(newEmail types.Email) error {
	if !u.IsActive() {
		return UserNotActive("User.ChangeEmail", u.id.String())
	}

	if u.email.Equals(newEmail) {
		return nil // No change needed
	}

	u.Raise(u, &UserEmailChangedEvent{
		BaseEvent: newUserBaseEvent(u.id),
		UserID:    u.id.String(),
		OldEmail:  u.email.String(),
		NewEmail:  newEmail.String(),
		ChangedAt: time.Now().UTC(),
	})

	return nil
}

// AddDID adds a DID to the user's identity.
func (u *User) AddDID(did string) error {
	if !u.IsActive() {
		return UserNotActive("User.AddDID", u.id.String())
	}

	if did == "" {
		return DIDNotFound("User.AddDID", did).
			WithMessage("DID cannot be empty")
	}

	if u.HasDID(did) {
		return DIDAlreadyLinked("User.AddDID", did)
	}

	u.Raise(u, &UserDIDAddedEvent{
		BaseEvent: newUserBaseEvent(u.id),
		UserID:    u.id.String(),
		DID:       did,
		AddedAt:   time.Now().UTC(),
	})

	return nil
}

// RemoveDID removes a DID from the user's identity.
func (u *User) RemoveDID(did string) error {
	if !u.IsActive() {
		return UserNotActive("User.RemoveDID", u.id.String())
	}

	if did == u.primaryDID {
		return PrimaryDIDCannotChange("User.RemoveDID")
	}

	if !u.HasDID(did) {
		return DIDNotFound("User.RemoveDID", did)
	}

	u.Raise(u, &UserDIDRemovedEvent{
		BaseEvent: newUserBaseEvent(u.id),
		UserID:    u.id.String(),
		DID:       did,
		RemovedAt: time.Now().UTC(),
	})

	return nil
}

// LinkOAuth links an OAuth account to the user.
func (u *User) LinkOAuth(subject OAuthSubject, email string) error {
	if !u.IsActive() {
		return UserNotActive("User.LinkOAuth", u.id.String())
	}

	if !subject.IsValid() {
		return OAuthProviderNotSupported("User.LinkOAuth", subject.Provider().String())
	}

	if u.HasOAuthLink(subject.Provider()) {
		return OAuthAccountLinked("User.LinkOAuth", subject.Provider().String())
	}

	u.Raise(u, &UserOAuthLinkedEvent{
		BaseEvent:  newUserBaseEvent(u.id),
		UserID:     u.id.String(),
		Provider:   subject.Provider().String(),
		ExternalID: subject.ExternalID(),
		Email:      email,
		LinkedAt:   time.Now().UTC(),
	})

	return nil
}

// UnlinkOAuth unlinks an OAuth account from the user.
func (u *User) UnlinkOAuth(provider OAuthProvider) error {
	if !u.IsActive() {
		return UserNotActive("User.UnlinkOAuth", u.id.String())
	}

	link, found := u.GetOAuthLink(provider)
	if !found {
		return OAuthProviderNotSupported("User.UnlinkOAuth", provider.String()).
			WithMessage("oauth provider not linked")
	}

	u.Raise(u, &UserOAuthUnlinkedEvent{
		BaseEvent:  newUserBaseEvent(u.id),
		UserID:     u.id.String(),
		Provider:   provider.String(),
		ExternalID: link.ExternalID(),
		UnlinkedAt: time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (u *User) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *UserRegisteredEvent:
		u.onUserRegistered(e)
	case *UserActivatedEvent:
		u.onUserActivated(e)
	case *UserSuspendedEvent:
		u.onUserSuspended(e)
	case *UserReactivatedEvent:
		u.onUserReactivated(e)
	case *UserDeletedEvent:
		u.onUserDeleted(e)
	case *UserDisplayNameChangedEvent:
		u.onUserDisplayNameChanged(e)
	case *UserEmailChangedEvent:
		u.onUserEmailChanged(e)
	case *UserDIDAddedEvent:
		u.onUserDIDAdded(e)
	case *UserDIDRemovedEvent:
		u.onUserDIDRemoved(e)
	case *UserOAuthLinkedEvent:
		u.onUserOAuthLinked(e)
	case *UserOAuthUnlinkedEvent:
		u.onUserOAuthUnlinked(e)
	}
}

func (u *User) onUserRegistered(e *UserRegisteredEvent) {
	var err error
	u.id, err = ParseUserID(e.UserID)
	if err != nil {
		panic("corrupt event store: UserRegistered has invalid UserID: " + e.UserID)
	}
	if e.Email != "" {
		u.email, err = types.NewEmail(e.Email)
		if err != nil {
			panic("corrupt event store: UserRegistered has invalid Email: " + e.Email)
		}
	}
	u.displayName, err = NewDisplayName(e.DisplayName)
	if err != nil {
		panic("corrupt event store: UserRegistered has invalid DisplayName: " + e.DisplayName)
	}
	u.status = UserStatusPending
	u.primaryDID = e.PrimaryDID
	u.dids = []string{e.PrimaryDID}
	u.oauthLinks = make([]OAuthSubject, 0)
	u.createdAt = e.RegisteredAt
	u.updatedAt = e.RegisteredAt
}

func (u *User) onUserActivated(e *UserActivatedEvent) {
	u.status = UserStatusActive
	u.updatedAt = e.ActivatedAt
}

func (u *User) onUserSuspended(e *UserSuspendedEvent) {
	u.status = UserStatusSuspended
	u.updatedAt = e.SuspendedAt
}

func (u *User) onUserReactivated(e *UserReactivatedEvent) {
	u.status = UserStatusActive
	u.updatedAt = e.ReactivatedAt
}

func (u *User) onUserDeleted(e *UserDeletedEvent) {
	u.status = UserStatusDeleted
	u.updatedAt = e.DeletedAt
}

func (u *User) onUserDisplayNameChanged(e *UserDisplayNameChangedEvent) {
	var err error
	u.displayName, err = NewDisplayName(e.NewDisplayName)
	if err != nil {
		panic("corrupt event store: UserDisplayNameChanged has invalid DisplayName: " + e.NewDisplayName)
	}
	u.updatedAt = e.ChangedAt
}

func (u *User) onUserEmailChanged(e *UserEmailChangedEvent) {
	var err error
	u.email, err = types.NewEmail(e.NewEmail)
	if err != nil {
		panic("corrupt event store: UserEmailChanged has invalid Email: " + e.NewEmail)
	}
	u.updatedAt = e.ChangedAt
}

func (u *User) onUserDIDAdded(e *UserDIDAddedEvent) {
	u.dids = append(u.dids, e.DID)
	u.updatedAt = e.AddedAt
}

func (u *User) onUserDIDRemoved(e *UserDIDRemovedEvent) {
	newDIDs := make([]string, 0, len(u.dids)-1)
	for _, did := range u.dids {
		if did != e.DID {
			newDIDs = append(newDIDs, did)
		}
	}
	u.dids = newDIDs
	u.updatedAt = e.RemovedAt
}

func (u *User) onUserOAuthLinked(e *UserOAuthLinkedEvent) {
	provider, err := ParseOAuthProvider(e.Provider)
	if err != nil {
		panic("corrupt event store: UserOAuthLinked has invalid Provider: " + e.Provider)
	}
	subject, err := NewOAuthSubject(provider, e.ExternalID)
	if err != nil {
		panic("corrupt event store: UserOAuthLinked has invalid OAuthSubject: " + e.Provider + ":" + e.ExternalID)
	}
	u.oauthLinks = append(u.oauthLinks, subject)
	u.updatedAt = e.LinkedAt
}

func (u *User) onUserOAuthUnlinked(e *UserOAuthUnlinkedEvent) {
	provider, err := ParseOAuthProvider(e.Provider)
	if err != nil {
		panic("corrupt event store: UserOAuthUnlinked has invalid Provider: " + e.Provider)
	}
	newLinks := make([]OAuthSubject, 0, len(u.oauthLinks)-1)
	for _, link := range u.oauthLinks {
		if link.Provider() != provider {
			newLinks = append(newLinks, link)
		}
	}
	u.oauthLinks = newLinks
	u.updatedAt = e.UnlinkedAt
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (u *User) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &u.AggregateRoot
}
