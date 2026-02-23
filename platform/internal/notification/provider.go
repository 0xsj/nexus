package notification

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/notification/app/command"
	"github.com/0xsj/nexus/platform/internal/notification/app/query"
	"github.com/0xsj/nexus/platform/internal/notification/domain"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/internal/notification/infrastructure/projections"
	v1 "github.com/0xsj/nexus/platform/internal/notification/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Notification bounded context.
type Provider struct {
	// Infrastructure
	NotificationRepo   *postgres.NotificationRepository
	PreferencesRepo    *postgres.PreferencesRepository
	NotificationLookup *postgres.NotificationLookup

	// Projections
	IdentityProjector *projections.IdentityProjector

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Notification provider.
type ProviderConfig struct {
	Pool           *pgxpool.Pool
	IdentityReader domain.IdentityReader
	EmailSender    domain.EmailSender
	PushSender     domain.PushSender
	Publisher      domain.EventPublisher
	Logger         log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Notification provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	notificationRepo := postgres.NewNotificationRepository(cfg.Pool)
	preferencesRepo := postgres.NewPreferencesRepository(cfg.Pool)
	lookup := postgres.NewNotificationLookup(cfg.Pool)

	// Projections
	projQueries := generated.New(cfg.Pool)
	identityProjector := projections.NewIdentityProjector(projQueries)

	// Use projection-backed reader if none provided
	identityReader := cfg.IdentityReader
	if identityReader == nil {
		identityReader = projections.NewIdentityReader(projQueries)
	}

	// Application - Command
	commandHandlers := command.NewHandlers(
		notificationRepo,
		preferencesRepo,
		identityReader,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		lookup,
		preferencesRepo,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		NotificationRepo:   notificationRepo,
		PreferencesRepo:    preferencesRepo,
		NotificationLookup: lookup,
		IdentityProjector:  identityProjector,
		CommandHandlers:    commandHandlers,
		QueryHandlers:      queryHandlers,
		HTTPHandler:        httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Notification HTTP routes on the given router.
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
		IdentityReader: domain.NewNullIdentityReader(),
		EmailSender:    domain.NewNullEmailSender(),
		PushSender:     domain.NewNullPushSender(),
		Publisher:      domain.NewNullEventPublisher(),
		Logger:         logger,
	})
}
