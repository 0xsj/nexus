package eventbus

import (
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// Config holds base configuration for any EventBus implementation.
type Config struct {
	// Logger is the logger used by the event bus.
	Logger log.Logger

	// Middlewares are applied to all handlers in order.
	Middlewares []Middleware

	// DefaultBufferSize is the default channel buffer for subscriptions.
	DefaultBufferSize int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Logger:            log.New(),
		DefaultBufferSize: 64,
	}
}

// WithLogger sets the logger.
func (c Config) WithLogger(logger log.Logger) Config {
	c.Logger = logger
	return c
}

// WithMiddleware appends middleware to the chain.
func (c Config) WithMiddleware(mw ...Middleware) Config {
	c.Middlewares = append(c.Middlewares, mw...)
	return c
}

// WithDefaultBufferSize sets the default buffer size for subscriptions.
func (c Config) WithDefaultBufferSize(size int) Config {
	c.DefaultBufferSize = size
	return c
}
