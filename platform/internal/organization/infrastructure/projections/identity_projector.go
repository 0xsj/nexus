package projections

import (
	"context"
	"encoding/json"

	identitydomain "github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// IdentityProjector handles User.* events and updates the organization_user_projections table.
type IdentityProjector struct {
	queries *generated.Queries
}

// NewIdentityProjector creates a new IdentityProjector.
func NewIdentityProjector(queries *generated.Queries) *IdentityProjector {
	return &IdentityProjector{queries: queries}
}

// Handle processes a single event envelope.
func (p *IdentityProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case identitydomain.EventTypeUserRegistered:
		var evt identitydomain.UserRegisteredEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpsertUserProjection(ctx, generated.UpsertUserProjectionParams{
			UserID:     evt.UserID,
			Email:      evt.Email,
			PrimaryDid: evt.PrimaryDID,
			Active:     true,
		})

	case identitydomain.EventTypeUserEmailChanged:
		var evt identitydomain.UserEmailChangedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpdateUserProjectionEmail(ctx, generated.UpdateUserProjectionEmailParams{
			UserID: evt.UserID,
			Email:  evt.NewEmail,
		})

	case identitydomain.EventTypeUserDIDAdded:
		var evt identitydomain.UserDIDAddedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpdateUserProjectionDID(ctx, generated.UpdateUserProjectionDIDParams{
			UserID:     evt.UserID,
			PrimaryDid: evt.DID,
		})

	case identitydomain.EventTypeUserDIDRemoved:
		var evt identitydomain.UserDIDRemovedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpdateUserProjectionDID(ctx, generated.UpdateUserProjectionDIDParams{
			UserID:     evt.UserID,
			PrimaryDid: "",
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
