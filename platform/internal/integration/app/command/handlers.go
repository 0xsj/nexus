package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Integration context.
type Handlers struct {
	integrationRepo domain.IntegrationRepository
	publisher       domain.EventPublisher
	logger          log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	integrationRepo domain.IntegrationRepository,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		integrationRepo: integrationRepo,
		publisher:       publisher,
		logger:          logger,
	}
}

// ============================================================================
// ConnectProvider Handler
// ============================================================================

// HandleConnectProvider handles the ConnectProvider command.
func (h *Handlers) HandleConnectProvider(ctx context.Context, cmd ConnectProvider) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleConnectProvider"

	// 1. Parse provider type
	providerType, err := domain.ParseProviderType(cmd.ProviderType)
	if err != nil {
		return nil, domain.ProviderNotSupported(op, cmd.ProviderType)
	}

	// 2. Generate integration ID
	integID := domain.NewIntegrationID()

	// 3. Create aggregate
	integ, err := domain.ConnectProvider(
		integID,
		cmd.UserID,
		providerType,
		cmd.ProviderUserID,
		cmd.ProviderUsername,
		cmd.Scopes,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.integrationRepo.Save(ctx, integ); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, integ.Changes()...)

	return &cqrs.CommandResult{
		ID:      integID.String(),
		Version: integ.Version(),
		Data: ConnectProviderResult{
			IntegrationID: integID.String(),
			ProviderType:  cmd.ProviderType,
			Status:        integ.Status().String(),
		},
	}, nil
}

// ============================================================================
// DisconnectProvider Handler
// ============================================================================

// HandleDisconnectProvider handles the DisconnectProvider command.
func (h *Handlers) HandleDisconnectProvider(ctx context.Context, cmd DisconnectProvider) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleDisconnectProvider"

	// 1. Load aggregate
	integID := domain.IntegrationIDFromTypesID(cmd.IntegrationID)
	integ, err := h.integrationRepo.Get(ctx, integID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Disconnect
	if err := integ.Disconnect(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.integrationRepo.Save(ctx, integ); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, integ.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.IntegrationID.String(),
		Version: integ.Version(),
		Data: DisconnectProviderResult{
			IntegrationID: cmd.IntegrationID.String(),
			Status:        integ.Status().String(),
		},
	}, nil
}

// ============================================================================
// RefreshCredentials Handler
// ============================================================================

// HandleRefreshCredentials handles the RefreshCredentials command.
func (h *Handlers) HandleRefreshCredentials(ctx context.Context, cmd RefreshCredentials) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRefreshCredentials"

	// 1. Load aggregate
	integID := domain.IntegrationIDFromTypesID(cmd.IntegrationID)
	integ, err := h.integrationRepo.Get(ctx, integID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Refresh credentials
	if err := integ.RefreshCredentials(cmd.Scopes); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.integrationRepo.Save(ctx, integ); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, integ.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.IntegrationID.String(),
		Version: integ.Version(),
		Data: RefreshCredentialsResult{
			IntegrationID: cmd.IntegrationID.String(),
			Status:        integ.Status().String(),
		},
	}, nil
}

// ============================================================================
// SuspendIntegration Handler
// ============================================================================

// HandleSuspendIntegration handles the SuspendIntegration command.
func (h *Handlers) HandleSuspendIntegration(ctx context.Context, cmd SuspendIntegration) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleSuspendIntegration"

	// 1. Load aggregate
	integID := domain.IntegrationIDFromTypesID(cmd.IntegrationID)
	integ, err := h.integrationRepo.Get(ctx, integID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Suspend
	if err := integ.Suspend(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.integrationRepo.Save(ctx, integ); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, integ.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.IntegrationID.String(),
		Version: integ.Version(),
		Data: SuspendIntegrationResult{
			IntegrationID: cmd.IntegrationID.String(),
			Status:        integ.Status().String(),
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
