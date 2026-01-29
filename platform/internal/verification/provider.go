package verification

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/wire"

	"github.com/0xsj/nexus/platform/internal/verification/application/command"
	"github.com/0xsj/nexus/platform/internal/verification/application/query"
	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/credential"
	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/oauth"
	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/persistence/postgres"
	httpv1 "github.com/0xsj/nexus/platform/internal/verification/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/id"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Wire Provider Set
// ============================================================================

// ProviderSet is the Wire provider set for the verification module.
var ProviderSet = wire.NewSet(
	NewModule,
)

// ============================================================================
// Module Configuration
// ============================================================================

// ModuleConfig holds configuration for the verification module.
type ModuleConfig struct {
	// Database is the PostgreSQL database connection.
	// Required - verification module requires persistence.
	Database *pgadapter.DB

	// OAuth contains OAuth provider configurations.
	OAuth oauth.Config

	// CredentialCommandBus is the credential module's command bus for issuing credentials.
	// Required for credential issuance.
	CredentialCommandBus cqrs.CommandBus

	// IssuerDID is the DID used for issuing credentials.
	IssuerDID string

	// UserDIDResolver resolves user DIDs for credential issuance.
	UserDIDResolver domain.UserDIDResolver

	// VerificationTTL is the time-to-live for verification flows.
	// Default: 15 minutes
	VerificationTTL time.Duration

	// AuthMiddleware is the authentication middleware for protected routes.
	// If nil, protected routes will not require authentication (development only).
	AuthMiddleware func(http.Handler) http.Handler
}

// DefaultModuleConfig returns default configuration.
func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		OAuth:           oauth.DefaultConfig(),
		VerificationTTL: 15 * time.Minute,
	}
}

// ============================================================================
// Module
// ============================================================================

// Module is the verification module that provides all verification functionality.
type Module struct {
	router *httpv1.Router

	// Repositories
	verificationRepo   domain.VerificationRepository
	oauthStateRepo     domain.OAuthStateRepository
	providerTokenRepo  domain.ProviderTokenRepository
	verificationReader domain.VerificationReadRepository

	// Services
	providerService   domain.ProviderService
	oauthStateService domain.OAuthStateService
	credentialIssuer  domain.CredentialIssuerService

	// CQRS
	commandBus *cqrs.InMemoryCommandBus
	queryBus   *cqrs.InMemoryQueryBus

	logger log.Logger
}

// NewModule creates a new verification module with default configuration.
func NewModule(logger log.Logger) (*Module, error) {
	return NewModuleWithConfig(logger, DefaultModuleConfig())
}

// NewModuleWithConfig creates a new verification module with custom configuration.
func NewModuleWithConfig(logger log.Logger, cfg ModuleConfig) (*Module, error) {
	if cfg.Database == nil {
		return nil, ErrMissingDependency("Database")
	}
	if cfg.CredentialCommandBus == nil {
		return nil, ErrMissingDependency("CredentialCommandBus")
	}
	if cfg.UserDIDResolver == nil {
		return nil, ErrMissingDependency("UserDIDResolver")
	}
	if cfg.IssuerDID == "" {
		return nil, ErrMissingDependency("IssuerDID")
	}

	// Validate OAuth config
	if err := cfg.OAuth.Validate(); err != nil {
		return nil, err
	}

	// Create ID generator
	idGenerator := id.NewGenerator()

	// Create repositories
	verificationRepo := postgres.NewVerificationRepository(cfg.Database)
	oauthStateRepo := postgres.NewOAuthStateRepository(cfg.Database)
	providerTokenRepo := postgres.NewProviderTokenRepository(cfg.Database)
	verificationReader := postgres.NewVerificationReader(cfg.Database)

	// Create OAuth services
	providerService := oauth.NewProviderService(&cfg.OAuth, logger)
	oauthStateService := oauth.NewOAuthStateService(oauthStateRepo, &cfg.OAuth, logger)

	// Create credential issuer adapter
	credentialModuleAdapter := credential.NewModuleAdapter(
		cfg.CredentialCommandBus,
		idGenerator,
		cfg.IssuerDID,
	)
	credentialIssuer := credential.NewIssuerService(credentialModuleAdapter, cfg.IssuerDID, logger)

	// Create CQRS buses
	commandBus := cqrs.NewCommandBus()
	queryBus := cqrs.NewQueryBus()

	// Register command handlers
	cmdDeps := command.HandlerDependencies{
		VerificationRepo:  verificationRepo,
		OAuthStateRepo:    oauthStateRepo,
		ProviderService:   providerService,
		CredentialIssuer:  credentialIssuer,
		OAuthStateService: oauthStateService,
		UserDIDResolver:   cfg.UserDIDResolver,
		IDGenerator:       idGenerator,
		VerificationTTL:   cfg.VerificationTTL,
	}

	if err := command.RegisterHandlers(commandBus, cmdDeps); err != nil {
		return nil, err
	}

	// Register query handlers
	queryDeps := query.HandlerDependencies{
		ReadRepo: verificationReader,
	}

	if err := query.RegisterHandlers(queryBus, queryDeps); err != nil {
		return nil, err
	}

	// Create HTTP router
	router := httpv1.NewRouter(commandBus, queryBus, idGenerator, logger)
	if cfg.AuthMiddleware != nil {
		router.WithAuthMiddleware(cfg.AuthMiddleware)
	}

	logger.Info("verification module initialized",
		log.String("storage", "PostgreSQL"),
		log.Int("enabled_providers", len(cfg.OAuth.EnabledProviders())),
	)

	return &Module{
		router:             router,
		verificationRepo:   verificationRepo,
		oauthStateRepo:     oauthStateRepo,
		providerTokenRepo:  providerTokenRepo,
		verificationReader: verificationReader,
		providerService:    providerService,
		oauthStateService:  oauthStateService,
		credentialIssuer:   credentialIssuer,
		commandBus:         commandBus,
		queryBus:           queryBus,
		logger:             logger,
	}, nil
}

// ============================================================================
// Route Accessors
// ============================================================================

// Routes returns the HTTP routes as a mountable chi.Router.
// Mount at: /api/v1/verifications
func (m *Module) Routes() chi.Router {
	return m.router.Routes()
}

// ============================================================================
// Service Accessors
// ============================================================================

// CommandBus returns the command bus for direct access.
func (m *Module) CommandBus() cqrs.CommandBus {
	return m.commandBus
}

// QueryBus returns the query bus for direct access.
func (m *Module) QueryBus() cqrs.QueryBus {
	return m.queryBus
}

// ProviderService returns the OAuth provider service.
func (m *Module) ProviderService() domain.ProviderService {
	return m.providerService
}

// OAuthStateService returns the OAuth state service.
func (m *Module) OAuthStateService() domain.OAuthStateService {
	return m.oauthStateService
}

// CredentialIssuer returns the credential issuer service.
func (m *Module) CredentialIssuer() domain.CredentialIssuerService {
	return m.credentialIssuer
}

// VerificationRepository returns the verification repository for direct access.
func (m *Module) VerificationRepository() domain.VerificationRepository {
	return m.verificationRepo
}

// ============================================================================
// Lifecycle
// ============================================================================

// Stop gracefully stops the module and releases resources.
func (m *Module) Stop() {
	m.logger.Info("verification module stopped")
}

// ============================================================================
// Errors
// ============================================================================

// ErrMissingDependency creates an error for missing dependencies.
func ErrMissingDependency(name string) error {
	return &MissingDependencyError{Name: name}
}

// MissingDependencyError indicates a required dependency is missing.
type MissingDependencyError struct {
	Name string
}

// Error implements error.
func (e *MissingDependencyError) Error() string {
	return "missing required dependency: " + e.Name
}
