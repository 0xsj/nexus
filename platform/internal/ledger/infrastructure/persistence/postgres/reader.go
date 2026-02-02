package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/0xsj/nexus/platform/internal/ledger/app/query"
	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ============================================================================
// Compile-time interface check
// ============================================================================

var _ query.Reader = (*Reader)(nil)

// ============================================================================
// Reader Implementation
// ============================================================================

// Reader implements query.Reader using PostgreSQL.
type Reader struct {
	queries *generated.Queries
}

// NewReader creates a new Reader.
func NewReader(queries *generated.Queries) *Reader {
	return &Reader{
		queries: queries,
	}
}

// GetByID retrieves a single entry by ID.
func (r *Reader) GetByID(ctx context.Context, id string) (*query.EntryView, error) {
	const op = "postgres.Reader.GetByID"

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.InvalidEntryID(op, id, "invalid UUID format")
	}

	row, err := r.queries.GetEntryByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.EntryNotFound(op, id)
		}
		return nil, pkgerrors.Infrastructure(op, err)
	}

	view := ToEntryView(row)
	return &view, nil
}

// List retrieves paginated entries with filters.
func (r *Reader) List(ctx context.Context, filter query.ListFilter) (*query.ActivityView, error) {
	const op = "postgres.Reader.List"

	page, pageSize := query.NormalizePagination(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize

	params := generated.ListEntriesParams{
		EventTypes:  emptyIfNil(filter.EventTypes),
		ActorID:     filter.ActorID,
		ActorType:   filter.ActorType,
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		ContextID:   filter.ContextID,
		FromTime:    zeroTimeIfEmpty(filter.FromTime),
		ToTime:      zeroTimeIfEmpty(filter.ToTime),
		PageOffset:  int32(offset),
		PageSize:    int32(pageSize),
	}

	rows, err := r.queries.ListEntries(ctx, params)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	countParams := generated.CountEntriesParams{
		EventTypes:  emptyIfNil(filter.EventTypes),
		ActorID:     filter.ActorID,
		ActorType:   filter.ActorType,
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		ContextID:   filter.ContextID,
		FromTime:    zeroTimeIfEmpty(filter.FromTime),
		ToTime:      zeroTimeIfEmpty(filter.ToTime),
	}

	totalCount, err := r.queries.CountEntries(ctx, countParams)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	return &query.ActivityView{
		Entries:    ToEntryViews(rows),
		TotalCount: int(totalCount),
		Page:       page,
		PageSize:   pageSize,
		HasMore:    int64(offset+pageSize) < totalCount,
	}, nil
}

// GetBySubject retrieves entries for a specific subject.
func (r *Reader) GetBySubject(ctx context.Context, filter query.SubjectFilter) (*query.SubjectHistoryView, error) {
	const op = "postgres.Reader.GetBySubject"

	page, pageSize := query.NormalizePagination(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize

	params := generated.GetEntriesBySubjectParams{
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		EventTypes:  emptyIfNil(filter.EventTypes),
		FromTime:    zeroTimeIfEmpty(filter.FromTime),
		ToTime:      zeroTimeIfEmpty(filter.ToTime),
		PageOffset:  int32(offset),
		PageSize:    int32(pageSize),
	}

	rows, err := r.queries.GetEntriesBySubject(ctx, params)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	countParams := generated.CountEntriesBySubjectParams{
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		EventTypes:  emptyIfNil(filter.EventTypes),
		FromTime:    zeroTimeIfEmpty(filter.FromTime),
		ToTime:      zeroTimeIfEmpty(filter.ToTime),
	}

	totalCount, err := r.queries.CountEntriesBySubject(ctx, countParams)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	return &query.SubjectHistoryView{
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		Entries:     ToEntryViews(rows),
		TotalCount:  int(totalCount),
	}, nil
}

// GetByActor retrieves entries for a specific actor.
func (r *Reader) GetByActor(ctx context.Context, filter query.ActorFilter) (*query.ActivityView, error) {
	const op = "postgres.Reader.GetByActor"

	page, pageSize := query.NormalizePagination(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize

	params := generated.GetEntriesByActorParams{
		ActorID:    filter.ActorID,
		ActorType:  filter.ActorType,
		EventTypes: emptyIfNil(filter.EventTypes),
		FromTime:   zeroTimeIfEmpty(filter.FromTime),
		ToTime:     zeroTimeIfEmpty(filter.ToTime),
		PageOffset: int32(offset),
		PageSize:   int32(pageSize),
	}

	rows, err := r.queries.GetEntriesByActor(ctx, params)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	countParams := generated.CountEntriesByActorParams{
		ActorID:    filter.ActorID,
		ActorType:  filter.ActorType,
		EventTypes: emptyIfNil(filter.EventTypes),
		FromTime:   zeroTimeIfEmpty(filter.FromTime),
		ToTime:     zeroTimeIfEmpty(filter.ToTime),
	}

	totalCount, err := r.queries.CountEntriesByActor(ctx, countParams)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	return &query.ActivityView{
		Entries:    ToEntryViews(rows),
		TotalCount: int(totalCount),
		Page:       page,
		PageSize:   pageSize,
		HasMore:    int64(offset+pageSize) < totalCount,
	}, nil
}

// GetVerificationLog retrieves verification attempts for a user.
func (r *Reader) GetVerificationLog(ctx context.Context, filter query.VerificationLogFilter) (*query.VerificationLogView, error) {
	const op = "postgres.Reader.GetVerificationLog"

	page, pageSize := query.NormalizePagination(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize

	params := generated.GetVerificationLogsParams{
		UserID:     filter.UserID,
		FromTime:   zeroTimeIfEmpty(filter.FromTime),
		ToTime:     zeroTimeIfEmpty(filter.ToTime),
		PageOffset: int32(offset),
		PageSize:   int32(pageSize),
	}

	rows, err := r.queries.GetVerificationLogs(ctx, params)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	countParams := generated.CountVerificationLogsParams{
		UserID:   filter.UserID,
		FromTime: zeroTimeIfEmpty(filter.FromTime),
		ToTime:   zeroTimeIfEmpty(filter.ToTime),
	}

	totalCount, err := r.queries.CountVerificationLogs(ctx, countParams)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	return &query.VerificationLogView{
		UserID:     filter.UserID,
		Entries:    ToVerificationLogEntries(rows),
		TotalCount: int(totalCount),
		Page:       page,
		PageSize:   pageSize,
		HasMore:    int64(offset+pageSize) < totalCount,
	}, nil
}

// GetStats retrieves aggregated statistics.
func (r *Reader) GetStats(ctx context.Context, filter query.StatsFilter) (*query.ActivityStatsView, error) {
	const op = "postgres.Reader.GetStats"

	statsParams := generated.GetEventTypeStatsParams{
		ActorID:     filter.ActorID,
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		FromTime:    zeroTimeIfEmpty(filter.FromTime),
		ToTime:      zeroTimeIfEmpty(filter.ToTime),
	}

	statsRows, err := r.queries.GetEventTypeStats(ctx, statsParams)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	countParams := generated.GetTotalCountParams{
		ActorID:     filter.ActorID,
		SubjectID:   filter.SubjectID,
		SubjectType: filter.SubjectType,
		FromTime:    zeroTimeIfEmpty(filter.FromTime),
		ToTime:      zeroTimeIfEmpty(filter.ToTime),
	}

	totalCount, err := r.queries.GetTotalCount(ctx, countParams)
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	eventTypeCounts := make([]query.EventTypeCount, len(statsRows))
	for i, row := range statsRows {
		eventTypeCounts[i] = query.EventTypeCount{
			EventType: row.EventType,
			Count:     int(row.Count),
		}
	}

	return &query.ActivityStatsView{
		TotalEntries:    int(totalCount),
		EventTypeCounts: eventTypeCounts,
		FromTime:        filter.FromTime,
		ToTime:          filter.ToTime,
	}, nil
}

// ============================================================================
// Helpers
// ============================================================================

// zeroTimeIfEmpty returns Go's zero time if the input is zero.
// This matches our SQL queries which use '0001-01-01' as a sentinel.
func zeroTimeIfEmpty(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return t
}

// emptyIfNil ensures a nil slice becomes an empty slice.
// PostgreSQL CARDINALITY(NULL) returns NULL, not 0, which breaks our
// "empty means match all" logic in queries.
func emptyIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
