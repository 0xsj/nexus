package command

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Verification context.
type Handlers struct {
	verificationRepo domain.VerificationRepository
	publisher        domain.EventPublisher
	logger           log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	verificationRepo domain.VerificationRepository,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		verificationRepo: verificationRepo,
		publisher:        publisher,
		logger:           logger,
	}
}

// ============================================================================
// StartVerification Handler
// ============================================================================

// HandleStartVerification handles the StartVerification command.
func (h *Handlers) HandleStartVerification(ctx context.Context, cmd StartVerification) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleStartVerification"

	// 1. Generate verification ID
	verificationID := domain.NewVerificationID()

	// 2. Generate OAuth state token
	oauthState, err := generateOAuthState()
	if err != nil {
		return nil, pkgerrors.Infrastructure(op, err)
	}

	// 3. Create aggregate
	v, err := domain.StartVerification(verificationID, cmd.UserID, cmd.ProviderType, oauthState)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.verificationRepo.Save(ctx, v); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, v.Changes()...)

	return &cqrs.CommandResult{
		ID:      verificationID.String(),
		Version: v.Version(),
		Data: StartVerificationResult{
			VerificationID: verificationID.String(),
			ProviderType:   cmd.ProviderType.String(),
			Status:         v.Status().String(),
			OAuthState:     oauthState,
		},
	}, nil
}

// ============================================================================
// ReceiveOAuthCallback Handler
// ============================================================================

// HandleReceiveOAuthCallback handles the ReceiveOAuthCallback command.
func (h *Handlers) HandleReceiveOAuthCallback(ctx context.Context, cmd ReceiveOAuthCallback) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleReceiveOAuthCallback"

	// 1. Parse verification ID
	verificationID, err := domain.ParseVerificationID(cmd.VerificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Load aggregate
	v, err := h.verificationRepo.Get(ctx, verificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Record callback
	if err := v.RecordOAuthCallback(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.verificationRepo.Save(ctx, v); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, v.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: v.Version(),
		Data: ReceiveOAuthCallbackResult{
			VerificationID: cmd.VerificationID,
			Status:         v.Status().String(),
		},
	}, nil
}

// ============================================================================
// CompleteVerification Handler
// ============================================================================

// HandleCompleteVerification handles the CompleteVerification command.
func (h *Handlers) HandleCompleteVerification(ctx context.Context, cmd CompleteVerification) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCompleteVerification"

	// 1. Parse verification ID
	verificationID, err := domain.ParseVerificationID(cmd.VerificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Load aggregate
	v, err := h.verificationRepo.Get(ctx, verificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Record data fetched (transition to DataFetched)
	if err := v.RecordDataFetched(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Record credential issued (transition to CredentialIssued)
	if err := v.RecordCredentialIssued(cmd.CredentialID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save
	if err := h.verificationRepo.Save(ctx, v); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, v.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: v.Version(),
		Data: CompleteVerificationResult{
			VerificationID: cmd.VerificationID,
			CredentialID:   cmd.CredentialID,
			Status:         v.Status().String(),
		},
	}, nil
}

// ============================================================================
// FailVerification Handler
// ============================================================================

// HandleFailVerification handles the FailVerification command.
func (h *Handlers) HandleFailVerification(ctx context.Context, cmd FailVerification) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleFailVerification"

	// 1. Parse verification ID
	verificationID, err := domain.ParseVerificationID(cmd.VerificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Load aggregate
	v, err := h.verificationRepo.Get(ctx, verificationID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Fail
	if err := v.Fail(cmd.ErrorCode, cmd.ErrorMessage); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.verificationRepo.Save(ctx, v); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, v.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: v.Version(),
		Data: FailVerificationResult{
			VerificationID: cmd.VerificationID,
			Status:         v.Status().String(),
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

// generateOAuthState generates a random OAuth state token.
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
