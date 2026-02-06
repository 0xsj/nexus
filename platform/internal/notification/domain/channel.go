package domain

import (
	"fmt"
)

// Channel represents a notification delivery mechanism.
type Channel int

const (
	// ChannelEmail represents email delivery.
	ChannelEmail Channel = 1

	// ChannelPush represents push notification delivery.
	ChannelPush Channel = 2

	// ChannelInApp represents in-app notification delivery.
	ChannelInApp Channel = 3
)

// String returns the string representation of the channel.
func (c Channel) String() string {
	switch c {
	case ChannelEmail:
		return "email"
	case ChannelPush:
		return "push"
	case ChannelInApp:
		return "in_app"
	default:
		return "unknown"
	}
}

// ParseChannel parses a string into a Channel.
func ParseChannel(s string) (Channel, error) {
	switch s {
	case "email":
		return ChannelEmail, nil
	case "push":
		return ChannelPush, nil
	case "in_app":
		return ChannelInApp, nil
	default:
		return 0, fmt.Errorf("invalid channel: %s", s)
	}
}

// IsValid returns true if the Channel is a known, valid channel.
func (c Channel) IsValid() bool {
	switch c {
	case ChannelEmail, ChannelPush, ChannelInApp:
		return true
	default:
		return false
	}
}
