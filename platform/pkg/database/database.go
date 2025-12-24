package database

import (
	"context"
	"io"
)

// DB represents a database connection.
type DB interface {
	TxBeginner
	io.Closer

	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error

	// Stats returns database connection pool statistics.
	Stats() Stats
}

// Stats represents database connection pool statistics.
type Stats struct {
	// MaxConnections is the maximum number of connections.
	MaxConnections int

	// OpenConnections is the current number of open connections.
	OpenConnections int

	// InUse is the number of connections currently in use.
	InUse int

	// Idle is the number of idle connections.
	Idle int

	// WaitCount is the total number of connections waited for.
	WaitCount int64

	// WaitDuration is the total time waited for connections.
	WaitDurationMs int64

	// MaxIdleClosed is the total number of connections closed due to max idle.
	MaxIdleClosed int64

	// MaxLifetimeClosed is the total number of connections closed due to max lifetime.
	MaxLifetimeClosed int64
}

// Config holds database configuration.
type Config struct {
	// Driver is the database driver name (postgres, mysql, sqlite).
	Driver string

	// Host is the database host.
	Host string

	// Port is the database port.
	Port int

	// Database is the database name.
	Database string

	// Username is the database user.
	Username string

	// Password is the database password.
	Password string

	// SSLMode is the SSL mode (disable, require, verify-ca, verify-full).
	SSLMode string

	// MaxOpenConns is the maximum number of open connections.
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int

	// ConnMaxLifetimeSec is the maximum connection lifetime in seconds.
	ConnMaxLifetimeSec int

	// ConnMaxIdleTimeSec is the maximum idle time in seconds.
	ConnMaxIdleTimeSec int

	// QueryTimeoutSec is the default query timeout in seconds.
	QueryTimeoutSec int
}

// DefaultConfig returns default database configuration.
func DefaultConfig() Config {
	return Config{
		Driver:             "postgres",
		Host:               "localhost",
		Port:               5432,
		SSLMode:            "disable",
		MaxOpenConns:       25,
		MaxIdleConns:       5,
		ConnMaxLifetimeSec: 300,
		ConnMaxIdleTimeSec: 60,
		QueryTimeoutSec:    30,
	}
}

// WithHost sets the host.
func (c Config) WithHost(host string) Config {
	c.Host = host
	return c
}

// WithPort sets the port.
func (c Config) WithPort(port int) Config {
	c.Port = port
	return c
}

// WithDatabase sets the database name.
func (c Config) WithDatabase(database string) Config {
	c.Database = database
	return c
}

// WithCredentials sets username and password.
func (c Config) WithCredentials(username, password string) Config {
	c.Username = username
	c.Password = password
	return c
}

// WithSSLMode sets the SSL mode.
func (c Config) WithSSLMode(mode string) Config {
	c.SSLMode = mode
	return c
}

// WithMaxOpenConns sets the maximum open connections.
func (c Config) WithMaxOpenConns(n int) Config {
	c.MaxOpenConns = n
	return c
}

// WithMaxIdleConns sets the maximum idle connections.
func (c Config) WithMaxIdleConns(n int) Config {
	c.MaxIdleConns = n
	return c
}

// ============================================================================
// Health Check
// ============================================================================

// HealthChecker can perform health checks.
type HealthChecker interface {
	// HealthCheck performs a health check.
	HealthCheck(ctx context.Context) HealthStatus
}

// HealthStatus represents the health of the database.
type HealthStatus struct {
	// Healthy indicates if the database is healthy.
	Healthy bool

	// Message provides additional details.
	Message string

	// LatencyMs is the ping latency in milliseconds.
	LatencyMs int64

	// Stats contains connection pool stats.
	Stats Stats
}

// ============================================================================
// Migrator
// ============================================================================

// Migrator handles database migrations.
type Migrator interface {
	// Up runs all pending migrations.
	Up(ctx context.Context) error

	// Down rolls back the last migration.
	Down(ctx context.Context) error

	// Version returns the current migration version.
	Version(ctx context.Context) (int, bool, error)

	// Reset rolls back all migrations and re-runs them.
	Reset(ctx context.Context) error
}

// ============================================================================
// Executor
// ============================================================================

// Executor can execute queries.
// Both DB and Tx implement this interface.
type Executor interface {
	// Exec executes a query without returning rows.
	Exec(ctx context.Context, query string, args ...any) (Result, error)

	// Query executes a query that returns rows.
	Query(ctx context.Context, query string, args ...any) (Rows, error)

	// QueryRow executes a query that returns at most one row.
	QueryRow(ctx context.Context, query string, args ...any) Row
}

// Result represents the result of an Exec operation.
type Result interface {
	// RowsAffected returns the number of rows affected.
	RowsAffected() (int64, error)

	// LastInsertId returns the last inserted ID.
	LastInsertId() (int64, error)
}

// Rows represents query result rows.
type Rows interface {
	io.Closer

	// Next prepares the next row for reading.
	Next() bool

	// Scan copies the columns into the values pointed at by dest.
	Scan(dest ...any) error

	// Err returns any error encountered during iteration.
	Err() error
}

// Row represents a single row result.
type Row interface {
	// Scan copies the columns into the values pointed at by dest.
	Scan(dest ...any) error
}
