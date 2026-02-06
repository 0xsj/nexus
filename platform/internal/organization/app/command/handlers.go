package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/organization/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Organization context.
type Handlers struct {
	orgRepo        domain.OrganizationRepository
	slugLookup     domain.SlugLookup
	identityReader domain.IdentityReader
	publisher      domain.EventPublisher
	logger         log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	orgRepo domain.OrganizationRepository,
	slugLookup domain.SlugLookup,
	identityReader domain.IdentityReader,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		orgRepo:        orgRepo,
		slugLookup:     slugLookup,
		identityReader: identityReader,
		publisher:      publisher,
		logger:         logger,
	}
}

// ============================================================================
// CreateOrganization Handler
// ============================================================================

// HandleCreateOrganization handles the CreateOrganization command.
func (h *Handlers) HandleCreateOrganization(ctx context.Context, cmd CreateOrganization) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCreateOrganization"

	// 1. Parse organization type
	orgType, err := domain.ParseOrganizationType(cmd.OrgType)
	if err != nil {
		return nil, domain.OrganizationInvalid(op, "unknown organization type: "+cmd.OrgType)
	}

	// 2. Check slug uniqueness
	slugExists, err := h.slugLookup.SlugExists(ctx, cmd.Slug)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if slugExists {
		return nil, domain.SlugAlreadyTaken(op, cmd.Slug)
	}

	// 3. Generate IDs
	orgID := domain.NewOrganizationID()
	ownerMemberID := domain.NewMemberID()

	// 4. Create aggregate
	org, err := domain.CreateOrganization(
		orgID,
		cmd.Name,
		cmd.Slug,
		orgType,
		cmd.OwnerUserID,
		ownerMemberID,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      orgID.String(),
		Version: org.Version(),
		Data: CreateOrganizationResult{
			OrganizationID: orgID.String(),
			Slug:           cmd.Slug,
			OwnerMemberID:  ownerMemberID.String(),
		},
	}, nil
}

// ============================================================================
// UpdateOrganization Handler
// ============================================================================

// HandleUpdateOrganization handles the UpdateOrganization command.
func (h *Handlers) HandleUpdateOrganization(ctx context.Context, cmd UpdateOrganization) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUpdateOrganization"

	// 1. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Update
	if err := org.UpdateMetadata(cmd.Name, cmd.Description); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: UpdateOrganizationResult{
			OrganizationID: cmd.OrganizationID.String(),
		},
	}, nil
}

// ============================================================================
// AddMember Handler
// ============================================================================

// HandleAddMember handles the AddMember command.
func (h *Handlers) HandleAddMember(ctx context.Context, cmd AddMember) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAddMember"

	// 1. Parse role
	role, err := domain.ParseRole(cmd.Role)
	if err != nil {
		return nil, domain.OrganizationInvalid(op, "unknown role: "+cmd.Role)
	}

	// 2. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Generate member ID and add
	memberID := domain.NewMemberID()
	if err := org.AddMember(memberID, cmd.UserID, role); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: AddMemberResult{
			OrganizationID: cmd.OrganizationID.String(),
			MemberID:       memberID.String(),
			Role:           cmd.Role,
		},
	}, nil
}

// ============================================================================
// RemoveMember Handler
// ============================================================================

// HandleRemoveMember handles the RemoveMember command.
func (h *Handlers) HandleRemoveMember(ctx context.Context, cmd RemoveMember) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRemoveMember"

	// 1. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Parse member ID and remove
	memberID, err := domain.ParseMemberID(cmd.MemberID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if err := org.RemoveMember(memberID, cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: RemoveMemberResult{
			OrganizationID: cmd.OrganizationID.String(),
			MemberID:       cmd.MemberID.String(),
		},
	}, nil
}

// ============================================================================
// ChangeMemberRole Handler
// ============================================================================

// HandleChangeMemberRole handles the ChangeMemberRole command.
func (h *Handlers) HandleChangeMemberRole(ctx context.Context, cmd ChangeMemberRole) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleChangeMemberRole"

	// 1. Parse new role
	newRole, err := domain.ParseRole(cmd.NewRole)
	if err != nil {
		return nil, domain.OrganizationInvalid(op, "unknown role: "+cmd.NewRole)
	}

	// 2. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Parse member ID and change role
	memberID, err := domain.ParseMemberID(cmd.MemberID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if err := org.ChangeMemberRole(memberID, newRole); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: ChangeMemberRoleResult{
			OrganizationID: cmd.OrganizationID.String(),
			MemberID:       cmd.MemberID.String(),
			NewRole:        cmd.NewRole,
		},
	}, nil
}

// ============================================================================
// TransferOwnership Handler
// ============================================================================

// HandleTransferOwnership handles the TransferOwnership command.
func (h *Handlers) HandleTransferOwnership(ctx context.Context, cmd TransferOwnership) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleTransferOwnership"

	// 1. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Parse target member ID and transfer
	toMemberID, err := domain.ParseMemberID(cmd.ToMemberID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if err := org.TransferOwnership(toMemberID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: TransferOwnershipResult{
			OrganizationID: cmd.OrganizationID.String(),
			NewOwnerID:     cmd.ToMemberID.String(),
		},
	}, nil
}

// ============================================================================
// CompleteVerification Handler
// ============================================================================

// HandleCompleteVerification handles the CompleteVerification command.
func (h *Handlers) HandleCompleteVerification(ctx context.Context, cmd CompleteVerification) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCompleteVerification"

	// 1. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Complete verification
	if err := org.CompleteVerification(cmd.DID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: CompleteVerificationResult{
			OrganizationID:     cmd.OrganizationID.String(),
			DID:                cmd.DID,
			VerificationStatus: org.VerificationStatus().String(),
		},
	}, nil
}

// ============================================================================
// DeleteOrganization Handler
// ============================================================================

// HandleDeleteOrganization handles the DeleteOrganization command.
func (h *Handlers) HandleDeleteOrganization(ctx context.Context, cmd DeleteOrganization) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleDeleteOrganization"

	// 1. Load aggregate
	orgID, err := domain.ParseOrganizationID(cmd.OrganizationID.String())
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	org, err := h.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Delete
	if err := org.Delete(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.orgRepo.Save(ctx, org); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, org.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.OrganizationID.String(),
		Version: org.Version(),
		Data: DeleteOrganizationResult{
			OrganizationID: cmd.OrganizationID.String(),
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
