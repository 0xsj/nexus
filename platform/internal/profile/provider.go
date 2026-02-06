package profile

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/profile/app/command"
	"github.com/0xsj/nexus/platform/internal/profile/app/query"
	"github.com/0xsj/nexus/platform/internal/profile/domain"
	"github.com/0xsj/nexus/platform/internal/profile/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/profile/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Profile bounded context.
type Provider struct {
	// Infrastructure
	Repository    domain.ProfileRepository
	ProfileLookup *postgres.ProfileLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Profile provider.
type ProviderConfig struct {
	Pool             *pgxpool.Pool
	CredentialReader domain.CredentialReader
	VanitySlugLookup domain.VanitySlugLookup
	Publisher        domain.EventPublisher
	Logger           log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Profile provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewProfileRepository(cfg.Pool)
	lookup := postgres.NewProfileLookup(cfg.Pool)

	// Use the lookup as vanity slug lookup if none provided
	vanitySlugLookup := cfg.VanitySlugLookup
	if vanitySlugLookup == nil {
		vanitySlugLookup = &profileLookupVanityAdapter{lookup: lookup}
	}

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		vanitySlugLookup,
		cfg.CredentialReader,
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
		ProfileLookup:   lookup,
		CommandHandlers: commandHandlers,
		QueryHandlers:   queryHandlers,
		HTTPHandler:     httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Profile HTTP routes on the given router.
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
		Pool:             pool,
		CredentialReader: domain.NewNullCredentialReader(),
		VanitySlugLookup: nil, // Will use the lookup adapter
		Publisher:        domain.NewNullEventPublisher(),
		Logger:           logger,
	})
}

// ============================================================================
// Vanity Slug Lookup Adapter
// ============================================================================

// profileLookupVanityAdapter adapts ProfileLookup to domain.VanitySlugLookup.
type profileLookupVanityAdapter struct {
	lookup *postgres.ProfileLookup
}

// SlugExists checks if a vanity slug is already taken.
func (a *profileLookupVanityAdapter) SlugExists(ctx context.Context, slug string) (bool, error) {
	return a.lookup.ExistsByVanitySlug(ctx, slug)
}
