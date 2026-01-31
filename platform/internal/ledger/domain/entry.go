package domain

import (
	"time"
)

// ============================================================================
// AuditEntry Entity
// ============================================================================

// AuditEntry represents an immutable record of a domain event in the audit trail.
// Unlike aggregates, AuditEntry is not event-sourced — it is a projection
// of events from other bounded contexts into a queryable audit log.
type AuditEntry struct {
	id         EntryID
	occurredAt time.Time
	recordedAt time.Time
	eventType  EventType
	actor      Actor
	subject    Subject
	metadata   Metadata
	contextID  string // Correlation ID for related entries
}

// ============================================================================
// Constructor
// ============================================================================

// NewAuditEntry creates a new AuditEntry.
func NewAuditEntry(
	id EntryID,
	occurredAt time.Time,
	eventType EventType,
	actor Actor,
	subject Subject,
	metadata Metadata,
) (*AuditEntry, error) {
	if id.IsZero() {
		return nil, EntryValidation("domain.NewAuditEntry", "id", "cannot be empty")
	}
	if occurredAt.IsZero() {
		return nil, EntryValidation("domain.NewAuditEntry", "occurred_at", "cannot be zero")
	}
	if eventType.IsZero() {
		return nil, EntryValidation("domain.NewAuditEntry", "event_type", "cannot be empty")
	}
	if actor.IsZero() {
		return nil, EntryValidation("domain.NewAuditEntry", "actor", "cannot be empty")
	}
	if subject.IsZero() {
		return nil, EntryValidation("domain.NewAuditEntry", "subject", "cannot be empty")
	}

	return &AuditEntry{
		id:         id,
		occurredAt: occurredAt,
		recordedAt: time.Now().UTC(),
		eventType:  eventType,
		actor:      actor,
		subject:    subject,
		metadata:   metadata,
		contextID:  metadata.CorrelationID(),
	}, nil
}

// Reconstitute creates an AuditEntry from persisted data.
// Used when loading from database — skips validation since data is already validated.
func Reconstitute(
	id EntryID,
	occurredAt time.Time,
	recordedAt time.Time,
	eventType EventType,
	actor Actor,
	subject Subject,
	metadata Metadata,
	contextID string,
) *AuditEntry {
	return &AuditEntry{
		id:         id,
		occurredAt: occurredAt,
		recordedAt: recordedAt,
		eventType:  eventType,
		actor:      actor,
		subject:    subject,
		metadata:   metadata,
		contextID:  contextID,
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the entry's unique identifier.
func (e *AuditEntry) ID() EntryID {
	return e.id
}

// OccurredAt returns when the original event occurred.
func (e *AuditEntry) OccurredAt() time.Time {
	return e.occurredAt
}

// RecordedAt returns when the entry was recorded in the ledger.
func (e *AuditEntry) RecordedAt() time.Time {
	return e.recordedAt
}

// EventType returns the type of event that occurred.
func (e *AuditEntry) EventType() EventType {
	return e.eventType
}

// Actor returns who caused the event.
func (e *AuditEntry) Actor() Actor {
	return e.actor
}

// Subject returns what the event is about.
func (e *AuditEntry) Subject() Subject {
	return e.subject
}

// Metadata returns additional context about the event.
func (e *AuditEntry) Metadata() Metadata {
	return e.metadata
}

// ContextID returns the correlation ID linking related entries.
func (e *AuditEntry) ContextID() string {
	return e.contextID
}

// ============================================================================
// Builder Pattern
// ============================================================================

// AuditEntryBuilder provides a fluent interface for constructing AuditEntry.
type AuditEntryBuilder struct {
	id         EntryID
	occurredAt time.Time
	eventType  EventType
	actor      Actor
	subject    Subject
	metadata   Metadata
}

// NewAuditEntryBuilder creates a new builder.
func NewAuditEntryBuilder() *AuditEntryBuilder {
	return &AuditEntryBuilder{
		metadata: NewMetadata(),
	}
}

// WithID sets the entry ID.
func (b *AuditEntryBuilder) WithID(id EntryID) *AuditEntryBuilder {
	b.id = id
	return b
}

// WithNewID generates and sets a new entry ID.
func (b *AuditEntryBuilder) WithNewID() *AuditEntryBuilder {
	b.id = NewEntryID()
	return b
}

// WithOccurredAt sets when the event occurred.
func (b *AuditEntryBuilder) WithOccurredAt(t time.Time) *AuditEntryBuilder {
	b.occurredAt = t
	return b
}

// WithOccurredNow sets the occurred time to now.
func (b *AuditEntryBuilder) WithOccurredNow() *AuditEntryBuilder {
	b.occurredAt = time.Now().UTC()
	return b
}

// WithEventType sets the event type.
func (b *AuditEntryBuilder) WithEventType(t EventType) *AuditEntryBuilder {
	b.eventType = t
	return b
}

// WithEventTypeString sets the event type from a string.
func (b *AuditEntryBuilder) WithEventTypeString(s string) *AuditEntryBuilder {
	et, err := NewEventType(s)
	if err == nil {
		b.eventType = et
	}
	return b
}

// WithActor sets the actor.
func (b *AuditEntryBuilder) WithActor(a Actor) *AuditEntryBuilder {
	b.actor = a
	return b
}

// WithUserActor sets a user actor.
func (b *AuditEntryBuilder) WithUserActor(userID string) *AuditEntryBuilder {
	actorID, err := NewActorID(userID)
	if err == nil {
		b.actor, _ = NewActor(ActorTypeUser, actorID)
	}
	return b
}

// WithSystemActor sets a system actor.
func (b *AuditEntryBuilder) WithSystemActor(serviceName string) *AuditEntryBuilder {
	b.actor = SystemActor(serviceName)
	return b
}

// WithSubject sets the subject.
func (b *AuditEntryBuilder) WithSubject(s Subject) *AuditEntryBuilder {
	b.subject = s
	return b
}

// WithSubjectUser sets a user subject.
func (b *AuditEntryBuilder) WithSubjectUser(userID string) *AuditEntryBuilder {
	subjectID, err := NewSubjectID(userID)
	if err == nil {
		b.subject, _ = NewSubject(SubjectTypeUser, subjectID)
	}
	return b
}

// WithSubjectCredential sets a credential subject.
func (b *AuditEntryBuilder) WithSubjectCredential(credentialID string) *AuditEntryBuilder {
	subjectID, err := NewSubjectID(credentialID)
	if err == nil {
		b.subject, _ = NewSubject(SubjectTypeCredential, subjectID)
	}
	return b
}

// WithMetadata sets the metadata.
func (b *AuditEntryBuilder) WithMetadata(m Metadata) *AuditEntryBuilder {
	b.metadata = m
	return b
}

// WithMeta adds a single metadata key-value pair.
func (b *AuditEntryBuilder) WithMeta(key string, value any) *AuditEntryBuilder {
	b.metadata = b.metadata.Set(key, value)
	return b
}

// WithCorrelationID sets the correlation ID in metadata.
func (b *AuditEntryBuilder) WithCorrelationID(id string) *AuditEntryBuilder {
	b.metadata = b.metadata.WithCorrelationID(id)
	return b
}

// Build creates the AuditEntry.
func (b *AuditEntryBuilder) Build() (*AuditEntry, error) {
	return NewAuditEntry(
		b.id,
		b.occurredAt,
		b.eventType,
		b.actor,
		b.subject,
		b.metadata,
	)
}

// MustBuild creates the AuditEntry and panics if invalid.
// Only use for tests.
func (b *AuditEntryBuilder) MustBuild() *AuditEntry {
	entry, err := b.Build()
	if err != nil {
		panic(err)
	}
	return entry
}
