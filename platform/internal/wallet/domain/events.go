package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Event Types
// ============================================================================

const (
	// EventTypeWalletLinked is emitted when a wallet is linked to a user.
	EventTypeWalletLinked = "Wallet.Linked"

	// EventTypeWalletUnlinked is emitted when a wallet is unlinked from a user.
	EventTypeWalletUnlinked = "Wallet.Unlinked"

	// EventTypeWalletVerified is emitted when wallet ownership is verified.
	EventTypeWalletVerified = "Wallet.Verified"

	// EventTypeWalletLabelUpdated is emitted when a wallet label is updated.
	EventTypeWalletLabelUpdated = "Wallet.LabelUpdated"

	// EventTypeWalletPrimarySet is emitted when a wallet is set as primary.
	EventTypeWalletPrimarySet = "Wallet.PrimarySet"

	// EventTypeWalletActivated is emitted when a wallet is activated.
	EventTypeWalletActivated = "Wallet.Activated"

	// EventTypeWalletDeactivated is emitted when a wallet is deactivated.
	EventTypeWalletDeactivated = "Wallet.Deactivated"

	// EventTypeWalletSuspended is emitted when a wallet is suspended.
	EventTypeWalletSuspended = "Wallet.Suspended"

	// EventTypeWalletUsed is emitted when a wallet is used for authentication.
	EventTypeWalletUsed = "Wallet.Used"

	// EventTypeChallengeCreated is emitted when a challenge is created.
	EventTypeChallengeCreated = "Wallet.ChallengeCreated"

	// EventTypeChallengeVerified is emitted when a challenge is verified.
	EventTypeChallengeVerified = "Wallet.ChallengeVerified"

	// EventTypeChallengeExpired is emitted when a challenge expires.
	EventTypeChallengeExpired = "Wallet.ChallengeExpired"
)

// ============================================================================
// Wallet Linked Event
// ============================================================================

// WalletLinkedEvent is emitted when a wallet is linked to a user.
type WalletLinkedEvent struct {
	eventsourcing.BaseEvent

	WalletID  string `json:"wallet_id"`
	UserID    string `json:"user_id"`
	Address   string `json:"address"`
	ChainID   string `json:"chain_id"`
	Family    string `json:"family"`
	DID       string `json:"did"`
	Label     string `json:"label,omitempty"`
	IsPrimary bool   `json:"is_primary"`
}

// EventType returns the event type.
func (e *WalletLinkedEvent) EventType() string {
	return EventTypeWalletLinked
}

// NewWalletLinkedEvent creates a new WalletLinkedEvent.
func NewWalletLinkedEvent(wallet *Wallet) *WalletLinkedEvent {
	return &WalletLinkedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:  wallet.ID(),
		UserID:    wallet.UserID(),
		Address:   wallet.AddressString(),
		ChainID:   wallet.ChainID().String(),
		Family:    wallet.Family().String(),
		DID:       wallet.DIDString(),
		Label:     wallet.Label(),
		IsPrimary: wallet.IsPrimary(),
	}
}

// ============================================================================
// Wallet Unlinked Event
// ============================================================================

// WalletUnlinkedEvent is emitted when a wallet is unlinked from a user.
type WalletUnlinkedEvent struct {
	eventsourcing.BaseEvent

	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	Address  string `json:"address"`
	ChainID  string `json:"chain_id"`
	Reason   string `json:"reason,omitempty"`
}

// EventType returns the event type.
func (e *WalletUnlinkedEvent) EventType() string {
	return EventTypeWalletUnlinked
}

// NewWalletUnlinkedEvent creates a new WalletUnlinkedEvent.
func NewWalletUnlinkedEvent(wallet *Wallet, reason string) *WalletUnlinkedEvent {
	return &WalletUnlinkedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:  wallet.ID(),
		UserID:    wallet.UserID(),
		Address:   wallet.AddressString(),
		ChainID:   wallet.ChainID().String(),
		Reason:    reason,
	}
}

// ============================================================================
// Wallet Verified Event
// ============================================================================

// WalletVerifiedEvent is emitted when wallet ownership is verified.
type WalletVerifiedEvent struct {
	eventsourcing.BaseEvent

	WalletID   string    `json:"wallet_id"`
	UserID     string    `json:"user_id"`
	Address    string    `json:"address"`
	ChainID    string    `json:"chain_id"`
	VerifiedAt time.Time `json:"verified_at"`
	Method     string    `json:"method"` // "siwe", "solana", "cosmos"
}

// EventType returns the event type.
func (e *WalletVerifiedEvent) EventType() string {
	return EventTypeWalletVerified
}

// NewWalletVerifiedEvent creates a new WalletVerifiedEvent.
func NewWalletVerifiedEvent(wallet *Wallet, method string) *WalletVerifiedEvent {
	return &WalletVerifiedEvent{
		BaseEvent:  eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:   wallet.ID(),
		UserID:     wallet.UserID(),
		Address:    wallet.AddressString(),
		ChainID:    wallet.ChainID().String(),
		VerifiedAt: wallet.VerifiedAt(),
		Method:     method,
	}
}

// ============================================================================
// Wallet Label Updated Event
// ============================================================================

// WalletLabelUpdatedEvent is emitted when a wallet label is updated.
type WalletLabelUpdatedEvent struct {
	eventsourcing.BaseEvent

	WalletID string `json:"wallet_id"`
	UserID   string `json:"user_id"`
	OldLabel string `json:"old_label"`
	NewLabel string `json:"new_label"`
}

// EventType returns the event type.
func (e *WalletLabelUpdatedEvent) EventType() string {
	return EventTypeWalletLabelUpdated
}

// NewWalletLabelUpdatedEvent creates a new WalletLabelUpdatedEvent.
func NewWalletLabelUpdatedEvent(wallet *Wallet, oldLabel, newLabel string) *WalletLabelUpdatedEvent {
	return &WalletLabelUpdatedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:  wallet.ID(),
		UserID:    wallet.UserID(),
		OldLabel:  oldLabel,
		NewLabel:  newLabel,
	}
}

// ============================================================================
// Wallet Primary Set Event
// ============================================================================

// WalletPrimarySetEvent is emitted when a wallet is set as primary.
type WalletPrimarySetEvent struct {
	eventsourcing.BaseEvent

	WalletID          string `json:"wallet_id"`
	UserID            string `json:"user_id"`
	PreviousPrimaryID string `json:"previous_primary_id,omitempty"`
}

// EventType returns the event type.
func (e *WalletPrimarySetEvent) EventType() string {
	return EventTypeWalletPrimarySet
}

// NewWalletPrimarySetEvent creates a new WalletPrimarySetEvent.
func NewWalletPrimarySetEvent(wallet *Wallet, previousPrimaryID string) *WalletPrimarySetEvent {
	return &WalletPrimarySetEvent{
		BaseEvent:         eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:          wallet.ID(),
		UserID:            wallet.UserID(),
		PreviousPrimaryID: previousPrimaryID,
	}
}

// ============================================================================
// Wallet Status Events
// ============================================================================

// WalletActivatedEvent is emitted when a wallet is activated.
type WalletActivatedEvent struct {
	eventsourcing.BaseEvent

	WalletID       string `json:"wallet_id"`
	UserID         string `json:"user_id"`
	PreviousStatus string `json:"previous_status"`
}

// EventType returns the event type.
func (e *WalletActivatedEvent) EventType() string {
	return EventTypeWalletActivated
}

// NewWalletActivatedEvent creates a new WalletActivatedEvent.
func NewWalletActivatedEvent(wallet *Wallet, previousStatus WalletStatus) *WalletActivatedEvent {
	return &WalletActivatedEvent{
		BaseEvent:      eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:       wallet.ID(),
		UserID:         wallet.UserID(),
		PreviousStatus: previousStatus.String(),
	}
}

// WalletDeactivatedEvent is emitted when a wallet is deactivated.
type WalletDeactivatedEvent struct {
	eventsourcing.BaseEvent

	WalletID       string `json:"wallet_id"`
	UserID         string `json:"user_id"`
	PreviousStatus string `json:"previous_status"`
	Reason         string `json:"reason,omitempty"`
}

// EventType returns the event type.
func (e *WalletDeactivatedEvent) EventType() string {
	return EventTypeWalletDeactivated
}

// NewWalletDeactivatedEvent creates a new WalletDeactivatedEvent.
func NewWalletDeactivatedEvent(wallet *Wallet, previousStatus WalletStatus, reason string) *WalletDeactivatedEvent {
	return &WalletDeactivatedEvent{
		BaseEvent:      eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:       wallet.ID(),
		UserID:         wallet.UserID(),
		PreviousStatus: previousStatus.String(),
		Reason:         reason,
	}
}

// WalletSuspendedEvent is emitted when a wallet is suspended.
type WalletSuspendedEvent struct {
	eventsourcing.BaseEvent

	WalletID       string `json:"wallet_id"`
	UserID         string `json:"user_id"`
	PreviousStatus string `json:"previous_status"`
	Reason         string `json:"reason"`
}

// EventType returns the event type.
func (e *WalletSuspendedEvent) EventType() string {
	return EventTypeWalletSuspended
}

// NewWalletSuspendedEvent creates a new WalletSuspendedEvent.
func NewWalletSuspendedEvent(wallet *Wallet, previousStatus WalletStatus, reason string) *WalletSuspendedEvent {
	return &WalletSuspendedEvent{
		BaseEvent:      eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:       wallet.ID(),
		UserID:         wallet.UserID(),
		PreviousStatus: previousStatus.String(),
		Reason:         reason,
	}
}

// ============================================================================
// Wallet Used Event
// ============================================================================

// WalletUsedEvent is emitted when a wallet is used for authentication.
type WalletUsedEvent struct {
	eventsourcing.BaseEvent

	WalletID  string    `json:"wallet_id"`
	UserID    string    `json:"user_id"`
	Address   string    `json:"address"`
	UsedAt    time.Time `json:"used_at"`
	Action    string    `json:"action"` // "login", "sign", "verify"
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// EventType returns the event type.
func (e *WalletUsedEvent) EventType() string {
	return EventTypeWalletUsed
}

// NewWalletUsedEvent creates a new WalletUsedEvent.
func NewWalletUsedEvent(wallet *Wallet, action, ipAddress, userAgent string) *WalletUsedEvent {
	return &WalletUsedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("Wallet", wallet.ID()),
		WalletID:  wallet.ID(),
		UserID:    wallet.UserID(),
		Address:   wallet.AddressString(),
		UsedAt:    time.Now(),
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// ============================================================================
// Challenge Events
// ============================================================================

// ChallengeCreatedEvent is emitted when a challenge is created.
type ChallengeCreatedEvent struct {
	eventsourcing.BaseEvent

	Nonce     string    `json:"nonce"`
	Address   string    `json:"address"`
	ChainID   string    `json:"chain_id"`
	Domain    string    `json:"domain"`
	URI       string    `json:"uri"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// EventType returns the event type.
func (e *ChallengeCreatedEvent) EventType() string {
	return EventTypeChallengeCreated
}

// NewChallengeCreatedEvent creates a new ChallengeCreatedEvent.
func NewChallengeCreatedEvent(challenge *Challenge) *ChallengeCreatedEvent {
	return &ChallengeCreatedEvent{
		BaseEvent: eventsourcing.NewBaseEvent("Wallet", challenge.Nonce),
		Nonce:     challenge.Nonce,
		Address:   challenge.Address.Normalized(),
		ChainID:   challenge.Address.ChainID().String(),
		Domain:    challenge.Domain,
		URI:       challenge.URI,
		IssuedAt:  challenge.IssuedAt,
		ExpiresAt: challenge.ExpiresAt,
	}
}

// ChallengeVerifiedEvent is emitted when a challenge is verified.
type ChallengeVerifiedEvent struct {
	eventsourcing.BaseEvent

	Nonce      string    `json:"nonce"`
	Address    string    `json:"address"`
	VerifiedAt time.Time `json:"verified_at"`
	WalletID   string    `json:"wallet_id,omitempty"`
}

// EventType returns the event type.
func (e *ChallengeVerifiedEvent) EventType() string {
	return EventTypeChallengeVerified
}

// NewChallengeVerifiedEvent creates a new ChallengeVerifiedEvent.
func NewChallengeVerifiedEvent(challenge *Challenge, walletID string) *ChallengeVerifiedEvent {
	return &ChallengeVerifiedEvent{
		BaseEvent:  eventsourcing.NewBaseEvent("Wallet", challenge.Nonce),
		Nonce:      challenge.Nonce,
		Address:    challenge.Address.Normalized(),
		VerifiedAt: time.Now(),
		WalletID:   walletID,
	}
}

// ChallengeExpiredEvent is emitted when a challenge expires.
type ChallengeExpiredEvent struct {
	eventsourcing.BaseEvent

	Nonce     string    `json:"nonce"`
	Address   string    `json:"address"`
	ExpiredAt time.Time `json:"expired_at"`
}

// EventType returns the event type.
func (e *ChallengeExpiredEvent) EventType() string {
	return EventTypeChallengeExpired
}

// NewChallengeExpiredEvent creates a new ChallengeExpiredEvent.
func NewChallengeExpiredEvent(challenge *Challenge) *ChallengeExpiredEvent {
	return &ChallengeExpiredEvent{
		BaseEvent: eventsourcing.NewBaseEvent("Wallet", challenge.Nonce),
		Nonce:     challenge.Nonce,
		Address:   challenge.Address.Normalized(),
		ExpiredAt: time.Now(),
	}
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterWalletEvents registers all wallet events with the registry.
func RegisterWalletEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeWalletLinked, func() eventsourcing.Event { return &WalletLinkedEvent{} })
	registry.Register(EventTypeWalletUnlinked, func() eventsourcing.Event { return &WalletUnlinkedEvent{} })
	registry.Register(EventTypeWalletVerified, func() eventsourcing.Event { return &WalletVerifiedEvent{} })
	registry.Register(EventTypeWalletLabelUpdated, func() eventsourcing.Event { return &WalletLabelUpdatedEvent{} })
	registry.Register(EventTypeWalletPrimarySet, func() eventsourcing.Event { return &WalletPrimarySetEvent{} })
	registry.Register(EventTypeWalletActivated, func() eventsourcing.Event { return &WalletActivatedEvent{} })
	registry.Register(EventTypeWalletDeactivated, func() eventsourcing.Event { return &WalletDeactivatedEvent{} })
	registry.Register(EventTypeWalletSuspended, func() eventsourcing.Event { return &WalletSuspendedEvent{} })
	registry.Register(EventTypeWalletUsed, func() eventsourcing.Event { return &WalletUsedEvent{} })
	registry.Register(EventTypeChallengeCreated, func() eventsourcing.Event { return &ChallengeCreatedEvent{} })
	registry.Register(EventTypeChallengeVerified, func() eventsourcing.Event { return &ChallengeVerifiedEvent{} })
	registry.Register(EventTypeChallengeExpired, func() eventsourcing.Event { return &ChallengeExpiredEvent{} })
}
