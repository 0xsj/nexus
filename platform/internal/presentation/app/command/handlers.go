package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/0xsj/nexus/platform/internal/presentation/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Presentation context.
type Handlers struct {
	presentationRepo domain.PresentationRepository
	shareLinkRepo    domain.ShareLinkRepository
	shareLinkLookup  *postgres.ShareLinkLookup
	publisher        domain.EventPublisher
	logger           log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	presentationRepo domain.PresentationRepository,
	shareLinkRepo domain.ShareLinkRepository,
	shareLinkLookup *postgres.ShareLinkLookup,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		presentationRepo: presentationRepo,
		shareLinkRepo:    shareLinkRepo,
		shareLinkLookup:  shareLinkLookup,
		publisher:        publisher,
		logger:           logger,
	}
}

// ============================================================================
// CreatePresentation Handler
// ============================================================================

// HandleCreatePresentation handles the CreatePresentation command.
func (h *Handlers) HandleCreatePresentation(ctx context.Context, cmd CreatePresentation) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCreatePresentation"

	// 1. Build DisclosurePolicy value object
	policy, err := domain.NewDisclosurePolicy(cmd.DisclosurePolicyType, cmd.AllowedClaims, cmd.BlockedClaims)
	if err != nil {
		return nil, domain.PresentationInvalid(op, err.Error())
	}

	// 2. Generate presentation ID
	presID := domain.NewPresentationID()

	// 3. Create aggregate
	pres, err := domain.CreatePresentation(presID, cmd.HolderDID, cmd.CredentialIDs, policy, cmd.Purpose)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.presentationRepo.Save(ctx, pres); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, pres.Changes()...)

	return &cqrs.CommandResult{
		ID:      presID.String(),
		Version: pres.Version(),
		Data: CreatePresentationResult{
			PresentationID: presID.String(),
			Status:         pres.Status().String(),
		},
	}, nil
}

// ============================================================================
// RevokePresentation Handler
// ============================================================================

// HandleRevokePresentation handles the RevokePresentation command.
func (h *Handlers) HandleRevokePresentation(ctx context.Context, cmd RevokePresentation) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRevokePresentation"

	// 1. Load aggregate
	presID := domain.PresentationIDFromTypesID(cmd.PresentationID)
	pres, err := h.presentationRepo.Get(ctx, presID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Revoke
	if err := pres.Revoke(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.presentationRepo.Save(ctx, pres); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, pres.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.PresentationID.String(),
		Version: pres.Version(),
		Data: RevokePresentationResult{
			PresentationID: cmd.PresentationID.String(),
			Status:         pres.Status().String(),
		},
	}, nil
}

// ============================================================================
// CreateShareLink Handler
// ============================================================================

// HandleCreateShareLink handles the CreateShareLink command.
func (h *Handlers) HandleCreateShareLink(ctx context.Context, cmd CreateShareLink) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCreateShareLink"

	// 1. Verify presentation exists
	presID := domain.PresentationIDFromTypesID(cmd.PresentationID)
	exists, err := h.presentationRepo.Exists(ctx, presID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if !exists {
		return nil, domain.PresentationNotFound(op, cmd.PresentationID.String())
	}

	// 2. Generate token and ID
	token := uuid.New().String()
	slID := domain.NewShareLinkID()

	// 3. Handle optional expiry
	var expiresAt time.Time
	if cmd.ExpiresAt != nil {
		expiresAt = *cmd.ExpiresAt
	}

	// 4. Hash PIN if provided (placeholder - in production use bcrypt)
	pinHash := cmd.Pin

	// 5. Create aggregate
	sl, err := domain.CreateShareLink(slID, presID, token, expiresAt, cmd.MaxViews, pinHash, cmd.Audience)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Save
	if err := h.shareLinkRepo.Save(ctx, sl); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 7. Publish events
	h.publishEvents(ctx, op, sl.Changes()...)

	return &cqrs.CommandResult{
		ID:      slID.String(),
		Version: sl.Version(),
		Data: CreateShareLinkResult{
			ShareLinkID:    slID.String(),
			PresentationID: cmd.PresentationID.String(),
			Token:          token,
			Status:         sl.Status().String(),
		},
	}, nil
}

// ============================================================================
// AccessShareLink Handler
// ============================================================================

// HandleAccessShareLink handles the AccessShareLink command.
func (h *Handlers) HandleAccessShareLink(ctx context.Context, cmd AccessShareLink) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAccessShareLink"

	// 1. Load aggregate
	slID := domain.ShareLinkIDFromTypesID(cmd.ShareLinkID)
	sl, err := h.shareLinkRepo.Get(ctx, slID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Access
	if err := sl.Access(cmd.VerifierDID, cmd.IPAddress); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.shareLinkRepo.Save(ctx, sl); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Insert access grant row
	now := time.Now().UTC()
	if err := h.shareLinkLookup.InsertAccessGrant(ctx, slID.String(), cmd.VerifierDID, cmd.IPAddress, now, nil); err != nil {
		h.logger.Error("failed to insert access grant",
			log.String("op", op),
			log.String("share_link_id", slID.String()),
			log.Err(err),
		)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, sl.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ShareLinkID.String(),
		Version: sl.Version(),
		Data: AccessShareLinkResult{
			ShareLinkID:  cmd.ShareLinkID.String(),
			CurrentViews: sl.CurrentViews(),
			Status:       sl.Status().String(),
		},
	}, nil
}

// ============================================================================
// RevokeShareLink Handler
// ============================================================================

// HandleRevokeShareLink handles the RevokeShareLink command.
func (h *Handlers) HandleRevokeShareLink(ctx context.Context, cmd RevokeShareLink) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRevokeShareLink"

	// 1. Load aggregate
	slID := domain.ShareLinkIDFromTypesID(cmd.ShareLinkID)
	sl, err := h.shareLinkRepo.Get(ctx, slID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Revoke
	if err := sl.Revoke(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.shareLinkRepo.Save(ctx, sl); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, sl.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ShareLinkID.String(),
		Version: sl.Version(),
		Data: RevokeShareLinkResult{
			ShareLinkID: cmd.ShareLinkID.String(),
			Status:      sl.Status().String(),
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
