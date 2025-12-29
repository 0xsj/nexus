package postgres

import (
	"fmt"
	"net/url"
	"time"

	"github.com/0xsj/nexus/platform/pkg/database"
)

// Config holds PostgreSQL-specific configuration.
type Config struct {
	database.Config

	// ApplicationName is the application name for connection identification.
	ApplicationName string

	// ConnectTimeoutSec is the connection timeout in seconds.
	ConnectTimeoutSec int
}

// DefaultConfig returns default PostgreSQL configuration.
func DefaultConfig() Config {
	return Config{
		Config:            database.DefaultConfig(),
		ApplicationName:   "nexus",
		ConnectTimeoutSec: 10,
	}
}

// NewConfig creates a new config from base database config.
func NewConfig(base database.Config) Config {
	return Config{
		Config:            base,
		ApplicationName:   "nexus",
		ConnectTimeoutSec: 10,
	}
}

// WithApplicationName sets the application name.
func (c Config) WithApplicationName(name string) Config {
	c.ApplicationName = name
	return c
}

// WithConnectTimeout sets the connection timeout.
func (c Config) WithConnectTimeout(seconds int) Config {
	c.ConnectTimeoutSec = seconds
	return c
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.MaxOpenConns < 1 {
		return fmt.Errorf("max open connections must be at least 1")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}
	return nil
}

// DSN returns the PostgreSQL connection string.
func (c Config) DSN() string {
	// Build query parameters
	params := url.Values{}
	params.Set("sslmode", c.sslMode())

	if c.ApplicationName != "" {
		params.Set("application_name", c.ApplicationName)
	}
	if c.ConnectTimeoutSec > 0 {
		params.Set("connect_timeout", fmt.Sprintf("%d", c.ConnectTimeoutSec))
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?%s",
		url.QueryEscape(c.Username),
		url.QueryEscape(c.Password),
		c.Host,
		c.Port,
		c.Database,
		params.Encode(),
	)

	return dsn
}

// DSNWithoutPassword returns the DSN with password masked (for logging).
func (c Config) DSNWithoutPassword() string {
	params := url.Values{}
	params.Set("sslmode", c.sslMode())

	if c.ApplicationName != "" {
		params.Set("application_name", c.ApplicationName)
	}

	return fmt.Sprintf(
		"postgres://%s:***@%s:%d/%s?%s",
		url.QueryEscape(c.Username),
		c.Host,
		c.Port,
		c.Database,
		params.Encode(),
	)
}

// sslMode returns the SSL mode, defaulting to "disable".
func (c Config) sslMode() string {
	if c.SSLMode == "" {
		return "disable"
	}
	return c.SSLMode
}

// ConnMaxLifetime returns the max connection lifetime as a Duration.
func (c Config) ConnMaxLifetime() time.Duration {
	if c.ConnMaxLifetimeSec <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(c.ConnMaxLifetimeSec) * time.Second
}

// ConnMaxIdleTime returns the max idle time as a Duration.
func (c Config) ConnMaxIdleTime() time.Duration {
	if c.ConnMaxIdleTimeSec <= 0 {
		return time.Minute
	}
	return time.Duration(c.ConnMaxIdleTimeSec) * time.Second
}

// QueryTimeout returns the query timeout as a Duration.
func (c Config) QueryTimeout() time.Duration {
	if c.QueryTimeoutSec <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.QueryTimeoutSec) * time.Second
}
