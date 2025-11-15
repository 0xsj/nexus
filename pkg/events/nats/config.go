// Package nats provides a NATS-based event bus implementation.
package nats

import (
	"time"

	"github.com/0xsj/nexus/pkg/config"
)

// Config holds NATS event bus configuration.
type Config struct {
	// URL is the NATS server URL(s).
	// Examples:
	//   - Single: "nats://localhost:4222"
	//   - Multiple: "nats://server1:4222,nats://server2:4222"
	URL string `env:"URL" default:"nats://localhost:4222"`

	// MaxReconnects is the maximum number of reconnection attempts.
	// -1 means unlimited reconnects.
	MaxReconnects int `env:"MAX_RECONNECTS" default:"10"`

	// ReconnectWait is the time to wait between reconnection attempts.
	ReconnectWait time.Duration `env:"RECONNECT_WAIT" default:"2s"`

	// ConnectionTimeout is the timeout for establishing a connection.
	ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT" default:"5s"`

	// PingInterval is the interval for server pings.
	PingInterval time.Duration `env:"PING_INTERVAL" default:"2m"`

	// MaxPingsOut is the maximum number of PINGs without a response before closing the connection.
	MaxPingsOut int `env:"MAX_PINGS_OUT" default:"2"`

	// StreamName is the JetStream stream name for persistent events.
	// If empty, uses core NATS (non-persistent).
	StreamName string `env:"STREAM_NAME" default:"events"`

	// UseJetStream enables JetStream for persistent event storage.
	UseJetStream bool `env:"USE_JETSTREAM" default:"true"`

	// StreamMaxAge is the maximum age for events in the JetStream stream.
	// Events older than this will be automatically deleted.
	// Should be aligned with your stale event filtering (typically 5-30 minutes).
	StreamMaxAge time.Duration `env:"STREAM_MAX_AGE" default:"30m"`

	// Credentials file path for authentication (optional).
	CredsFile string `env:"CREDS_FILE"`

	// Token for authentication (optional).
	Token string `env:"TOKEN"`

	// Username and password for authentication (optional).
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
}

// LoadConfig loads NATS configuration from environment variables with the given prefix.
//
// Example:
//
//	cfg, err := nats.LoadConfig("NATS_")
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	return config.LoadWithPrefix(cfg, prefix)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrInvalidConfig{Field: "URL", Reason: "cannot be empty"}
	}
	if c.MaxReconnects < -1 {
		return ErrInvalidConfig{Field: "MaxReconnects", Reason: "must be -1 or greater"}
	}
	return nil
}

// ErrInvalidConfig is returned when configuration is invalid.
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid NATS config: " + e.Field + " " + e.Reason
}
