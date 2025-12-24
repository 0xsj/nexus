package database

import "context"

// Tx represents a database transaction.
type Tx interface {
	// Commit commits the transaction.
	Commit(ctx context.Context) error

	// Rollback aborts the transaction.
	Rollback(ctx context.Context) error
}

// TxBeginner can begin transactions.
type TxBeginner interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (Tx, error)

	// BeginTx starts a new transaction with options.
	BeginTx(ctx context.Context, opts TxOptions) (Tx, error)
}

// TxOptions configures transaction behavior.
type TxOptions struct {
	// Isolation is the transaction isolation level.
	Isolation IsolationLevel

	// ReadOnly marks the transaction as read-only.
	ReadOnly bool
}

// IsolationLevel represents transaction isolation levels.
type IsolationLevel int

const (
	// IsolationDefault uses the database default.
	IsolationDefault IsolationLevel = iota

	// IsolationReadUncommitted allows dirty reads.
	IsolationReadUncommitted

	// IsolationReadCommitted prevents dirty reads.
	IsolationReadCommitted

	// IsolationRepeatableRead prevents non-repeatable reads.
	IsolationRepeatableRead

	// IsolationSerializable provides full isolation.
	IsolationSerializable
)

// String returns the isolation level name.
func (i IsolationLevel) String() string {
	switch i {
	case IsolationReadUncommitted:
		return "READ UNCOMMITTED"
	case IsolationReadCommitted:
		return "READ COMMITTED"
	case IsolationRepeatableRead:
		return "REPEATABLE READ"
	case IsolationSerializable:
		return "SERIALIZABLE"
	default:
		return "DEFAULT"
	}
}

// DefaultTxOptions returns default transaction options.
func DefaultTxOptions() TxOptions {
	return TxOptions{
		Isolation: IsolationDefault,
		ReadOnly:  false,
	}
}

// ReadOnlyTxOptions returns read-only transaction options.
func ReadOnlyTxOptions() TxOptions {
	return TxOptions{
		Isolation: IsolationDefault,
		ReadOnly:  true,
	}
}

// ============================================================================
// Transaction Helper
// ============================================================================

// TxFunc is a function that runs within a transaction.
type TxFunc func(ctx context.Context, tx Tx) error

// WithTx executes a function within a transaction.
// Commits on success, rolls back on error or panic.
func WithTx(ctx context.Context, beginner TxBeginner, fn TxFunc) error {
	return WithTxOptions(ctx, beginner, DefaultTxOptions(), fn)
}

// WithTxOptions executes a function within a transaction with options.
func WithTxOptions(ctx context.Context, beginner TxBeginner, opts TxOptions, fn TxFunc) (err error) {
	tx, err := beginner.BeginTx(ctx, opts)
	if err != nil {
		return ErrTransactionFailed("WithTxOptions", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback(ctx)
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			// Rollback on error
			_ = tx.Rollback(ctx)
		}
	}()

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return ErrTransactionFailed("WithTxOptions.Commit", err)
	}

	return nil
}

// ============================================================================
// Context-based Transaction
// ============================================================================

// txContextKey is the context key for transactions.
type txContextKey struct{}

// ContextWithTx returns a context with the transaction attached.
func ContextWithTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext retrieves a transaction from context.
// Returns nil if no transaction is present.
func TxFromContext(ctx context.Context) Tx {
	tx, _ := ctx.Value(txContextKey{}).(Tx)
	return tx
}

// HasTx returns true if the context has a transaction.
func HasTx(ctx context.Context) bool {
	return TxFromContext(ctx) != nil
}
