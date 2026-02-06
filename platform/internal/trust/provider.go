package trust

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/trust/app/command"
	"github.com/0xsj/nexus/platform/internal/trust/app/query"
	"github.com/0xsj/nexus/platform/internal/trust/domain"
	"github.com/0xsj/nexus/platform/internal/trust/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/trust/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Trust bounded context.
type Provider struct {
	// Infrastructure
	VouchRepo        domain.VouchRepository
	VouchLookup      *postgres.VouchLookup
	ReputationLookup *postgres.ReputationLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Trust provider.
type ProviderConfig struct {
	Pool               *pgxpool.Pool
	IdentityReader     domain.IdentityReader
	CredentialReader   domain.CredentialReader
	OrganizationReader domain.OrganizationReader
	Publisher          domain.EventPublisher
	Logger             log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Trust provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewVouchRepository(cfg.Pool)
	vouchLookup := postgres.NewVouchLookup(cfg.Pool)
	reputationLookup := postgres.NewReputationLookup(cfg.Pool)

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		vouchLookup,
		reputationLookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		VouchRepo:        repository,
		VouchLookup:      vouchLookup,
		ReputationLookup: reputationLookup,
		CommandHandlers:  commandHandlers,
		QueryHandlers:    queryHandlers,
		HTTPHandler:      httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Trust HTTP routes on the given router.
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
		Pool:               pool,
		IdentityReader:     &domain.NullIdentityReader{},
		CredentialReader:   &domain.NullCredentialReader{},
		OrganizationReader: &domain.NullOrganizationReader{},
		Publisher:          &domain.NullEventPublisher{},
		Logger:             logger,
	})
}
