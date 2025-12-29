package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/0xsj/nexus/platform/pkg/database"
)

// Tx wraps a pgx transaction.
type Tx struct {
	tx pgx.Tx
}

// Ensure Tx implements database.Tx.
var _ database.Tx = (*Tx)(nil)

// newTx creates a new transaction wrapper.
func newTx(tx pgx.Tx) *Tx {
	return &Tx{tx: tx}
}

// Commit commits the transaction.
func (t *Tx) Commit(ctx context.Context) error {
	const op = "postgres.Tx.Commit"

	if err := t.tx.Commit(ctx); err != nil {
		return MapError(op, err)
	}
	return nil
}

// Rollback aborts the transaction.
func (t *Tx) Rollback(ctx context.Context) error {
	const op = "postgres.Tx.Rollback"

	if err := t.tx.Rollback(ctx); err != nil {
		// Ignore rollback errors if transaction is already closed
		if err == pgx.ErrTxClosed {
			return nil
		}
		return MapError(op, err)
	}
	return nil
}

// Underlying returns the underlying pgx.Tx.
// Use this when you need direct access to pgx features.
func (t *Tx) Underlying() pgx.Tx {
	return t.tx
}

// ============================================================================
// Transaction Options Conversion
// ============================================================================

// toPgxTxOptions converts database.TxOptions to pgx.TxOptions.
func toPgxTxOptions(opts database.TxOptions) pgx.TxOptions {
	return pgx.TxOptions{
		IsoLevel:       toPgxIsolationLevel(opts.Isolation),
		AccessMode:     toPgxAccessMode(opts.ReadOnly),
		DeferrableMode: pgx.NotDeferrable,
	}
}

// toPgxIsolationLevel converts database.IsolationLevel to pgx.TxIsoLevel.
func toPgxIsolationLevel(level database.IsolationLevel) pgx.TxIsoLevel {
	switch level {
	case database.IsolationReadUncommitted:
		return pgx.ReadUncommitted
	case database.IsolationReadCommitted:
		return pgx.ReadCommitted
	case database.IsolationRepeatableRead:
		return pgx.RepeatableRead
	case database.IsolationSerializable:
		return pgx.Serializable
	default:
		return pgx.ReadCommitted // PostgreSQL default
	}
}

// toPgxAccessMode converts read-only flag to pgx.TxAccessMode.
func toPgxAccessMode(readOnly bool) pgx.TxAccessMode {
	if readOnly {
		return pgx.ReadOnly
	}
	return pgx.ReadWrite
}

// ============================================================================
// Executor Interface
// ============================================================================

// Exec executes a query without returning rows.
func (t *Tx) Exec(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	const op = "postgres.Tx.Exec"

	tag, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, MapError(op, err)
	}

	return &result{tag: tag}, nil
}

// Query executes a query that returns rows.
func (t *Tx) Query(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	const op = "postgres.Tx.Query"

	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, MapError(op, err)
	}

	return &rowsWrapper{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row.
func (t *Tx) QueryRow(ctx context.Context, query string, args ...interface{}) database.Row {
	row := t.tx.QueryRow(ctx, query, args...)
	return &rowWrapper{row: row}
}

// ============================================================================
// Result Wrapper
// ============================================================================

// result wraps pgx command tag.
type result struct {
	tag pgconn.CommandTag
}

// Ensure result implements database.Result.
var _ database.Result = (*result)(nil)

// RowsAffected returns the number of rows affected.
func (r *result) RowsAffected() (int64, error) {
	return r.tag.RowsAffected(), nil
}

// LastInsertId is not supported by PostgreSQL.
// Use RETURNING clause instead.
func (r *result) LastInsertId() (int64, error) {
	return 0, database.ErrQueryFailed("postgres.Result.LastInsertId",
		errLastInsertIdNotSupported)
}

var errLastInsertIdNotSupported = &pgxError{msg: "LastInsertId is not supported; use RETURNING clause"}

// pgxError is a simple error type for pgx-specific errors.
type pgxError struct {
	msg string
}

func (e *pgxError) Error() string {
	return e.msg
}

// ============================================================================
// Rows Wrapper
// ============================================================================

// rowsWrapper wraps pgx.Rows.
type rowsWrapper struct {
	rows pgx.Rows
}

// Ensure rowsWrapper implements database.Rows.
var _ database.Rows = (*rowsWrapper)(nil)

// Close closes the rows.
func (r *rowsWrapper) Close() error {
	r.rows.Close()
	return nil
}

// Next prepares the next row for reading.
func (r *rowsWrapper) Next() bool {
	return r.rows.Next()
}

// Scan copies the columns into the values pointed at by dest.
func (r *rowsWrapper) Scan(dest ...interface{}) error {
	const op = "postgres.Rows.Scan"

	if err := r.rows.Scan(dest...); err != nil {
		return MapError(op, err)
	}
	return nil
}

// Err returns any error encountered during iteration.
func (r *rowsWrapper) Err() error {
	const op = "postgres.Rows.Err"

	if err := r.rows.Err(); err != nil {
		return MapError(op, err)
	}
	return nil
}

// ============================================================================
// Row Wrapper
// ============================================================================

// rowWrapper wraps pgx.Row.
type rowWrapper struct {
	row pgx.Row
}

// Ensure rowWrapper implements database.Row.
var _ database.Row = (*rowWrapper)(nil)

// Scan copies the columns into the values pointed at by dest.
func (r *rowWrapper) Scan(dest ...interface{}) error {
	const op = "postgres.Row.Scan"

	if err := r.row.Scan(dest...); err != nil {
		if err == pgx.ErrNoRows {
			return database.ErrNotFound(op, "row")
		}
		return MapError(op, err)
	}
	return nil
}
