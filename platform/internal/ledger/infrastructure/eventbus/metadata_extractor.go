package eventbus

import (
	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// MetadataExtractor extracts metadata from domain events that wrap EventEnvelopes.
type MetadataExtractor struct{}

// Compile-time check.
var _ domain.MetadataExtractor = (*MetadataExtractor)(nil)

// NewMetadataExtractor creates a new MetadataExtractor.
func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

// Extract retrieves metadata from a domain event.
// If the event wraps an EventEnvelope (via our adapter), extracts from its metadata.
// Otherwise returns empty metadata.
func (e *MetadataExtractor) Extract(event domain.DomainEvent) domain.EventMetadata {
	// Check if the event is our adapter type that carries an envelope.
	type envelopeCarrier interface {
		Envelope() *eventsourcing.EventEnvelope
	}

	carrier, ok := event.(envelopeCarrier)
	if !ok {
		return domain.EventMetadata{}
	}

	envelope := carrier.Envelope()
	return domain.EventMetadata{
		CorrelationID: envelope.Metadata.CorrelationID,
		CausationID:   envelope.Metadata.CausationID,
		UserID:        envelope.Metadata.UserID,
		TenantID:      envelope.Metadata.TenantID,
		TraceID:       envelope.Metadata.TraceID,
		SpanID:        envelope.Metadata.SpanID,
		Custom:        envelope.Metadata.Custom,
	}
}
