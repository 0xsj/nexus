package projections

import (
	"context"
	"encoding/json"

	credentialdomain "github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// CredentialProjector handles Credential.* events and updates the
// presentation_credential_projections table so the Presentation context
// can verify credential existence without coupling to the Credential context.
type CredentialProjector struct {
	queries *generated.Queries
}

// NewCredentialProjector creates a new CredentialProjector.
func NewCredentialProjector(queries *generated.Queries) *CredentialProjector {
	return &CredentialProjector{queries: queries}
}

// Handle processes an event envelope and updates the credential projection.
func (p *CredentialProjector) Handle(ctx context.Context, envelope *eventsourcing.EventEnvelope) error {
	switch envelope.Type {
	case credentialdomain.EventTypeCredentialIssued:
		var evt credentialdomain.CredentialIssuedEvent
		if err := json.Unmarshal(envelope.Data, &evt); err != nil {
			return err
		}
		// Look up user_id from identity projection via subject DID
		userID, _ := p.queries.GetUserIDByDID(ctx, evt.SubjectDID)
		return p.queries.UpsertCredentialProjection(ctx, generated.UpsertCredentialProjectionParams{
			CredentialID:   evt.CredentialID,
			CredentialType: evt.CredentialType,
			SubjectDid:     evt.SubjectDID,
			UserID:         userID,
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
