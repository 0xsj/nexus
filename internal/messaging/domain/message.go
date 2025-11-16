package domain

import (
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

const (
	MaxMessageLength = 10000
	MaxAttachments   = 10
)

// Message represents a chat message (aggregate root).
type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	Type           MessageType
	Content        string
	Status         MessageStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EditedAt       *time.Time
	DeletedAt      *time.Time

	// Reply to another message
	ReplyToMessageID *string

	// Attachments
	Attachments []*MediaAttachment

	// Read receipts
	ReadBy map[string]time.Time // userID -> read timestamp

	// Reactions
	Reactions map[string]string // userID -> emoji

	// Translation (optional)
	TranslatedContent *string
	TranslatedTo      *string
}

// NewMessage creates a new message.
func NewMessage(
	conversationID, senderID string,
	messageType MessageType,
	content string,
) result.Result[*Message] {
	// Validate conversation ID
	if conversationID == "" {
		return result.Err[*Message](ErrInvalidConversationID())
	}

	// Validate sender ID
	if senderID == "" {
		return result.Err[*Message](ErrInvalidMessageID())
	}

	// Validate message type
	if !messageType.IsValid() {
		return result.Err[*Message](ErrInvalidMessageType())
	}

	// Validate content for text messages
	if messageType.RequiresContent() {
		if content == "" {
			return result.Err[*Message](ErrEmptyMessageContent())
		}

		if len(content) > MaxMessageLength {
			return result.Err[*Message](ErrMessageContentTooLong(len(content), MaxMessageLength))
		}
	}

	now := time.Now().UTC()

	message := &Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Type:           messageType,
		Content:        content,
		Status:         MessageStatusSending,
		CreatedAt:      now,
		UpdatedAt:      now,
		Attachments:    []*MediaAttachment{},
		ReadBy:         make(map[string]time.Time),
		Reactions:      make(map[string]string),
	}

	return result.Ok(message)
}

// NewSystemMessage creates a system message.
func NewSystemMessage(conversationID, content string) result.Result[*Message] {
	return NewMessage(conversationID, "system", MessageTypeSystem, content)
}

// ============================================================================
// Message Updates
// ============================================================================

// UpdateContent updates the message content.
func (m *Message) UpdateContent(content string) error {
	// Cannot edit system messages
	if m.Type.IsSystem() {
		return ErrCannotModifySystemMessage()
	}

	// Cannot edit deleted messages
	if m.IsDeleted() {
		return ErrMessageAlreadyDeleted()
	}

	// Validate content
	if content == "" {
		return ErrEmptyMessageContent()
	}

	if len(content) > MaxMessageLength {
		return ErrMessageContentTooLong(len(content), MaxMessageLength)
	}

	m.Content = content
	now := time.Now().UTC()
	m.EditedAt = &now
	m.UpdatedAt = now
	return nil
}

// Delete soft-deletes the message.
func (m *Message) Delete() error {
	// Cannot delete system messages
	if m.Type.IsSystem() {
		return ErrCannotDeleteSystemMessage()
	}

	// Already deleted
	if m.IsDeleted() {
		return ErrMessageAlreadyDeleted()
	}

	now := time.Now().UTC()
	m.DeletedAt = &now
	m.UpdatedAt = now
	return nil
}

// ============================================================================
// Status Management
// ============================================================================

// MarkAsSent marks the message as sent.
func (m *Message) MarkAsSent() {
	if m.Status == MessageStatusSending {
		m.Status = MessageStatusSent
		m.UpdatedAt = time.Now().UTC()
	}
}

// MarkAsDelivered marks the message as delivered.
func (m *Message) MarkAsDelivered() {
	if m.Status == MessageStatusSent {
		m.Status = MessageStatusDelivered
		m.UpdatedAt = time.Now().UTC()
	}
}

// MarkAsRead marks the message as read by a user.
func (m *Message) MarkAsRead(userID string) error {
	// Cannot mark own messages as read
	if userID == m.SenderID {
		return nil
	}

	// Check if already read
	if _, exists := m.ReadBy[userID]; exists {
		return ErrMessageAlreadyRead()
	}

	m.ReadBy[userID] = time.Now().UTC()
	m.Status = MessageStatusRead
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkAsFailed marks the message as failed to send.
func (m *Message) MarkAsFailed() {
	m.Status = MessageStatusFailed
	m.UpdatedAt = time.Now().UTC()
}

// ============================================================================
// Attachments
// ============================================================================

// AddAttachment adds a media attachment.
func (m *Message) AddAttachment(attachment *MediaAttachment) error {
	if len(m.Attachments) >= MaxAttachments {
		return ErrTooManyAttachments(len(m.Attachments)+1, MaxAttachments)
	}

	m.Attachments = append(m.Attachments, attachment)
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// SetAttachments sets all attachments at once.
func (m *Message) SetAttachments(attachments []*MediaAttachment) error {
	if len(attachments) > MaxAttachments {
		return ErrTooManyAttachments(len(attachments), MaxAttachments)
	}

	m.Attachments = attachments
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// HasAttachments checks if the message has attachments.
func (m *Message) HasAttachments() bool {
	return len(m.Attachments) > 0
}

// ============================================================================
// Reply
// ============================================================================

// SetReplyTo sets the message this is replying to.
func (m *Message) SetReplyTo(messageID string) {
	m.ReplyToMessageID = &messageID
	m.UpdatedAt = time.Now().UTC()
}

// IsReply checks if this message is a reply.
func (m *Message) IsReply() bool {
	return m.ReplyToMessageID != nil
}

// ============================================================================
// Reactions
// ============================================================================

// AddReaction adds a reaction from a user.
func (m *Message) AddReaction(userID, emoji string) {
	m.Reactions[userID] = emoji
	m.UpdatedAt = time.Now().UTC()
}

// RemoveReaction removes a reaction from a user.
func (m *Message) RemoveReaction(userID string) {
	delete(m.Reactions, userID)
	m.UpdatedAt = time.Now().UTC()
}

// GetReaction gets a user's reaction.
func (m *Message) GetReaction(userID string) (string, bool) {
	emoji, exists := m.Reactions[userID]
	return emoji, exists
}

// HasReactions checks if the message has any reactions.
func (m *Message) HasReactions() bool {
	return len(m.Reactions) > 0
}

// ============================================================================
// Translation
// ============================================================================

// SetTranslation sets the translated content.
func (m *Message) SetTranslation(translatedContent, language string) {
	m.TranslatedContent = &translatedContent
	m.TranslatedTo = &language
	m.UpdatedAt = time.Now().UTC()
}

// HasTranslation checks if the message has a translation.
func (m *Message) HasTranslation() bool {
	return m.TranslatedContent != nil
}

// ============================================================================
// Query Methods
// ============================================================================

// IsDeleted checks if the message is deleted.
func (m *Message) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsEdited checks if the message has been edited.
func (m *Message) IsEdited() bool {
	return m.EditedAt != nil
}

// IsReadBy checks if the message has been read by a specific user.
func (m *Message) IsReadBy(userID string) bool {
	_, exists := m.ReadBy[userID]
	return exists
}

// GetReadCount returns the number of users who have read the message.
func (m *Message) GetReadCount() int {
	return len(m.ReadBy)
}

// IsOwnedBy checks if the message was sent by a specific user.
func (m *Message) IsOwnedBy(userID string) bool {
	return m.SenderID == userID
}

// CanBeEditedBy checks if a user can edit this message.
func (m *Message) CanBeEditedBy(userID string) bool {
	return m.IsOwnedBy(userID) && !m.IsDeleted() && !m.Type.IsSystem()
}

// CanBeDeletedBy checks if a user can delete this message.
func (m *Message) CanBeDeletedBy(userID string) bool {
	return m.IsOwnedBy(userID) && !m.IsDeleted() && !m.Type.IsSystem()
}

// IsSystemMessage checks if this is a system message.
func (m *Message) IsSystemMessage() bool {
	return m.Type.IsSystem()
}
