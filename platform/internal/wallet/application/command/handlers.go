package command

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/id"
)

// ============================================================================
// Link Wallet Handler
// ============================================================================

// LinkWalletHandler handles LinkWallet commands.
type LinkWalletHandler struct {
	walletRepo        domain.WalletRepository
	challengeRepo     domain.ChallengeRepository
	signatureVerifier domain.SignatureVerificationService
	didDerivation     domain.DIDDerivationService
	idGenerator       id.Generator
}

// NewLinkWalletHandler creates a new LinkWalletHandler.
func NewLinkWalletHandler(
	walletRepo domain.WalletRepository,
	challengeRepo domain.ChallengeRepository,
	signatureVerifier domain.SignatureVerificationService,
	didDerivation domain.DIDDerivationService,
	idGenerator id.Generator,
) *LinkWalletHandler {
	return &LinkWalletHandler{
		walletRepo:        walletRepo,
		challengeRepo:     challengeRepo,
		signatureVerifier: signatureVerifier,
		didDerivation:     didDerivation,
		idGenerator:       idGenerator,
	}
}

// Handle handles the LinkWallet command.
func (h *LinkWalletHandler) Handle(ctx context.Context, cmd *LinkWallet) (*cqrs.CommandResult, error) {
	const op = "LinkWalletHandler.Handle"

	// Validate and consume challenge
	challenge, err := h.challengeRepo.FindByNonce(ctx, cmd.Nonce)
	if err != nil {
		return nil, err
	}

	if !challenge.IsValid() {
		if challenge.IsExpired() {
			return nil, domain.ErrSignatureExpired(op)
		}
		return nil, domain.ErrSignatureUsed(op, cmd.Nonce)
	}

	// Create address value object
	address, err := domain.NewAddress(cmd.Address, cmd.ChainID)
	if err != nil {
		return nil, err
	}

	// Verify signature
	signature, err := domain.NewSignature(cmd.Signature, domain.AlgorithmForFamily(address.Family()))
	if err != nil {
		return nil, err
	}

	err = h.signatureVerifier.VerifySignature(ctx, domain.VerifySignatureParams{
		Address:   address,
		Message:   cmd.Message,
		Signature: signature,
	})
	if err != nil {
		return nil, domain.ErrVerificationFailed(op, err.Error())
	}

	// Check wallet doesn't already exist
	exists, err := h.walletRepo.ExistsByAddress(ctx, address)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrWalletAlreadyExists(op, address.Normalized(), cmd.ChainID.String())
	}

	// Derive DID
	walletDID, err := h.didDerivation.DeriveDID(ctx, address)
	if err != nil {
		return nil, err
	}

	// Create wallet
	walletID := h.idGenerator.Generate().String()
	wallet, err := domain.NewWallet(walletID, cmd.UserID, address, walletDID)
	if err != nil {
		return nil, err
	}

	// Set label if provided
	if cmd.Label != "" {
		wallet.SetLabel(cmd.Label)
	}

	// Check if this is the user's first wallet (make it primary)
	existingWallets, err := h.walletRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if len(existingWallets) == 0 {
		wallet.SetPrimary(true)
	}

	// Save wallet
	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	// Mark challenge as used
	_ = h.challengeRepo.MarkUsed(ctx, cmd.Nonce)

	return &cqrs.CommandResult{
		ID: wallet.ID(),
		Data: &LinkWalletResult{
			WalletID:  wallet.ID(),
			UserID:    wallet.UserID(),
			Address:   wallet.AddressString(),
			ChainID:   wallet.ChainID().String(),
			DID:       wallet.DIDString(),
			IsPrimary: wallet.IsPrimary(),
			CreatedAt: wallet.CreatedAt(),
		},
	}, nil
}

// ============================================================================
// Unlink Wallet Handler
// ============================================================================

// UnlinkWalletHandler handles UnlinkWallet commands.
type UnlinkWalletHandler struct {
	walletRepo domain.WalletRepository
}

// NewUnlinkWalletHandler creates a new UnlinkWalletHandler.
func NewUnlinkWalletHandler(walletRepo domain.WalletRepository) *UnlinkWalletHandler {
	return &UnlinkWalletHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the UnlinkWallet command.
func (h *UnlinkWalletHandler) Handle(ctx context.Context, cmd *UnlinkWallet) (*cqrs.CommandResult, error) {
	const op = "UnlinkWalletHandler.Handle"

	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if wallet.UserID() != cmd.UserID {
		return nil, domain.ErrWalletNotFound(op, cmd.WalletID)
	}

	// Store info for result before deletion
	result := &UnlinkWalletResult{
		WalletID: wallet.ID(),
		UserID:   wallet.UserID(),
		Address:  wallet.AddressString(),
		ChainID:  wallet.ChainID().String(),
	}

	// Delete wallet
	if err := h.walletRepo.Delete(ctx, cmd.WalletID); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID:   cmd.WalletID,
		Data: result,
	}, nil
}

// ============================================================================
// Update Label Handler
// ============================================================================

// UpdateLabelHandler handles UpdateLabel commands.
type UpdateLabelHandler struct {
	walletRepo domain.WalletRepository
}

// NewUpdateLabelHandler creates a new UpdateLabelHandler.
func NewUpdateLabelHandler(walletRepo domain.WalletRepository) *UpdateLabelHandler {
	return &UpdateLabelHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the UpdateLabel command.
func (h *UpdateLabelHandler) Handle(ctx context.Context, cmd *UpdateLabel) (*cqrs.CommandResult, error) {
	const op = "UpdateLabelHandler.Handle"

	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if wallet.UserID() != cmd.UserID {
		return nil, domain.ErrWalletNotFound(op, cmd.WalletID)
	}

	oldLabel := wallet.Label()
	wallet.SetLabel(cmd.Label)

	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.WalletID,
		Data: &UpdateLabelResult{
			WalletID: wallet.ID(),
			OldLabel: oldLabel,
			NewLabel: cmd.Label,
		},
	}, nil
}

// ============================================================================
// Set Primary Handler
// ============================================================================

// SetPrimaryHandler handles SetPrimary commands.
type SetPrimaryHandler struct {
	walletRepo domain.WalletRepository
}

// NewSetPrimaryHandler creates a new SetPrimaryHandler.
func NewSetPrimaryHandler(walletRepo domain.WalletRepository) *SetPrimaryHandler {
	return &SetPrimaryHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the SetPrimary command.
func (h *SetPrimaryHandler) Handle(ctx context.Context, cmd *SetPrimary) (*cqrs.CommandResult, error) {
	const op = "SetPrimaryHandler.Handle"

	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if wallet.UserID() != cmd.UserID {
		return nil, domain.ErrWalletNotFound(op, cmd.WalletID)
	}

	// Find current primary wallet
	var previousPrimaryID string
	currentPrimary, err := h.walletRepo.FindPrimaryByUserID(ctx, cmd.UserID)
	if err == nil && currentPrimary != nil {
		previousPrimaryID = currentPrimary.ID()

		// Unset current primary
		if currentPrimary.ID() != wallet.ID() {
			currentPrimary.SetPrimary(false)
			if err := h.walletRepo.Save(ctx, currentPrimary); err != nil {
				return nil, err
			}
		}
	}

	// Set new primary
	wallet.SetPrimary(true)
	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.WalletID,
		Data: &SetPrimaryResult{
			WalletID:          wallet.ID(),
			PreviousPrimaryID: previousPrimaryID,
		},
	}, nil
}

// ============================================================================
// Activate Wallet Handler
// ============================================================================

// ActivateWalletHandler handles ActivateWallet commands.
type ActivateWalletHandler struct {
	walletRepo domain.WalletRepository
}

// NewActivateWalletHandler creates a new ActivateWalletHandler.
func NewActivateWalletHandler(walletRepo domain.WalletRepository) *ActivateWalletHandler {
	return &ActivateWalletHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the ActivateWallet command.
func (h *ActivateWalletHandler) Handle(ctx context.Context, cmd *ActivateWallet) (*cqrs.CommandResult, error) {
	const op = "ActivateWalletHandler.Handle"

	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if wallet.UserID() != cmd.UserID {
		return nil, domain.ErrWalletNotFound(op, cmd.WalletID)
	}

	previousStatus := wallet.Status()

	if err := wallet.Activate(); err != nil {
		return nil, err
	}

	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.WalletID,
		Data: &ActivateWalletResult{
			WalletID:       wallet.ID(),
			PreviousStatus: previousStatus.String(),
			CurrentStatus:  wallet.Status().String(),
		},
	}, nil
}

// ============================================================================
// Deactivate Wallet Handler
// ============================================================================

// DeactivateWalletHandler handles DeactivateWallet commands.
type DeactivateWalletHandler struct {
	walletRepo domain.WalletRepository
}

// NewDeactivateWalletHandler creates a new DeactivateWalletHandler.
func NewDeactivateWalletHandler(walletRepo domain.WalletRepository) *DeactivateWalletHandler {
	return &DeactivateWalletHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the DeactivateWallet command.
func (h *DeactivateWalletHandler) Handle(ctx context.Context, cmd *DeactivateWallet) (*cqrs.CommandResult, error) {
	const op = "DeactivateWalletHandler.Handle"

	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if wallet.UserID() != cmd.UserID {
		return nil, domain.ErrWalletNotFound(op, cmd.WalletID)
	}

	previousStatus := wallet.Status()

	if err := wallet.Deactivate(); err != nil {
		return nil, err
	}

	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.WalletID,
		Data: &DeactivateWalletResult{
			WalletID:       wallet.ID(),
			PreviousStatus: previousStatus.String(),
			CurrentStatus:  wallet.Status().String(),
		},
	}, nil
}

// ============================================================================
// Suspend Wallet Handler
// ============================================================================

// SuspendWalletHandler handles SuspendWallet commands.
type SuspendWalletHandler struct {
	walletRepo domain.WalletRepository
}

// NewSuspendWalletHandler creates a new SuspendWalletHandler.
func NewSuspendWalletHandler(walletRepo domain.WalletRepository) *SuspendWalletHandler {
	return &SuspendWalletHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the SuspendWallet command.
func (h *SuspendWalletHandler) Handle(ctx context.Context, cmd *SuspendWallet) (*cqrs.CommandResult, error) {
	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	previousStatus := wallet.Status()

	if err := wallet.Suspend(); err != nil {
		return nil, err
	}

	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.WalletID,
		Data: &SuspendWalletResult{
			WalletID:       wallet.ID(),
			PreviousStatus: previousStatus.String(),
			CurrentStatus:  wallet.Status().String(),
			Reason:         cmd.Reason,
		},
	}, nil
}

// ============================================================================
// Create Challenge Handler
// ============================================================================

// CreateChallengeHandler handles CreateChallenge commands.
type CreateChallengeHandler struct {
	challengeService domain.ChallengeService
	addressValidator domain.AddressValidationService
}

// NewCreateChallengeHandler creates a new CreateChallengeHandler.
func NewCreateChallengeHandler(
	challengeService domain.ChallengeService,
	addressValidator domain.AddressValidationService,
) *CreateChallengeHandler {
	return &CreateChallengeHandler{
		challengeService: challengeService,
		addressValidator: addressValidator,
	}
}

// Handle handles the CreateChallenge command.
func (h *CreateChallengeHandler) Handle(ctx context.Context, cmd *CreateChallenge) (*cqrs.CommandResult, error) {
	// Validate address
	address, err := h.addressValidator.Validate(ctx, cmd.Address, cmd.ChainID)
	if err != nil {
		return nil, err
	}

	// Create challenge
	challenge, err := h.challengeService.CreateChallenge(ctx, domain.CreateChallengeParams{
		Address:   address,
		Domain:    cmd.Domain,
		URI:       cmd.URI,
		Statement: cmd.Statement,
		Resources: cmd.Resources,
	})
	if err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: challenge.Nonce,
		Data: &CreateChallengeResult{
			Nonce:     challenge.Nonce,
			Message:   challenge.Message,
			Domain:    challenge.Domain,
			URI:       challenge.URI,
			IssuedAt:  challenge.IssuedAt,
			ExpiresAt: challenge.ExpiresAt,
		},
	}, nil
}

// ============================================================================
// Verify Challenge Handler
// ============================================================================

// VerifyChallengeHandler handles VerifyChallenge commands.
type VerifyChallengeHandler struct {
	challengeRepo     domain.ChallengeRepository
	signatureVerifier domain.SignatureVerificationService
}

// NewVerifyChallengeHandler creates a new VerifyChallengeHandler.
func NewVerifyChallengeHandler(
	challengeRepo domain.ChallengeRepository,
	signatureVerifier domain.SignatureVerificationService,
) *VerifyChallengeHandler {
	return &VerifyChallengeHandler{
		challengeRepo:     challengeRepo,
		signatureVerifier: signatureVerifier,
	}
}

// Handle handles the VerifyChallenge command.
func (h *VerifyChallengeHandler) Handle(ctx context.Context, cmd *VerifyChallenge) (*cqrs.CommandResult, error) {
	const op = "VerifyChallengeHandler.Handle"

	// Find challenge
	challenge, err := h.challengeRepo.FindByNonce(ctx, cmd.Nonce)
	if err != nil {
		return nil, err
	}

	// Check challenge is valid
	if !challenge.IsValid() {
		if challenge.IsExpired() {
			return nil, domain.ErrSignatureExpired(op)
		}
		return nil, domain.ErrSignatureUsed(op, cmd.Nonce)
	}

	// Create signature
	signature, err := domain.NewSignature(cmd.Signature, domain.AlgorithmForFamily(challenge.Address.Family()))
	if err != nil {
		return nil, err
	}

	// Verify SIWE signature
	result, err := h.signatureVerifier.VerifySIWE(ctx, domain.VerifySIWEParams{
		Message:         cmd.Message,
		Signature:       signature,
		ExpectedAddress: &challenge.Address,
		ExpectedNonce:   cmd.Nonce,
	})
	if err != nil {
		return nil, domain.ErrVerificationFailed(op, err.Error())
	}

	// Mark challenge as used
	_ = h.challengeRepo.MarkUsed(ctx, cmd.Nonce)

	return &cqrs.CommandResult{
		ID: cmd.Nonce,
		Data: &VerifyChallengeResult{
			Valid:   result.Valid,
			Address: result.Address.Normalized(),
			ChainID: result.Address.ChainID().String(),
			Nonce:   cmd.Nonce,
		},
	}, nil
}

// ============================================================================
// Record Wallet Usage Handler
// ============================================================================

// RecordWalletUsageHandler handles RecordWalletUsage commands.
type RecordWalletUsageHandler struct {
	walletRepo domain.WalletRepository
}

// NewRecordWalletUsageHandler creates a new RecordWalletUsageHandler.
func NewRecordWalletUsageHandler(walletRepo domain.WalletRepository) *RecordWalletUsageHandler {
	return &RecordWalletUsageHandler{
		walletRepo: walletRepo,
	}
}

// Handle handles the RecordWalletUsage command.
func (h *RecordWalletUsageHandler) Handle(ctx context.Context, cmd *RecordWalletUsage) (*cqrs.CommandResult, error) {
	wallet, err := h.walletRepo.FindByID(ctx, cmd.WalletID)
	if err != nil {
		return nil, err
	}

	wallet.RecordUsage()

	if err := h.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.WalletID,
		Data: &RecordWalletUsageResult{
			WalletID:   wallet.ID(),
			Action:     cmd.Action,
			RecordedAt: time.Now(),
		},
	}, nil
}

// ============================================================================
// Handler Dependencies
// ============================================================================

// HandlerDependencies contains all dependencies needed for command handlers.
type HandlerDependencies struct {
	WalletRepo        domain.WalletRepository
	ChallengeRepo     domain.ChallengeRepository
	ChallengeService  domain.ChallengeService
	SignatureVerifier domain.SignatureVerificationService
	AddressValidator  domain.AddressValidationService
	DIDDerivation     domain.DIDDerivationService
	IDGenerator       id.Generator
}

// RegisterHandlers registers all wallet command handlers with the command bus.
func RegisterHandlers(bus *cqrs.InMemoryCommandBus, deps HandlerDependencies) error {
	handlers := map[string]any{
		TypeLinkWallet: NewLinkWalletHandler(
			deps.WalletRepo,
			deps.ChallengeRepo,
			deps.SignatureVerifier,
			deps.DIDDerivation,
			deps.IDGenerator,
		),
		TypeUnlinkWallet:      NewUnlinkWalletHandler(deps.WalletRepo),
		TypeUpdateLabel:       NewUpdateLabelHandler(deps.WalletRepo),
		TypeSetPrimary:        NewSetPrimaryHandler(deps.WalletRepo),
		TypeActivateWallet:    NewActivateWalletHandler(deps.WalletRepo),
		TypeDeactivateWallet:  NewDeactivateWalletHandler(deps.WalletRepo),
		TypeSuspendWallet:     NewSuspendWalletHandler(deps.WalletRepo),
		TypeCreateChallenge:   NewCreateChallengeHandler(deps.ChallengeService, deps.AddressValidator),
		TypeVerifyChallenge:   NewVerifyChallengeHandler(deps.ChallengeRepo, deps.SignatureVerifier),
		TypeRecordWalletUsage: NewRecordWalletUsageHandler(deps.WalletRepo),
	}

	for cmdType, handler := range handlers {
		if err := bus.Register(cmdType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.CommandHandler[*LinkWallet]        = (*LinkWalletHandler)(nil)
	_ cqrs.CommandHandler[*UnlinkWallet]      = (*UnlinkWalletHandler)(nil)
	_ cqrs.CommandHandler[*UpdateLabel]       = (*UpdateLabelHandler)(nil)
	_ cqrs.CommandHandler[*SetPrimary]        = (*SetPrimaryHandler)(nil)
	_ cqrs.CommandHandler[*ActivateWallet]    = (*ActivateWalletHandler)(nil)
	_ cqrs.CommandHandler[*DeactivateWallet]  = (*DeactivateWalletHandler)(nil)
	_ cqrs.CommandHandler[*SuspendWallet]     = (*SuspendWalletHandler)(nil)
	_ cqrs.CommandHandler[*CreateChallenge]   = (*CreateChallengeHandler)(nil)
	_ cqrs.CommandHandler[*VerifyChallenge]   = (*VerifyChallengeHandler)(nil)
	_ cqrs.CommandHandler[*RecordWalletUsage] = (*RecordWalletUsageHandler)(nil)
)
