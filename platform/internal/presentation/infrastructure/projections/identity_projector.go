package projections

import (
	"context"
	"encoding/json"

	identitydomain "github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// IdentityProjector handles User.* events and updates the
// presentation_user_projections table so the Presentation context
// can resolve holder DIDs without coupling to the Identity context.
type IdentityProjector struct {
	queries *generated.Queries
}

// NewIdentityProjector creates a new IdentityProjector.
func NewIdentityProjector(queries *generated.Queries) *IdentityProjector {
	return &IdentityProjector{queries: queries}
}

// Handle processes an event envelope and updates the user projection.
func (p *IdentityProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case identitydomain.EventTypeUserRegistered:
		var evt identitydomain.UserRegisteredEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpsertUserProjection(ctx, generated.UpsertUserProjectionParams{
			UserID:     evt.UserID,
			PrimaryDid: evt.PrimaryDID,
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
	}
	return nil
}
