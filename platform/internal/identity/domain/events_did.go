package domain

import (
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Event Types
// ============================================================================

const (
	EventTypeDIDLinked         = "identity.did.linked"
	EventTypeDIDUnlinked       = "identity.did.unlinked"
	EventTypeDIDPrimaryChanged = "identity.did.primary_changed"
	EventTypeDIDLabelUpdated   = "identity.did.label_updated"
)

// ============================================================================
// DID Linked Event
// ============================================================================

// DIDLinkedEvent is raised when a DID is linked to a user.
type DIDLinkedEvent struct {
	eventsourcing.BaseEvent

	// LinkedDIDID is the unique ID of the LinkedDID entry.
	LinkedDIDID string `json:"linked_did_id"`

	// DID is the decentralized identifier string.
	DID string `json:"did"`

	// Source indicates how the DID was created (custodial, wallet, imported).
	Source string `json:"source"`

	// IsPrimary indicates if this DID is set as primary.
	IsPrimary bool `json:"is_primary"`

	// Label is the user-friendly name for this DID.
	Label string `json:"label"`

	// Metadata contains source-specific data.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EventType returns the event type.
func (e *DIDLinkedEvent) EventType() string {
	return EventTypeDIDLinked
}

// NewDIDLinkedEvent creates a new DIDLinkedEvent.
func NewDIDLinkedEvent(
	userID string,
	linkedDIDID string,
	did string,
	source DIDSource,
	isPrimary bool,
	label string,
	metadata map[string]string,
) *DIDLinkedEvent {
	return &DIDLinkedEvent{
		BaseEvent:   eventsourcing.NewBaseEvent(UserAggregateType, userID),
		LinkedDIDID: linkedDIDID,
		DID:         did,
		Source:      source.String(),
		IsPrimary:   isPrimary,
		Label:       label,
		Metadata:    metadata,
	}
}

// ============================================================================
// DID Unlinked Event
// ============================================================================

// DIDUnlinkedEvent is raised when a DID is unlinked from a user.
type DIDUnlinkedEvent struct {
	eventsourcing.BaseEvent

	// LinkedDIDID is the unique ID of the LinkedDID entry being removed.
	LinkedDIDID string `json:"linked_did_id"`

	// DID is the decentralized identifier string being unlinked.
	DID string `json:"did"`

	// Reason is an optional reason for unlinking.
	Reason string `json:"reason,omitempty"`
}

// EventType returns the event type.
func (e *DIDUnlinkedEvent) EventType() string {
	return EventTypeDIDUnlinked
}

// NewDIDUnlinkedEvent creates a new DIDUnlinkedEvent.
func NewDIDUnlinkedEvent(userID, linkedDIDID, did, reason string) *DIDUnlinkedEvent {
	return &DIDUnlinkedEvent{
		BaseEvent:   eventsourcing.NewBaseEvent(UserAggregateType, userID),
		LinkedDIDID: linkedDIDID,
		DID:         did,
		Reason:      reason,
	}
}

// ============================================================================
// DID Primary Changed Event
// ============================================================================

// DIDPrimaryChangedEvent is raised when the primary DID is changed.
type DIDPrimaryChangedEvent struct {
	eventsourcing.BaseEvent

	// OldPrimaryDIDID is the LinkedDID ID that was previously primary (empty if none).
	OldPrimaryDIDID string `json:"old_primary_did_id,omitempty"`

	// OldPrimaryDID is the DID string that was previously primary (empty if none).
	OldPrimaryDID string `json:"old_primary_did,omitempty"`

	// NewPrimaryDIDID is the LinkedDID ID that is now primary.
	NewPrimaryDIDID string `json:"new_primary_did_id"`

	// NewPrimaryDID is the DID string that is now primary.
	NewPrimaryDID string `json:"new_primary_did"`
}

// EventType returns the event type.
func (e *DIDPrimaryChangedEvent) EventType() string {
	return EventTypeDIDPrimaryChanged
}

// NewDIDPrimaryChangedEvent creates a new DIDPrimaryChangedEvent.
func NewDIDPrimaryChangedEvent(
	userID string,
	oldPrimaryDIDID, oldPrimaryDID string,
	newPrimaryDIDID, newPrimaryDID string,
) *DIDPrimaryChangedEvent {
	return &DIDPrimaryChangedEvent{
		BaseEvent:       eventsourcing.NewBaseEvent(UserAggregateType, userID),
		OldPrimaryDIDID: oldPrimaryDIDID,
		OldPrimaryDID:   oldPrimaryDID,
		NewPrimaryDIDID: newPrimaryDIDID,
		NewPrimaryDID:   newPrimaryDID,
	}
}

// ============================================================================
// DID Label Updated Event
// ============================================================================

// DIDLabelUpdatedEvent is raised when a DID's label is updated.
type DIDLabelUpdatedEvent struct {
	eventsourcing.BaseEvent

	// LinkedDIDID is the unique ID of the LinkedDID entry.
	LinkedDIDID string `json:"linked_did_id"`

	// OldLabel is the previous label.
	OldLabel string `json:"old_label"`

	// NewLabel is the new label.
	NewLabel string `json:"new_label"`
}

// EventType returns the event type.
func (e *DIDLabelUpdatedEvent) EventType() string {
	return EventTypeDIDLabelUpdated
}

// NewDIDLabelUpdatedEvent creates a new DIDLabelUpdatedEvent.
func NewDIDLabelUpdatedEvent(userID, linkedDIDID, oldLabel, newLabel string) *DIDLabelUpdatedEvent {
	return &DIDLabelUpdatedEvent{
		BaseEvent:   eventsourcing.NewBaseEvent(UserAggregateType, userID),
		LinkedDIDID: linkedDIDID,
		OldLabel:    oldLabel,
		NewLabel:    newLabel,
	}
}

// ============================================================================
// Event Registration
// ============================================================================

// RegisterDIDEvents registers DID-related events with the given registry.
// Call this in addition to RegisterEvents for complete event registration.
func RegisterDIDEvents(registry *eventsourcing.EventRegistry) {
	registry.Register(EventTypeDIDLinked, func() eventsourcing.Event { return &DIDLinkedEvent{} })
	registry.Register(EventTypeDIDUnlinked, func() eventsourcing.Event { return &DIDUnlinkedEvent{} })
	registry.Register(EventTypeDIDPrimaryChanged, func() eventsourcing.Event { return &DIDPrimaryChangedEvent{} })
	registry.Register(EventTypeDIDLabelUpdated, func() eventsourcing.Event { return &DIDLabelUpdatedEvent{} })
}
