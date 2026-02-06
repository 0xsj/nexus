package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Wallet context.
type Handlers struct {
	walletRepo        domain.WalletRepository
	walletLookup      domain.WalletLookup
	signatureVerifier domain.SignatureVerifier
	didDeriver        domain.DIDDeriver
	publisher         domain.EventPublisher
	logger            log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	walletRepo domain.WalletRepository,
	walletLookup domain.WalletLookup,
	signatureVerifier domain.SignatureVerifier,
	didDeriver domain.DIDDeriver,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		walletRepo:        walletRepo,
		walletLookup:      walletLookup,
		signatureVerifier: signatureVerifier,
		didDeriver:        didDeriver,
		publisher:         publisher,
		logger:            logger,
	}
}

// ============================================================================
// LinkWallet Handler
// ============================================================================

// HandleLinkWallet handles the LinkWallet command.
func (h *Handlers) HandleLinkWallet(ctx context.Context, cmd LinkWallet) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleLinkWallet"

	// 1. Parse address value object
	address, err := domain.NewWalletAddress(cmd.Address)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Check address uniqueness via lookup
	exists, err := h.walletLookup.ExistsByAddress(ctx, address)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.WalletAlreadyLinked(op, cmd.Address)
	}

	// 3. Parse chain
	chain, err := domain.NewChain(cmd.ChainID, chainNameFromID(cmd.ChainID), "eip155")
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Parse label
	label, err := domain.NewWalletLabel(cmd.Label)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Derive DID (non-fatal)
	did := ""
	derivedDID, err := h.didDeriver.DeriveDID(ctx, address, chain)
	if err != nil {
		h.logger.Warn("DID derivation skipped",
			log.String("op", op),
			log.String("address", cmd.Address),
			log.Err(err),
		)
	} else {
		did = derivedDID
	}

	// 6. If DID is still empty, generate a fallback did:pkh
	if did == "" {
		did = "did:pkh:eip155:" + cmd.Address
	}

	// 7. Generate wallet ID
	walletID := domain.NewWalletID()

	// 8. Create aggregate
	w, err := domain.LinkWallet(walletID, cmd.UserID, address, chain, did, label)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 9. Save
	if err := h.walletRepo.Save(ctx, w); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 10. Publish events
	h.publishEvents(ctx, op, w.Changes()...)

	return &cqrs.CommandResult{
		ID:      walletID.String(),
		Version: w.Version(),
		Data: LinkWalletResult{
			WalletID: walletID.String(),
			Address:  cmd.Address,
			Status:   w.Status().String(),
		},
	}, nil
}

// ============================================================================
// VerifyWallet Handler
// ============================================================================

// HandleVerifyWallet handles the VerifyWallet command.
func (h *Handlers) HandleVerifyWallet(ctx context.Context, cmd VerifyWallet) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleVerifyWallet"

	// 1. Load aggregate
	walletID := domain.WalletIDFromTypesID(cmd.WalletID)
	w, err := h.walletRepo.Get(ctx, walletID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Verify signature
	valid, err := h.signatureVerifier.VerifySignature(ctx, w.Address(), cmd.Message, cmd.Signature)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if !valid {
		return nil, domain.SignatureInvalid(op, w.Address().String())
	}

	// 3. Verify aggregate
	if err := w.Verify(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.walletRepo.Save(ctx, w); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, w.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.WalletID.String(),
		Version: w.Version(),
		Data: VerifyWalletResult{
			WalletID: cmd.WalletID.String(),
			Status:   w.Status().String(),
		},
	}, nil
}

// ============================================================================
// UnlinkWallet Handler
// ============================================================================

// HandleUnlinkWallet handles the UnlinkWallet command.
func (h *Handlers) HandleUnlinkWallet(ctx context.Context, cmd UnlinkWallet) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUnlinkWallet"

	// 1. Load aggregate
	walletID := domain.WalletIDFromTypesID(cmd.WalletID)
	w, err := h.walletRepo.Get(ctx, walletID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Unlink
	if err := w.Unlink(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.walletRepo.Save(ctx, w); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, w.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.WalletID.String(),
		Version: w.Version(),
		Data: UnlinkWalletResult{
			WalletID: cmd.WalletID.String(),
			Status:   w.Status().String(),
		},
	}, nil
}

// ============================================================================
// SetPrimaryWallet Handler
// ============================================================================

// HandleSetPrimaryWallet handles the SetPrimaryWallet command.
func (h *Handlers) HandleSetPrimaryWallet(ctx context.Context, cmd SetPrimaryWallet) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleSetPrimaryWallet"

	// 1. Load aggregate
	walletID := domain.WalletIDFromTypesID(cmd.WalletID)
	w, err := h.walletRepo.Get(ctx, walletID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Set primary
	if err := w.SetPrimary(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.walletRepo.Save(ctx, w); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, w.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.WalletID.String(),
		Version: w.Version(),
		Data: SetPrimaryWalletResult{
			WalletID: cmd.WalletID.String(),
			Status:   w.Status().String(),
		},
	}, nil
}

// ============================================================================
// UpdateWalletLabel Handler
// ============================================================================

// HandleUpdateWalletLabel handles the UpdateWalletLabel command.
func (h *Handlers) HandleUpdateWalletLabel(ctx context.Context, cmd UpdateWalletLabel) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUpdateWalletLabel"

	// 1. Load aggregate
	walletID := domain.WalletIDFromTypesID(cmd.WalletID)
	w, err := h.walletRepo.Get(ctx, walletID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Parse new label
	label, err := domain.NewWalletLabel(cmd.Label)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Update label
	if err := w.UpdateLabel(label); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.walletRepo.Save(ctx, w); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, w.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.WalletID.String(),
		Version: w.Version(),
		Data: UpdateWalletLabelResult{
			WalletID: cmd.WalletID.String(),
			Label:    cmd.Label,
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

// chainNameFromID maps common chain IDs to names.
func chainNameFromID(chainID int) string {
	switch chainID {
	case 1:
		return "Ethereum"
	case 137:
		return "Polygon"
	case 10:
		return "Optimism"
	case 42161:
		return "Arbitrum"
	case 8453:
		return "Base"
	case 56:
		return "BNB Chain"
	case 43114:
		return "Avalanche"
	default:
		return "Unknown"
	}
}
