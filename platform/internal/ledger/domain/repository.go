package domain

import (
	"context"
)

// ============================================================================
// Writer Interface (Command Side)
// ============================================================================

// Writer defines the interface for appending audit entries.
// Ledger is append-only — entries cannot be updated or deleted.
type Writer interface {
	// Append persists a new audit entry.
	Append(ctx context.Context, entry *AuditEntry) error

	// AppendBatch persists multiple audit entries atomically.
	AppendBatch(ctx context.Context, entries []*AuditEntry) error
}
