// Package postgres provides PostgreSQL database connection and utilities.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Transact executes the given function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// Example:
//
//	err := db.Transact(ctx, func(tx pgx.Tx) error {
//	    queries := sqlc.New(tx)
//	    err := queries.CreateUser(ctx, params)
//	    if err != nil {
//	        return err // Rolls back
//	    }
//	    return nil // Commits
//	})
func (db *DB) Transact(ctx context.Context, fn func(pgx.Tx) error) error {
	return db.TransactTx(ctx, pgx.TxOptions{}, fn)
}

// TransactTx is like Transact but allows specifying transaction options.
//
// Example:
//
//	opts := pgx.TxOptions{
//	    IsoLevel:   pgx.ReadCommitted,
//	    AccessMode: pgx.ReadOnly,
//	}
//	err := db.TransactTx(ctx, opts, func(tx pgx.Tx) error {
//	    // Read-only transaction
//	    return nil
//	})
func (db *DB) TransactTx(ctx context.Context, opts pgx.TxOptions, fn func(pgx.Tx) error) error {
	// Begin transaction
	tx, err := db.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if p := recover(); p != nil {
			// Panic occurred, rollback and re-panic
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Execute function
	err = fn(tx)
	if err != nil {
		// Function returned error, rollback
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
