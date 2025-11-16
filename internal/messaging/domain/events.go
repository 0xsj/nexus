package domain

import (
	"time"

	"github.com/0xsj/nexus/pkg/events"
)

const (
	// Event types - Conversations
	EventTypeConversationCreated = "messaging.conversation.created"
	EventTypeConversationUpdated = "messaging.conversation.updated"
	EventTypeConversationDeleted = "messaging.conversation.deleted"
	EventTypeConversationMuted   = "messaging.conversation.muted"
	EventTypeConversationUnmuted = "messaging.conversation.unmuted"

	// Event types - Participants
	EventTypeParticipantAdded   = "messaging.participant.added"
	EventTypeParticipantRemoved = "messaging.participant.removed"
	EventTypeParticipantLeft    = "messaging.participant.left"

	// Event types - Messages
	EventTypeMessageSent      = "messaging.message.sent"
	EventTypeMessageEdited    = "messaging.message.edited"
	EventTypeMessageDeleted   = "messaging.message.deleted"
	EventTypeMessageRead      = "messaging.message.read"
	EventTypeMessageDelivered = "messaging.message.delivered"
	EventTypeMessageFailed    = "messaging.message.failed"

	// Event types - Interactions
	EventTypeReactionAdded   = "messaging.reaction.added"
	EventTypeReactionRemoved = "messaging.reaction.removed"
	EventTypeTypingStarted   = "messaging.typing.started"
	EventTypeTypingStopped   = "messaging.typing.stopped"

	// Aggregate type
	AggregateTypeMessaging = "messaging"
)

// ============================================================================
// Conversation Events
// ============================================================================

// ConversationCreatedPayload is the payload for conversation created event.
type ConversationCreatedPayload struct {
	ConversationID   string           `json:"conversation_id"`
	Type             ConversationType `json:"type"`
	CreatedBy        string           `json:"created_by"`
	ParticipantIDs   []string         `json:"participant_ids"`
	ParticipantCount int              `json:"participant_count"`
}

// NewConversationCreatedEvent creates a new conversation created event.
func NewConversationCreatedEvent(conversation *Conversation) events.Event {
	payload := ConversationCreatedPayload{
		ConversationID:   conversation.ID,
		Type:             conversation.Type,
		CreatedBy:        conversation.CreatedBy,
		ParticipantIDs:   conversation.ParticipantIDs,
		ParticipantCount: len(conversation.ParticipantIDs),
	}

	event := events.NewBaseEvent(
		EventTypeConversationCreated,
		AggregateTypeMessaging,
		conversation.ID,
		payload,
	)

	event.WithMetadata("conversation_type", string(conversation.Type))
	event.WithMetadata("created_by", conversation.CreatedBy)

	return event
}

// ConversationUpdatedPayload is the payload for conversation updated event.
type ConversationUpdatedPayload struct {
	ConversationID string   `json:"conversation_id"`
	UpdatedFields  []string `json:"updated_fields"`
}

// NewConversationUpdatedEvent creates a new conversation updated event.
func NewConversationUpdatedEvent(conversation *Conversation, updatedFields []string) events.Event {
	payload := ConversationUpdatedPayload{
		ConversationID: conversation.ID,
		UpdatedFields:  updatedFields,
	}

	event := events.NewBaseEvent(
		EventTypeConversationUpdated,
		AggregateTypeMessaging,
		conversation.ID,
		payload,
	)

	return event
}

// ConversationDeletedPayload is the payload for conversation deleted event.
type ConversationDeletedPayload struct {
	ConversationID string `json:"conversation_id"`
}

// NewConversationDeletedEvent creates a new conversation deleted event.
func NewConversationDeletedEvent(conversationID string) events.Event {
	payload := ConversationDeletedPayload{
		ConversationID: conversationID,
	}

	event := events.NewBaseEvent(
		EventTypeConversationDeleted,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	return event
}

// ConversationMutedPayload is the payload for conversation muted event.
type ConversationMutedPayload struct {
	ConversationID string     `json:"conversation_id"`
	UserID         string     `json:"user_id"`
	MutedUntil     *time.Time `json:"muted_until,omitempty"`
}

// NewConversationMutedEvent creates a new conversation muted event.
func NewConversationMutedEvent(conversation *Conversation, userID string) events.Event {
	payload := ConversationMutedPayload{
		ConversationID: conversation.ID,
		UserID:         userID,
		MutedUntil:     conversation.MutedUntil,
	}

	event := events.NewBaseEvent(
		EventTypeConversationMuted,
		AggregateTypeMessaging,
		conversation.ID,
		payload,
	)

	event.WithMetadata("user_id", userID)

	return event
}

// ConversationUnmutedPayload is the payload for conversation unmuted event.
type ConversationUnmutedPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

// NewConversationUnmutedEvent creates a new conversation unmuted event.
func NewConversationUnmutedEvent(conversationID, userID string) events.Event {
	payload := ConversationUnmutedPayload{
		ConversationID: conversationID,
		UserID:         userID,
	}

	event := events.NewBaseEvent(
		EventTypeConversationUnmuted,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	event.WithMetadata("user_id", userID)

	return event
}

// ============================================================================
// Participant Events
// ============================================================================

// ParticipantAddedPayload is the payload for participant added event.
type ParticipantAddedPayload struct {
	ConversationID string          `json:"conversation_id"`
	UserID         string          `json:"user_id"`
	AddedBy        string          `json:"added_by"`
	Role           ParticipantRole `json:"role"`
}

// NewParticipantAddedEvent creates a new participant added event.
func NewParticipantAddedEvent(conversationID, userID, addedBy string, role ParticipantRole) events.Event {
	payload := ParticipantAddedPayload{
		ConversationID: conversationID,
		UserID:         userID,
		AddedBy:        addedBy,
		Role:           role,
	}

	event := events.NewBaseEvent(
		EventTypeParticipantAdded,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	event.WithMetadata("user_id", userID)
	event.WithMetadata("added_by", addedBy)

	return event
}

// ParticipantRemovedPayload is the payload for participant removed event.
type ParticipantRemovedPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	RemovedBy      string `json:"removed_by"`
}

// NewParticipantRemovedEvent creates a new participant removed event.
func NewParticipantRemovedEvent(conversationID, userID, removedBy string) events.Event {
	payload := ParticipantRemovedPayload{
		ConversationID: conversationID,
		UserID:         userID,
		RemovedBy:      removedBy,
	}

	event := events.NewBaseEvent(
		EventTypeParticipantRemoved,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	event.WithMetadata("user_id", userID)
	event.WithMetadata("removed_by", removedBy)

	return event
}

// ParticipantLeftPayload is the payload for participant left event.
type ParticipantLeftPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

// NewParticipantLeftEvent creates a new participant left event.
func NewParticipantLeftEvent(conversationID, userID string) events.Event {
	payload := ParticipantLeftPayload{
		ConversationID: conversationID,
		UserID:         userID,
	}

	event := events.NewBaseEvent(
		EventTypeParticipantLeft,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	event.WithMetadata("user_id", userID)

	return event
}

// ============================================================================
// Message Events
// ============================================================================

// MessageSentPayload is the payload for message sent event.
type MessageSentPayload struct {
	MessageID      string      `json:"message_id"`
	ConversationID string      `json:"conversation_id"`
	SenderID       string      `json:"sender_id"`
	Type           MessageType `json:"type"`
	HasAttachments bool        `json:"has_attachments"`
	IsReply        bool        `json:"is_reply"`
}

// NewMessageSentEvent creates a new message sent event.
func NewMessageSentEvent(message *Message) events.Event {
	payload := MessageSentPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Type:           message.Type,
		HasAttachments: message.HasAttachments(),
		IsReply:        message.IsReply(),
	}

	event := events.NewBaseEvent(
		EventTypeMessageSent,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("sender_id", message.SenderID)
	event.WithMetadata("message_type", string(message.Type))

	return event
}

// MessageEditedPayload is the payload for message edited event.
type MessageEditedPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
}

// NewMessageEditedEvent creates a new message edited event.
func NewMessageEditedEvent(message *Message) events.Event {
	payload := MessageEditedPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
	}

	event := events.NewBaseEvent(
		EventTypeMessageEdited,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("sender_id", message.SenderID)

	return event
}

// MessageDeletedPayload is the payload for message deleted event.
type MessageDeletedPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
}

// NewMessageDeletedEvent creates a new message deleted event.
func NewMessageDeletedEvent(message *Message) events.Event {
	payload := MessageDeletedPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
	}

	event := events.NewBaseEvent(
		EventTypeMessageDeleted,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("sender_id", message.SenderID)

	return event
}

// MessageReadPayload is the payload for message read event.
type MessageReadPayload struct {
	MessageID      string    `json:"message_id"`
	ConversationID string    `json:"conversation_id"`
	UserID         string    `json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}

// NewMessageReadEvent creates a new message read event.
func NewMessageReadEvent(message *Message, userID string) events.Event {
	readAt := time.Now().UTC()
	if timestamp, exists := message.ReadBy[userID]; exists {
		readAt = timestamp
	}

	payload := MessageReadPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		UserID:         userID,
		ReadAt:         readAt,
	}

	event := events.NewBaseEvent(
		EventTypeMessageRead,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("user_id", userID)

	return event
}

// MessageDeliveredPayload is the payload for message delivered event.
type MessageDeliveredPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
}

// NewMessageDeliveredEvent creates a new message delivered event.
func NewMessageDeliveredEvent(message *Message) events.Event {
	payload := MessageDeliveredPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
	}

	event := events.NewBaseEvent(
		EventTypeMessageDelivered,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)

	return event
}

// MessageFailedPayload is the payload for message failed event.
type MessageFailedPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
}

// NewMessageFailedEvent creates a new message failed event.
func NewMessageFailedEvent(message *Message) events.Event {
	payload := MessageFailedPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
	}

	event := events.NewBaseEvent(
		EventTypeMessageFailed,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("sender_id", message.SenderID)

	return event
}

// ============================================================================
// Interaction Events
// ============================================================================

// ReactionAddedPayload is the payload for reaction added event.
type ReactionAddedPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Emoji          string `json:"emoji"`
}

// NewReactionAddedEvent creates a new reaction added event.
func NewReactionAddedEvent(message *Message, userID, emoji string) events.Event {
	payload := ReactionAddedPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		UserID:         userID,
		Emoji:          emoji,
	}

	event := events.NewBaseEvent(
		EventTypeReactionAdded,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("user_id", userID)

	return event
}

// ReactionRemovedPayload is the payload for reaction removed event.
type ReactionRemovedPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

// NewReactionRemovedEvent creates a new reaction removed event.
func NewReactionRemovedEvent(message *Message, userID string) events.Event {
	payload := ReactionRemovedPayload{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		UserID:         userID,
	}

	event := events.NewBaseEvent(
		EventTypeReactionRemoved,
		AggregateTypeMessaging,
		message.ConversationID,
		payload,
	)

	event.WithMetadata("message_id", message.ID)
	event.WithMetadata("user_id", userID)

	return event
}

// TypingStartedPayload is the payload for typing started event.
type TypingStartedPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

// NewTypingStartedEvent creates a new typing started event.
func NewTypingStartedEvent(conversationID, userID string) events.Event {
	payload := TypingStartedPayload{
		ConversationID: conversationID,
		UserID:         userID,
	}

	event := events.NewBaseEvent(
		EventTypeTypingStarted,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	event.WithMetadata("user_id", userID)

	return event
}

// TypingStoppedPayload is the payload for typing stopped event.
type TypingStoppedPayload struct {
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
}

// NewTypingStoppedEvent creates a new typing stopped event.
func NewTypingStoppedEvent(conversationID, userID string) events.Event {
	payload := TypingStoppedPayload{
		ConversationID: conversationID,
		UserID:         userID,
	}

	event := events.NewBaseEvent(
		EventTypeTypingStopped,
		AggregateTypeMessaging,
		conversationID,
		payload,
	)

	event.WithMetadata("user_id", userID)

	return event
}
