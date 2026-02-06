package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/profile/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Profile context.
type Handlers struct {
	profileRepo      domain.ProfileRepository
	vanitySlugLookup domain.VanitySlugLookup
	credentialReader domain.CredentialReader
	publisher        domain.EventPublisher
	logger           log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	profileRepo domain.ProfileRepository,
	vanitySlugLookup domain.VanitySlugLookup,
	credentialReader domain.CredentialReader,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		profileRepo:      profileRepo,
		vanitySlugLookup: vanitySlugLookup,
		credentialReader: credentialReader,
		publisher:        publisher,
		logger:           logger,
	}
}

// ============================================================================
// CreateProfile Handler
// ============================================================================

// HandleCreateProfile handles the CreateProfile command.
func (h *Handlers) HandleCreateProfile(ctx context.Context, cmd CreateProfile) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCreateProfile"

	// 1. Check if vanity slug is already taken
	slugTaken, err := h.vanitySlugLookup.SlugExists(ctx, cmd.VanitySlug)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if slugTaken {
		return nil, domain.VanityURLTaken(op, cmd.VanitySlug)
	}

	// 2. Generate profile ID
	profileID := domain.NewProfileID()

	// 3. Create aggregate
	p, err := domain.CreateProfile(profileID, cmd.UserID, cmd.DisplayName, cmd.Headline, cmd.VanitySlug)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.profileRepo.Save(ctx, p); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, p.Changes()...)

	return &cqrs.CommandResult{
		ID:      profileID.String(),
		Version: p.Version(),
		Data: CreateProfileResult{
			ProfileID:  profileID.String(),
			VanitySlug: cmd.VanitySlug,
		},
	}, nil
}

// ============================================================================
// UpdateProfile Handler
// ============================================================================

// HandleUpdateProfile handles the UpdateProfile command.
func (h *Handlers) HandleUpdateProfile(ctx context.Context, cmd UpdateProfile) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUpdateProfile"

	// 1. Load aggregate
	profileID := domain.ProfileIDFromTypesID(cmd.ProfileID)
	p, err := h.profileRepo.Get(ctx, profileID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Update metadata
	if err := p.UpdateMetadata(cmd.DisplayName, cmd.Headline, cmd.Bio); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.profileRepo.Save(ctx, p); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, p.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ProfileID.String(),
		Version: p.Version(),
		Data: UpdateProfileResult{
			ProfileID: cmd.ProfileID.String(),
		},
	}, nil
}

// ============================================================================
// AddBadge Handler
// ============================================================================

// HandleAddBadge handles the AddBadge command.
func (h *Handlers) HandleAddBadge(ctx context.Context, cmd AddBadge) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAddBadge"

	// 1. Load aggregate
	profileID := domain.ProfileIDFromTypesID(cmd.ProfileID)
	p, err := h.profileRepo.Get(ctx, profileID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Optionally verify credential exists
	exists, _, err := h.credentialReader.GetCredentialByID(ctx, cmd.CredentialID)
	if err != nil {
		h.logger.Warn("credential lookup failed",
			log.String("op", op),
			log.String("credential_id", cmd.CredentialID),
			log.Err(err),
		)
	} else if !exists {
		h.logger.Warn("credential not found for badge, proceeding anyway",
			log.String("op", op),
			log.String("credential_id", cmd.CredentialID),
		)
	}

	// 3. Create badge value object
	badgeID := domain.NewBadgeID()
	badge := domain.NewBadge(
		badgeID,
		cmd.CredentialID,
		cmd.BadgeType,
		cmd.DisplayName,
		cmd.PrimaryValue,
		cmd.VerifiedAt,
	).WithSecondaryValue(cmd.SecondaryValue).
		WithIcon(cmd.Icon)

	// 4. Add badge to profile
	if err := p.AddBadge(badge); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save
	if err := h.profileRepo.Save(ctx, p); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, p.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ProfileID.String(),
		Version: p.Version(),
		Data: AddBadgeResult{
			ProfileID: cmd.ProfileID.String(),
			BadgeID:   badgeID.String(),
		},
	}, nil
}

// ============================================================================
// RemoveBadge Handler
// ============================================================================

// HandleRemoveBadge handles the RemoveBadge command.
func (h *Handlers) HandleRemoveBadge(ctx context.Context, cmd RemoveBadge) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRemoveBadge"

	// 1. Load aggregate
	profileID := domain.ProfileIDFromTypesID(cmd.ProfileID)
	p, err := h.profileRepo.Get(ctx, profileID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Remove badge
	badgeID := domain.BadgeIDFromTypesID(cmd.BadgeID)
	if err := p.RemoveBadge(badgeID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.profileRepo.Save(ctx, p); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, p.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ProfileID.String(),
		Version: p.Version(),
		Data: RemoveBadgeResult{
			ProfileID: cmd.ProfileID.String(),
			BadgeID:   cmd.BadgeID.String(),
		},
	}, nil
}

// ============================================================================
// ChangeBadgeVisibility Handler
// ============================================================================

// HandleChangeBadgeVisibility handles the ChangeBadgeVisibility command.
func (h *Handlers) HandleChangeBadgeVisibility(ctx context.Context, cmd ChangeBadgeVisibility) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleChangeBadgeVisibility"

	// 1. Parse visibility
	visibility, err := domain.ParseVisibility(cmd.Visibility)
	if err != nil {
		return nil, domain.ProfileInvalid(op, "invalid visibility: "+cmd.Visibility)
	}

	// 2. Load aggregate
	profileID := domain.ProfileIDFromTypesID(cmd.ProfileID)
	p, err := h.profileRepo.Get(ctx, profileID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Change visibility
	badgeID := domain.BadgeIDFromTypesID(cmd.BadgeID)
	if err := p.ChangeBadgeVisibility(badgeID, visibility); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.profileRepo.Save(ctx, p); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, p.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ProfileID.String(),
		Version: p.Version(),
		Data: ChangeBadgeVisibilityResult{
			ProfileID:  cmd.ProfileID.String(),
			BadgeID:    cmd.BadgeID.String(),
			Visibility: cmd.Visibility,
		},
	}, nil
}

// ============================================================================
// ClaimVanityURL Handler
// ============================================================================

// HandleClaimVanityURL handles the ClaimVanityURL command.
func (h *Handlers) HandleClaimVanityURL(ctx context.Context, cmd ClaimVanityURL) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleClaimVanityURL"

	// 1. Check if slug is already taken
	slugTaken, err := h.vanitySlugLookup.SlugExists(ctx, cmd.Slug)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if slugTaken {
		return nil, domain.VanityURLTaken(op, cmd.Slug)
	}

	// 2. Load aggregate
	profileID := domain.ProfileIDFromTypesID(cmd.ProfileID)
	p, err := h.profileRepo.Get(ctx, profileID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Claim vanity URL
	if err := p.ClaimVanityURL(cmd.Slug); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.profileRepo.Save(ctx, p); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, p.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.ProfileID.String(),
		Version: p.Version(),
		Data: ClaimVanityURLResult{
			ProfileID: cmd.ProfileID.String(),
			Slug:      cmd.Slug,
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
