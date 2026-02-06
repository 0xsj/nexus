package issuer

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/issuer/app/command"
	"github.com/0xsj/nexus/platform/internal/issuer/app/query"
	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/internal/issuer/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/issuer/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Issuer bounded context.
type Provider struct {
	// Infrastructure
	IssuerRepo     domain.IssuerRepository
	TemplateRepo   domain.TemplateRepository
	IssuerLookup   *postgres.IssuerLookup
	TemplateLookup *postgres.TemplateLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Issuer provider.
type ProviderConfig struct {
	Pool               *pgxpool.Pool
	OrganizationReader domain.OrganizationReader
	SchemaReader       domain.SchemaReader
	Publisher          domain.EventPublisher
	Logger             log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Issuer provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	issuerRepo := postgres.NewIssuerRepository(cfg.Pool)
	templateRepo := postgres.NewTemplateRepository(cfg.Pool)
	issuerLookup := postgres.NewIssuerLookup(cfg.Pool)
	templateLookup := postgres.NewTemplateLookup(cfg.Pool)

	// Application - Command
	commandHandlers := command.NewHandlers(
		issuerRepo,
		templateRepo,
		cfg.OrganizationReader,
		cfg.SchemaReader,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		issuerLookup,
		templateLookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		IssuerRepo:      issuerRepo,
		TemplateRepo:    templateRepo,
		IssuerLookup:    issuerLookup,
		TemplateLookup:  templateLookup,
		CommandHandlers: commandHandlers,
		QueryHandlers:   queryHandlers,
		HTTPHandler:     httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Issuer HTTP routes on the given router.
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
		OrganizationReader: &domain.NullOrganizationReader{},
		SchemaReader:       &domain.NullSchemaReader{},
		Publisher:          &domain.NullEventPublisher{},
		Logger:             logger,
	})
}
