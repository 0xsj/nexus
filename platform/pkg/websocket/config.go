package websocket

import "time"

// Config holds configuration for WebSocket connections and the hub.
type Config struct {
	// ReadBufferSize is the buffer size for reading from the connection.
	ReadBufferSize int

	// WriteBufferSize is the buffer size for writing to the connection.
	WriteBufferSize int

	// WriteTimeout is the deadline for writing a message.
	WriteTimeout time.Duration

	// ReadTimeout is the deadline for reading a message.
	ReadTimeout time.Duration

	// PingInterval is how often to send ping frames.
	PingInterval time.Duration

	// PongWait is the deadline for receiving a pong after a ping.
	PongWait time.Duration

	// SendBufferSize is the channel buffer size for outgoing messages per connection.
	SendBufferSize int

	// AllowedOrigins is the list of allowed WebSocket origins. Empty means allow all.
	AllowedOrigins []string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		WriteTimeout:    10 * time.Second,
		ReadTimeout:     60 * time.Second,
		PingInterval:    30 * time.Second,
		PongWait:        60 * time.Second,
		SendBufferSize:  256,
	}
}

// WithReadBufferSize sets the read buffer size.
func (c Config) WithReadBufferSize(size int) Config {
	c.ReadBufferSize = size
	return c
}

// WithWriteBufferSize sets the write buffer size.
func (c Config) WithWriteBufferSize(size int) Config {
	c.WriteBufferSize = size
	return c
}

// WithWriteTimeout sets the write deadline.
func (c Config) WithWriteTimeout(d time.Duration) Config {
	c.WriteTimeout = d
	return c
}

// WithReadTimeout sets the read deadline.
func (c Config) WithReadTimeout(d time.Duration) Config {
	c.ReadTimeout = d
	return c
}

// WithPingInterval sets the ping interval.
func (c Config) WithPingInterval(d time.Duration) Config {
	c.PingInterval = d
	return c
}

// WithPongWait sets the pong wait deadline.
func (c Config) WithPongWait(d time.Duration) Config {
	c.PongWait = d
	return c
}

// WithSendBufferSize sets the per-connection send buffer size.
func (c Config) WithSendBufferSize(size int) Config {
	c.SendBufferSize = size
	return c
}

// WithAllowedOrigins sets the allowed WebSocket origins.
func (c Config) WithAllowedOrigins(origins ...string) Config {
	c.AllowedOrigins = origins
	return c
}
