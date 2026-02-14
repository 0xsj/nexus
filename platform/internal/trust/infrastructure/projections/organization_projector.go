package projections

import (
	"context"
	"encoding/json"

	orgdomain "github.com/0xsj/nexus/platform/internal/organization/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// organizationProjectorQueries defines the query methods needed by OrganizationProjector.
type organizationProjectorQueries interface {
	UpsertOrganizationProjection(ctx context.Context, arg generated.UpsertOrganizationProjectionParams) error
	UpdateOrganizationProjectionVerified(ctx context.Context, arg generated.UpdateOrganizationProjectionVerifiedParams) error
	UpdateOrganizationProjectionActive(ctx context.Context, arg generated.UpdateOrganizationProjectionActiveParams) error
}

// OrganizationProjector projects Organization domain events into the trust_organization_projections table.
type OrganizationProjector struct {
	queries organizationProjectorQueries
}

// NewOrganizationProjector creates a new OrganizationProjector.
func NewOrganizationProjector(queries *generated.Queries) *OrganizationProjector {
	return &OrganizationProjector{queries: queries}
}

// Handle processes an Organization domain event envelope.
func (p *OrganizationProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case orgdomain.EventTypeOrganizationCreated:
		var evt orgdomain.OrganizationCreatedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpsertOrganizationProjection(ctx, generated.UpsertOrganizationProjectionParams{
			OrganizationID:     evt.OrganizationID,
			VerificationStatus: "unverified",
			Active:             true,
		})

	case orgdomain.EventTypeOrganizationVerified:
		return p.queries.UpdateOrganizationProjectionVerified(ctx, generated.UpdateOrganizationProjectionVerifiedParams{
			OrganizationID:     envelope.AggregateID,
			VerificationStatus: "verified",
		})

	case orgdomain.EventTypeOrganizationDeleted:
		return p.queries.UpdateOrganizationProjectionActive(ctx, generated.UpdateOrganizationProjectionActiveParams{
			OrganizationID: envelope.AggregateID,
			Active:         false,
		})
	}

	return nil
}
