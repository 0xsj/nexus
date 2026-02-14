package projections

import (
	"context"
	"encoding/json"

	schemadomain "github.com/0xsj/nexus/platform/internal/schema/domain"

	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// schemaProjectorQueries defines the query methods needed by SchemaProjector.
type schemaProjectorQueries interface {
	UpsertSchemaProjection(ctx context.Context, arg generated.UpsertSchemaProjectionParams) error
	UpdateSchemaProjectionStatus(ctx context.Context, arg generated.UpdateSchemaProjectionStatusParams) error
	UpdateSchemaProjectionClaims(ctx context.Context, arg generated.UpdateSchemaProjectionClaimsParams) error
}

// SchemaProjector projects Schema domain events into the credential_schema_projections table.
type SchemaProjector struct {
	queries schemaProjectorQueries
}

// NewSchemaProjector creates a new SchemaProjector.
func NewSchemaProjector(queries *generated.Queries) *SchemaProjector {
	return &SchemaProjector{queries: queries}
}

// Handle processes a Schema domain event envelope.
func (p *SchemaProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case schemadomain.EventTypeSchemaRegistered:
		var evt schemadomain.SchemaRegisteredEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		claimsJSON, _ := json.Marshal(evt.Claims)
		return p.queries.UpsertSchemaProjection(ctx, generated.UpsertSchemaProjectionParams{
			SchemaID:   evt.SchemaID,
			SchemaType: evt.SchemaType,
			Status:     "active",
			Claims:     claimsJSON,
		})

	case schemadomain.EventTypeSchemaVersionAdded:
		var evt schemadomain.SchemaVersionAddedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		claimsJSON, _ := json.Marshal(evt.Claims)
		return p.queries.UpdateSchemaProjectionClaims(ctx, generated.UpdateSchemaProjectionClaimsParams{
			SchemaID: evt.SchemaID,
			Claims:   claimsJSON,
		})

	case schemadomain.EventTypeSchemaDeprecated:
		return p.queries.UpdateSchemaProjectionStatus(ctx, generated.UpdateSchemaProjectionStatusParams{
			SchemaID: envelope.AggregateID,
			Status:   "deprecated",
		})

	case schemadomain.EventTypeSchemaActivated:
		return p.queries.UpdateSchemaProjectionStatus(ctx, generated.UpdateSchemaProjectionStatusParams{
			SchemaID: envelope.AggregateID,
			Status:   "active",
		})
	}

	return nil
}
