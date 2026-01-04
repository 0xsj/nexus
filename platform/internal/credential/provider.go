package credential

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"

	"github.com/0xsj/nexus/platform/internal/credential/application/command"
	"github.com/0xsj/nexus/platform/internal/credential/application/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/memory"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/signing"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/verification"
	httpv1 "github.com/0xsj/nexus/platform/internal/credential/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Wire Provider Set
// ============================================================================

// ProviderSet is the Wire provider set for the credential module.
var ProviderSet = wire.NewSet(
	NewModule,
	wire.Bind(new(domain.CredentialRepository), new(*memory.Repository)),
	wire.Bind(new(query.CredentialReadRepository), new(*memory.Repository)),
)

// ============================================================================
// Module Configuration
// ============================================================================

// ModuleConfig holds configuration for the credential module.
type ModuleConfig struct {
	// IssuerDID is the DID of the credential issuer.
	// If empty, a new DID will be generated.
	IssuerDID string

	// PrivateKey is the issuer's private key bytes (64 bytes for Ed25519).
	// If empty, a new key pair will be generated.
	PrivateKey []byte

	// EnableSigning enables credential signing.
	// If false, credentials will not have signed VCs.
	EnableSigning bool

	// Database is the PostgreSQL database connection.
	// If nil, in-memory storage is used.
	Database *pgadapter.DB
}

// DefaultModuleConfig returns default configuration.
func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		EnableSigning: true,
	}
}

// ============================================================================
// Module
// ============================================================================

// Module is the credential module that provides all credential functionality.
type Module struct {
	router              *httpv1.Router
	writeRepository     domain.CredentialRepository
	readRepository      query.CredentialReadRepository
	commandBus          *cqrs.InMemoryCommandBus
	queryBus            *cqrs.InMemoryQueryBus
	signingService      domain.SigningService
	verificationService domain.VerificationService
	issuerDID           did.DID
	logger              log.Logger
}

// NewModule creates a new credential module with default configuration.
func NewModule(logger log.Logger) (*Module, error) {
	return NewModuleWithConfig(logger, DefaultModuleConfig())
}

// NewModuleWithConfig creates a new credential module with custom configuration.
func NewModuleWithConfig(logger log.Logger, cfg ModuleConfig) (*Module, error) {
	// Create repositories based on configuration
	var writeRepo domain.CredentialRepository
	var readRepo query.CredentialReadRepository

	if cfg.Database != nil {
		// Use PostgreSQL
		pgRepo := postgres.NewRepository(cfg.Database)
		writeRepo = pgRepo
		readRepo = pgRepo
		logger.Info("credential module using PostgreSQL persistence")
	} else {
		// Use in-memory
		memRepo := memory.NewRepository()
		writeRepo = memRepo
		readRepo = memRepo
		logger.Info("credential module using in-memory persistence")
	}

	// Create CQRS buses
	commandBus := cqrs.NewCommandBus()
	queryBus := cqrs.NewQueryBus()

	// Create signing service if enabled
	var signingService domain.SigningService
	var issuerDID did.DID

	if cfg.EnableSigning {
		signer, issuer, err := createSigningService(cfg, logger)
		if err != nil {
			return nil, err
		}
		signingService = signer
		issuerDID = issuer
	}

	// Create verification service
	verificationService := verification.NewVerifier()

	// Register command handlers
	if err := command.RegisterHandlers(commandBus, writeRepo, signingService); err != nil {
		return nil, err
	}

	// Register query handlers
	if err := query.RegisterHandlers(queryBus, readRepo, verificationService); err != nil {
		return nil, err
	}

	// Create HTTP router
	router := httpv1.NewRouter(commandBus, queryBus, logger)

	return &Module{
		router:              router,
		writeRepository:     writeRepo,
		readRepository:      readRepo,
		commandBus:          commandBus,
		queryBus:            queryBus,
		signingService:      signingService,
		verificationService: verificationService,
		issuerDID:           issuerDID,
		logger:              logger,
	}, nil
}

// createSigningService creates the signing service based on configuration.
func createSigningService(cfg ModuleConfig, logger log.Logger) (domain.SigningService, did.DID, error) {
	var keyPair *ed25519.KeyPair
	var issuerDID did.DID
	var err error

	if len(cfg.PrivateKey) > 0 && cfg.IssuerDID != "" {
		// Use provided key and DID
		keyPair, err = ed25519.FromPrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, did.DID{}, err
		}

		issuerDID, err = did.Parse(cfg.IssuerDID)
		if err != nil {
			return nil, did.DID{}, err
		}
	} else {
		// Generate new key pair and DID
		issuerDID, keyPair, err = key.Generate()
		if err != nil {
			return nil, did.DID{}, err
		}

		logger.Info("generated new issuer DID",
			log.String("issuer_did", issuerDID.String()),
		)
	}

	// Create crypto signer
	cryptoSigner, err := ed25519.NewSigner(keyPair)
	if err != nil {
		return nil, did.DID{}, err
	}

	// Create signing service
	signer, err := signing.NewSigner(issuerDID, keyPair, cryptoSigner)
	if err != nil {
		return nil, did.DID{}, err
	}

	return signer, issuerDID, nil
}

// ============================================================================
// Accessors
// ============================================================================

// Routes returns the HTTP routes for mounting.
func (m *Module) Routes() chi.Router {
	return m.router.Routes()
}

// CommandBus returns the command bus for direct access if needed.
func (m *Module) CommandBus() cqrs.CommandBus {
	return m.commandBus
}

// QueryBus returns the query bus for direct access if needed.
func (m *Module) QueryBus() cqrs.QueryBus {
	return m.queryBus
}

// WriteRepository returns the write repository.
func (m *Module) WriteRepository() domain.CredentialRepository {
	return m.writeRepository
}

// ReadRepository returns the read repository.
func (m *Module) ReadRepository() query.CredentialReadRepository {
	return m.readRepository
}

// SigningService returns the signing service.
func (m *Module) SigningService() domain.SigningService {
	return m.signingService
}

// VerificationService returns the verification service.
func (m *Module) VerificationService() domain.VerificationService {
	return m.verificationService
}

// IssuerDID returns the issuer DID.
func (m *Module) IssuerDID() did.DID {
	return m.issuerDID
}
