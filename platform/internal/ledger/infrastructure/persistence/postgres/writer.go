package postgres

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Compile-time interface check
// ============================================================================

var _ domain.Writer = (*Writer)(nil)

// ============================================================================
// Writer Implementation
// ============================================================================

// Writer implements domain.Writer using PostgreSQL.
type Writer struct {
	queries *generated.Queries
}

// NewWriter creates a new Writer.
func NewWriter(queries *generated.Queries) *Writer {
	return &Writer{
		queries: queries,
	}
}

// Append persists a new audit entry.
func (w *Writer) Append(ctx context.Context, entry *domain.AuditEntry) error {
	const op = "postgres.Writer.Append"

	params, err := ToInsertParams(entry)
	if err != nil {
		return pkgerrors.Wrap(err, op)
	}

	if err := w.queries.InsertEntry(ctx, params); err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	return nil
}

// AppendBatch persists multiple audit entries atomically.
func (w *Writer) AppendBatch(ctx context.Context, entries []*domain.AuditEntry) error {
	const op = "postgres.Writer.AppendBatch"

	if len(entries) == 0 {
		return nil
	}

	params := make([]generated.InsertEntryBatchParams, 0, len(entries))
	for _, entry := range entries {
		p, err := ToBatchParams(entry)
		if err != nil {
			return pkgerrors.Wrap(err, op)
		}
		params = append(params, p)
	}

	count, err := w.queries.InsertEntryBatch(ctx, params)
	if err != nil {
		return pkgerrors.Infrastructure(op, err)
	}

	if count != int64(len(entries)) {
		return pkgerrors.Internalf(op, "expected to insert %d entries, inserted %d", len(entries), count)
	}

	return nil
}
