package projections

import (
	"context"
	"encoding/json"

	issuerdomain "github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// issuerProjectorQueries defines the query methods needed by IssuerProjector.
type issuerProjectorQueries interface {
	UpsertIssuerProjection(ctx context.Context, arg generated.UpsertIssuerProjectionParams) error
	UpdateIssuerProjectionActive(ctx context.Context, arg generated.UpdateIssuerProjectionActiveParams) error
}

// IssuerProjector projects Issuer domain events into the schema_issuer_projections table.
type IssuerProjector struct {
	queries issuerProjectorQueries
}

// NewIssuerProjector creates a new IssuerProjector.
func NewIssuerProjector(queries *generated.Queries) *IssuerProjector {
	return &IssuerProjector{queries: queries}
}

// Handle processes an Issuer domain event envelope.
func (p *IssuerProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case issuerdomain.EventTypeIssuerRegistered:
		var evt issuerdomain.IssuerRegisteredEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpsertIssuerProjection(ctx, generated.UpsertIssuerProjectionParams{
			IssuerID: evt.IssuerID,
			Name:     evt.Name,
			Active:   true,
		})

	case issuerdomain.EventTypeIssuerActivated:
		return p.queries.UpsertIssuerProjection(ctx, generated.UpsertIssuerProjectionParams{
			IssuerID: envelope.AggregateID,
			Name:     "",
			Active:   true,
		})

	case issuerdomain.EventTypeIssuerSuspended:
		return p.queries.UpdateIssuerProjectionActive(ctx, generated.UpdateIssuerProjectionActiveParams{
			IssuerID: envelope.AggregateID,
			Active:   false,
		})
	}

	return nil
}
