package domain

import (
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Wallet is the aggregate root for linked blockchain wallets.
// It manages wallet address, chain information, verification status,
// and metadata such as labels and primary designation.
type Wallet struct {
	eventsourcing.AggregateRoot

	id         WalletID
	userID     string
	address    WalletAddress
	chain      Chain
	did        string
	label      WalletLabel
	isPrimary  bool
	status     WalletStatus
	verifiedAt time.Time
	createdAt  time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// LinkWallet creates a new Wallet aggregate for linking a wallet to a user.
func LinkWallet(
	id WalletID,
	userID string,
	address WalletAddress,
	chain Chain,
	did string,
	label WalletLabel,
) (*Wallet, error) {
	if id.IsZero() {
		return nil, WalletNotFound("Wallet.Link", "").
			WithMessage("wallet ID is required")
	}

	if userID == "" {
		return nil, WalletNotActive("Wallet.Link", id.String()).
			WithMessage("user ID is required")
	}

	if address.IsZero() {
		return nil, InvalidAddress("Wallet.Link", "", "address is required")
	}

	if chain.IsZero() {
		return nil, ChainNotSupported("Wallet.Link", 0).
			WithMessage("chain is required")
	}

	if did == "" {
		return nil, WalletNotActive("Wallet.Link", id.String()).
			WithMessage("DID is required")
	}

	w := &Wallet{}
	w.InitAggregate(AggregateTypeWallet, id.String())

	now := time.Now().UTC()

	w.Raise(w, &WalletLinkedEvent{
		BaseEvent: newWalletBaseEvent(id),
		WalletID:  id.String(),
		UserID:    userID,
		Address:   address.String(),
		ChainID:   chain.ID(),
		ChainName: chain.Name(),
		DID:       did,
		Label:     label.String(),
		LinkedAt:  now,
	})

	return w, nil
}

// NewWalletFromEvents reconstructs a Wallet from events (for hydration).
func NewWalletFromEvents(id string) *Wallet {
	w := &Wallet{}
	w.InitAggregate(AggregateTypeWallet, id)
	return w
}

// WalletFactory creates a factory for Wallet aggregates.
func WalletFactory() eventsourcing.AggregateFactory {
	return eventsourcing.AggregateFactoryFunc(func(aggregateID string) eventsourcing.Aggregate {
		return NewWalletFromEvents(aggregateID)
	})
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the wallet's ID.
func (w *Wallet) ID() WalletID {
	return w.id
}

// UserID returns the user ID that owns this wallet.
func (w *Wallet) UserID() string {
	return w.userID
}

// Address returns the wallet's blockchain address.
func (w *Wallet) Address() WalletAddress {
	return w.address
}

// Chain returns the wallet's chain.
func (w *Wallet) Chain() Chain {
	return w.chain
}

// DID returns the wallet's derived DID.
func (w *Wallet) DID() string {
	return w.did
}

// Label returns the wallet's user-assigned label.
func (w *Wallet) Label() WalletLabel {
	return w.label
}

// IsPrimary returns true if this is the user's primary wallet.
func (w *Wallet) IsPrimary() bool {
	return w.isPrimary
}

// Status returns the wallet's current status.
func (w *Wallet) Status() WalletStatus {
	return w.status
}

// VerifiedAt returns when the wallet was verified.
func (w *Wallet) VerifiedAt() time.Time {
	return w.verifiedAt
}

// CreatedAt returns when the wallet was linked.
func (w *Wallet) CreatedAt() time.Time {
	return w.createdAt
}

// ============================================================================
// Command Methods
// ============================================================================

// Verify verifies the wallet after successful signature validation.
func (w *Wallet) Verify() error {
	if !w.status.IsUnverified() {
		return WalletNotActive("Wallet.Verify", w.id.String()).
			WithMessage("cannot verify wallet in current status: " + w.status.String())
	}

	w.Raise(w, &WalletVerifiedEvent{
		BaseEvent:  newWalletBaseEvent(w.id),
		WalletID:   w.id.String(),
		VerifiedAt: time.Now().UTC(),
	})

	return nil
}

// Unlink unlinks the wallet from the user.
func (w *Wallet) Unlink() error {
	if w.status.IsRevoked() {
		return WalletNotActive("Wallet.Unlink", w.id.String()).
			WithMessage("cannot unlink wallet in current status: " + w.status.String())
	}

	w.Raise(w, &WalletUnlinkedEvent{
		BaseEvent:  newWalletBaseEvent(w.id),
		WalletID:   w.id.String(),
		UnlinkedAt: time.Now().UTC(),
	})

	return nil
}

// SetPrimary designates this wallet as the user's primary wallet.
func (w *Wallet) SetPrimary() error {
	if !w.status.IsActive() {
		return WalletNotActive("Wallet.SetPrimary", w.id.String()).
			WithMessage("cannot set primary: wallet is not active")
	}

	w.Raise(w, &WalletSetPrimaryEvent{
		BaseEvent: newWalletBaseEvent(w.id),
		WalletID:  w.id.String(),
		SetAt:     time.Now().UTC(),
	})

	return nil
}

// UpdateLabel updates the wallet's user-assigned label.
func (w *Wallet) UpdateLabel(label WalletLabel) error {
	if !w.status.IsActive() {
		return WalletNotActive("Wallet.UpdateLabel", w.id.String()).
			WithMessage("cannot update label: wallet is not active")
	}

	w.Raise(w, &WalletLabelUpdatedEvent{
		BaseEvent: newWalletBaseEvent(w.id),
		WalletID:  w.id.String(),
		OldLabel:  w.label.String(),
		NewLabel:  label.String(),
		UpdatedAt: time.Now().UTC(),
	})

	return nil
}

// ============================================================================
// Event Application
// ============================================================================

// ApplyEvent applies an event to update the aggregate state.
func (w *Wallet) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *WalletLinkedEvent:
		w.onWalletLinked(e)
	case *WalletVerifiedEvent:
		w.onWalletVerified(e)
	case *WalletUnlinkedEvent:
		w.onWalletUnlinked(e)
	case *WalletSetPrimaryEvent:
		w.onWalletSetPrimary(e)
	case *WalletLabelUpdatedEvent:
		w.onWalletLabelUpdated(e)
	}
}

func (w *Wallet) onWalletLinked(e *WalletLinkedEvent) {
	var err error

	w.id, err = ParseWalletID(e.WalletID)
	if err != nil {
		panic("corrupt event store: WalletLinked has invalid WalletID: " + e.WalletID)
	}

	w.userID = e.UserID

	w.address, err = NewWalletAddress(e.Address)
	if err != nil {
		panic("corrupt event store: WalletLinked has invalid Address: " + e.Address)
	}

	w.chain, err = NewChain(e.ChainID, e.ChainName, "eip155")
	if err != nil {
		panic(fmt.Sprintf("corrupt event store: WalletLinked has invalid Chain: %d %s", e.ChainID, e.ChainName))
	}

	w.did = e.DID

	w.label, err = NewWalletLabel(e.Label)
	if err != nil {
		panic("corrupt event store: WalletLinked has invalid Label: " + e.Label)
	}

	w.isPrimary = false
	w.status = WalletStatusUnverified
	w.createdAt = e.LinkedAt
}

func (w *Wallet) onWalletVerified(e *WalletVerifiedEvent) {
	w.status = WalletStatusActive
	w.verifiedAt = e.VerifiedAt
}

func (w *Wallet) onWalletUnlinked(e *WalletUnlinkedEvent) {
	w.status = WalletStatusRevoked
}

func (w *Wallet) onWalletSetPrimary(e *WalletSetPrimaryEvent) {
	w.isPrimary = true
}

func (w *Wallet) onWalletLabelUpdated(e *WalletLabelUpdatedEvent) {
	var err error
	w.label, err = NewWalletLabel(e.NewLabel)
	if err != nil {
		panic("corrupt event store: WalletLabelUpdated has invalid Label: " + e.NewLabel)
	}
}

// ============================================================================
// Aggregate Root Access
// ============================================================================

// GetAggregateRoot returns the embedded AggregateRoot.
func (w *Wallet) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &w.AggregateRoot
}
