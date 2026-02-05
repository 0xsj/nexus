package nats

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventbus"
)

// Config holds NATS-specific configuration on top of the base eventbus.Config.
type Config struct {
	eventbus.Config

	// URL is the NATS server URL.
	URL string

	// StreamName is the JetStream stream name for events.
	StreamName string

	// StreamSubjects are the subjects the stream captures.
	// Defaults to ["events.>"] if empty.
	StreamSubjects []string

	// CredentialsFile is the path to NATS credentials (optional).
	CredentialsFile string

	// TLSCertFile is the path to a TLS client certificate (optional).
	TLSCertFile string

	// TLSKeyFile is the path to a TLS client key (optional).
	TLSKeyFile string

	// MaxReconnects is the maximum number of reconnection attempts.
	MaxReconnects int

	// ReconnectWait is the delay between reconnection attempts.
	ReconnectWait time.Duration

	// AckWait is the default acknowledgment timeout for JetStream consumers.
	AckWait time.Duration

	// MaxDeliver is the maximum delivery attempts before moving to dead letter.
	MaxDeliver int
}

// DefaultNATSConfig returns a Config with sensible defaults matching the Docker Compose setup.
func DefaultNATSConfig() Config {
	return Config{
		Config:         eventbus.DefaultConfig(),
		URL:            "nats://localhost:4222",
		StreamName:     "NEXUS_EVENTS",
		StreamSubjects: []string{"events.>"},
		MaxReconnects:  60,
		ReconnectWait:  2 * time.Second,
		AckWait:        30 * time.Second,
		MaxDeliver:     5,
	}
}

// WithURL sets the NATS server URL.
func (c Config) WithURL(url string) Config {
	c.URL = url
	return c
}

// WithStreamName sets the JetStream stream name.
func (c Config) WithStreamName(name string) Config {
	c.StreamName = name
	return c
}

// WithStreamSubjects sets the subjects the stream captures.
func (c Config) WithStreamSubjects(subjects ...string) Config {
	c.StreamSubjects = subjects
	return c
}

// WithCredentialsFile sets the NATS credentials file path.
func (c Config) WithCredentialsFile(path string) Config {
	c.CredentialsFile = path
	return c
}

// WithTLS sets TLS certificate and key paths.
func (c Config) WithTLS(certFile, keyFile string) Config {
	c.TLSCertFile = certFile
	c.TLSKeyFile = keyFile
	return c
}

// WithMaxReconnects sets the maximum reconnection attempts.
func (c Config) WithMaxReconnects(n int) Config {
	c.MaxReconnects = n
	return c
}

// WithReconnectWait sets the delay between reconnection attempts.
func (c Config) WithReconnectWait(d time.Duration) Config {
	c.ReconnectWait = d
	return c
}

// WithAckWait sets the default acknowledgment timeout.
func (c Config) WithAckWait(d time.Duration) Config {
	c.AckWait = d
	return c
}

// WithMaxDeliver sets the maximum delivery attempts.
func (c Config) WithMaxDeliver(n int) Config {
	c.MaxDeliver = n
	return c
}
