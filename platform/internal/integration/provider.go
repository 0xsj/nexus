package integration

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/integration/app/command"
	"github.com/0xsj/nexus/platform/internal/integration/app/query"
	"github.com/0xsj/nexus/platform/internal/integration/domain"
	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/integration/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Integration bounded context.
type Provider struct {
	// Infrastructure
	IntegrationRepo   domain.IntegrationRepository
	IntegrationLookup *postgres.IntegrationLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Integration provider.
type ProviderConfig struct {
	Pool      *pgxpool.Pool
	Publisher domain.EventPublisher
	Logger    log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Integration provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewIntegrationRepository(cfg.Pool)
	lookup := postgres.NewIntegrationLookup(cfg.Pool)

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
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
		IntegrationRepo:   repository,
		IntegrationLookup: lookup,
		CommandHandlers:   commandHandlers,
		QueryHandlers:     queryHandlers,
		HTTPHandler:       httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Integration HTTP routes on the given router.
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
		Pool:      pool,
		Publisher: &domain.NullEventPublisher{},
		Logger:    logger,
	})
}
