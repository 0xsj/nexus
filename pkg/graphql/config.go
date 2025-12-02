// pkg/graphql/config.go

package graphql

import "time"

// ============================================================================
// Server Configuration
// ============================================================================

// Config holds GraphQL server configuration.
type Config struct {
	// Path is the endpoint path for GraphQL queries.
	Path string `env:"PATH"`

	// PlaygroundPath is the endpoint path for GraphQL playground.
	// Set to empty string to disable.
	PlaygroundPath string `env:"PLAYGROUND_PATH"`

	// PlaygroundEnabled enables the GraphQL playground UI.
	PlaygroundEnabled bool `env:"PLAYGROUND_ENABLED"`

	// IntrospectionEnabled enables GraphQL introspection.
	// Should be disabled in production.
	IntrospectionEnabled bool `env:"INTROSPECTION_ENABLED"`

	// MaxComplexity is the maximum allowed query complexity.
	// Set to 0 to disable complexity limiting.
	MaxComplexity int `env:"MAX_COMPLEXITY"`

	// MaxDepth is the maximum allowed query depth.
	// Set to 0 to disable depth limiting.
	MaxDepth int `env:"MAX_DEPTH"`

	// MaxUploadSize is the maximum file upload size in bytes.
	MaxUploadSize int64 `env:"MAX_UPLOAD_SIZE"`

	// QueryTimeout is the maximum time allowed for query execution.
	QueryTimeout time.Duration `env:"QUERY_TIMEOUT"`

	// WebsocketKeepAliveDuration is the interval for websocket keep-alive pings.
	// Used for subscriptions.
	WebsocketKeepAliveDuration time.Duration `env:"WEBSOCKET_KEEP_ALIVE"`

	// APQEnabled enables Automatic Persisted Queries.
	APQEnabled bool `env:"APQ_ENABLED"`

	// TracingEnabled enables tracing in responses (Apollo Tracing).
	TracingEnabled bool `env:"TRACING_ENABLED"`
}

// DefaultConfig returns default GraphQL configuration.
func DefaultConfig() Config {
	return Config{
		Path:                       "/graphql",
		PlaygroundPath:             "/playground",
		PlaygroundEnabled:          true,
		IntrospectionEnabled:       true,
		MaxComplexity:              300,
		MaxDepth:                   15,
		MaxUploadSize:              10 * 1024 * 1024, // 10MB
		QueryTimeout:               30 * time.Second,
		WebsocketKeepAliveDuration: 15 * time.Second,
		APQEnabled:                 false,
		TracingEnabled:             false,
	}
}

// ProductionConfig returns production-safe GraphQL configuration.
func ProductionConfig() Config {
	return Config{
		Path:                       "/graphql",
		PlaygroundPath:             "",
		PlaygroundEnabled:          false,
		IntrospectionEnabled:       false,
		MaxComplexity:              200,
		MaxDepth:                   10,
		MaxUploadSize:              5 * 1024 * 1024, // 5MB
		QueryTimeout:               15 * time.Second,
		WebsocketKeepAliveDuration: 30 * time.Second,
		APQEnabled:                 true,
		TracingEnabled:             false,
	}
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if c.Path == "" {
		return ErrInvalidConfig("path is required")
	}
	if c.QueryTimeout <= 0 {
		return ErrInvalidConfig("query_timeout must be positive")
	}
	return nil
}

// ============================================================================
// Config Errors
// ============================================================================

// ConfigError represents a configuration error.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return "graphql config: " + e.Message
}

// ErrInvalidConfig creates a new configuration error.
func ErrInvalidConfig(message string) error {
	return &ConfigError{Message: message}
}
