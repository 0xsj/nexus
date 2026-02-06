package domain

import (
	"fmt"
)

// ============================================================================
// DigestFrequency
// ============================================================================

// DigestFrequency represents how often digest notifications are sent.
type DigestFrequency string

const (
	// DigestFrequencyDaily sends digests once per day.
	DigestFrequencyDaily DigestFrequency = "daily"

	// DigestFrequencyWeekly sends digests once per week.
	DigestFrequencyWeekly DigestFrequency = "weekly"
)

// ParseDigestFrequency parses a string into a DigestFrequency.
func ParseDigestFrequency(s string) (DigestFrequency, error) {
	switch s {
	case "daily":
		return DigestFrequencyDaily, nil
	case "weekly":
		return DigestFrequencyWeekly, nil
	default:
		return "", fmt.Errorf("invalid digest frequency: %s", s)
	}
}

// IsValid returns true if the DigestFrequency is a known, valid frequency.
func (f DigestFrequency) IsValid() bool {
	switch f {
	case DigestFrequencyDaily, DigestFrequencyWeekly:
		return true
	default:
		return false
	}
}

// String returns the string representation of the DigestFrequency.
func (f DigestFrequency) String() string {
	return string(f)
}

// ============================================================================
// NotificationPreferences
// ============================================================================

// NotificationPreferences is an immutable value object representing a user's
// notification preferences across channels and categories.
type NotificationPreferences struct {
	userID           string
	globalEnabled    bool
	channelEnabled   map[Channel]bool
	categoryChannels map[Category]map[Channel]bool
	digestEnabled    bool
	digestFrequency  DigestFrequency
	timezone         string
}

// NewDefaultPreferences creates a new NotificationPreferences with all channels
// and categories enabled by default.
func NewDefaultPreferences(userID string) NotificationPreferences {
	channelEnabled := map[Channel]bool{
		ChannelEmail: true,
		ChannelPush:  true,
		ChannelInApp: true,
	}

	defaultChannels := map[Channel]bool{
		ChannelEmail: true,
		ChannelPush:  true,
		ChannelInApp: true,
	}

	categoryChannels := map[Category]map[Channel]bool{
		CategorySecurity:     copyChannelMap(defaultChannels),
		CategoryCredentials:  copyChannelMap(defaultChannels),
		CategoryVerification: copyChannelMap(defaultChannels),
		CategorySocial:       copyChannelMap(defaultChannels),
		CategoryOrganization: copyChannelMap(defaultChannels),
		CategoryIssuer:       copyChannelMap(defaultChannels),
		CategorySystem:       copyChannelMap(defaultChannels),
	}

	return NotificationPreferences{
		userID:           userID,
		globalEnabled:    true,
		channelEnabled:   channelEnabled,
		categoryChannels: categoryChannels,
		digestEnabled:    false,
		digestFrequency:  DigestFrequencyDaily,
		timezone:         "UTC",
	}
}

// ShouldNotify checks whether a notification should be delivered for the given
// category and channel, based on the user's preferences.
func (p NotificationPreferences) ShouldNotify(category Category, channel Channel) bool {
	if !p.globalEnabled {
		return false
	}

	if enabled, ok := p.channelEnabled[channel]; ok && !enabled {
		return false
	}

	if catChannels, ok := p.categoryChannels[category]; ok {
		if enabled, ok := catChannels[channel]; ok {
			return enabled
		}
	}

	return true
}

// ============================================================================
// Getters
// ============================================================================

// UserID returns the user ID these preferences belong to.
func (p NotificationPreferences) UserID() string {
	return p.userID
}

// GlobalEnabled returns whether notifications are globally enabled.
func (p NotificationPreferences) GlobalEnabled() bool {
	return p.globalEnabled
}

// ChannelEnabled returns a copy of the channel enabled map.
func (p NotificationPreferences) ChannelEnabled() map[Channel]bool {
	return copyChannelMap(p.channelEnabled)
}

// CategoryChannels returns a deep copy of the category-channel preferences.
func (p NotificationPreferences) CategoryChannels() map[Category]map[Channel]bool {
	result := make(map[Category]map[Channel]bool, len(p.categoryChannels))
	for cat, channels := range p.categoryChannels {
		result[cat] = copyChannelMap(channels)
	}
	return result
}

// DigestEnabled returns whether digest mode is enabled.
func (p NotificationPreferences) DigestEnabled() bool {
	return p.digestEnabled
}

// DigestFrequency returns the digest frequency.
func (p NotificationPreferences) DigestFrequency() DigestFrequency {
	return p.digestFrequency
}

// Timezone returns the user's timezone.
func (p NotificationPreferences) Timezone() string {
	return p.timezone
}

// ============================================================================
// With* Methods (immutable updates)
// ============================================================================

// WithGlobalEnabled returns a new NotificationPreferences with the global enabled flag set.
func (p NotificationPreferences) WithGlobalEnabled(enabled bool) NotificationPreferences {
	result := p.copy()
	result.globalEnabled = enabled
	return result
}

// WithChannelEnabled returns a new NotificationPreferences with the channel enabled flag set.
func (p NotificationPreferences) WithChannelEnabled(channel Channel, enabled bool) NotificationPreferences {
	result := p.copy()
	result.channelEnabled[channel] = enabled
	return result
}

// WithCategoryChannel returns a new NotificationPreferences with the category-channel preference set.
func (p NotificationPreferences) WithCategoryChannel(category Category, channel Channel, enabled bool) NotificationPreferences {
	result := p.copy()
	if _, ok := result.categoryChannels[category]; !ok {
		result.categoryChannels[category] = make(map[Channel]bool)
	}
	result.categoryChannels[category][channel] = enabled
	return result
}

// WithDigestEnabled returns a new NotificationPreferences with the digest enabled flag set.
func (p NotificationPreferences) WithDigestEnabled(enabled bool) NotificationPreferences {
	result := p.copy()
	result.digestEnabled = enabled
	return result
}

// WithDigestFrequency returns a new NotificationPreferences with the digest frequency set.
func (p NotificationPreferences) WithDigestFrequency(frequency DigestFrequency) NotificationPreferences {
	result := p.copy()
	result.digestFrequency = frequency
	return result
}

// WithTimezone returns a new NotificationPreferences with the timezone set.
func (p NotificationPreferences) WithTimezone(timezone string) NotificationPreferences {
	result := p.copy()
	result.timezone = timezone
	return result
}

// ============================================================================
// Internal Helpers
// ============================================================================

// copy creates a deep copy of the NotificationPreferences.
func (p NotificationPreferences) copy() NotificationPreferences {
	channelEnabled := copyChannelMap(p.channelEnabled)

	categoryChannels := make(map[Category]map[Channel]bool, len(p.categoryChannels))
	for cat, channels := range p.categoryChannels {
		categoryChannels[cat] = copyChannelMap(channels)
	}

	return NotificationPreferences{
		userID:           p.userID,
		globalEnabled:    p.globalEnabled,
		channelEnabled:   channelEnabled,
		categoryChannels: categoryChannels,
		digestEnabled:    p.digestEnabled,
		digestFrequency:  p.digestFrequency,
		timezone:         p.timezone,
	}
}

// copyChannelMap creates a shallow copy of a channel bool map.
func copyChannelMap(m map[Channel]bool) map[Channel]bool {
	result := make(map[Channel]bool, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
