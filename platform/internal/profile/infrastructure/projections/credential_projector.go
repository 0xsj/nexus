package projections

import (
	"context"
	"encoding/json"

	credentialdomain "github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// credentialProjectorQueries defines the query methods needed by CredentialProjector.
type credentialProjectorQueries interface {
	UpsertCredentialProjection(ctx context.Context, arg generated.UpsertCredentialProjectionParams) error
	UpdateCredentialProjectionStatus(ctx context.Context, arg generated.UpdateCredentialProjectionStatusParams) error
}

// CredentialProjector projects Credential domain events into the profile_credential_projections table.
type CredentialProjector struct {
	queries credentialProjectorQueries
}

// NewCredentialProjector creates a new CredentialProjector.
func NewCredentialProjector(queries *generated.Queries) *CredentialProjector {
	return &CredentialProjector{queries: queries}
}

// Handle processes a Credential domain event envelope.
func (p *CredentialProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case credentialdomain.EventTypeCredentialIssued:
		var evt credentialdomain.CredentialIssuedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		return p.queries.UpsertCredentialProjection(ctx, generated.UpsertCredentialProjectionParams{
			CredentialID:   evt.CredentialID,
			CredentialType: evt.CredentialType,
			Status:         "active",
		})

	case credentialdomain.EventTypeCredentialRevoked:
		return p.queries.UpdateCredentialProjectionStatus(ctx, generated.UpdateCredentialProjectionStatusParams{
			CredentialID: envelope.AggregateID,
			Status:       "revoked",
		})

	case credentialdomain.EventTypeCredentialExpired:
		return p.queries.UpdateCredentialProjectionStatus(ctx, generated.UpdateCredentialProjectionStatusParams{
			CredentialID: envelope.AggregateID,
			Status:       "expired",
		})
	}

	return nil
}
