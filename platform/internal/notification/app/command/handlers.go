package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Notification context.
type Handlers struct {
	notificationRepo domain.NotificationRepository
	preferencesRepo  domain.PreferencesRepository
	identityReader   domain.IdentityReader
	publisher        domain.EventPublisher
	logger           log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	notificationRepo domain.NotificationRepository,
	preferencesRepo domain.PreferencesRepository,
	identityReader domain.IdentityReader,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		identityReader:   identityReader,
		publisher:        publisher,
		logger:           logger,
	}
}

// ============================================================================
// CreateNotification Handler
// ============================================================================

// HandleCreateNotification handles the CreateNotification command.
func (h *Handlers) HandleCreateNotification(ctx context.Context, cmd CreateNotification) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCreateNotification"

	// 1. Parse category
	category, err := domain.ParseCategory(cmd.Category)
	if err != nil {
		return nil, domain.NotificationInvalid(op, "unknown category: "+cmd.Category)
	}

	// 2. Parse channel
	channel, err := domain.ParseChannel(cmd.Channel)
	if err != nil {
		return nil, domain.NotificationInvalid(op, "unknown channel: "+cmd.Channel)
	}

	// 3. Validate recipient exists (non-fatal — projection may lag)
	if exists, err := h.identityReader.UserExists(ctx, cmd.RecipientID); err != nil {
		h.logger.Warn("recipient identity verification failed",
			log.String("op", op),
			log.String("recipient_id", cmd.RecipientID),
			log.Err(err),
		)
	} else if !exists {
		h.logger.Warn("recipient not found in projection",
			log.String("op", op),
			log.String("recipient_id", cmd.RecipientID),
		)
	}

	// 4. Generate notification ID
	notifID := domain.NewNotificationID()

	// 5. Create aggregate
	notif, err := domain.CreateNotification(
		notifID,
		cmd.RecipientID,
		category,
		channel,
		cmd.TemplateID,
		cmd.Subject,
		cmd.Body,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Set optional action URL
	if cmd.ActionURL != "" {
		notif.WithActionURL(cmd.ActionURL)
	}

	// 7. Save
	if err := h.notificationRepo.Save(ctx, notif); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 8. Publish events
	h.publishEvents(ctx, op, notif.Changes()...)

	return &cqrs.CommandResult{
		ID:      notifID.String(),
		Version: notif.Version(),
		Data: CreateNotificationResult{
			NotificationID: notifID.String(),
			Status:         notif.Status().String(),
		},
	}, nil
}

// ============================================================================
// MarkNotificationSent Handler
// ============================================================================

// HandleMarkNotificationSent handles the MarkNotificationSent command.
func (h *Handlers) HandleMarkNotificationSent(ctx context.Context, cmd MarkNotificationSent) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleMarkNotificationSent"

	// 1. Load aggregate
	notifID := domain.NotificationIDFromTypesID(cmd.NotificationID)
	notif, err := h.notificationRepo.Get(ctx, notifID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Mark as sent
	if err := notif.MarkAsSent(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.notificationRepo.Save(ctx, notif); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, notif.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.NotificationID.String(),
		Version: notif.Version(),
		Data: MarkNotificationSentResult{
			NotificationID: cmd.NotificationID.String(),
			Status:         notif.Status().String(),
		},
	}, nil
}

// ============================================================================
// MarkNotificationFailed Handler
// ============================================================================

// HandleMarkNotificationFailed handles the MarkNotificationFailed command.
func (h *Handlers) HandleMarkNotificationFailed(ctx context.Context, cmd MarkNotificationFailed) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleMarkNotificationFailed"

	// 1. Load aggregate
	notifID := domain.NotificationIDFromTypesID(cmd.NotificationID)
	notif, err := h.notificationRepo.Get(ctx, notifID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Mark as failed
	if err := notif.MarkAsFailed(cmd.ErrorMessage); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.notificationRepo.Save(ctx, notif); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, notif.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.NotificationID.String(),
		Version: notif.Version(),
		Data: MarkNotificationFailedResult{
			NotificationID: cmd.NotificationID.String(),
			Status:         notif.Status().String(),
		},
	}, nil
}

// ============================================================================
// MarkNotificationRead Handler
// ============================================================================

// HandleMarkNotificationRead handles the MarkNotificationRead command.
func (h *Handlers) HandleMarkNotificationRead(ctx context.Context, cmd MarkNotificationRead) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleMarkNotificationRead"

	// 1. Load aggregate
	notifID := domain.NotificationIDFromTypesID(cmd.NotificationID)
	notif, err := h.notificationRepo.Get(ctx, notifID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Mark as read
	if err := notif.MarkAsRead(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.notificationRepo.Save(ctx, notif); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, notif.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.NotificationID.String(),
		Version: notif.Version(),
		Data: MarkNotificationReadResult{
			NotificationID: cmd.NotificationID.String(),
			Status:         notif.Status().String(),
		},
	}, nil
}

// ============================================================================
// SuppressNotification Handler
// ============================================================================

// HandleSuppressNotification handles the SuppressNotification command.
func (h *Handlers) HandleSuppressNotification(ctx context.Context, cmd SuppressNotification) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleSuppressNotification"

	// 1. Load aggregate
	notifID := domain.NotificationIDFromTypesID(cmd.NotificationID)
	notif, err := h.notificationRepo.Get(ctx, notifID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Suppress
	if err := notif.Suppress(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.notificationRepo.Save(ctx, notif); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, notif.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.NotificationID.String(),
		Version: notif.Version(),
		Data: SuppressNotificationResult{
			NotificationID: cmd.NotificationID.String(),
			Status:         notif.Status().String(),
		},
	}, nil
}

// ============================================================================
// UpdatePreferences Handler
// ============================================================================

// HandleUpdatePreferences handles the UpdatePreferences command.
func (h *Handlers) HandleUpdatePreferences(ctx context.Context, cmd UpdatePreferences) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUpdatePreferences"

	// 1. Load existing preferences (defaults if none)
	prefs, err := h.preferencesRepo.GetPreferences(ctx, cmd.UserID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Apply updates
	if cmd.GlobalEnabled != nil {
		prefs = prefs.WithGlobalEnabled(*cmd.GlobalEnabled)
	}

	if cmd.ChannelUpdates != nil {
		for chStr, enabled := range cmd.ChannelUpdates {
			ch, err := domain.ParseChannel(chStr)
			if err != nil {
				return nil, domain.NotificationInvalid(op, "unknown channel: "+chStr)
			}
			prefs = prefs.WithChannelEnabled(ch, enabled)
		}
	}

	if cmd.CategoryUpdates != nil {
		for catStr, channels := range cmd.CategoryUpdates {
			cat, err := domain.ParseCategory(catStr)
			if err != nil {
				return nil, domain.NotificationInvalid(op, "unknown category: "+catStr)
			}
			for chStr, enabled := range channels {
				ch, err := domain.ParseChannel(chStr)
				if err != nil {
					return nil, domain.NotificationInvalid(op, "unknown channel: "+chStr)
				}
				prefs = prefs.WithCategoryChannel(cat, ch, enabled)
			}
		}
	}

	if cmd.DigestEnabled != nil {
		prefs = prefs.WithDigestEnabled(*cmd.DigestEnabled)
	}

	if cmd.DigestFrequency != "" {
		freq, err := domain.ParseDigestFrequency(cmd.DigestFrequency)
		if err != nil {
			return nil, domain.NotificationInvalid(op, "unknown digest frequency: "+cmd.DigestFrequency)
		}
		prefs = prefs.WithDigestFrequency(freq)
	}

	if cmd.Timezone != "" {
		prefs = prefs.WithTimezone(cmd.Timezone)
	}

	// 3. Save
	if err := h.preferencesRepo.SavePreferences(ctx, prefs); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &cqrs.CommandResult{
		ID: cmd.UserID,
		Data: UpdatePreferencesResult{
			UserID: cmd.UserID,
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
