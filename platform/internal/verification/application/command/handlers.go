package command

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/id"
)

// ============================================================================
// Handler Dependencies
// ============================================================================

// HandlerDependencies contains all dependencies for command handlers.
type HandlerDependencies struct {
	VerificationRepo  domain.VerificationRepository
	OAuthStateRepo    domain.OAuthStateRepository
	ProviderService   domain.ProviderService
	CredentialIssuer  domain.CredentialIssuerService
	OAuthStateService domain.OAuthStateService
	UserDIDResolver   domain.UserDIDResolver
	IDGenerator       id.Generator
	VerificationTTL   time.Duration
}

// DefaultVerificationTTL is the default TTL for verification flows.
const DefaultVerificationTTL = 15 * time.Minute

// ============================================================================
// Initiate Verification Handler
// ============================================================================

// InitiateVerificationHandler handles InitiateVerification commands.
type InitiateVerificationHandler struct {
	verificationRepo  domain.VerificationRepository
	providerService   domain.ProviderService
	oauthStateService domain.OAuthStateService
	idGenerator       id.Generator
	verificationTTL   time.Duration
}

// NewInitiateVerificationHandler creates a new InitiateVerificationHandler.
func NewInitiateVerificationHandler(
	verificationRepo domain.VerificationRepository,
	providerService domain.ProviderService,
	oauthStateService domain.OAuthStateService,
	idGenerator id.Generator,
	verificationTTL time.Duration,
) *InitiateVerificationHandler {
	if verificationTTL == 0 {
		verificationTTL = DefaultVerificationTTL
	}
	return &InitiateVerificationHandler{
		verificationRepo:  verificationRepo,
		providerService:   providerService,
		oauthStateService: oauthStateService,
		idGenerator:       idGenerator,
		verificationTTL:   verificationTTL,
	}
}

// Handle handles the InitiateVerification command.
func (h *InitiateVerificationHandler) Handle(ctx context.Context, cmd *InitiateVerification) (*cqrs.CommandResult, error) {
	const op = "InitiateVerificationHandler.Handle"

	// Validate provider is supported
	if !h.providerService.SupportsProvider(cmd.Provider) {
		return nil, domain.ErrProviderNotSupported(op, cmd.Provider.String())
	}

	// 1. Create verification aggregate with temporary OAuth state
	verification := domain.NewVerification(cmd.VerificationID)

	// Use verification ID as temporary placeholder for OAuth state
	// This will be updated after we generate the real state
	tempState := cmd.VerificationID

	if err := verification.Initiate(
		cmd.UserID,
		cmd.Provider,
		cmd.CredentialType,
		tempState,
		cmd.RedirectURL,
		h.verificationTTL,
	); err != nil {
		return nil, err
	}

	// 2. Save verification FIRST (FK constraint requires verification to exist)
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	// 3. Generate OAuth state (now that verification exists)
	oauthState, err := h.oauthStateService.GenerateState(ctx, domain.GenerateStateParams{
		UserID:         cmd.UserID,
		Provider:       cmd.Provider,
		CredentialType: cmd.CredentialType,
		RedirectURL:    cmd.RedirectURL,
		TTL:            int64(h.verificationTTL.Seconds()),
		VerificationID: cmd.VerificationID,
	})
	if err != nil {
		// Cleanup: delete the verification we just created
		_ = h.verificationRepo.Delete(ctx, cmd.VerificationID)
		return nil, err
	}

	// 4. Update verification with the real OAuth state
	if err := verification.UpdateOAuthState(oauthState.Value); err != nil {
		_ = h.verificationRepo.Delete(ctx, cmd.VerificationID)
		return nil, err
	}

	// 5. Save the updated verification
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	// 6. Get authorization URL from provider
	authURL, err := h.providerService.GetAuthorizationURL(ctx, domain.AuthorizationURLParams{
		Provider:    cmd.Provider,
		State:       oauthState.Value,
		RedirectURI: cmd.RedirectURL,
		Scopes:      nil, // Provider will use defaults for credential type
	})
	if err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: verification.Version(),
		Data: &InitiateVerificationResult{
			VerificationID:   cmd.VerificationID,
			AuthorizationURL: authURL,
			OAuthState:       oauthState.Value,
			ExpiresIn:        int64(h.verificationTTL.Seconds()),
		},
	}, nil
}

// ============================================================================
// Handle OAuth Callback Handler
// ============================================================================

// HandleOAuthCallbackHandler handles HandleOAuthCallback commands.
type HandleOAuthCallbackHandler struct {
	verificationRepo  domain.VerificationRepository
	oauthStateService domain.OAuthStateService
	providerService   domain.ProviderService
}

// NewHandleOAuthCallbackHandler creates a new HandleOAuthCallbackHandler.
func NewHandleOAuthCallbackHandler(
	verificationRepo domain.VerificationRepository,
	oauthStateService domain.OAuthStateService,
	providerService domain.ProviderService,
) *HandleOAuthCallbackHandler {
	return &HandleOAuthCallbackHandler{
		verificationRepo:  verificationRepo,
		oauthStateService: oauthStateService,
		providerService:   providerService,
	}
}

// Handle handles the HandleOAuthCallback command.
func (h *HandleOAuthCallbackHandler) Handle(ctx context.Context, cmd *HandleOAuthCallback) (*cqrs.CommandResult, error) {
	const op = "HandleOAuthCallbackHandler.Handle"

	// Handle OAuth error from provider
	if cmd.Error != "" {
		// Try to find verification by state to fail it
		oauthState, err := h.oauthStateService.ConsumeState(ctx, cmd.State)
		if err == nil {
			verification, verr := h.verificationRepo.FindByOAuthState(ctx, oauthState.Value)
			if verr == nil {
				_ = verification.Fail("OAuth error: "+cmd.Error, cmd.Error)
				_ = h.verificationRepo.Save(ctx, verification)
			}
		}
		return nil, domain.ErrProviderAuthFailed(op, "unknown", cmd.Error)
	}

	// Validate and consume OAuth state
	oauthState, err := h.oauthStateService.ConsumeState(ctx, cmd.State)
	if err != nil {
		return nil, domain.ErrOAuthStateMismatch(op)
	}

	// Find verification by OAuth state
	verification, err := h.verificationRepo.FindByOAuthState(ctx, oauthState.Value)
	if err != nil {
		return nil, err
	}

	// Exchange code for tokens
	tokens, err := h.providerService.ExchangeCode(ctx, domain.ExchangeCodeParams{
		Provider:    oauthState.Provider,
		Code:        cmd.Code,
		RedirectURI: oauthState.RedirectURL,
	})
	if err != nil {
		// Mark verification as failed
		_ = verification.Fail("Failed to exchange authorization code", domain.CodeOAuthCodeInvalid.String())
		_ = h.verificationRepo.Save(ctx, verification)
		return nil, domain.ErrOAuthCodeInvalid(op, oauthState.Provider.String())
	}

	// Mark as authorized
	if err := verification.Authorize(cmd.Code); err != nil {
		return nil, err
	}

	// Start fetching (hash the token for audit purposes)
	tokenHash := hashToken(tokens.AccessToken)
	if err := verification.StartFetching(tokenHash); err != nil {
		return nil, err
	}

	// Persist
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      verification.AggregateID(),
		Version: verification.Version(),
		Data: &HandleOAuthCallbackResult{
			VerificationID: verification.AggregateID(),
			UserID:         oauthState.UserID,
			Provider:       oauthState.Provider.String(),
			RedirectURL:    oauthState.RedirectURL,
			Success:        true,
		},
	}, nil
}

// ============================================================================
// Complete Verification Handler
// ============================================================================

// CompleteVerificationHandler handles CompleteVerification commands.
type CompleteVerificationHandler struct {
	verificationRepo domain.VerificationRepository
	credentialIssuer domain.CredentialIssuerService
	userDIDResolver  domain.UserDIDResolver
	idGenerator      id.Generator
}

// NewCompleteVerificationHandler creates a new CompleteVerificationHandler.
func NewCompleteVerificationHandler(
	verificationRepo domain.VerificationRepository,
	credentialIssuer domain.CredentialIssuerService,
	userDIDResolver domain.UserDIDResolver,
	idGenerator id.Generator,
) *CompleteVerificationHandler {
	return &CompleteVerificationHandler{
		verificationRepo: verificationRepo,
		credentialIssuer: credentialIssuer,
		userDIDResolver:  userDIDResolver,
		idGenerator:      idGenerator,
	}
}

// Handle handles the CompleteVerification command.
func (h *CompleteVerificationHandler) Handle(ctx context.Context, cmd *CompleteVerification) (*cqrs.CommandResult, error) {
	const op = "CompleteVerificationHandler.Handle"

	// Load verification
	verification, err := h.verificationRepo.FindByID(ctx, cmd.VerificationID)
	if err != nil {
		return nil, err
	}

	// Resolve user's DID for credential issuance
	holderDID, err := h.userDIDResolver.ResolvePrimaryDID(ctx, verification.UserID())
	if err != nil {
		_ = verification.Fail("Failed to resolve user DID", "DID_RESOLUTION_FAILED")
		_ = h.verificationRepo.Save(ctx, verification)
		return nil, err
	}

	// Issue credential
	issuedCredential, err := h.credentialIssuer.IssueCredential(ctx, domain.IssueCredentialParams{
		HolderDID:      holderDID,
		CredentialType: verification.CredentialType(),
		Claims:         cmd.Claims,
		Evidence: []domain.CredentialEvidence{
			{
				Type:       "ProviderVerification",
				Source:     verification.Provider().String(),
				VerifiedAt: time.Now().UTC().Format(time.RFC3339),
				Data: map[string]any{
					"provider_user_id": cmd.ProviderUserID,
					"username":         cmd.Username,
				},
			},
		},
	})
	if err != nil {
		_ = verification.Fail("Failed to issue credential: "+err.Error(), domain.CodeCredentialIssuanceFailed.String())
		_ = h.verificationRepo.Save(ctx, verification)
		return nil, domain.ErrCredentialIssuanceFailed(op, "issuance failed", err)
	}

	// Complete the verification
	if err := verification.Complete(
		cmd.ProviderUserID,
		cmd.Username,
		issuedCredential.ID,
		cmd.Claims,
	); err != nil {
		return nil, err
	}

	// Persist
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: verification.Version(),
		Data: &CompleteVerificationResult{
			VerificationID: cmd.VerificationID,
			CredentialID:   issuedCredential.ID,
			CredentialType: string(verification.CredentialType()),
			Provider:       verification.Provider().String(),
			SignedVC:       issuedCredential.SignedVC,
		},
	}, nil
}

// ============================================================================
// Fail Verification Handler
// ============================================================================

// FailVerificationHandler handles FailVerification commands.
type FailVerificationHandler struct {
	verificationRepo domain.VerificationRepository
}

// NewFailVerificationHandler creates a new FailVerificationHandler.
func NewFailVerificationHandler(verificationRepo domain.VerificationRepository) *FailVerificationHandler {
	return &FailVerificationHandler{
		verificationRepo: verificationRepo,
	}
}

// Handle handles the FailVerification command.
func (h *FailVerificationHandler) Handle(ctx context.Context, cmd *FailVerification) (*cqrs.CommandResult, error) {
	// Load verification
	verification, err := h.verificationRepo.FindByID(ctx, cmd.VerificationID)
	if err != nil {
		return nil, err
	}

	// Execute domain logic
	if err := verification.Fail(cmd.Reason, cmd.ErrorCode); err != nil {
		return nil, err
	}

	// Persist
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: verification.Version(),
		Data: &FailVerificationResult{
			VerificationID: cmd.VerificationID,
			Failed:         true,
		},
	}, nil
}

// ============================================================================
// Expire Verification Handler
// ============================================================================

// ExpireVerificationHandler handles ExpireVerification commands.
type ExpireVerificationHandler struct {
	verificationRepo domain.VerificationRepository
}

// NewExpireVerificationHandler creates a new ExpireVerificationHandler.
func NewExpireVerificationHandler(verificationRepo domain.VerificationRepository) *ExpireVerificationHandler {
	return &ExpireVerificationHandler{
		verificationRepo: verificationRepo,
	}
}

// Handle handles the ExpireVerification command.
func (h *ExpireVerificationHandler) Handle(ctx context.Context, cmd *ExpireVerification) (*cqrs.CommandResult, error) {
	// Load verification
	verification, err := h.verificationRepo.FindByID(ctx, cmd.VerificationID)
	if err != nil {
		return nil, err
	}

	// Execute domain logic
	if err := verification.Expire(); err != nil {
		return nil, err
	}

	// Persist
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: verification.Version(),
		Data: &ExpireVerificationResult{
			VerificationID: cmd.VerificationID,
			Expired:        true,
		},
	}, nil
}

// ============================================================================
// Cancel Verification Handler
// ============================================================================

// CancelVerificationHandler handles CancelVerification commands.
type CancelVerificationHandler struct {
	verificationRepo  domain.VerificationRepository
	oauthStateService domain.OAuthStateService
}

// NewCancelVerificationHandler creates a new CancelVerificationHandler.
func NewCancelVerificationHandler(
	verificationRepo domain.VerificationRepository,
	oauthStateService domain.OAuthStateService,
) *CancelVerificationHandler {
	return &CancelVerificationHandler{
		verificationRepo:  verificationRepo,
		oauthStateService: oauthStateService,
	}
}

// Handle handles the CancelVerification command.
func (h *CancelVerificationHandler) Handle(ctx context.Context, cmd *CancelVerification) (*cqrs.CommandResult, error) {
	const op = "CancelVerificationHandler.Handle"

	// Load verification
	verification, err := h.verificationRepo.FindByID(ctx, cmd.VerificationID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if verification.UserID() != cmd.UserID {
		return nil, domain.ErrVerificationNotFound(op, cmd.VerificationID)
	}

	// Invalidate OAuth state if still valid
	if verification.OAuthState() != "" {
		_ = h.oauthStateService.InvalidateState(ctx, verification.OAuthState())
	}

	// Mark as failed with cancellation reason
	reason := cmd.Reason
	if reason == "" {
		reason = "Cancelled by user"
	}
	if err := verification.Fail(reason, "CANCELLED"); err != nil {
		return nil, err
	}

	// Persist
	if err := h.verificationRepo.Save(ctx, verification); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:      cmd.VerificationID,
		Version: verification.Version(),
		Data: &CancelVerificationResult{
			VerificationID: cmd.VerificationID,
			Cancelled:      true,
		},
	}, nil
}

// ============================================================================
// Handler Registration
// ============================================================================

// RegisterHandlers registers all verification command handlers with the command bus.
func RegisterHandlers(bus *cqrs.InMemoryCommandBus, deps HandlerDependencies) error {
	ttl := deps.VerificationTTL
	if ttl == 0 {
		ttl = DefaultVerificationTTL
	}

	handlers := map[string]any{
		TypeInitiateVerification: NewInitiateVerificationHandler(
			deps.VerificationRepo,
			deps.ProviderService,
			deps.OAuthStateService,
			deps.IDGenerator,
			ttl,
		),
		TypeHandleOAuthCallback: NewHandleOAuthCallbackHandler(
			deps.VerificationRepo,
			deps.OAuthStateService,
			deps.ProviderService,
		),
		TypeCompleteVerification: NewCompleteVerificationHandler(
			deps.VerificationRepo,
			deps.CredentialIssuer,
			deps.UserDIDResolver,
			deps.IDGenerator,
		),
		TypeFailVerification:   NewFailVerificationHandler(deps.VerificationRepo),
		TypeExpireVerification: NewExpireVerificationHandler(deps.VerificationRepo),
		TypeCancelVerification: NewCancelVerificationHandler(
			deps.VerificationRepo,
			deps.OAuthStateService,
		),
	}

	for cmdType, handler := range handlers {
		if err := bus.Register(cmdType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// hashToken creates a hash of a token for audit purposes.
func hashToken(token string) string {
	if len(token) < 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.CommandHandler[*InitiateVerification] = (*InitiateVerificationHandler)(nil)
	_ cqrs.CommandHandler[*HandleOAuthCallback]  = (*HandleOAuthCallbackHandler)(nil)
	_ cqrs.CommandHandler[*CompleteVerification] = (*CompleteVerificationHandler)(nil)
	_ cqrs.CommandHandler[*FailVerification]     = (*FailVerificationHandler)(nil)
	_ cqrs.CommandHandler[*ExpireVerification]   = (*ExpireVerificationHandler)(nil)
	_ cqrs.CommandHandler[*CancelVerification]   = (*CancelVerificationHandler)(nil)
)
