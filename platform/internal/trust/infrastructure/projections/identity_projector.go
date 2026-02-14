package projections

import (
	"context"
	"encoding/json"

	identitydomain "github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// identityProjectorQueries defines the query methods needed by IdentityProjector.
type identityProjectorQueries interface {
	UpsertUserProjection(ctx context.Context, arg generated.UpsertUserProjectionParams) error
	UpdateUserProjectionActive(ctx context.Context, arg generated.UpdateUserProjectionActiveParams) error
}

// IdentityProjector projects Identity domain events into the trust_user_projections table.
type IdentityProjector struct {
	queries identityProjectorQueries
}

// NewIdentityProjector creates a new IdentityProjector.
func NewIdentityProjector(queries *generated.Queries) *IdentityProjector {
	return &IdentityProjector{queries: queries}
}

// Handle processes an Identity domain event envelope.
func (p *IdentityProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case identitydomain.EventTypeUserRegistered:
		var evt identitydomain.UserRegisteredEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpsertUserProjection(ctx, generated.UpsertUserProjectionParams{
			UserID: evt.UserID,
			Active: true,
		})

	case identitydomain.EventTypeUserActivated, identitydomain.EventTypeUserReactivated:
		return p.queries.UpdateUserProjectionActive(ctx, generated.UpdateUserProjectionActiveParams{
			UserID: envelope.AggregateID,
			Active: true,
		})

	case identitydomain.EventTypeUserSuspended, identitydomain.EventTypeUserDeleted:
		return p.queries.UpdateUserProjectionActive(ctx, generated.UpdateUserProjectionActiveParams{
			UserID: envelope.AggregateID,
			Active: false,
		})
	}

	return nil
}
