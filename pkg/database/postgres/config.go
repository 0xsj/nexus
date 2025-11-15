// Package postgres provides PostgreSQL database connection and utilities.
package postgres

import (
	"fmt"
	"time"

	"github.com/0xsj/nexus/pkg/config"
)

// Config holds PostgreSQL database configuration.
type Config struct {
	// Connection parameters
	Host     string `env:"HOST" default:"localhost"`
	Port     int    `env:"PORT" default:"5432"`
	Database string `env:"DATABASE" required:"true"`
	Username string `env:"USERNAME" required:"true"`
	Password string `env:"PASSWORD" required:"true"`

	// Connection pool settings
	MaxConnections    int           `env:"MAX_CONNECTIONS" default:"25"`
	MinConnections    int           `env:"MIN_CONNECTIONS" default:"5"`
	MaxConnLifetime   time.Duration `env:"MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime   time.Duration `env:"MAX_CONN_IDLE_TIME" default:"30m"`
	HealthCheckPeriod time.Duration `env:"HEALTH_CHECK_PERIOD" default:"1m"`
	ConnectTimeout    time.Duration `env:"CONNECT_TIMEOUT" default:"5s"`

	// SSL/TLS settings
	SSLMode     string `env:"SSL_MODE" default:"disable"`
	SSLCert     string `env:"SSL_CERT"`
	SSLKey      string `env:"SSL_KEY"`
	SSLRootCert string `env:"SSL_ROOT_CERT"`

	// Application identification
	ApplicationName string `env:"APPLICATION_NAME" default:"nexus"`

	// Query timeout
	StatementTimeout time.Duration `env:"STATEMENT_TIMEOUT" default:"30s"`
}

// LoadConfig loads PostgreSQL configuration from environment variables.
// Variables are prefixed with the given prefix (e.g., "DATABASE_").
//
// Example:
//
//	cfg, err := postgres.LoadConfig("DATABASE_")
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	return config.LoadWithPrefix(cfg, prefix)
}

// DSN returns the PostgreSQL connection string in keyword/value format.
// This format is compatible with pgx and most PostgreSQL tools.
//
// Example output:
//
//	host=localhost port=5432 dbname=nexus user=postgres password=secret sslmode=disable
func (c *Config) DSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s application_name=%s",
		c.Host,
		c.Port,
		c.Database,
		c.Username,
		c.Password,
		c.SSLMode,
		c.ApplicationName,
	)

	// Add SSL certificate paths if provided
	if c.SSLCert != "" {
		dsn += fmt.Sprintf(" sslcert=%s", c.SSLCert)
	}
	if c.SSLKey != "" {
		dsn += fmt.Sprintf(" sslkey=%s", c.SSLKey)
	}
	if c.SSLRootCert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", c.SSLRootCert)
	}

	return dsn
}

// ConnectionString returns a PostgreSQL connection URL.
// This format is useful for some tools and libraries.
//
// Example output:
//
//	postgres://user:pass@localhost:5432/dbname?sslmode=disable&application_name=nexus
func (c *Config) ConnectionString() string {
	// Build query parameters
	params := fmt.Sprintf("sslmode=%s&application_name=%s", c.SSLMode, c.ApplicationName)

	if c.SSLCert != "" {
		params += fmt.Sprintf("&sslcert=%s", c.SSLCert)
	}
	if c.SSLKey != "" {
		params += fmt.Sprintf("&sslkey=%s", c.SSLKey)
	}
	if c.SSLRootCert != "" {
		params += fmt.Sprintf("&sslrootcert=%s", c.SSLRootCert)
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		params,
	)
}

// Validate checks if the configuration is valid.
// Returns an error if any required field is missing or invalid.
func (c *Config) Validate() error {
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if c.MaxConnections < 1 {
		return fmt.Errorf("max connections must be at least 1")
	}
	if c.MinConnections < 0 {
		return fmt.Errorf("min connections cannot be negative")
	}
	if c.MinConnections > c.MaxConnections {
		return fmt.Errorf("min connections (%d) cannot exceed max connections (%d)",
			c.MinConnections, c.MaxConnections)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}
	return nil
}
