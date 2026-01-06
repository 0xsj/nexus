package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Wallet Aggregate
// ============================================================================

// Wallet represents a blockchain wallet linked to a user.
type Wallet struct {
	eventsourcing.AggregateRoot

	// id is the unique wallet identifier.
	id string

	// userID is the user this wallet belongs to.
	userID string

	// address is the validated wallet address.
	address Address

	// did is the derived did:pkh for this wallet.
	did did.DID

	// label is the user-friendly name for the wallet.
	label string

	// isPrimary indicates if this is the user's primary wallet.
	isPrimary bool

	// status is the wallet status.
	status WalletStatus

	// verifiedAt is when the wallet ownership was verified.
	verifiedAt time.Time

	// lastUsedAt is when the wallet was last used.
	lastUsedAt *time.Time

	// createdAt is when the wallet was created.
	createdAt time.Time

	// updatedAt is when the wallet was last updated.
	updatedAt time.Time
}

// ============================================================================
// Constructor
// ============================================================================

// NewWallet creates a new Wallet aggregate.
func NewWallet(id string, userID string, address Address, walletDID did.DID) (*Wallet, error) {
	const op = "Wallet.New"

	if id == "" {
		return nil, ErrInvalidAddress(op, "", "wallet ID cannot be empty")
	}

	if userID == "" {
		return nil, ErrInvalidAddress(op, "", "user ID cannot be empty")
	}

	if address.IsZero() {
		return nil, ErrInvalidAddress(op, "", "address cannot be empty")
	}

	now := time.Now()

	w := &Wallet{
		id:         id,
		userID:     userID,
		address:    address,
		did:        walletDID,
		status:     WalletStatusActive,
		verifiedAt: now,
		createdAt:  now,
		updatedAt:  now,
	}

	w.InitAggregate("Wallet", id)

	return w, nil
}

// RehydrateWallet reconstructs a Wallet from persisted state.
// Use only when loading from database.
func RehydrateWallet(
	id string,
	userID string,
	address Address,
	walletDID did.DID,
	label string,
	isPrimary bool,
	status WalletStatus,
	verifiedAt time.Time,
	lastUsedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Wallet {
	w := &Wallet{
		id:         id,
		userID:     userID,
		address:    address,
		did:        walletDID,
		label:      label,
		isPrimary:  isPrimary,
		status:     status,
		verifiedAt: verifiedAt,
		lastUsedAt: lastUsedAt,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}

	w.InitAggregate("Wallet", id)

	return w
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the wallet ID.
func (w *Wallet) ID() string {
	return w.id
}

// UserID returns the user ID.
func (w *Wallet) UserID() string {
	return w.userID
}

// Address returns the wallet address.
func (w *Wallet) Address() Address {
	return w.address
}

// AddressString returns the normalized address string.
func (w *Wallet) AddressString() string {
	return w.address.Normalized()
}

// ChainID returns the chain ID.
func (w *Wallet) ChainID() ChainID {
	return w.address.ChainID()
}

// Family returns the chain family.
func (w *Wallet) Family() ChainFamily {
	return w.address.Family()
}

// DID returns the derived did:pkh.
func (w *Wallet) DID() did.DID {
	return w.did
}

// DIDString returns the DID as a string.
func (w *Wallet) DIDString() string {
	return w.did.String()
}

// Label returns the wallet label.
func (w *Wallet) Label() string {
	return w.label
}

// IsPrimary returns true if this is the primary wallet.
func (w *Wallet) IsPrimary() bool {
	return w.isPrimary
}

// Status returns the wallet status.
func (w *Wallet) Status() WalletStatus {
	return w.status
}

// VerifiedAt returns when the wallet was verified.
func (w *Wallet) VerifiedAt() time.Time {
	return w.verifiedAt
}

// LastUsedAt returns when the wallet was last used.
func (w *Wallet) LastUsedAt() *time.Time {
	return w.lastUsedAt
}

// CreatedAt returns when the wallet was created.
func (w *Wallet) CreatedAt() time.Time {
	return w.createdAt
}

// UpdatedAt returns when the wallet was last updated.
func (w *Wallet) UpdatedAt() time.Time {
	return w.updatedAt
}

// ============================================================================
// Predicates
// ============================================================================

// IsActive returns true if the wallet is active.
func (w *Wallet) IsActive() bool {
	return w.status == WalletStatusActive
}

// IsInactive returns true if the wallet is inactive.
func (w *Wallet) IsInactive() bool {
	return w.status == WalletStatusInactive
}

// IsSuspended returns true if the wallet is suspended.
func (w *Wallet) IsSuspended() bool {
	return w.status == WalletStatusSuspended
}

// IsEVM returns true if this is an EVM wallet.
func (w *Wallet) IsEVM() bool {
	return w.address.IsEVM()
}

// IsSolana returns true if this is a Solana wallet.
func (w *Wallet) IsSolana() bool {
	return w.address.IsSolana()
}

// IsCosmos returns true if this is a Cosmos wallet.
func (w *Wallet) IsCosmos() bool {
	return w.address.IsCosmos()
}

// ============================================================================
// Commands
// ============================================================================

// SetLabel sets the wallet label.
func (w *Wallet) SetLabel(label string) {
	w.label = label
	w.updatedAt = time.Now()
}

// SetPrimary sets whether this is the primary wallet.
func (w *Wallet) SetPrimary(isPrimary bool) {
	w.isPrimary = isPrimary
	w.updatedAt = time.Now()
}

// Activate activates the wallet.
func (w *Wallet) Activate() error {
	if w.status == WalletStatusActive {
		return nil // Already active
	}

	w.status = WalletStatusActive
	w.updatedAt = time.Now()
	return nil
}

// Deactivate deactivates the wallet.
func (w *Wallet) Deactivate() error {
	if w.status == WalletStatusInactive {
		return nil // Already inactive
	}

	w.status = WalletStatusInactive
	w.updatedAt = time.Now()
	return nil
}

// Suspend suspends the wallet.
func (w *Wallet) Suspend() error {
	if w.status == WalletStatusSuspended {
		return nil // Already suspended
	}

	w.status = WalletStatusSuspended
	w.updatedAt = time.Now()
	return nil
}

// RecordUsage records that the wallet was used.
func (w *Wallet) RecordUsage() {
	now := time.Now()
	w.lastUsedAt = &now
	w.updatedAt = now
}

// ============================================================================
// Validation
// ============================================================================

// Validate validates the wallet state.
func (w *Wallet) Validate() error {
	const op = "Wallet.Validate"

	if w.id == "" {
		return ErrInvalidAddress(op, "", "wallet ID cannot be empty")
	}

	if w.userID == "" {
		return ErrInvalidAddress(op, "", "user ID cannot be empty")
	}

	if w.address.IsZero() {
		return ErrInvalidAddress(op, "", "address cannot be empty")
	}

	if !w.status.IsValid() {
		return ErrInvalidAddress(op, "", "invalid wallet status")
	}

	return nil
}

// ============================================================================
// Event Sourcing
// ============================================================================

// ApplyEvent applies an event to update aggregate state.
func (w *Wallet) ApplyEvent(event eventsourcing.Event) {
	// TODO: Implement event handlers when events are defined
}

// GetAggregateRoot returns the embedded AggregateRoot.
func (w *Wallet) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &w.AggregateRoot
}
