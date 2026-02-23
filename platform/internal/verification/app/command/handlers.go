package command

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

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
	verificationRepo   domain.VerificationRepository
	credentialIssuer   domain.CredentialIssuer
	dataFetcher        domain.DataFetcher
	oauthURLGenerator  domain.OAuthURLGenerator
	oauthCodeExchanger domain.OAuthCodeExchanger
	publisher          domain.EventPublisher
	logger             log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	verificationRepo domain.VerificationRepository,
	credentialIssuer domain.CredentialIssuer,
	dataFetcher domain.DataFetcher,
	oauthURLGenerator domain.OAuthURLGenerator,
	oauthCodeExchanger domain.OAuthCodeExchanger,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		verificationRepo:   verificationRepo,
		credentialIssuer:   credentialIssuer,
		dataFetcher:        dataFetcher,
		oauthURLGenerator:  oauthURLGenerator,
		oauthCodeExchanger: oauthCodeExchanger,
		publisher:          publisher,
		logger:             logger,
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

	// 6. Generate OAuth authorization URL (non-fatal on failure)
	var authURL string
	if cmd.RedirectURI != "" {
		url, err := h.oauthURLGenerator.GenerateAuthURL(ctx, cmd.ProviderType, oauthState, cmd.RedirectURI)
		if err != nil {
			h.logger.Error("failed to generate auth URL",
				log.String("op", op),
				log.String("provider", cmd.ProviderType.String()),
				log.Err(err),
			)
		} else {
			authURL = url
		}
	}

	return &cqrs.CommandResult{
		ID:      verificationID.String(),
		Version: v.Version(),
		Data: StartVerificationResult{
			VerificationID: verificationID.String(),
			ProviderType:   cmd.ProviderType.String(),
			Status:         v.Status().String(),
			OAuthState:     oauthState,
			AuthURL:        authURL,
		},
	}, nil
}

// ============================================================================
// ReceiveOAuthCallback Handler
// ============================================================================

// HandleReceiveOAuthCallback handles the ReceiveOAuthCallback command.
// Orchestrates the full verification flow: validate state → record callback →
// exchange code → fetch data → issue credential.
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

	// 3. Validate OAuth state (CSRF protection)
	if v.OAuthState() != cmd.State {
		if err := v.Fail("INVALID_STATE", "OAuth state mismatch"); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if err := h.verificationRepo.Save(ctx, v); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		h.publishEvents(ctx, op, v.Changes()...)
		return nil, pkgerrors.Validation(op, "OAuth state mismatch")
	}

	// 4. Record OAuth callback received → OAuthCompleted
	if err := v.RecordOAuthCallback(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Exchange code for access token
	accessToken, err := h.oauthCodeExchanger.ExchangeCode(ctx, v.Provider(), cmd.Code, cmd.RedirectURI)
	if err != nil {
		h.logger.Error("code exchange failed", log.String("op", op), log.Err(err))
		if dfErr := v.RecordDataFetchFailed("CODE_EXCHANGE_FAILED", err.Error()); dfErr != nil {
			return nil, pkgerrors.Wrap(dfErr, op)
		}
		if err := h.verificationRepo.Save(ctx, v); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		h.publishEvents(ctx, op, v.Changes()...)
		return nil, pkgerrors.Infrastructure(op, fmt.Errorf("code exchange failed: %w", err))
	}

	// 6. Fetch provider data
	data, err := h.dataFetcher.FetchData(ctx, v.Provider(), accessToken)
	if err != nil {
		h.logger.Error("data fetch failed", log.String("op", op), log.Err(err))
		if dfErr := v.RecordDataFetchFailed("DATA_FETCH_FAILED", err.Error()); dfErr != nil {
			return nil, pkgerrors.Wrap(dfErr, op)
		}
		if err := h.verificationRepo.Save(ctx, v); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		h.publishEvents(ctx, op, v.Changes()...)
		return nil, pkgerrors.Infrastructure(op, fmt.Errorf("data fetch failed: %w", err))
	}

	// 7. Record data fetched → DataFetched
	if err := v.RecordDataFetched(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 8. Issue credential
	credentialID, err := h.credentialIssuer.IssueCredential(ctx, v.UserID(), v.Provider(), data)
	if err != nil {
		h.logger.Error("credential issuance failed", log.String("op", op), log.Err(err))
		if fErr := v.Fail("CREDENTIAL_ISSUANCE_FAILED", err.Error()); fErr != nil {
			return nil, pkgerrors.Wrap(fErr, op)
		}
		if err := h.verificationRepo.Save(ctx, v); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		h.publishEvents(ctx, op, v.Changes()...)
		return nil, pkgerrors.Infrastructure(op, fmt.Errorf("credential issuance failed: %w", err))
	}

	// 9. Record credential issued → CredentialIssued (terminal success)
	if err := v.RecordCredentialIssued(credentialID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 10. Save all events atomically
	if err := h.verificationRepo.Save(ctx, v); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 11. Publish events
	h.publishEvents(ctx, op, v.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: v.Version(),
		Data: ReceiveOAuthCallbackResult{
			VerificationID: cmd.VerificationID,
			Status:         v.Status().String(),
			CredentialID:   credentialID,
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
