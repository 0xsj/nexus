package domain

import (
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

const (
	MinParticipants = 2
	MaxParticipants = 256
)

// Conversation represents a chat conversation (aggregate root).
type Conversation struct {
	ID             string
	Type           ConversationType
	Name           *string
	AvatarURL      *string
	ParticipantIDs []string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	MutedUntil     *time.Time
	LastMessageID  *string
	LastMessageAt  *time.Time
}

// NewDirectConversation creates a new direct (1-on-1) conversation.
func NewDirectConversation(user1ID, user2ID string) result.Result[*Conversation] {
	// Validate: must have exactly 2 participants
	if user1ID == "" || user2ID == "" {
		return result.Err[*Conversation](ErrInvalidConversationID())
	}

	if user1ID == user2ID {
		return result.Err[*Conversation](ErrInvalidDirectConversation())
	}

	now := time.Now().UTC()

	conversation := &Conversation{
		ID:             uuid.New().String(),
		Type:           ConversationTypeDirect,
		ParticipantIDs: []string{user1ID, user2ID},
		CreatedBy:      user1ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return result.Ok(conversation)
}

// NewGroupConversation creates a new group conversation.
func NewGroupConversation(name string, creatorID string, participantIDs []string) result.Result[*Conversation] {
	// Validate: group must have a name
	if name == "" {
		return result.Err[*Conversation](ErrGroupNameRequired())
	}

	// Validate: minimum participants
	if len(participantIDs) < MinParticipants {
		return result.Err[*Conversation](ErrTooFewParticipants())
	}

	// Validate: maximum participants
	if len(participantIDs) > MaxParticipants {
		return result.Err[*Conversation](ErrTooManyParticipants(len(participantIDs), MaxParticipants))
	}

	// Validate: no duplicates
	if hasDuplicates(participantIDs) {
		return result.Err[*Conversation](ErrDuplicateParticipants())
	}

	// Ensure creator is in the participant list
	if !contains(participantIDs, creatorID) {
		participantIDs = append(participantIDs, creatorID)
	}

	now := time.Now().UTC()

	conversation := &Conversation{
		ID:             uuid.New().String(),
		Type:           ConversationTypeGroup,
		Name:           &name,
		ParticipantIDs: participantIDs,
		CreatedBy:      creatorID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return result.Ok(conversation)
}

// ============================================================================
// Participant Management
// ============================================================================

// AddParticipant adds a participant to the conversation.
func (c *Conversation) AddParticipant(userID string) error {
	// Only groups can add participants
	if c.Type.IsDirect() {
		return ErrCannotAddParticipants()
	}

	// Check if already a participant
	if c.IsParticipant(userID) {
		return ErrAlreadyParticipant(userID)
	}

	// Check max participants
	if len(c.ParticipantIDs) >= MaxParticipants {
		return ErrTooManyParticipants(len(c.ParticipantIDs)+1, MaxParticipants)
	}

	c.ParticipantIDs = append(c.ParticipantIDs, userID)
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// RemoveParticipant removes a participant from the conversation.
func (c *Conversation) RemoveParticipant(userID string) error {
	// Only groups can remove participants
	if c.Type.IsDirect() {
		return ErrCannotLeaveDirectConversation()
	}

	// Check if participant exists
	if !c.IsParticipant(userID) {
		return ErrParticipantNotFound(userID)
	}

	// Remove participant
	c.ParticipantIDs = removeFromSlice(c.ParticipantIDs, userID)
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// IsParticipant checks if a user is a participant.
func (c *Conversation) IsParticipant(userID string) bool {
	return contains(c.ParticipantIDs, userID)
}

// GetParticipantCount returns the number of participants.
func (c *Conversation) GetParticipantCount() int {
	return len(c.ParticipantIDs)
}

// GetOtherParticipant returns the other participant in a direct conversation.
func (c *Conversation) GetOtherParticipant(userID string) (string, error) {
	if !c.Type.IsDirect() {
		return "", ErrInvalidDirectConversation()
	}

	for _, id := range c.ParticipantIDs {
		if id != userID {
			return id, nil
		}
	}

	return "", ErrParticipantNotFound(userID)
}

// ============================================================================
// Conversation Updates
// ============================================================================

// UpdateName updates the conversation name (group only).
func (c *Conversation) UpdateName(name string) error {
	if c.Type.IsDirect() {
		return ErrInvalidDirectConversation()
	}

	if name == "" {
		return ErrGroupNameRequired()
	}

	c.Name = &name
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateAvatar updates the conversation avatar URL.
func (c *Conversation) UpdateAvatar(avatarURL string) error {
	if c.Type.IsDirect() {
		return ErrInvalidDirectConversation()
	}

	c.AvatarURL = &avatarURL
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateLastMessage updates the last message info.
func (c *Conversation) UpdateLastMessage(messageID string) {
	now := time.Now().UTC()
	c.LastMessageID = &messageID
	c.LastMessageAt = &now
	c.UpdatedAt = now
}

// ============================================================================
// Mute Management
// ============================================================================

// Mute mutes the conversation until a specific time.
func (c *Conversation) Mute(mutedUntil *time.Time) {
	c.MutedUntil = mutedUntil
	c.UpdatedAt = time.Now().UTC()
}

// Unmute unmutes the conversation.
func (c *Conversation) Unmute() {
	c.MutedUntil = nil
	c.UpdatedAt = time.Now().UTC()
}

// IsMuted checks if the conversation is currently muted.
func (c *Conversation) IsMuted() bool {
	if c.MutedUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*c.MutedUntil)
}

// ============================================================================
// Query Methods
// ============================================================================

// IsDirect checks if the conversation is direct (1-on-1).
func (c *Conversation) IsDirect() bool {
	return c.Type.IsDirect()
}

// IsGroup checks if the conversation is a group.
func (c *Conversation) IsGroup() bool {
	return c.Type.IsGroup()
}

// HasLastMessage checks if the conversation has a last message.
func (c *Conversation) HasLastMessage() bool {
	return c.LastMessageID != nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasDuplicates(slice []string) bool {
	seen := make(map[string]bool)
	for _, item := range slice {
		if seen[item] {
			return true
		}
		seen[item] = true
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
