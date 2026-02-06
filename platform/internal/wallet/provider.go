package wallet

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/wallet/app/command"
	"github.com/0xsj/nexus/platform/internal/wallet/app/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/wallet/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Wallet bounded context.
type Provider struct {
	// Infrastructure
	Repository   domain.WalletRepository
	WalletLookup *postgres.WalletLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Wallet provider.
type ProviderConfig struct {
	Pool              *pgxpool.Pool
	WalletLookup      domain.WalletLookup
	ChallengeRepo     domain.ChallengeRepository
	SignatureVerifier domain.SignatureVerifier
	DIDDeriver        domain.DIDDeriver
	Publisher         domain.EventPublisher
	Logger            log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Wallet provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewWalletRepository(cfg.Pool)
	lookup := postgres.NewWalletLookup(cfg.Pool)

	// Use the provided WalletLookup or fall back to the postgres lookup adapter
	var walletLookup domain.WalletLookup
	if cfg.WalletLookup != nil {
		walletLookup = cfg.WalletLookup
	} else {
		walletLookup = newWalletLookupAdapter(lookup)
	}

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		walletLookup,
		cfg.SignatureVerifier,
		cfg.DIDDeriver,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		lookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		Repository:      repository,
		WalletLookup:    lookup,
		CommandHandlers: commandHandlers,
		QueryHandlers:   queryHandlers,
		HTTPHandler:     httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Wallet HTTP routes on the given router.
func (p *Provider) RegisterRoutes(r chi.Router) {
	p.HTTPHandler.RegisterRoutes(r)
}

// ============================================================================
// Optional Dependencies
// ============================================================================

// NewProviderWithDefaults creates a Provider with default/null implementations
// for optional dependencies. Useful for testing or standalone operation.
func NewProviderWithDefaults(pool *pgxpool.Pool, logger log.Logger) *Provider {
	return NewProvider(ProviderConfig{
		Pool:              pool,
		SignatureVerifier: domain.NewNullSignatureVerifier(),
		DIDDeriver:        domain.NewNullDIDDeriver(),
		ChallengeRepo:     domain.NewNullChallengeRepository(),
		Publisher:         domain.NewNullEventPublisher(),
		Logger:            logger,
	})
}

// ============================================================================
// Wallet Lookup Adapter
// ============================================================================

// walletLookupAdapter bridges the postgres.WalletLookup to domain.WalletLookup.
type walletLookupAdapter struct {
	lookup *postgres.WalletLookup
}

// newWalletLookupAdapter creates a new walletLookupAdapter.
func newWalletLookupAdapter(lookup *postgres.WalletLookup) *walletLookupAdapter {
	return &walletLookupAdapter{lookup: lookup}
}

// ExistsByAddress implements domain.WalletLookup.
func (a *walletLookupAdapter) ExistsByAddress(ctx context.Context, address domain.WalletAddress) (bool, error) {
	return a.lookup.ExistsByAddress(ctx, address.String())
}

// GetWalletIDByAddress implements domain.WalletLookup.
func (a *walletLookupAdapter) GetWalletIDByAddress(ctx context.Context, address domain.WalletAddress) (domain.WalletID, error) {
	proj, err := a.lookup.GetByAddress(ctx, address.String())
	if err != nil {
		return domain.WalletID{}, err
	}
	return domain.ParseWalletID(proj.ID)
}

// Ensure walletLookupAdapter implements domain.WalletLookup.
var _ domain.WalletLookup = (*walletLookupAdapter)(nil)
