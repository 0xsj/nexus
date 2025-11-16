package domain

// ConversationType represents the type of conversation.
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

// IsValid checks if the conversation type is valid.
func (ct ConversationType) IsValid() bool {
	switch ct {
	case ConversationTypeDirect, ConversationTypeGroup:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (ct ConversationType) String() string {
	return string(ct)
}

// IsDirect checks if the conversation is direct (1-on-1).
func (ct ConversationType) IsDirect() bool {
	return ct == ConversationTypeDirect
}

// IsGroup checks if the conversation is a group.
func (ct ConversationType) IsGroup() bool {
	return ct == ConversationTypeGroup
}

// RequiredParticipants returns the required number of participants.
func (ct ConversationType) RequiredParticipants() int {
	if ct.IsDirect() {
		return 2
	}
	return 2 // Groups also need at least 2 participants
}
