package projection

import (
	"time"

	"github.com/0xsj/nexus/platform/internal/ledger/domain"
)

// ============================================================================
// Event Mapper
// ============================================================================

// Mapper transforms domain events into audit entries.
type Mapper struct{}

// NewMapper creates a new Mapper instance.
func NewMapper() *Mapper {
	return &Mapper{}
}

// MapToEntry transforms a domain event into an AuditEntry.
func (m *Mapper) MapToEntry(
	event domain.DomainEvent,
	actor domain.Actor,
	subject domain.Subject,
	metadata domain.Metadata,
) (*domain.AuditEntry, error) {
	eventType, err := domain.NewEventType(event.EventType())
	if err != nil {
		return nil, err
	}

	occurredAt := time.UnixMilli(event.OccurredAt()).UTC()

	return domain.NewAuditEntry(
		domain.NewEntryID(),
		occurredAt,
		eventType,
		actor,
		subject,
		metadata,
	)
}

// ============================================================================
// Actor Extraction
// ============================================================================

// ExtractActor extracts the Actor from event metadata.
func (m *Mapper) ExtractActor(meta domain.EventMetadata) (domain.Actor, error) {
	// Default to system actor if no user ID
	if meta.UserID == "" {
		return domain.SystemActor("event-processor"), nil
	}

	actorID, err := domain.NewActorID(meta.UserID)
	if err != nil {
		return domain.Actor{}, err
	}

	return domain.NewActor(domain.ActorTypeUser, actorID)
}

// ============================================================================
// Subject Extraction
// ============================================================================

// SubjectMapping defines how to extract subject from an event.
type SubjectMapping struct {
	SubjectType domain.SubjectType
	IDExtractor func(event domain.DomainEvent) string
}

// DefaultSubjectMappings maps aggregate types to subject types.
var DefaultSubjectMappings = map[string]domain.SubjectType{
	"User":         domain.SubjectTypeUser,
	"Credential":   domain.SubjectTypeCredential,
	"Presentation": domain.SubjectTypePresentation,
	"Verification": domain.SubjectTypeVerification,
	"Wallet":       domain.SubjectTypeWallet,
	"Session":      domain.SubjectTypeSession,
	"Organization": domain.SubjectTypeOrganization,
	"DID":          domain.SubjectTypeDID,
	"Vouch":        domain.SubjectTypeVouch,
}

// ExtractSubject extracts the Subject from a domain event.
func (m *Mapper) ExtractSubject(event domain.DomainEvent) (domain.Subject, error) {
	aggregateType := event.AggregateType()
	aggregateID := event.AggregateID()

	subjectType, ok := DefaultSubjectMappings[aggregateType]
	if !ok {
		subjectType = domain.SubjectTypeUnknown
	}

	subjectID, err := domain.NewSubjectID(aggregateID)
	if err != nil {
		return domain.Subject{}, err
	}

	return domain.NewSubject(subjectType, subjectID)
}

// ============================================================================
// Metadata Extraction
// ============================================================================

// ExtractMetadata extracts Metadata from event metadata.
func (m *Mapper) ExtractMetadata(meta domain.EventMetadata) domain.Metadata {
	metadata := domain.NewMetadata()

	if meta.CorrelationID != "" {
		metadata = metadata.WithCorrelationID(meta.CorrelationID)
	}
	if meta.CausationID != "" {
		metadata = metadata.WithCausationID(meta.CausationID)
	}
	if meta.TraceID != "" {
		metadata = metadata.Set("trace_id", meta.TraceID)
	}
	if meta.SpanID != "" {
		metadata = metadata.Set("span_id", meta.SpanID)
	}
	if meta.TenantID != "" {
		metadata = metadata.Set("tenant_id", meta.TenantID)
	}

	// Copy custom metadata
	for k, v := range meta.Custom {
		metadata = metadata.Set(k, v)
	}

	return metadata
}

// ============================================================================
// Full Mapping
// ============================================================================

// MapEvent performs full mapping from a domain event to an audit entry.
// This is a convenience method that extracts actor, subject, and metadata.
func (m *Mapper) MapEvent(event domain.DomainEvent, meta domain.EventMetadata) (*domain.AuditEntry, error) {
	actor, err := m.ExtractActor(meta)
	if err != nil {
		return nil, err
	}

	subject, err := m.ExtractSubject(event)
	if err != nil {
		return nil, err
	}

	metadata := m.ExtractMetadata(meta)

	return m.MapToEntry(event, actor, subject, metadata)
}
