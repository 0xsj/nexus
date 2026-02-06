package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Credential context.
type Handlers struct {
	credentialRepo domain.CredentialRepository
	signer         domain.CredentialSigner
	schemaResolver domain.SchemaResolver
	publisher      domain.EventPublisher
	logger         log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	credentialRepo domain.CredentialRepository,
	signer domain.CredentialSigner,
	schemaResolver domain.SchemaResolver,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		credentialRepo: credentialRepo,
		signer:         signer,
		schemaResolver: schemaResolver,
		publisher:      publisher,
		logger:         logger,
	}
}

// ============================================================================
// IssueCredential Handler
// ============================================================================

// HandleIssueCredential handles the IssueCredential command.
func (h *Handlers) HandleIssueCredential(ctx context.Context, cmd IssueCredential) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleIssueCredential"

	// 1. Parse credential type
	credType, err := domain.ParseCredentialType(cmd.CredentialType)
	if err != nil {
		return nil, domain.CredentialInvalid(op, "unknown credential type: "+cmd.CredentialType)
	}

	// 2. Build Claims value object
	claims, err := domain.NewClaims(cmd.Claims)
	if err != nil {
		return nil, domain.ClaimsValidationFailed(op, err.Error())
	}

	// 3. Try schema validation (non-fatal)
	if err := h.schemaResolver.ValidateClaims(ctx, credType, claims); err != nil {
		h.logger.Warn("schema validation skipped",
			log.String("op", op),
			log.String("credential_type", cmd.CredentialType),
			log.Err(err),
		)
	}

	// 4. Generate credential ID
	credID := domain.NewCredentialID()

	// 5. Try signing (non-fatal)
	jwt := ""
	tempCred, err := domain.IssueCredential(
		credID, credType, cmd.IssuerDID, cmd.SubjectDID,
		claims, cmd.ExpiresAt, cmd.VerificationID, "",
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	signedJWT, err := h.signer.Sign(ctx, tempCred)
	if err != nil {
		h.logger.Warn("credential signing skipped",
			log.String("op", op),
			log.Err(err),
		)
	} else {
		jwt = signedJWT
	}

	// 6. Create aggregate with JWT
	cred, err := domain.IssueCredential(
		credID, credType, cmd.IssuerDID, cmd.SubjectDID,
		claims, cmd.ExpiresAt, cmd.VerificationID, jwt,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 7. Save
	if err := h.credentialRepo.Save(ctx, cred); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 8. Publish events
	h.publishEvents(ctx, op, cred.Changes()...)

	return &cqrs.CommandResult{
		ID:      credID.String(),
		Version: cred.Version(),
		Data: IssueCredentialResult{
			CredentialID:   credID.String(),
			CredentialType: cmd.CredentialType,
			Status:         cred.Status().String(),
		},
	}, nil
}

// ============================================================================
// RevokeCredential Handler
// ============================================================================

// HandleRevokeCredential handles the RevokeCredential command.
func (h *Handlers) HandleRevokeCredential(ctx context.Context, cmd RevokeCredential) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRevokeCredential"

	// 1. Load aggregate
	credID := domain.CredentialIDFromTypesID(cmd.CredentialID)
	cred, err := h.credentialRepo.Get(ctx, credID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Revoke
	if err := cred.Revoke(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.credentialRepo.Save(ctx, cred); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, cred.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID.String(),
		Version: cred.Version(),
		Data: RevokeCredentialResult{
			CredentialID: cmd.CredentialID.String(),
			Status:       cred.Status().String(),
		},
	}, nil
}

// ============================================================================
// ExpireCredential Handler
// ============================================================================

// HandleExpireCredential handles the ExpireCredential command.
func (h *Handlers) HandleExpireCredential(ctx context.Context, cmd ExpireCredential) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleExpireCredential"

	// 1. Load aggregate
	credID := domain.CredentialIDFromTypesID(cmd.CredentialID)
	cred, err := h.credentialRepo.Get(ctx, credID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Expire
	if err := cred.Expire(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.credentialRepo.Save(ctx, cred); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, cred.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.CredentialID.String(),
		Version: cred.Version(),
		Data: ExpireCredentialResult{
			CredentialID: cmd.CredentialID.String(),
			Status:       cred.Status().String(),
		},
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// publishEvents publishes domain events (fire-and-forget with logging).
func (h *Handlers) publishEvents(ctx context.Context, op string, events ...eventsourcing.Event) {
	if len(events) == 0 {
		return
	}

	if err := h.publisher.Publish(ctx, events...); err != nil {
		h.logger.Error("failed to publish events",
			log.String("op", op),
			log.Err(err),
		)
	}
}
