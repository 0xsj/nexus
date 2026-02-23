package schema

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/schema/app/command"
	"github.com/0xsj/nexus/platform/internal/schema/app/query"
	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/internal/schema/infrastructure/projections"
	v1 "github.com/0xsj/nexus/platform/internal/schema/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Schema bounded context.
type Provider struct {
	// Infrastructure
	Repository   domain.SchemaRepository
	SchemaLookup domain.SchemaLookup

	// Projections
	IssuerProjector *projections.IssuerProjector

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Schema provider.
type ProviderConfig struct {
	Pool           *pgxpool.Pool
	IssuerReader   domain.IssuerReader
	EventPublisher domain.EventPublisher
	Logger         log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Schema provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewRepository(cfg.Pool)

	// The repository implements both SchemaRepository and SchemaLookup
	schemaLookup := repository

	// Projections
	projQueries := generated.New(cfg.Pool)
	issuerProjector := projections.NewIssuerProjector(projQueries)

	// Use projection-backed reader if none provided
	issuerReader := cfg.IssuerReader
	if issuerReader == nil {
		issuerReader = projections.NewIssuerReader(projQueries)
	}

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		schemaLookup,
		issuerReader,
		cfg.EventPublisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		repository,
		schemaLookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		Repository:      repository,
		SchemaLookup:    schemaLookup,
		IssuerProjector: issuerProjector,
		CommandHandlers: commandHandlers,
		QueryHandlers:   queryHandlers,
		HTTPHandler:     httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Schema HTTP routes on the given router.
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
		Pool:           pool,
		IssuerReader:   domain.NewNullIssuerReader(),
		EventPublisher: domain.NewNullEventPublisher(),
		Logger:         logger,
	})
}
