package websocket

import (
	"encoding/json"
	"time"
)

// MessageType identifies the kind of WebSocket message.
type MessageType string

const (
	MessageTypeEvent       MessageType = "event"
	MessageTypeSubscribe   MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"
	MessageTypeError       MessageType = "error"
	MessageTypePing        MessageType = "ping"
	MessageTypePong        MessageType = "pong"
)

// Message is the envelope for all WebSocket communication.
type Message struct {
	// Type identifies the message kind.
	Type MessageType `json:"type"`

	// Topic is the event topic or subscription pattern.
	Topic string `json:"topic,omitempty"`

	// Payload carries the message data.
	Payload json.RawMessage `json:"payload,omitempty"`

	// Timestamp is when the message was created.
	Timestamp time.Time `json:"timestamp"`
}

// NewEventMessage creates a new event message with the given topic and payload.
func NewEventMessage(topic string, payload any) (Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Type:      MessageTypeEvent,
		Topic:     topic,
		Payload:   data,
		Timestamp: time.Now().UTC(),
	}, nil
}

// NewErrorMessage creates a new error message.
func NewErrorMessage(errMsg string) Message {
	payload, _ := json.Marshal(map[string]string{"error": errMsg})
	return Message{
		Type:      MessageTypeError,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}
