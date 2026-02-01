package postgres

import (
	"encoding/json"

	"github.com/0xsj/nexus/platform/internal/ledger/app/query"
	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Domain to Database Mapping
// ============================================================================

// ToInsertParams converts a domain AuditEntry to database insert params.
func ToInsertParams(entry *domain.AuditEntry) (generated.InsertEntryParams, error) {
	metadata, err := entry.Metadata().ToJSON()
	if err != nil {
		return generated.InsertEntryParams{}, err
	}

	var contextID *string
	if cid := entry.ContextID(); cid != "" {
		contextID = &cid
	}

	return generated.InsertEntryParams{
		ID:          entry.ID().Value().UUID(),
		OccurredAt:  entry.OccurredAt(),
		RecordedAt:  entry.RecordedAt(),
		EventType:   entry.EventType().String(),
		ActorID:     entry.Actor().ID().String(),
		ActorType:   entry.Actor().Type().String(),
		SubjectID:   entry.Subject().ID().String(),
		SubjectType: entry.Subject().Type().String(),
		Metadata:    metadata,
		ContextID:   contextID,
	}, nil
}

// ToBatchParams converts a domain AuditEntry to batch insert params.
func ToBatchParams(entry *domain.AuditEntry) (generated.InsertEntryBatchParams, error) {
	metadata, err := entry.Metadata().ToJSON()
	if err != nil {
		return generated.InsertEntryBatchParams{}, err
	}

	var contextID *string
	if cid := entry.ContextID(); cid != "" {
		contextID = &cid
	}

	return generated.InsertEntryBatchParams{
		ID:          entry.ID().Value().UUID(),
		OccurredAt:  entry.OccurredAt(),
		RecordedAt:  entry.RecordedAt(),
		EventType:   entry.EventType().String(),
		ActorID:     entry.Actor().ID().String(),
		ActorType:   entry.Actor().Type().String(),
		SubjectID:   entry.Subject().ID().String(),
		SubjectType: entry.Subject().Type().String(),
		Metadata:    metadata,
		ContextID:   contextID,
	}, nil
}

// ============================================================================
// Database to Domain Mapping
// ============================================================================

// ToDomainEntry converts a database LedgerEntry to domain AuditEntry.
func ToDomainEntry(row generated.LedgerEntry) (*domain.AuditEntry, error) {
	entryID := domain.FromTypesID(mustParseTypesID(row.ID.String()))

	eventType, err := domain.NewEventType(row.EventType)
	if err != nil {
		return nil, err
	}

	actorType, err := domain.ParseActorType(row.ActorType)
	if err != nil {
		return nil, err
	}

	actorID, err := domain.NewActorID(row.ActorID)
	if err != nil {
		return nil, err
	}

	actor, err := domain.NewActor(actorType, actorID)
	if err != nil {
		return nil, err
	}

	subjectType, err := domain.ParseSubjectType(row.SubjectType)
	if err != nil {
		return nil, err
	}

	subjectID, err := domain.NewSubjectID(row.SubjectID)
	if err != nil {
		return nil, err
	}

	subject, err := domain.NewSubject(subjectType, subjectID)
	if err != nil {
		return nil, err
	}

	metadata, err := domain.MetadataFromJSON(row.Metadata)
	if err != nil {
		return nil, err
	}

	contextID := ""
	if row.ContextID != nil {
		contextID = *row.ContextID
	}

	return domain.Reconstitute(
		entryID,
		row.OccurredAt,
		row.RecordedAt,
		eventType,
		actor,
		subject,
		metadata,
		contextID,
	), nil
}

// ============================================================================
// Database to View Mapping
// ============================================================================

// ToEntryView converts a database LedgerEntry to query EntryView.
func ToEntryView(row generated.LedgerEntry) query.EntryView {
	var metadata map[string]any
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &metadata)
	}

	contextID := ""
	if row.ContextID != nil {
		contextID = *row.ContextID
	}

	return query.EntryView{
		ID:          row.ID.String(),
		OccurredAt:  row.OccurredAt,
		RecordedAt:  row.RecordedAt,
		EventType:   row.EventType,
		ActorID:     row.ActorID,
		ActorType:   row.ActorType,
		SubjectID:   row.SubjectID,
		SubjectType: row.SubjectType,
		Metadata:    metadata,
		ContextID:   contextID,
	}
}

// ToEntryViews converts multiple database rows to query EntryViews.
func ToEntryViews(rows []generated.LedgerEntry) []query.EntryView {
	views := make([]query.EntryView, len(rows))
	for i, row := range rows {
		views[i] = ToEntryView(row)
	}
	return views
}

// ToVerificationLogEntry converts a database LedgerEntry to VerificationLogEntry.
func ToVerificationLogEntry(row generated.LedgerEntry) query.VerificationLogEntry {
	var metadata map[string]any
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &metadata)
	}

	outcome := ""
	if metadata != nil {
		if o, ok := metadata["outcome"].(string); ok {
			outcome = o
		}
	}

	return query.VerificationLogEntry{
		ID:           row.ID.String(),
		OccurredAt:   row.OccurredAt,
		VerifierID:   row.ActorID,
		VerifierType: row.ActorType,
		CredentialID: row.SubjectID,
		Outcome:      outcome,
		Metadata:     metadata,
	}
}

// ToVerificationLogEntries converts multiple database rows to VerificationLogEntries.
func ToVerificationLogEntries(rows []generated.LedgerEntry) []query.VerificationLogEntry {
	entries := make([]query.VerificationLogEntry, len(rows))
	for i, row := range rows {
		entries[i] = ToVerificationLogEntry(row)
	}
	return entries
}

// ============================================================================
// Helpers
// ============================================================================

func mustParseTypesID(s string) types.ID {
	id, err := types.ParseID(s)
	if err != nil {
		panic(err)
	}
	return id
}
