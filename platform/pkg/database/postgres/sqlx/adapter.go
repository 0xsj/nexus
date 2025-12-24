package sqlx

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/0xsj/nexus/platform/pkg/database"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
)

// Adapter wraps sqlx for PostgreSQL queries.
type Adapter struct {
	db       *sqlx.DB
	pgDB     *postgres.DB
	executor executor
}

// executor is the common interface for sqlx.DB and sqlx.Tx.
type executor interface {
	sqlx.ExtContext
	sqlx.ExecerContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Ensure Adapter implements postgres.Adapter.
var _ postgres.Adapter = (*Adapter)(nil)

// New creates a new sqlx adapter from a postgres.DB.
func New(pgDB *postgres.DB) (*Adapter, error) {
	const op = "sqlx.New"

	// Register pgx driver and get sql.DB
	connStr := pgDB.Config().DSN()
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, database.ErrConnectionFailed(op, err)
	}

	// Configure connection pool
	config := pgDB.Config()
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime())
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime())

	// Wrap with sqlx
	db := sqlx.NewDb(sqlDB, "pgx")

	return &Adapter{
		db:       db,
		pgDB:     pgDB,
		executor: db,
	}, nil
}

// NewFromPool creates a new sqlx adapter using the existing pgx pool.
func NewFromPool(pgDB *postgres.DB) (*Adapter, error) {
	const op = "sqlx.NewFromPool"

	// Get sql.DB from pgx pool
	sqlDB := stdlib.OpenDBFromPool(pgDB.Pool())

	// Wrap with sqlx
	db := sqlx.NewDb(sqlDB, "pgx")

	return &Adapter{
		db:       db,
		pgDB:     pgDB,
		executor: db,
	}, nil
}

// Close closes the sqlx connection.
func (a *Adapter) Close() error {
	return a.db.Close()
}

// Executor returns the underlying postgres executor.
func (a *Adapter) Executor() postgres.Executor {
	return a.pgDB
}

// WithTx returns an adapter using the given transaction.
func (a *Adapter) WithTx(tx *postgres.Tx) postgres.Adapter {
	return &txAdapter{
		adapter: a,
		pgTx:    tx,
	}
}

// DB returns the underlying sqlx.DB.
func (a *Adapter) DB() *sqlx.DB {
	return a.db
}

// ============================================================================
// Query Methods
// ============================================================================

// Get executes a query and scans a single row into dest.
func (a *Adapter) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	const op = "sqlx.Adapter.Get"

	if err := a.executor.GetContext(ctx, dest, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return database.ErrNotFound(op, "row")
		}
		return postgres.MapError(op, err)
	}
	return nil
}

// Select executes a query and scans multiple rows into dest.
func (a *Adapter) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	const op = "sqlx.Adapter.Select"

	if err := a.executor.SelectContext(ctx, dest, query, args...); err != nil {
		return postgres.MapError(op, err)
	}
	return nil
}

// Exec executes a query without returning rows.
func (a *Adapter) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	const op = "sqlx.Adapter.Exec"

	result, err := a.executor.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return result, nil
}

// Query executes a query that returns rows.
func (a *Adapter) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	const op = "sqlx.Adapter.Query"

	rows, err := a.executor.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row.
func (a *Adapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return a.db.QueryRowxContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows.
func (a *Adapter) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	const op = "sqlx.Adapter.NamedExec"

	result, err := a.db.NamedExecContext(ctx, query, arg)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return result, nil
}

// NamedQuery executes a named query that returns rows.
func (a *Adapter) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	const op = "sqlx.Adapter.NamedQuery"

	rows, err := a.db.NamedQueryContext(ctx, query, arg)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return rows, nil
}

// ============================================================================
// Transaction Methods
// ============================================================================

// BeginTx starts a new sqlx transaction.
func (a *Adapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	const op = "sqlx.Adapter.BeginTx"

	tx, err := a.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return tx, nil
}

// RunInTx executes a function within a transaction.
func (a *Adapter) RunInTx(ctx context.Context, fn func(ctx context.Context, tx *sqlx.Tx) error) error {
	return a.RunInTxWithOptions(ctx, nil, fn)
}

// RunInTxWithOptions executes a function within a transaction with options.
func (a *Adapter) RunInTxWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx *sqlx.Tx) error) (err error) {
	tx, err := a.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return postgres.MapError("sqlx.RunInTxWithOptions.Commit", err)
	}

	return nil
}

// ============================================================================
// Transaction Adapter
// ============================================================================

// txAdapter wraps a transaction for the Adapter interface.
type txAdapter struct {
	adapter *Adapter
	pgTx    *postgres.Tx
	sqlxTx  *sqlx.Tx
}

// Executor returns the underlying postgres executor.
func (t *txAdapter) Executor() postgres.Executor {
	return t.pgTx
}

// WithTx returns self (already in transaction).
func (t *txAdapter) WithTx(tx *postgres.Tx) postgres.Adapter {
	return &txAdapter{
		adapter: t.adapter,
		pgTx:    tx,
	}
}

// ============================================================================
// Prepared Statements
// ============================================================================

// Stmt wraps a prepared statement.
type Stmt struct {
	stmt *sqlx.Stmt
}

// Prepare creates a prepared statement.
func (a *Adapter) Prepare(ctx context.Context, query string) (*Stmt, error) {
	const op = "sqlx.Adapter.Prepare"

	stmt, err := a.db.PreparexContext(ctx, query)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return &Stmt{stmt: stmt}, nil
}

// Close closes the prepared statement.
func (s *Stmt) Close() error {
	return s.stmt.Close()
}

// Get executes the prepared statement and scans a single row.
func (s *Stmt) Get(ctx context.Context, dest interface{}, args ...interface{}) error {
	const op = "sqlx.Stmt.Get"

	if err := s.stmt.GetContext(ctx, dest, args...); err != nil {
		if err == sql.ErrNoRows {
			return database.ErrNotFound(op, "row")
		}
		return postgres.MapError(op, err)
	}
	return nil
}

// Select executes the prepared statement and scans multiple rows.
func (s *Stmt) Select(ctx context.Context, dest interface{}, args ...interface{}) error {
	const op = "sqlx.Stmt.Select"

	if err := s.stmt.SelectContext(ctx, dest, args...); err != nil {
		return postgres.MapError(op, err)
	}
	return nil
}

// Exec executes the prepared statement without returning rows.
func (s *Stmt) Exec(ctx context.Context, args ...interface{}) (sql.Result, error) {
	const op = "sqlx.Stmt.Exec"

	result, err := s.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, postgres.MapError(op, err)
	}
	return result, nil
}