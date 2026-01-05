package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

const (
	// UserAggregateType is the type name for User aggregates.
	UserAggregateType = "identity.User"
)

// ============================================================================
// User Aggregate
// ============================================================================

// User represents a self-sovereign identity in the system.
// The primary identifier is a DID, not email or username.
// Users can link additional identities (wallets, email, OAuth) for convenience.
type User struct {
	eventsourcing.AggregateRoot

	// Primary identity
	primaryDID did.DID

	// Status
	status UserStatus

	// Linked identities (wallets, email, OAuth connections)
	linkedIdentities []LinkedIdentity

	// Wallet addresses (for multi-chain support)
	wallets []WalletAddress

	// Metadata
	createdAt       time.Time
	updatedAt       time.Time
	lastLoginAt     time.Time
	lastLoginMethod AuthMethod
}

// ============================================================================
// Constructor
// ============================================================================

// NewUser creates a new user with a primary DID.
func NewUser(id string, primaryDID did.DID) (*User, error) {
	if id == "" {
		return nil, ErrInvalidCredentials("User.New")
	}

	if primaryDID.IsZero() {
		return nil, ErrInvalidCredentials("User.New")
	}

	now := time.Now()

	user := &User{
		primaryDID:       primaryDID,
		status:           UserStatusActive,
		linkedIdentities: make([]LinkedIdentity, 0),
		wallets:          make([]WalletAddress, 0),
		createdAt:        now,
		updatedAt:        now,
	}

	user.InitAggregate(UserAggregateType, id)

	// Raise creation event
	user.Raise(user, NewUserCreatedEvent(id, primaryDID.String()))

	return user, nil
}

// NewUserFromWallet creates a new user from a wallet signature.
// Generates a did:pkh from the wallet address as the primary DID.
func NewUserFromWallet(id string, wallet WalletAddress) (*User, error) {
	if id == "" {
		return nil, ErrInvalidCredentials("User.NewFromWallet")
	}

	// Create did:pkh from wallet address
	primaryDID, err := did.Parse(wallet.ToDID())
	if err != nil {
		return nil, ErrInvalidCredentials("User.NewFromWallet")
	}

	user, err := NewUser(id, primaryDID)
	if err != nil {
		return nil, err
	}

	// Add wallet to linked wallets
	user.wallets = append(user.wallets, wallet)

	// Add wallet as linked identity
	user.linkedIdentities = append(user.linkedIdentities, NewWalletIdentity(wallet.Address, wallet.Chain))

	return user, nil
}

// NewUserFromEmail creates a new user from an email address.
// Generates a did:key as the primary DID since email doesn't have a native DID.
func NewUserFromEmail(id string, email string) (*User, error) {
	if id == "" {
		return nil, ErrInvalidCredentials("User.NewFromEmail")
	}

	if email == "" {
		return nil, ErrInvalidEmail("User.NewFromEmail", email)
	}

	// Generate a did:key for email users
	// In production, this would use proper key generation
	primaryDID, err := did.Parse("did:key:" + id)
	if err != nil {
		return nil, ErrInvalidCredentials("User.NewFromEmail")
	}

	user, err := NewUser(id, primaryDID)
	if err != nil {
		return nil, err
	}

	// Add email as linked identity (verified since they clicked magic link)
	if err := user.LinkEmail(email, true); err != nil {
		return nil, err
	}

	return user, nil
}

// Reconstitute creates a User from persisted data (no events raised).
func Reconstitute(
	id string,
	version int,
	primaryDID did.DID,
	status UserStatus,
	linkedIdentities []LinkedIdentity,
	wallets []WalletAddress,
	createdAt time.Time,
	updatedAt time.Time,
	lastLoginAt time.Time,
	lastLoginMethod AuthMethod,
) *User {
	user := &User{
		primaryDID:       primaryDID,
		status:           status,
		linkedIdentities: linkedIdentities,
		wallets:          wallets,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		lastLoginAt:      lastLoginAt,
		lastLoginMethod:  lastLoginMethod,
	}
	user.InitAggregate(UserAggregateType, id)
	user.SetVersion(version)
	return user
}

// ============================================================================
// Aggregate Interface Implementation
// ============================================================================

// ApplyEvent applies an event to update aggregate state.
func (u *User) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *UserCreatedEvent:
		u.applyUserCreated(e)
	case *UserLoggedInEvent:
		u.applyUserLoggedIn(e)
	case *UserSuspendedEvent:
		u.applyUserSuspended(e)
	case *UserActivatedEvent:
		u.applyUserActivated(e)
	case *IdentityLinkedEvent:
		u.applyIdentityLinked(e)
	case *EmailVerifiedEvent:
		u.applyEmailVerified(e)
	case *WalletLinkedEvent:
		u.applyWalletLinked(e)
	case *WalletUnlinkedEvent:
		u.applyWalletUnlinked(e)
	case *OAuthLinkedEvent:
		u.applyOAuthLinked(e)
	case *OAuthUnlinkedEvent:
		u.applyOAuthUnlinked(e)
	case *PrimaryDIDUpdatedEvent:
		u.applyPrimaryDIDUpdated(e)
	}
}

func (u *User) applyUserCreated(e *UserCreatedEvent) {
	parsedDID, _ := did.Parse(e.PrimaryDID)
	u.primaryDID = parsedDID
	u.status = UserStatusActive
	u.createdAt = e.OccurredAt()
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyUserLoggedIn(e *UserLoggedInEvent) {
	u.lastLoginAt = e.OccurredAt()
	u.lastLoginMethod = AuthMethod(e.Method)
}

func (u *User) applyUserSuspended(e *UserSuspendedEvent) {
	u.status = UserStatusSuspended
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyUserActivated(e *UserActivatedEvent) {
	u.status = UserStatusActive
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyIdentityLinked(e *IdentityLinkedEvent) {
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyEmailVerified(e *EmailVerifiedEvent) {
	for i, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && identity.Value == e.Email {
			u.linkedIdentities[i].Verified = true
			u.linkedIdentities[i].VerifiedAt = e.OccurredAt()
			break
		}
	}
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyWalletLinked(e *WalletLinkedEvent) {
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyWalletUnlinked(e *WalletUnlinkedEvent) {
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyOAuthLinked(e *OAuthLinkedEvent) {
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyOAuthUnlinked(e *OAuthUnlinkedEvent) {
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyPrimaryDIDUpdated(e *PrimaryDIDUpdatedEvent) {
	parsedDID, _ := did.Parse(e.NewDID)
	u.primaryDID = parsedDID
	u.updatedAt = e.OccurredAt()
}

// GetAggregateRoot returns the embedded AggregateRoot.
// Required for eventsourcing.Hydrate to work correctly.
func (u *User) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &u.AggregateRoot
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the user's unique identifier.
func (u *User) ID() string {
	return u.AggregateID()
}

// PrimaryDID returns the user's primary decentralized identifier.
func (u *User) PrimaryDID() did.DID {
	return u.primaryDID
}

// Status returns the user's status.
func (u *User) Status() UserStatus {
	return u.status
}

// LinkedIdentities returns all linked identities.
func (u *User) LinkedIdentities() []LinkedIdentity {
	result := make([]LinkedIdentity, len(u.linkedIdentities))
	copy(result, u.linkedIdentities)
	return result
}

// Wallets returns all linked wallet addresses.
func (u *User) Wallets() []WalletAddress {
	result := make([]WalletAddress, len(u.wallets))
	copy(result, u.wallets)
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

// LastLoginAt returns when the user last logged in.
func (u *User) LastLoginAt() time.Time {
	return u.lastLoginAt
}

// LastLoginMethod returns the method used for last login.
func (u *User) LastLoginMethod() AuthMethod {
	return u.lastLoginMethod
}

// ============================================================================
// Identity Queries
// ============================================================================

// HasLinkedEmail returns true if the user has a verified email linked.
func (u *User) HasLinkedEmail() bool {
	for _, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && identity.Verified {
			return true
		}
	}
	return false
}

// LinkedEmail returns the first verified linked email, if any.
func (u *User) LinkedEmail() (string, bool) {
	for _, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && identity.Verified {
			return identity.Value, true
		}
	}
	return "", false
}

// HasLinkedWallet returns true if the user has a wallet on the specified chain.
func (u *User) HasLinkedWallet(chain Chain) bool {
	for _, wallet := range u.wallets {
		if wallet.Chain == chain {
			return true
		}
	}
	return false
}

// WalletForChain returns the wallet address for a specific chain, if any.
func (u *User) WalletForChain(chain Chain) (WalletAddress, bool) {
	for _, wallet := range u.wallets {
		if wallet.Chain == chain {
			return wallet, true
		}
	}
	return WalletAddress{}, false
}

// HasOAuthConnection returns true if the user has connected the specified provider.
func (u *User) HasOAuthConnection(provider OAuthProvider) bool {
	for _, identity := range u.linkedIdentities {
		if identity.Provider == provider.String() {
			return true
		}
	}
	return false
}

// OAuthConnection returns the OAuth connection for a provider, if any.
func (u *User) OAuthConnection(provider OAuthProvider) (LinkedIdentity, bool) {
	for _, identity := range u.linkedIdentities {
		if identity.Provider == provider.String() {
			return identity, true
		}
	}
	return LinkedIdentity{}, false
}

// AllDIDs returns all DIDs associated with this user (primary + wallet-derived).
func (u *User) AllDIDs() []string {
	dids := make([]string, 0, 1+len(u.wallets))
	dids = append(dids, u.primaryDID.String())

	for _, wallet := range u.wallets {
		dids = append(dids, wallet.ToDID())
	}

	return dids
}

// ============================================================================
// Commands
// ============================================================================

// CanAuthenticate returns true if the user can authenticate.
func (u *User) CanAuthenticate() bool {
	return u.status.CanAuthenticate()
}

// RecordLogin records a successful login.
func (u *User) RecordLogin(method AuthMethod) {
	u.Raise(u, NewUserLoggedInEvent(u.ID(), method.String()))
}

// Suspend suspends the user account.
func (u *User) Suspend(reason string) error {
	if u.status == UserStatusSuspended {
		return ErrUserDisabled("User.Suspend", u.ID())
	}

	u.Raise(u, NewUserSuspendedEvent(u.ID(), reason))
	return nil
}

// Activate activates a suspended user account.
func (u *User) Activate() error {
	if u.status == UserStatusActive {
		return nil // Already active
	}

	u.Raise(u, NewUserActivatedEvent(u.ID()))
	return nil
}

// ============================================================================
// Link Identities
// ============================================================================

// LinkEmail links an email address to this user.
func (u *User) LinkEmail(email string, verified bool) error {
	// Check if email already linked
	for _, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && identity.Value == email {
			return ErrConnectionAlreadyExists("User.LinkEmail", "email", u.ID())
		}
	}

	identity, err := NewEmailIdentity(email, verified)
	if err != nil {
		return err
	}

	u.linkedIdentities = append(u.linkedIdentities, identity)
	u.Raise(u, NewIdentityLinkedEvent(u.ID(), IdentityTypeEmail.String(), email))
	return nil
}

// VerifyEmail marks an email as verified.
func (u *User) VerifyEmail(email string) error {
	for i, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && identity.Value == email {
			u.linkedIdentities[i].Verified = true
			u.linkedIdentities[i].VerifiedAt = time.Now()

			u.Raise(u, NewEmailVerifiedEvent(u.ID(), email))
			return nil
		}
	}

	return ErrConnectionNotFound("User.VerifyEmail", "email", u.ID())
}

// VerifyPrimaryEmail marks the user's first unverified email as verified.
func (u *User) VerifyPrimaryEmail() {
	for i, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && !identity.Verified {
			u.linkedIdentities[i].Verified = true
			u.linkedIdentities[i].VerifiedAt = time.Now()
			u.Raise(u, NewEmailVerifiedEvent(u.ID(), identity.Value))
			return
		}
	}
}

// LinkWallet links a wallet address to this user.
func (u *User) LinkWallet(address string, chain Chain) error {
	wallet := NewWalletAddress(address, chain)

	// Check if wallet already linked
	for _, w := range u.wallets {
		if w.Equals(wallet) {
			return ErrConnectionAlreadyExists("User.LinkWallet", chain.String(), u.ID())
		}
	}

	u.wallets = append(u.wallets, wallet)
	u.linkedIdentities = append(u.linkedIdentities, NewWalletIdentity(address, chain))

	u.Raise(u, NewWalletLinkedEvent(u.ID(), address, chain.String()))
	return nil
}

// UnlinkWallet removes a wallet address from this user.
func (u *User) UnlinkWallet(address string, chain Chain) error {
	wallet := NewWalletAddress(address, chain)

	// Cannot unlink if it's the only identity
	if len(u.wallets) == 1 && u.wallets[0].Equals(wallet) && !u.HasLinkedEmail() {
		return ErrInvalidCredentials("User.UnlinkWallet")
	}

	// Find and remove wallet
	found := false
	for i, w := range u.wallets {
		if w.Equals(wallet) {
			u.wallets = append(u.wallets[:i], u.wallets[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrConnectionNotFound("User.UnlinkWallet", chain.String(), u.ID())
	}

	// Remove from linked identities
	for i, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeWallet && identity.Value == wallet.Address && identity.Provider == chain.String() {
			u.linkedIdentities = append(u.linkedIdentities[:i], u.linkedIdentities[i+1:]...)
			break
		}
	}

	u.Raise(u, NewWalletUnlinkedEvent(u.ID(), address, chain.String()))
	return nil
}

// LinkOAuth links an OAuth provider connection.
func (u *User) LinkOAuth(provider OAuthProvider, providerUserID string) error {
	// Check if provider already linked
	if u.HasOAuthConnection(provider) {
		return ErrConnectionAlreadyExists("User.LinkOAuth", provider.String(), u.ID())
	}

	identity := NewOAuthIdentity(provider, providerUserID)
	u.linkedIdentities = append(u.linkedIdentities, identity)

	u.Raise(u, NewOAuthLinkedEvent(u.ID(), provider.String(), providerUserID))
	return nil
}

// UnlinkOAuth removes an OAuth provider connection.
func (u *User) UnlinkOAuth(provider OAuthProvider) error {
	// Cannot unlink if it's the only identity
	if len(u.linkedIdentities) == 1 && u.linkedIdentities[0].Provider == provider.String() {
		return ErrInvalidCredentials("User.UnlinkOAuth")
	}

	found := false
	for i, identity := range u.linkedIdentities {
		if identity.Provider == provider.String() {
			u.linkedIdentities = append(u.linkedIdentities[:i], u.linkedIdentities[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrConnectionNotFound("User.UnlinkOAuth", provider.String(), u.ID())
	}

	u.Raise(u, NewOAuthUnlinkedEvent(u.ID(), provider.String()))
	return nil
}

// ============================================================================
// DID Management
// ============================================================================

// UpdatePrimaryDID updates the user's primary DID.
// This is a significant operation and should be used carefully.
func (u *User) UpdatePrimaryDID(newDID did.DID) error {
	if newDID.IsZero() {
		return ErrInvalidCredentials("User.UpdatePrimaryDID")
	}

	oldDID := u.primaryDID

	u.Raise(u, NewPrimaryDIDUpdatedEvent(u.ID(), oldDID.String(), newDID.String()))
	return nil
}
