package config

import (
	"os"
	"strings"
)

// Source represents a configuration source.
type Source string

const (
	// SourceEnv loads configuration from environment variables.
	SourceEnv Source = "env"

	// SourceFile loads configuration from a file.
	SourceFile Source = "file"

	// SourceDefault uses default values only.
	SourceDefault Source = "default"
)

// Options configures how configuration is loaded.
type Options struct {
	// Sources defines the configuration sources in priority order.
	// Later sources override earlier ones.
	Sources []Source

	// EnvPrefix is the prefix for environment variables.
	// e.g., "APP" means APP_SERVER_PORT maps to Server.Port
	EnvPrefix string

	// EnvFile is the path to a .env file to load.
	// Only used if SourceEnv is in Sources.
	EnvFile string

	// ConfigFile is the path to a config file (JSON, YAML, TOML).
	// Only used if SourceFile is in Sources.
	ConfigFile string

	// Required fields that must be set.
	Required []string

	// AllowEmpty permits empty string values for fields.
	AllowEmpty bool

	// ExpandEnv expands ${VAR} or $VAR in values.
	ExpandEnv bool

	// Tag is the struct tag to use for field mapping.
	// Default is "env".
	Tag string
}

// DefaultOptions returns the default loader options.
func DefaultOptions() Options {
	return Options{
		Sources:    []Source{SourceDefault, SourceEnv},
		EnvPrefix:  "",
		EnvFile:    "",
		ConfigFile: "",
		Required:   nil,
		AllowEmpty: false,
		ExpandEnv:  true,
		Tag:        "env",
	}
}

// WithSources sets the configuration sources.
func (o Options) WithSources(sources ...Source) Options {
	o.Sources = sources
	return o
}

// WithEnvPrefix sets the environment variable prefix.
func (o Options) WithEnvPrefix(prefix string) Options {
	o.EnvPrefix = prefix
	return o
}

// WithEnvFile sets the .env file path.
func (o Options) WithEnvFile(path string) Options {
	o.EnvFile = path
	o.Sources = appendSource(o.Sources, SourceEnv)
	return o
}

// WithConfigFile sets the config file path.
func (o Options) WithConfigFile(path string) Options {
	o.ConfigFile = path
	o.Sources = appendSource(o.Sources, SourceFile)
	return o
}

// WithRequired sets required field names.
func (o Options) WithRequired(fields ...string) Options {
	o.Required = fields
	return o
}

// WithAllowEmpty permits empty string values.
func (o Options) WithAllowEmpty(allow bool) Options {
	o.AllowEmpty = allow
	return o
}

// WithExpandEnv enables/disables environment variable expansion.
func (o Options) WithExpandEnv(expand bool) Options {
	o.ExpandEnv = expand
	return o
}

// WithTag sets the struct tag name.
func (o Options) WithTag(tag string) Options {
	o.Tag = tag
	return o
}

// appendSource adds a source if not already present.
func appendSource(sources []Source, source Source) []Source {
	for _, s := range sources {
		if s == source {
			return sources
		}
	}
	return append(sources, source)
}

// ============================================================================
// Environment Helpers
// ============================================================================

// Environment represents the application environment.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
	EnvTest        Environment = "test"
)

// String returns the string representation.
func (e Environment) String() string {
	return string(e)
}

// IsDevelopment returns true if this is a development environment.
func (e Environment) IsDevelopment() bool {
	return e == EnvDevelopment
}

// IsProduction returns true if this is a production environment.
func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

// IsTest returns true if this is a test environment.
func (e Environment) IsTest() bool {
	return e == EnvTest
}

// GetEnvironment returns the current environment from ENV or APP_ENV.
func GetEnvironment() Environment {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = os.Getenv("GO_ENV")
	}

	switch strings.ToLower(env) {
	case "prod", "production":
		return EnvProduction
	case "staging", "stage":
		return EnvStaging
	case "test", "testing":
		return EnvTest
	default:
		return EnvDevelopment
	}
}

// IsDevelopment returns true if running in development.
func IsDevelopment() bool {
	return GetEnvironment().IsDevelopment()
}

// IsProduction returns true if running in production.
func IsProduction() bool {
	return GetEnvironment().IsProduction()
}

// IsTest returns true if running in test.
func IsTest() bool {
	return GetEnvironment().IsTest()
}
