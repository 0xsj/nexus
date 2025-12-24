package postgres

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/database"
)

// Adapter provides a query execution interface.
// Implementations can use sqlx, sqlc, or raw queries.
type Adapter interface {
	// Executor returns the underlying executor (DB or Tx).
	Executor() Executor

	// WithTx returns an adapter using the given transaction.
	WithTx(tx *Tx) Adapter
}

// Executor is the common interface for DB and Tx query execution.
type Executor interface {
	Exec(ctx context.Context, query string, args ...interface{}) (database.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (database.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) database.Row
}

// Ensure DB and Tx implement Executor.
var (
	_ Executor = (*DB)(nil)
	_ Executor = (*Tx)(nil)
)

// ============================================================================
// Base Adapter
// ============================================================================

// BaseAdapter provides common adapter functionality.
type BaseAdapter struct {
	db       *DB
	executor Executor
}

// NewBaseAdapter creates a new base adapter.
func NewBaseAdapter(db *DB) *BaseAdapter {
	return &BaseAdapter{
		db:       db,
		executor: db,
	}
}

// Executor returns the underlying executor.
func (a *BaseAdapter) Executor() Executor {
	return a.executor
}

// WithTx returns an adapter using the given transaction.
func (a *BaseAdapter) WithTx(tx *Tx) *BaseAdapter {
	return &BaseAdapter{
		db:       a.db,
		executor: tx,
	}
}

// DB returns the underlying database connection.
func (a *BaseAdapter) DB() *DB {
	return a.db
}

// InTransaction returns true if the adapter is using a transaction.
func (a *BaseAdapter) InTransaction() bool {
	_, ok := a.executor.(*Tx)
	return ok
}

// ============================================================================
// Query Helpers
// ============================================================================

// Get executes a query and scans a single row into dest.
func (a *BaseAdapter) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	const op = "postgres.BaseAdapter.Get"

	row := a.executor.QueryRow(ctx, query, args...)
	if err := row.Scan(dest); err != nil {
		return err // Already mapped by rowWrapper
	}
	return nil
}

// Select executes a query and returns multiple rows.
// The caller is responsible for iterating and closing rows.
func (a *BaseAdapter) Select(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	return a.executor.Query(ctx, query, args...)
}

// Exec executes a query without returning rows.
func (a *BaseAdapter) Exec(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	return a.executor.Exec(ctx, query, args...)
}

// ============================================================================
// Transaction Helpers
// ============================================================================

// RunInTx executes a function within a transaction.
func (a *BaseAdapter) RunInTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error {
	return a.RunInTxWithOptions(ctx, database.DefaultTxOptions(), fn)
}

// RunInTxWithOptions executes a function within a transaction with options.
func (a *BaseAdapter) RunInTxWithOptions(ctx context.Context, opts database.TxOptions, fn func(ctx context.Context, tx *Tx) error) (err error) {
	const op = "postgres.BaseAdapter.RunInTxWithOptions"

	dbTx, err := a.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	tx := dbTx.(*Tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Scanner Interface
// ============================================================================

// Scanner can scan rows into structs.
type Scanner interface {
	// ScanRow scans a single row into dest.
	ScanRow(row database.Row, dest interface{}) error

	// ScanRows scans multiple rows into dest slice.
	ScanRows(rows database.Rows, dest interface{}) error
}

// ============================================================================
// Query Builder Helpers
// ============================================================================

// Placeholder returns a PostgreSQL placeholder ($1, $2, etc.).
func Placeholder(index int) string {
	return "$" + itoa(index)
}

// Placeholders returns a slice of PostgreSQL placeholders.
func Placeholders(count int) []string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = Placeholder(i + 1)
	}
	return placeholders
}

// PlaceholderList returns a comma-separated placeholder list.
// e.g., "$1, $2, $3"
func PlaceholderList(count int) string {
	if count <= 0 {
		return ""
	}

	result := "$1"
	for i := 2; i <= count; i++ {
		result += ", $" + itoa(i)
	}
	return result
}

// itoa converts an int to string without importing strconv.
func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}

	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
