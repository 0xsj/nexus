package config

import (
	"fmt"
	"time"
)

// Config is a base configuration that can be embedded in application configs.
type Config struct {
	// App is the application name.
	App string `env:"APP_NAME" default:"nexus"`

	// Environment is the runtime environment.
	Environment Environment `env:"APP_ENV" default:"development"`

	// Debug enables debug mode.
	Debug bool `env:"DEBUG" default:"false"`

	// LogLevel is the logging level.
	LogLevel string `env:"LOG_LEVEL" default:"info"`
}

// Validate validates the base configuration.
func (c *Config) Validate() error {
	if c.App == "" {
		return fmt.Errorf("app name is required")
	}
	return nil
}

// IsDevelopment returns true if running in development.
func (c *Config) IsDevelopment() bool {
	return c.Environment.IsDevelopment()
}

// IsProduction returns true if running in production.
func (c *Config) IsProduction() bool {
	return c.Environment.IsProduction()
}

// IsTest returns true if running in test.
func (c *Config) IsTest() bool {
	return c.Environment.IsTest()
}

// ============================================================================
// Common Configuration Types
// ============================================================================

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host            string        `env:"HOST" default:"0.0.0.0"`
	Port            int           `env:"PORT" default:"8080"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" default:"30s"`
	MaxHeaderBytes  int           `env:"MAX_HEADER_BYTES" default:"1048576"`
}

// Address returns the server address as host:port.
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate validates the server configuration.
func (c *ServerConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.ReadTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}
	if c.WriteTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}
	return nil
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host            string        `env:"DB_HOST" default:"localhost"`
	Port            int           `env:"DB_PORT" default:"5432"`
	Name            string        `env:"DB_NAME" default:"nexus"`
	User            string        `env:"DB_USER" default:"postgres"`
	Password        string        `env:"DB_PASSWORD"`
	SSLMode         string        `env:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" default:"5m"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" default:"1m"`
	QueryTimeout    time.Duration `env:"DB_QUERY_TIMEOUT" default:"30s"`
	MigrationPath   string        `env:"DB_MIGRATION_PATH" default:"migrations"`
	AutoMigrate     bool          `env:"DB_AUTO_MIGRATE" default:"false"`
}

// DSN returns the database connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// Validate validates the database configuration.
func (c *DatabaseConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Port)
	}
	if c.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.MaxOpenConns < 1 {
		return fmt.Errorf("max open connections must be at least 1")
	}
	return nil
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecret         string        `env:"JWT_SECRET"`
	JWTExpiration     time.Duration `env:"JWT_EXPIRATION" default:"24h"`
	RefreshExpiration time.Duration `env:"REFRESH_EXPIRATION" default:"168h"`
	Issuer            string        `env:"JWT_ISSUER" default:"nexus"`
	Audience          string        `env:"JWT_AUDIENCE" default:"nexus"`
	BCryptCost        int           `env:"BCRYPT_COST" default:"10"`
	SessionTimeout    time.Duration `env:"SESSION_TIMEOUT" default:"24h"`
	MaxLoginAttempts  int           `env:"MAX_LOGIN_ATTEMPTS" default:"5"`
	LockoutDuration   time.Duration `env:"LOCKOUT_DURATION" default:"15m"`
}

// Validate validates the auth configuration.
func (c *AuthConfig) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}
	if c.JWTExpiration < time.Minute {
		return fmt.Errorf("JWT expiration must be at least 1 minute")
	}
	if c.BCryptCost < 4 || c.BCryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 4 and 31")
	}
	return nil
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	Enabled          bool     `env:"CORS_ENABLED" default:"true"`
	AllowedOrigins   []string `env:"CORS_ALLOWED_ORIGINS" default:"*"`
	AllowedMethods   []string `env:"CORS_ALLOWED_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowedHeaders   []string `env:"CORS_ALLOWED_HEADERS" default:"Content-Type,Authorization"`
	ExposedHeaders   []string `env:"CORS_EXPOSED_HEADERS"`
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS" default:"false"`
	MaxAge           int      `env:"CORS_MAX_AGE" default:"86400"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled       bool          `env:"RATE_LIMIT_ENABLED" default:"true"`
	RequestsPerIP int           `env:"RATE_LIMIT_REQUESTS_PER_IP" default:"100"`
	Window        time.Duration `env:"RATE_LIMIT_WINDOW" default:"1m"`
	BurstSize     int           `env:"RATE_LIMIT_BURST_SIZE" default:"20"`
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host         string        `env:"REDIS_HOST" default:"localhost"`
	Port         int           `env:"REDIS_PORT" default:"6379"`
	Password     string        `env:"REDIS_PASSWORD"`
	DB           int           `env:"REDIS_DB" default:"0"`
	MaxRetries   int           `env:"REDIS_MAX_RETRIES" default:"3"`
	PoolSize     int           `env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" default:"5"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" default:"3s"`
}

// Address returns the Redis address as host:port.
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ============================================================================
// Application Configuration Example
// ============================================================================

// AppConfig is an example of a complete application configuration.
type AppConfig struct {
	Config

	Server    ServerConfig    `env:"SERVER"`
	Database  DatabaseConfig  `env:"DB"`
	Auth      AuthConfig      `env:"AUTH"`
	CORS      CORSConfig      `env:"CORS"`
	RateLimit RateLimitConfig `env:"RATE_LIMIT"`
	Redis     RedisConfig     `env:"REDIS"`
}

// Validate validates the entire application configuration.
func (c *AppConfig) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if err := c.Database.Validate(); err != nil {
		return err
	}
	// Auth validation is optional in development
	if c.Environment.IsProduction() {
		if err := c.Auth.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// MustLoad loads configuration and panics on error.
func MustLoad[T any]() *T {
	cfg, err := Load[T]()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// MustLoadWithOptions loads configuration with options and panics on error.
func MustLoadWithOptions[T any](opts Options) *T {
	cfg, err := LoadWithOptions[T](opts)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// LoadFromEnvFile loads configuration from a .env file.
func LoadFromEnvFile[T any](path string) (*T, error) {
	opts := DefaultOptions().WithEnvFile(path)
	return LoadWithOptions[T](opts)
}

// LoadWithPrefix loads configuration with an environment variable prefix.
func LoadWithPrefix[T any](prefix string) (*T, error) {
	opts := DefaultOptions().WithEnvPrefix(prefix)
	return LoadWithOptions[T](opts)
}
