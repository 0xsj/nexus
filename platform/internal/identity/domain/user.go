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
// Users can have multiple DIDs (custodial, wallet-derived, imported) and
// can link additional identities (email, OAuth) for authentication convenience.
type User struct {
	eventsourcing.AggregateRoot

	// Primary DID (points to one of the LinkedDIDs)
	primaryDID did.DID

	// All DIDs associated with this user
	linkedDIDs LinkedDIDs

	// Status
	status UserStatus

	// Linked identities (email, OAuth connections) - for authentication
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
// Constructors
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
		linkedDIDs:       make(LinkedDIDs, 0),
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

// NewUserWithCustodialDID creates a new user with a Nexus-generated custodial DID.
// Used for email/OAuth users who don't have a wallet.
func NewUserWithCustodialDID(id string, linkedDIDID string, primaryDID did.DID) (*User, error) {
	user, err := NewUser(id, primaryDID)
	if err != nil {
		return nil, err
	}

	// Add the custodial DID as the first linked DID
	custodialDID := NewCustodialDID(linkedDIDID, primaryDID, true)
	user.linkedDIDs = append(user.linkedDIDs, custodialDID)

	// Raise DID linked event
	user.Raise(user, NewDIDLinkedEvent(
		id,
		linkedDIDID,
		primaryDID.String(),
		DIDSourceCustodial,
		true,
		custodialDID.Label,
		custodialDID.Metadata,
	))

	return user, nil
}

// NewUserFromWallet creates a new user from a wallet signature.
// Generates a did:pkh from the wallet address as the primary DID.
func NewUserFromWallet(id string, linkedDIDID string, wallet WalletAddress) (*User, error) {
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

	// Add the wallet-derived DID
	walletDID := NewWalletDID(linkedDIDID, primaryDID, wallet.Address, wallet.Chain, true)
	user.linkedDIDs = append(user.linkedDIDs, walletDID)

	// Add wallet to linked wallets
	user.wallets = append(user.wallets, wallet)

	// Add wallet as linked identity
	user.linkedIdentities = append(user.linkedIdentities, NewWalletIdentity(wallet.Address, wallet.Chain))

	// Raise DID linked event
	user.Raise(user, NewDIDLinkedEvent(
		id,
		linkedDIDID,
		primaryDID.String(),
		DIDSourceWallet,
		true,
		walletDID.Label,
		walletDID.Metadata,
	))

	return user, nil
}

// NewUserFromEmail creates a new user from an email address.
// Generates a did:key as the primary DID since email doesn't have a native DID.
func NewUserFromEmail(id string, linkedDIDID string, email string, custodialDID did.DID) (*User, error) {
	if id == "" {
		return nil, ErrInvalidCredentials("User.NewFromEmail")
	}

	if email == "" {
		return nil, ErrInvalidEmail("User.NewFromEmail", email)
	}

	if custodialDID.IsZero() {
		return nil, ErrInvalidCredentials("User.NewFromEmail")
	}

	user, err := NewUserWithCustodialDID(id, linkedDIDID, custodialDID)
	if err != nil {
		return nil, err
	}

	// Add email as linked identity (verified since they clicked magic link)
	if err := user.linkEmailInternal(email, true); err != nil {
		return nil, err
	}

	return user, nil
}

// Reconstitute creates a User from persisted data (no events raised).
func Reconstitute(
	id string,
	version int,
	primaryDID did.DID,
	linkedDIDs LinkedDIDs,
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
		linkedDIDs:       linkedDIDs,
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
	// New DID events
	case *DIDLinkedEvent:
		u.applyDIDLinked(e)
	case *DIDUnlinkedEvent:
		u.applyDIDUnlinked(e)
	case *DIDPrimaryChangedEvent:
		u.applyDIDPrimaryChanged(e)
	case *DIDLabelUpdatedEvent:
		u.applyDIDLabelUpdated(e)
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

func (u *User) applyDIDLinked(e *DIDLinkedEvent) {
	parsedDID, _ := did.Parse(e.DID)

	linkedDID := LinkedDID{
		ID:        e.LinkedDIDID,
		DID:       parsedDID,
		Source:    DIDSource(e.Source),
		IsPrimary: e.IsPrimary,
		Label:     e.Label,
		Metadata:  e.Metadata,
		LinkedAt:  e.OccurredAt(),
	}

	u.linkedDIDs = append(u.linkedDIDs, linkedDID)

	// If this is primary, update primaryDID and unset others
	if e.IsPrimary {
		u.primaryDID = parsedDID
		for i := range u.linkedDIDs {
			if u.linkedDIDs[i].ID != e.LinkedDIDID {
				u.linkedDIDs[i].IsPrimary = false
			}
		}
	}

	u.updatedAt = e.OccurredAt()
}

func (u *User) applyDIDUnlinked(e *DIDUnlinkedEvent) {
	for i, ld := range u.linkedDIDs {
		if ld.ID == e.LinkedDIDID {
			u.linkedDIDs = append(u.linkedDIDs[:i], u.linkedDIDs[i+1:]...)
			break
		}
	}
	u.updatedAt = e.OccurredAt()
}

func (u *User) applyDIDPrimaryChanged(e *DIDPrimaryChangedEvent) {
	// Unset old primary
	for i := range u.linkedDIDs {
		if u.linkedDIDs[i].ID == e.OldPrimaryDIDID {
			u.linkedDIDs[i].IsPrimary = false
		}
		if u.linkedDIDs[i].ID == e.NewPrimaryDIDID {
			u.linkedDIDs[i].IsPrimary = true
		}
	}

	// Update primaryDID
	parsedDID, _ := did.Parse(e.NewPrimaryDID)
	u.primaryDID = parsedDID

	u.updatedAt = e.OccurredAt()
}

func (u *User) applyDIDLabelUpdated(e *DIDLabelUpdatedEvent) {
	for i := range u.linkedDIDs {
		if u.linkedDIDs[i].ID == e.LinkedDIDID {
			u.linkedDIDs[i].Label = e.NewLabel
			break
		}
	}
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

// LinkedDIDs returns all linked DIDs.
func (u *User) LinkedDIDs() LinkedDIDs {
	result := make(LinkedDIDs, len(u.linkedDIDs))
	copy(result, u.linkedDIDs)
	return result
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
// DID Queries
// ============================================================================

// PrimaryLinkedDID returns the primary LinkedDID entry.
func (u *User) PrimaryLinkedDID() (LinkedDID, bool) {
	return u.linkedDIDs.Primary()
}

// HasDID checks if the user has a specific DID linked.
func (u *User) HasDID(d did.DID) bool {
	return u.linkedDIDs.HasDID(d)
}

// FindLinkedDID finds a linked DID by its DID value.
func (u *User) FindLinkedDID(d did.DID) (LinkedDID, bool) {
	return u.linkedDIDs.FindByDID(d)
}

// FindLinkedDIDByID finds a linked DID by its ID.
func (u *User) FindLinkedDIDByID(id string) (LinkedDID, bool) {
	return u.linkedDIDs.FindByID(id)
}

// WalletDIDs returns all wallet-derived DIDs.
func (u *User) WalletDIDs() LinkedDIDs {
	return u.linkedDIDs.WalletDIDs()
}

// CustodialDIDs returns all custodial DIDs.
func (u *User) CustodialDIDs() LinkedDIDs {
	return u.linkedDIDs.CustodialDIDs()
}

// HasCustodialDID returns true if the user has a custodial DID.
func (u *User) HasCustodialDID() bool {
	return len(u.linkedDIDs.CustodialDIDs()) > 0
}

// HasWalletDID returns true if the user has a wallet-derived DID.
func (u *User) HasWalletDID() bool {
	return len(u.linkedDIDs.WalletDIDs()) > 0
}

// AllDIDs returns all DID strings associated with this user.
func (u *User) AllDIDs() []string {
	return u.linkedDIDs.DIDs()
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
// DID Management
// ============================================================================

// LinkDID links a new DID to this user.
func (u *User) LinkDID(linkedDID LinkedDID) error {
	// Check if DID already linked
	if u.linkedDIDs.HasDID(linkedDID.DID) {
		return ErrConnectionAlreadyExists("User.LinkDID", "did", u.ID())
	}

	// Check if ID already exists
	if _, found := u.linkedDIDs.FindByID(linkedDID.ID); found {
		return ErrConnectionAlreadyExists("User.LinkDID", "did", u.ID())
	}

	// If this is set as primary, we need to update the current primary
	if linkedDID.IsPrimary {
		if currentPrimary, found := u.linkedDIDs.Primary(); found {
			u.Raise(u, NewDIDPrimaryChangedEvent(
				u.ID(),
				currentPrimary.ID,
				currentPrimary.DID.String(),
				linkedDID.ID,
				linkedDID.DID.String(),
			))
		}
	}

	u.Raise(u, NewDIDLinkedEvent(
		u.ID(),
		linkedDID.ID,
		linkedDID.DID.String(),
		linkedDID.Source,
		linkedDID.IsPrimary,
		linkedDID.Label,
		linkedDID.Metadata,
	))

	return nil
}

// LinkCustodialDID creates and links a new custodial DID.
func (u *User) LinkCustodialDID(linkedDIDID string, d did.DID, makePrimary bool) error {
	custodialDID := NewCustodialDID(linkedDIDID, d, makePrimary)
	return u.LinkDID(custodialDID)
}

// LinkWalletDID creates and links a new wallet-derived DID.
func (u *User) LinkWalletDID(linkedDIDID string, d did.DID, walletAddress string, chain Chain, makePrimary bool) error {
	walletDID := NewWalletDID(linkedDIDID, d, walletAddress, chain, makePrimary)
	return u.LinkDID(walletDID)
}

// UnlinkDID removes a DID from this user.
func (u *User) UnlinkDID(linkedDIDID string, reason string) error {
	linkedDID, found := u.linkedDIDs.FindByID(linkedDIDID)
	if !found {
		return ErrConnectionNotFound("User.UnlinkDID", "did", u.ID())
	}

	// Cannot unlink the only DID
	if len(u.linkedDIDs) == 1 {
		return ErrInvalidCredentials("User.UnlinkDID")
	}

	// Cannot unlink primary DID without setting a new primary first
	if linkedDID.IsPrimary {
		return ErrInvalidCredentials("User.UnlinkDID")
	}

	u.Raise(u, NewDIDUnlinkedEvent(u.ID(), linkedDIDID, linkedDID.DID.String(), reason))
	return nil
}

// SetPrimaryDID changes which linked DID is the primary.
func (u *User) SetPrimaryDID(linkedDIDID string) error {
	newPrimary, found := u.linkedDIDs.FindByID(linkedDIDID)
	if !found {
		return ErrConnectionNotFound("User.SetPrimaryDID", "did", u.ID())
	}

	// Already primary
	if newPrimary.IsPrimary {
		return nil
	}

	oldPrimary, _ := u.linkedDIDs.Primary()

	u.Raise(u, NewDIDPrimaryChangedEvent(
		u.ID(),
		oldPrimary.ID,
		oldPrimary.DID.String(),
		newPrimary.ID,
		newPrimary.DID.String(),
	))

	return nil
}

// UpdateDIDLabel updates the label for a linked DID.
func (u *User) UpdateDIDLabel(linkedDIDID string, newLabel string) error {
	linkedDID, found := u.linkedDIDs.FindByID(linkedDIDID)
	if !found {
		return ErrConnectionNotFound("User.UpdateDIDLabel", "did", u.ID())
	}

	if linkedDID.Label == newLabel {
		return nil // No change
	}

	u.Raise(u, NewDIDLabelUpdatedEvent(u.ID(), linkedDIDID, linkedDID.Label, newLabel))
	return nil
}

// UpdatePrimaryDID updates the user's primary DID.
// Deprecated: Use SetPrimaryDID with a linkedDIDID instead.
func (u *User) UpdatePrimaryDID(newDID did.DID) error {
	if newDID.IsZero() {
		return ErrInvalidCredentials("User.UpdatePrimaryDID")
	}

	oldDID := u.primaryDID

	u.Raise(u, NewPrimaryDIDUpdatedEvent(u.ID(), oldDID.String(), newDID.String()))
	return nil
}

// ============================================================================
// Link Identities
// ============================================================================

// linkEmailInternal links an email without raising IdentityLinkedEvent.
// Used by constructors to avoid duplicate events.
func (u *User) linkEmailInternal(email string, verified bool) error {
	identity, err := NewEmailIdentity(email, verified)
	if err != nil {
		return err
	}

	u.linkedIdentities = append(u.linkedIdentities, identity)
	return nil
}

// LinkEmail links an email address to this user.
func (u *User) LinkEmail(email string, verified bool) error {
	// Check if email already linked
	for _, identity := range u.linkedIdentities {
		if identity.Type == IdentityTypeEmail && identity.Value == email {
			return ErrConnectionAlreadyExists("User.LinkEmail", "email", u.ID())
		}
	}

	if err := u.linkEmailInternal(email, verified); err != nil {
		return err
	}

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
// Also creates a wallet-derived DID if linkedDIDID is provided.
func (u *User) LinkWallet(address string, chain Chain, linkedDIDID string) error {
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

	// Create wallet-derived DID if linkedDIDID provided
	if linkedDIDID != "" {
		walletDID, err := did.Parse(wallet.ToDID())
		if err == nil {
			// Don't make it primary by default - user can explicitly set it
			_ = u.LinkWalletDID(linkedDIDID, walletDID, address, chain, false)
		}
	}

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

	// Note: We don't automatically unlink the wallet DID here.
	// The caller should explicitly call UnlinkDID if needed.

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
