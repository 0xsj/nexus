package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// Aggregate type constants
const (
	AggregateTypeWallet = "Wallet"
)

// Event type constants
const (
	EventTypeWalletLinked       = "Wallet.Linked"
	EventTypeWalletVerified     = "Wallet.Verified"
	EventTypeWalletUnlinked     = "Wallet.Unlinked"
	EventTypeWalletSetPrimary   = "Wallet.SetPrimary"
	EventTypeWalletLabelUpdated = "Wallet.LabelUpdated"
)

// ============================================================================
// Wallet Events
// ============================================================================

// WalletLinkedEvent is emitted when a new wallet is linked to a user.
type WalletLinkedEvent struct {
	eventsourcing.BaseEvent

	WalletID  string    `json:"wallet_id"`
	UserID    string    `json:"user_id"`
	Address   string    `json:"address"`
	ChainID   int       `json:"chain_id"`
	ChainName string    `json:"chain_name"`
	DID       string    `json:"did"`
	Label     string    `json:"label,omitempty"`
	LinkedAt  time.Time `json:"linked_at"`
}

// EventType returns the event type.
func (e WalletLinkedEvent) EventType() string {
	return EventTypeWalletLinked
}

// WalletVerifiedEvent is emitted when a wallet is verified via signature.
type WalletVerifiedEvent struct {
	eventsourcing.BaseEvent

	WalletID   string    `json:"wallet_id"`
	VerifiedAt time.Time `json:"verified_at"`
}

// EventType returns the event type.
func (e WalletVerifiedEvent) EventType() string {
	return EventTypeWalletVerified
}

// WalletUnlinkedEvent is emitted when a wallet is unlinked from a user.
type WalletUnlinkedEvent struct {
	eventsourcing.BaseEvent

	WalletID   string    `json:"wallet_id"`
	UnlinkedAt time.Time `json:"unlinked_at"`
}

// EventType returns the event type.
func (e WalletUnlinkedEvent) EventType() string {
	return EventTypeWalletUnlinked
}

// WalletSetPrimaryEvent is emitted when a wallet is set as the primary wallet.
type WalletSetPrimaryEvent struct {
	eventsourcing.BaseEvent

	WalletID string    `json:"wallet_id"`
	SetAt    time.Time `json:"set_at"`
}

// EventType returns the event type.
func (e WalletSetPrimaryEvent) EventType() string {
	return EventTypeWalletSetPrimary
}

// WalletLabelUpdatedEvent is emitted when a wallet's label is changed.
type WalletLabelUpdatedEvent struct {
	eventsourcing.BaseEvent

	WalletID  string    `json:"wallet_id"`
	OldLabel  string    `json:"old_label,omitempty"`
	NewLabel  string    `json:"new_label"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventType returns the event type.
func (e WalletLabelUpdatedEvent) EventType() string {
	return EventTypeWalletLabelUpdated
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterWalletEvents registers all Wallet domain events with the event registry.
func RegisterWalletEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeWalletLinked, func() eventsourcing.Event { return &WalletLinkedEvent{} })
	registry.Register(EventTypeWalletVerified, func() eventsourcing.Event { return &WalletVerifiedEvent{} })
	registry.Register(EventTypeWalletUnlinked, func() eventsourcing.Event { return &WalletUnlinkedEvent{} })
	registry.Register(EventTypeWalletSetPrimary, func() eventsourcing.Event { return &WalletSetPrimaryEvent{} })
	registry.Register(EventTypeWalletLabelUpdated, func() eventsourcing.Event { return &WalletLabelUpdatedEvent{} })
}

// init registers events with the default registry.
func init() {
	RegisterWalletEvents(eventsourcing.DefaultRegistry)
}

// ============================================================================
// Event Helpers
// ============================================================================

// newWalletBaseEvent creates a base event for wallet aggregate.
func newWalletBaseEvent(walletID WalletID) eventsourcing.BaseEvent {
	return eventsourcing.NewBaseEvent(AggregateTypeWallet, walletID.String())
}
