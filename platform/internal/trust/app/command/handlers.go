package command

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Trust context.
type Handlers struct {
	vouchRepo          domain.VouchRepository
	identityReader     domain.IdentityReader
	credentialReader   domain.CredentialReader
	organizationReader domain.OrganizationReader
	publisher          domain.EventPublisher
	logger             log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	vouchRepo domain.VouchRepository,
	identityReader domain.IdentityReader,
	credentialReader domain.CredentialReader,
	organizationReader domain.OrganizationReader,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		vouchRepo:          vouchRepo,
		identityReader:     identityReader,
		credentialReader:   credentialReader,
		organizationReader: organizationReader,
		publisher:          publisher,
		logger:             logger,
	}
}

// ============================================================================
// GiveVouch Handler
// ============================================================================

// HandleGiveVouch handles the GiveVouch command.
func (h *Handlers) HandleGiveVouch(ctx context.Context, cmd GiveVouch) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleGiveVouch"

	// 1. Parse relationship type
	relationship, err := domain.ParseRelationshipType(cmd.Relationship)
	if err != nil {
		return nil, domain.VouchInvalid(op, "unknown relationship type: "+cmd.Relationship)
	}

	// 2. Validate voucher exists (non-fatal — projection may lag)
	if err := h.identityReader.GetUser(ctx, cmd.VoucherID); err != nil {
		h.logger.Warn("voucher identity verification skipped",
			log.String("op", op),
			log.String("voucher_id", cmd.VoucherID),
			log.Err(err),
		)
	}

	// 3. Validate vouchee exists (non-fatal — projection may lag)
	if err := h.identityReader.GetUser(ctx, cmd.VoucheeID); err != nil {
		h.logger.Warn("vouchee identity verification skipped",
			log.String("op", op),
			log.String("vouchee_id", cmd.VoucheeID),
			log.Err(err),
		)
	}

	// 4. Validate credential exists (non-fatal — projection may lag)
	if cmd.CredentialID != "" {
		if err := h.credentialReader.GetCredential(ctx, cmd.CredentialID); err != nil {
			h.logger.Warn("credential verification skipped",
				log.String("op", op),
				log.String("credential_id", cmd.CredentialID),
				log.Err(err),
			)
		}
	}

	// 5. Parse vouch strength
	strength, err := domain.NewVouchStrength(cmd.Strength)
	if err != nil {
		return nil, domain.VouchInvalid(op, err.Error())
	}

	// 6. Generate vouch ID
	vouchID := domain.NewVouchID()

	// 7. Determine expiration
	var expiresAt time.Time
	if cmd.ExpiresAt != nil {
		expiresAt = *cmd.ExpiresAt
	}

	// 8. Create aggregate
	vouch, err := domain.GiveVouch(
		vouchID,
		cmd.VoucherID,
		cmd.VoucheeID,
		cmd.CredentialID,
		cmd.ClaimKey,
		relationship,
		strength,
		cmd.Statement,
		cmd.Context,
		expiresAt,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 9. Save
	if err := h.vouchRepo.Save(ctx, vouch); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 10. Publish events
	h.publishEvents(ctx, op, vouch.Changes()...)

	return &cqrs.CommandResult{
		ID:      vouchID.String(),
		Version: vouch.Version(),
		Data: GiveVouchResult{
			VouchID: vouchID.String(),
			Status:  vouch.Status().String(),
		},
	}, nil
}

// ============================================================================
// AcceptVouch Handler
// ============================================================================

// HandleAcceptVouch handles the AcceptVouch command.
func (h *Handlers) HandleAcceptVouch(ctx context.Context, cmd AcceptVouch) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAcceptVouch"

	// 1. Load aggregate
	vouchID := domain.VouchIDFromTypesID(cmd.VouchID)
	vouch, err := h.vouchRepo.Get(ctx, vouchID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Accept (validates voucheeID match)
	if err := vouch.Accept(cmd.VoucheeID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.vouchRepo.Save(ctx, vouch); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, vouch.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.VouchID.String(),
		Version: vouch.Version(),
		Data: AcceptVouchResult{
			VouchID: cmd.VouchID.String(),
			Status:  vouch.Status().String(),
		},
	}, nil
}

// ============================================================================
// RevokeVouch Handler
// ============================================================================

// HandleRevokeVouch handles the RevokeVouch command.
func (h *Handlers) HandleRevokeVouch(ctx context.Context, cmd RevokeVouch) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRevokeVouch"

	// 1. Load aggregate
	vouchID := domain.VouchIDFromTypesID(cmd.VouchID)
	vouch, err := h.vouchRepo.Get(ctx, vouchID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Revoke
	if err := vouch.Revoke(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.vouchRepo.Save(ctx, vouch); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, vouch.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.VouchID.String(),
		Version: vouch.Version(),
		Data: RevokeVouchResult{
			VouchID: cmd.VouchID.String(),
			Status:  vouch.Status().String(),
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
