package domain

import (
	"fmt"

	"github.com/0xsj/result"
)

// ============================================================================
// Validation Errors (Kind: KindValidation)
// ============================================================================

func ErrInvalidConversationID() error {
	return result.Validation("messaging.conversation.validate", "invalid conversation ID", nil)
}

func ErrInvalidMessageID() error {
	return result.Validation("messaging.message.validate", "invalid message ID", nil)
}

func ErrEmptyMessageContent() error {
	return result.Validation("messaging.message.validate", "message content cannot be empty", nil)
}

func ErrMessageContentTooLong(length, max int) error {
	return result.Validation(
		"messaging.message.validate",
		fmt.Sprintf("message content too long: %d characters (max: %d)", length, max),
		map[string]any{
			"length": length,
			"max":    max,
		},
	)
}

func ErrInvalidConversationType() error {
	return result.Validation("messaging.conversation.validate", "invalid conversation type", nil)
}

func ErrInvalidMessageType() error {
	return result.Validation("messaging.message.validate", "invalid message type", nil)
}

func ErrTooFewParticipants() error {
	return result.Validation("messaging.conversation.validate", "conversation must have at least 2 participants", nil)
}

func ErrTooManyParticipants(count, max int) error {
	return result.Validation(
		"messaging.conversation.validate",
		fmt.Sprintf("too many participants: %d (max: %d)", count, max),
		map[string]any{
			"count": count,
			"max":   max,
		},
	)
}

func ErrDuplicateParticipants() error {
	return result.Validation("messaging.conversation.validate", "duplicate participants not allowed", nil)
}

func ErrInvalidDirectConversation() error {
	return result.Validation("messaging.conversation.validate", "direct conversation must have exactly 2 participants", nil)
}

func ErrGroupNameRequired() error {
	return result.Validation("messaging.conversation.validate", "group conversation must have a name", nil)
}

func ErrTooManyAttachments(count, max int) error {
	return result.Validation(
		"messaging.message.validate",
		fmt.Sprintf("too many attachments: %d (max: %d)", count, max),
		map[string]any{
			"count": count,
			"max":   max,
		},
	)
}

// ============================================================================
// Not Found Errors (Kind: KindNotFound)
// ============================================================================

func ErrConversationNotFound(conversationID string) error {
	return result.NotFound("messaging.conversation.find", "conversation not found")
}

func ErrMessageNotFound(messageID string) error {
	return result.NotFound("messaging.message.find", "message not found")
}

func ErrParticipantNotFound(participantID string) error {
	return result.NotFound("messaging.participant.find", "participant not found in conversation")
}

// ============================================================================
// Conflict Errors (Kind: KindConflict)
// ============================================================================

func ErrConversationAlreadyExists(participantIDs []string) error {
	return result.Conflict("messaging.conversation.create", "conversation already exists between these participants")
}

func ErrAlreadyParticipant(userID string) error {
	return result.Conflict("messaging.participant.add", "user is already a participant")
}

func ErrMessageAlreadyDeleted() error {
	return result.Conflict("messaging.message.delete", "message is already deleted")
}

// ============================================================================
// Domain Errors (Kind: KindDomain)
// ============================================================================

func ErrCannotModifySystemMessage() error {
	return result.Domain("messaging.message.modify", "cannot modify system messages")
}

func ErrCannotDeleteSystemMessage() error {
	return result.Domain("messaging.message.delete", "cannot delete system messages")
}

func ErrMessageAlreadyRead() error {
	return result.Domain("messaging.message.read", "message already marked as read")
}

func ErrCannotReplyToDeletedMessage() error {
	return result.Domain("messaging.message.reply", "cannot reply to a deleted message")
}

func ErrConversationMuted() error {
	return result.Domain("messaging.conversation.notify", "conversation is muted")
}

func ErrCannotLeaveDirectConversation() error {
	return result.Domain("messaging.conversation.leave", "cannot leave direct conversation")
}

// ============================================================================
// Unauthorized Errors (Kind: KindUnauthorized)
// ============================================================================

func ErrNotParticipant() error {
	return result.Unauthorized("messaging.conversation.access", "user is not a participant in this conversation")
}

func ErrNotMessageAuthor() error {
	return result.Unauthorized("messaging.message.modify", "user is not the message author")
}

func ErrCannotAddParticipants() error {
	return result.Unauthorized("messaging.conversation.modify", "user does not have permission to add participants")
}

func ErrCannotRemoveParticipants() error {
	return result.Unauthorized("messaging.conversation.modify", "user does not have permission to remove participants")
}
