package domain

// MessageType represents the type of message content.
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeFile     MessageType = "file"
	MessageTypeLocation MessageType = "location"
	MessageTypeSystem   MessageType = "system"
)

// IsValid checks if the message type is valid.
func (mt MessageType) IsValid() bool {
	switch mt {
	case MessageTypeText, MessageTypeImage, MessageTypeVideo, MessageTypeAudio,
		MessageTypeFile, MessageTypeLocation, MessageTypeSystem:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (mt MessageType) String() string {
	return string(mt)
}

// IsMedia checks if the message contains media.
func (mt MessageType) IsMedia() bool {
	return mt == MessageTypeImage || mt == MessageTypeVideo || mt == MessageTypeAudio
}

// IsSystem checks if the message is a system message.
func (mt MessageType) IsSystem() bool {
	return mt == MessageTypeSystem
}

// RequiresContent checks if the message type requires text content.
func (mt MessageType) RequiresContent() bool {
	return mt == MessageTypeText || mt == MessageTypeSystem
}

// MessageStatus represents the delivery status of a message.
type MessageStatus string

const (
	MessageStatusSending   MessageStatus = "sending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// IsValid checks if the message status is valid.
func (ms MessageStatus) IsValid() bool {
	switch ms {
	case MessageStatusSending, MessageStatusSent, MessageStatusDelivered,
		MessageStatusRead, MessageStatusFailed:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (ms MessageStatus) String() string {
	return string(ms)
}

// IsFinal checks if the status is final (read or failed).
func (ms MessageStatus) IsFinal() bool {
	return ms == MessageStatusRead || ms == MessageStatusFailed
}
