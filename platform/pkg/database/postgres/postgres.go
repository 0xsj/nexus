package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/pkg/database"
)

// DB is a PostgreSQL database connection pool.
type DB struct {
	pool   *pgxpool.Pool
	config Config
}

// Ensure DB implements database.DB.
var _ database.DB = (*DB)(nil)

// New creates a new PostgreSQL database connection.
func New(ctx context.Context, config Config) (*DB, error) {
	const op = "postgres.New"

	if err := config.Validate(); err != nil {
		return nil, database.ErrConnectionFailed(op, err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		return nil, database.ErrConnectionFailed(op, err)
	}

	// Apply pool settings
	poolConfig.MaxConns = int32(config.MaxOpenConns)
	poolConfig.MinConns = int32(config.MaxIdleConns)
	poolConfig.MaxConnLifetime = config.ConnMaxLifetime()
	poolConfig.MaxConnIdleTime = config.ConnMaxIdleTime()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, database.ErrConnectionFailed(op, err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, database.ErrConnectionFailed(op, err)
	}

	return &DB{
		pool:   pool,
		config: config,
	}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

// Ping verifies the database connection is alive.
func (db *DB) Ping(ctx context.Context) error {
	const op = "postgres.DB.Ping"

	if err := db.pool.Ping(ctx); err != nil {
		return MapError(op, err)
	}
	return nil
}

// Stats returns database connection pool statistics.
func (db *DB) Stats() database.Stats {
	stat := db.pool.Stat()

	return database.Stats{
		MaxConnections:    int(stat.MaxConns()),
		OpenConnections:   int(stat.TotalConns()),
		InUse:             int(stat.AcquiredConns()),
		Idle:              int(stat.IdleConns()),
		WaitCount:         stat.EmptyAcquireCount(),
		WaitDurationMs:    stat.AcquireDuration().Milliseconds(),
		MaxIdleClosed:     stat.MaxIdleDestroyCount(),
		MaxLifetimeClosed: stat.MaxLifetimeDestroyCount(),
	}
}

// Begin starts a new transaction.
func (db *DB) Begin(ctx context.Context) (database.Tx, error) {
	return db.BeginTx(ctx, database.DefaultTxOptions())
}

// BeginTx starts a new transaction with options.
func (db *DB) BeginTx(ctx context.Context, opts database.TxOptions) (database.Tx, error) {
	const op = "postgres.DB.BeginTx"

	tx, err := db.pool.BeginTx(ctx, toPgxTxOptions(opts))
	if err != nil {
		return nil, MapError(op, err)
	}

	return newTx(tx), nil
}

// ============================================================================
// Executor Interface
// ============================================================================

// Exec executes a query without returning rows.
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	const op = "postgres.DB.Exec"

	tag, err := db.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, MapError(op, err)
	}

	return &result{tag: tag}, nil
}

// Query executes a query that returns rows.
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	const op = "postgres.DB.Query"

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, MapError(op, err)
	}

	return &rowsWrapper{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row.
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) database.Row {
	row := db.pool.QueryRow(ctx, query, args...)
	return &rowWrapper{row: row}
}

// ============================================================================
// Health Check
// ============================================================================

// HealthCheck performs a health check on the database.
func (db *DB) HealthCheck(ctx context.Context) database.HealthStatus {
	start := time.Now()

	err := db.Ping(ctx)
	latency := time.Since(start)

	status := database.HealthStatus{
		Healthy:   err == nil,
		LatencyMs: latency.Milliseconds(),
		Stats:     db.Stats(),
	}

	if err != nil {
		status.Message = err.Error()
	} else {
		status.Message = "connected"
	}

	return status
}

// ============================================================================
// Pool Access
// ============================================================================

// Pool returns the underlying pgxpool.Pool.
// Use this when you need direct access to pgx features.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Config returns the database configuration.
func (db *DB) Config() Config {
	return db.config
}

// ============================================================================
// Acquire Connection
// ============================================================================

// Acquire acquires a connection from the pool.
// The connection must be released back to the pool.
func (db *DB) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	const op = "postgres.DB.Acquire"

	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, MapError(op, err)
	}

	return conn, nil
}
