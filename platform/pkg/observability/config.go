package observability

import (
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// Config holds configuration for all observability components.
type Config struct {
	// Log configures the logger.
	Log LogConfig `env:"LOG"`

	// Metrics configures metrics collection.
	// Metrics MetricsConfig `env:"METRICS"` // Future

	// Trace configures distributed tracing.
	// Trace TraceConfig `env:"TRACE"` // Future

	// Health configures health checks.
	// Health HealthConfig `env:"HEALTH"` // Future
}

// LogConfig holds logger configuration.
type LogConfig struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string `env:"LEVEL" default:"info"`

	// Format is the log format (console, json).
	Format string `env:"FORMAT" default:"console"`

	// Colorize enables colored output (console format only).
	Colorize bool `env:"COLORIZE" default:"true"`

	// IncludeTimestamp includes timestamp in logs.
	IncludeTimestamp bool `env:"TIMESTAMP" default:"true"`

	// TimestampFormat is the timestamp format.
	TimestampFormat string `env:"TIMESTAMP_FORMAT" default:"15:04:05.000"`

	// IncludeCaller includes caller information.
	IncludeCaller bool `env:"CALLER" default:"false"`

	// Component is the default component name.
	Component string `env:"COMPONENT" default:""`
}

// DefaultConfig returns a default observability configuration.
func DefaultConfig() Config {
	return Config{
		Log: LogConfig{
			Level:            "info",
			Format:           "console",
			Colorize:         true,
			IncludeTimestamp: true,
			TimestampFormat:  "15:04:05.000",
			IncludeCaller:    false,
			Component:        "",
		},
	}
}

// DevelopmentConfig returns a configuration suitable for development.
func DevelopmentConfig() Config {
	return Config{
		Log: LogConfig{
			Level:            "debug",
			Format:           "console",
			Colorize:         true,
			IncludeTimestamp: true,
			TimestampFormat:  "15:04:05.000",
			IncludeCaller:    true,
			Component:        "",
		},
	}
}

// ProductionConfig returns a configuration suitable for production.
func ProductionConfig() Config {
	return Config{
		Log: LogConfig{
			Level:            "info",
			Format:           "json",
			Colorize:         false,
			IncludeTimestamp: true,
			TimestampFormat:  "2006-01-02T15:04:05.000Z07:00",
			IncludeCaller:    false,
			Component:        "",
		},
	}
}

// ToLogOptions converts LogConfig to log.Options.
func (c LogConfig) ToLogOptions() log.Options {
	level, err := log.ParseLevel(c.Level)
	if err != nil {
		level = log.LevelInfo
	}

	opts := log.DefaultOptions().
		WithLevel(level).
		WithColorize(c.Colorize).
		WithTimestamp(c.IncludeTimestamp).
		WithTimestampFormat(c.TimestampFormat).
		WithCaller(c.IncludeCaller)

	if c.Component != "" {
		opts = opts.WithFields(log.Component(c.Component))
	}

	return opts
}
