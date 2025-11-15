// Package postgres provides PostgreSQL database connection and utilities.
package postgres

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgxpool.Pool with additional functionality.
type DB struct {
	pool   *pgxpool.Pool
	config *Config
	logger logger.Logger
}

// New creates a new PostgreSQL connection pool.
// It connects to the database and verifies the connection with a ping.
//
// Example:
//
//	cfg, _ := postgres.LoadConfig("DATABASE_")
//	db, err := postgres.New(ctx, cfg, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func New(ctx context.Context, cfg *Config, log logger.Logger) (*DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build pgxpool config from our config
	poolConfig, err := buildPoolConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build pool config: %w", err)
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connection established",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("database", cfg.Database),
		logger.Int("max_connections", cfg.MaxConnections),
	)

	return &DB{
		pool:   pool,
		config: cfg,
		logger: log,
	}, nil
}

// Pool returns the underlying pgxpool.Pool.
// Use this to execute queries, transactions, and work with sqlc.
//
// Example:
//
//	queries := sqlc.New(db.Pool())
//	user, err := queries.GetUserByID(ctx, userID)
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool.
// It waits for all connections to be returned to the pool.
func (db *DB) Close() {
	if db.pool != nil {
		db.logger.Info("Closing database connection pool")
		db.pool.Close()
	}
}

// Ping verifies the connection to the database is still alive.
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Stats returns connection pool statistics.
func (db *DB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}

// buildPoolConfig creates a pgxpool.Config from our Config.
func buildPoolConfig(cfg *Config) (*pgxpool.Config, error) {
	// Parse connection string
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Set connection timeout
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	// Set statement timeout
	if cfg.StatementTimeout > 0 {
		poolConfig.ConnConfig.RuntimeParams = map[string]string{
			"statement_timeout": fmt.Sprintf("%d", cfg.StatementTimeout.Milliseconds()),
		}
	}

	return poolConfig, nil
}
