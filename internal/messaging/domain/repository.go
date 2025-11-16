package domain

import (
	"context"
	"time"

	"github.com/0xsj/result"
)

// ConversationRepository defines operations for conversation persistence.
type ConversationRepository interface {
	// Create persists a new conversation.
	Create(ctx context.Context, conversation *Conversation) result.Result[*Conversation]

	// FindByID retrieves a conversation by its ID.
	FindByID(ctx context.Context, id string) result.Result[*Conversation]

	// FindDirectConversation finds a direct conversation between two users.
	// Returns ErrConversationNotFound if no conversation exists.
	FindDirectConversation(ctx context.Context, user1ID, user2ID string) result.Result[*Conversation]

	// Update updates an existing conversation.
	Update(ctx context.Context, conversation *Conversation) result.Result[*Conversation]

	// Delete soft-deletes a conversation.
	Delete(ctx context.Context, id string) result.Result[struct{}]

	// ListByUser retrieves conversations for a user with pagination.
	ListByUser(ctx context.Context, userID string, params ConversationListParams) result.Result[*ConversationListResult]

	// ExistsByID checks if a conversation exists.
	ExistsByID(ctx context.Context, id string) result.Result[bool]

	// GetUnreadCount gets the total unread message count for a user across all conversations.
	GetUnreadCount(ctx context.Context, userID string) result.Result[int]

	// UpdateLastMessage updates the last message info for a conversation.
	UpdateLastMessage(ctx context.Context, conversationID, messageID string) result.Result[struct{}]
}

// MessageRepository defines operations for message persistence.
type MessageRepository interface {
	// Create persists a new message.
	Create(ctx context.Context, message *Message) result.Result[*Message]

	// FindByID retrieves a message by its ID.
	FindByID(ctx context.Context, conversationID, messageID string) result.Result[*Message]

	// Update updates an existing message.
	Update(ctx context.Context, message *Message) result.Result[*Message]

	// Delete soft-deletes a message.
	Delete(ctx context.Context, conversationID, messageID string) result.Result[struct{}]

	// ListByConversation retrieves messages in a conversation with pagination.
	ListByConversation(ctx context.Context, conversationID string, params MessageListParams) result.Result[*MessageListResult]

	// GetUnreadMessages retrieves unread messages for a user in a conversation.
	GetUnreadMessages(ctx context.Context, conversationID, userID string) result.Result[[]*Message]

	// MarkAsRead marks messages as read by a user.
	MarkAsRead(ctx context.Context, conversationID, userID string, messageIDs []string) result.Result[struct{}]

	// MarkAllAsRead marks all messages in a conversation as read by a user.
	MarkAllAsRead(ctx context.Context, conversationID, userID string) result.Result[struct{}]

	// GetMessageCount returns the total message count in a conversation.
	GetMessageCount(ctx context.Context, conversationID string) result.Result[int]

	// GetUnreadCount returns the unread message count for a user in a conversation.
	GetUnreadCount(ctx context.Context, conversationID, userID string) result.Result[int]

	// SearchMessages searches messages in a conversation.
	SearchMessages(ctx context.Context, conversationID, query string, params MessageListParams) result.Result[*MessageListResult]
}

// ParticipantRepository defines operations for participant persistence.
type ParticipantRepository interface {
	// Add adds a participant to a conversation.
	Add(ctx context.Context, conversationID string, participant *Participant) result.Result[*Participant]

	// Remove removes a participant from a conversation.
	Remove(ctx context.Context, conversationID, userID string) result.Result[struct{}]

	// FindByConversation retrieves all participants in a conversation.
	FindByConversation(ctx context.Context, conversationID string) result.Result[[]*Participant]

	// FindByUser retrieves a participant by user ID and conversation ID.
	FindByUser(ctx context.Context, conversationID, userID string) result.Result[*Participant]

	// Update updates a participant.
	Update(ctx context.Context, conversationID string, participant *Participant) result.Result[*Participant]

	// UpdateLastSeen updates the last seen timestamp for a participant.
	UpdateLastSeen(ctx context.Context, conversationID, userID string) result.Result[struct{}]

	// UpdateLastRead updates the last read message ID for a participant.
	UpdateLastRead(ctx context.Context, conversationID, userID, messageID string) result.Result[struct{}]

	// IsParticipant checks if a user is a participant in a conversation.
	IsParticipant(ctx context.Context, conversationID, userID string) result.Result[bool]
}

// ConversationListParams defines parameters for listing conversations.
type ConversationListParams struct {
	Limit  int
	Offset int
	Type   *ConversationType
	SortBy string // "updated_at", "created_at"
	Order  string // "asc", "desc"
}

// ConversationListResult contains paginated conversation results.
type ConversationListResult struct {
	Conversations []*Conversation
	Total         int
	Limit         int
	Offset        int
	HasMore       bool
}

// MessageListParams defines parameters for listing messages.
type MessageListParams struct {
	Limit  int
	Offset int
	Before *time.Time // Get messages before this time
	After  *time.Time // Get messages after this time
	SortBy string     // "created_at"
	Order  string     // "asc", "desc"
}

// MessageListResult contains paginated message results.
type MessageListResult struct {
	Messages []*Message
	Total    int
	Limit    int
	Offset   int
	HasMore  bool
}

// NewConversationListParams creates ConversationListParams with defaults.
func NewConversationListParams(limit, offset int) ConversationListParams {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return ConversationListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: "updated_at",
		Order:  "desc",
	}
}

// WithType filters by conversation type.
func (p ConversationListParams) WithType(conversationType ConversationType) ConversationListParams {
	p.Type = &conversationType
	return p
}

// WithSortBy sets the sort field.
func (p ConversationListParams) WithSortBy(sortBy string) ConversationListParams {
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
	}

	if validSortFields[sortBy] {
		p.SortBy = sortBy
	}

	return p
}

// WithOrder sets the sort order.
func (p ConversationListParams) WithOrder(order string) ConversationListParams {
	if order == "asc" || order == "desc" {
		p.Order = order
	}

	return p
}

// NewMessageListParams creates MessageListParams with defaults.
func NewMessageListParams(limit, offset int) MessageListParams {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	return MessageListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: "created_at",
		Order:  "desc",
	}
}

// WithBefore sets the before timestamp filter.
func (p MessageListParams) WithBefore(before time.Time) MessageListParams {
	p.Before = &before
	return p
}

// WithAfter sets the after timestamp filter.
func (p MessageListParams) WithAfter(after time.Time) MessageListParams {
	p.After = &after
	return p
}

// WithSortBy sets the sort field.
func (p MessageListParams) WithSortBy(sortBy string) MessageListParams {
	if sortBy == "created_at" {
		p.SortBy = sortBy
	}

	return p
}

// WithOrder sets the sort order.
func (p MessageListParams) WithOrder(order string) MessageListParams {
	if order == "asc" || order == "desc" {
		p.Order = order
	}

	return p
}
