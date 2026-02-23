package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/credential"
	credentialadapters "github.com/0xsj/nexus/platform/internal/credential/infrastructure/adapters"
	credentialeventbus "github.com/0xsj/nexus/platform/internal/credential/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/identity"
	identityadapters "github.com/0xsj/nexus/platform/internal/identity/infrastructure/adapters"
	identityeventbus "github.com/0xsj/nexus/platform/internal/identity/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/integration"
	integrationadapters "github.com/0xsj/nexus/platform/internal/integration/infrastructure/adapters"
	integrationeventbus "github.com/0xsj/nexus/platform/internal/integration/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/issuer"
	issuereventbus "github.com/0xsj/nexus/platform/internal/issuer/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/ledger"
	ledgereventbus "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/notification"
	notificationdomain "github.com/0xsj/nexus/platform/internal/notification/domain"
	notificationeventbus "github.com/0xsj/nexus/platform/internal/notification/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/organization"
	organizationdomain "github.com/0xsj/nexus/platform/internal/organization/domain"
	organizationeventbus "github.com/0xsj/nexus/platform/internal/organization/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/presentation"
	presentationeventbus "github.com/0xsj/nexus/platform/internal/presentation/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/profile"
	profileeventbus "github.com/0xsj/nexus/platform/internal/profile/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/schema"
	schemaeventbus "github.com/0xsj/nexus/platform/internal/schema/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/trust"
	trusteventbus "github.com/0xsj/nexus/platform/internal/trust/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/verification"
	verifadapters "github.com/0xsj/nexus/platform/internal/verification/infrastructure/adapters"
	verificationeventbus "github.com/0xsj/nexus/platform/internal/verification/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/wallet"
	walletadapters "github.com/0xsj/nexus/platform/internal/wallet/infrastructure/adapters"
	walleteventbus "github.com/0xsj/nexus/platform/internal/wallet/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/pkg/database"
	"github.com/0xsj/nexus/platform/pkg/email"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventbus/memory"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/middleware"
	"github.com/0xsj/nexus/platform/pkg/observability"
	"github.com/0xsj/nexus/platform/pkg/observability/health"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// ========================================================================
	// Configuration
	// ========================================================================

	port := envInt("PORT", 8080)
	dbHost := envStr("DATABASE_HOST", "localhost")
	dbPort := envInt("DATABASE_PORT", 5432)
	dbName := envStr("DATABASE_NAME", "nexus")
	dbUser := envStr("DATABASE_USER", "nexus")
	dbPassword := envStr("DATABASE_PASSWORD", "nexus")
	logLevel := envStr("LOG_LEVEL", "debug")
	appEnv := envStr("APP_ENV", "development")

	smtpHost := envStr("SMTP_HOST", "localhost")
	smtpPort := envInt("SMTP_PORT", 1025)
	smtpFrom := envStr("SMTP_FROM", "noreply@nexus.local")
	appBaseURL := envStr("APP_BASE_URL", "http://localhost:3010")

	githubClientID := envStr("OAUTH_GITHUB_CLIENT_ID", "")
	githubClientSecret := envStr("OAUTH_GITHUB_CLIENT_SECRET", "")
	googleClientID := envStr("OAUTH_GOOGLE_CLIENT_ID", "")
	googleClientSecret := envStr("OAUTH_GOOGLE_CLIENT_SECRET", "")

	// ========================================================================
	// Observability
	// ========================================================================

	var obs *observability.Provider
	if appEnv == "production" {
		obs = observability.NewProduction()
	} else {
		obs = observability.NewDevelopment()
	}
	_ = logLevel // Used by observability config when customized

	logger := obs.Logger()
	logger.Info("starting nexus platform",
		log.String("env", appEnv),
		log.Int("port", port),
	)

	// ========================================================================
	// Database
	// ========================================================================

	pgConfig := postgres.NewConfig(database.Config{
		Host:         dbHost,
		Port:         dbPort,
		Database:     dbName,
		Username:     dbUser,
		Password:     dbPassword,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
	})

	db, err := postgres.New(ctx, pgConfig)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer db.Close()

	pool := db.Pool()
	logger.Info("database connected",
		log.String("host", dbHost),
		log.Int("port", dbPort),
		log.String("database", dbName),
	)

	// ========================================================================
	// Event Bus
	// ========================================================================

	bus := memory.New(eventbus.Config{
		Logger: obs.ComponentLogger("eventbus"),
	})
	if err := bus.Start(ctx); err != nil {
		return fmt.Errorf("starting event bus: %w", err)
	}
	defer bus.Close()

	// ========================================================================
	// Bounded Contexts
	// ========================================================================

	// Identity
	identityPublisher := identityeventbus.NewAdapter(bus)

	smtpSender := email.NewSMTPSender(email.SMTPConfig{
		Host: smtpHost,
		Port: smtpPort,
		From: smtpFrom,
	})

	oauthProviders := make(map[string]identityadapters.OAuthProviderConfig)
	if githubClientID != "" {
		oauthProviders["github"] = identityadapters.OAuthProviderConfig{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			Scopes:       []string{"read:user", "user:email"},
		}
	}
	if googleClientID != "" {
		oauthProviders["google"] = identityadapters.OAuthProviderConfig{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
			Scopes:       []string{"openid", "email", "profile"},
		}
	}

	oauthService := identityadapters.NewOAuthService(identityadapters.OAuthServiceConfig{
		Providers: oauthProviders,
	})
	emailService := identityadapters.NewEmailService(identityadapters.EmailServiceConfig{
		SMTPSender: smtpSender,
		BaseURL:    appBaseURL,
	})

	identityProvider := identity.NewProvider(identity.ProviderConfig{
		Pool:         pool,
		DIDGenerator: identityadapters.NewDIDGenerator(),
		WalletReader: identityadapters.NewWalletReader(),
		OAuthService: oauthService,
		EmailService: emailService,
		Publisher:    identityPublisher,
		Logger:       obs.ComponentLogger("identity"),
	})

	// Schema (nil IssuerReader → uses projection-backed reader)
	schemaProvider := schema.NewProvider(schema.ProviderConfig{
		Pool:           pool,
		EventPublisher: schemaeventbus.NewAdapter(bus),
		Logger:         obs.ComponentLogger("schema"),
	})

	// Credential (nil SchemaResolver → uses projection-backed reader)
	jwtSigner, err := credentialadapters.NewJWTCredentialSigner()
	if err != nil {
		return fmt.Errorf("creating credential signer: %w", err)
	}
	credentialProvider := credential.NewProvider(credential.ProviderConfig{
		Pool:      pool,
		Signer:    jwtSigner,
		Publisher: credentialeventbus.NewAdapter(bus),
		Logger:    obs.ComponentLogger("credential"),
	})

	// Integration (must be created before Verification — it's a dependency)
	integrationAdapterConfig := &integrationadapters.AdapterConfig{}
	if githubClientID != "" {
		integrationAdapterConfig.GitHub = &integrationadapters.GitHubConfig{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURI:  appBaseURL + "/auth/callback/github",
		}
	}
	if googleClientID != "" {
		integrationAdapterConfig.Google = &integrationadapters.GoogleConfig{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURI:  appBaseURL + "/auth/callback/google",
		}
	}
	integrationRegistry, err := integrationadapters.RegisterDefaultAdapters(integrationAdapterConfig)
	if err != nil {
		return fmt.Errorf("creating integration adapter registry: %w", err)
	}

	integrationBridge := verifadapters.NewIntegrationBridge(integrationRegistry)

	// Verification (with cross-context bridges)
	credentialIssuerAdapter := verifadapters.NewCredentialIssuerAdapter(credentialProvider.CommandHandlers)
	verificationProvider := verification.NewProvider(verification.ProviderConfig{
		Pool:               pool,
		CredentialIssuer:   credentialIssuerAdapter,
		DataFetcher:        integrationBridge,
		OAuthURLGenerator:  integrationBridge,
		OAuthCodeExchanger: integrationBridge,
		Publisher:          verificationeventbus.NewAdapter(bus),
		Logger:             obs.ComponentLogger("verification"),
	})

	// Wallet
	walletProvider := wallet.NewProvider(wallet.ProviderConfig{
		Pool:              pool,
		SignatureVerifier: walletadapters.NewSIWESignatureVerifier(),
		DIDDeriver:        walletadapters.NewPKHDIDDeriver(),
		ChallengeRepo:     walletadapters.NewInMemoryChallengeRepository(),
		Publisher:         walleteventbus.NewAdapter(bus),
		Logger:            obs.ComponentLogger("wallet"),
	})

	// Organization (nil IdentityReader → uses projection-backed reader,
	// nil SlugLookup → falls back to postgres OrganizationLookup)
	organizationProvider := organization.NewProvider(organization.ProviderConfig{
		Pool:                pool,
		SlugLookup:          nil,
		DIDService:          organizationdomain.NewNullDIDService(),
		NotificationService: organizationdomain.NewNullNotificationService(),
		Publisher:           organizationeventbus.NewAdapter(bus),
		Logger:              obs.ComponentLogger("organization"),
	})

	// Profile (nil CredentialReader → uses projection-backed reader)
	profileProvider := profile.NewProvider(profile.ProviderConfig{
		Pool:      pool,
		Publisher: profileeventbus.NewAdapter(bus),
		Logger:    obs.ComponentLogger("profile"),
	})

	// Notification
	notificationProvider := notification.NewProvider(notification.ProviderConfig{
		Pool:        pool,
		EmailSender: notificationdomain.NewNullEmailSender(),
		PushSender:  notificationdomain.NewNullPushSender(),
		Publisher:   notificationeventbus.NewAdapter(bus),
		Logger:      obs.ComponentLogger("notification"),
	})

	// Presentation (nil readers → projections populate tables for future use)
	presentationProvider := presentation.NewProvider(presentation.ProviderConfig{
		Pool:      pool,
		Publisher: presentationeventbus.NewAdapter(bus),
		Logger:    obs.ComponentLogger("presentation"),
	})

	// Trust (nil readers → projections populate tables for future use)
	trustProvider := trust.NewProvider(trust.ProviderConfig{
		Pool:      pool,
		Publisher: trusteventbus.NewAdapter(bus),
		Logger:    obs.ComponentLogger("trust"),
	})

	// Integration (adapter registry created above, provider for HTTP routes + persistence)
	integrationProvider := integration.NewProvider(integration.ProviderConfig{
		Pool:      pool,
		Publisher: integrationeventbus.NewAdapter(bus),
		Logger:    obs.ComponentLogger("integration"),
	})

	// Issuer (nil readers → uses projection-backed readers)
	issuerProvider := issuer.NewProvider(issuer.ProviderConfig{
		Pool:      pool,
		Publisher: issuereventbus.NewAdapter(bus),
		Logger:    obs.ComponentLogger("issuer"),
	})

	// Ledger
	metadataExtractor := ledgereventbus.NewMetadataExtractor()
	ledgerProvider := ledger.NewProvider(pool, metadataExtractor)

	// Subscribe ledger to domain events via eventbus adapter
	ledgerSubscriber := ledgereventbus.NewAdapter(bus)
	if err := ledgerProvider.SubscribeToEvents(ledgerSubscriber); err != nil {
		return fmt.Errorf("subscribing ledger to events: %w", err)
	}

	// ========================================================================
	// Cross-Context Projector Subscriptions
	// ========================================================================

	// Identity events → Organization, Notification, Presentation, Trust
	if _, err := bus.Subscribe(ctx, "User.*", organizationProvider.IdentityProjector.Handle); err != nil {
		return fmt.Errorf("subscribing organization identity projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "User.*", notificationProvider.IdentityProjector.Handle); err != nil {
		return fmt.Errorf("subscribing notification identity projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "User.*", presentationProvider.IdentityProjector.Handle); err != nil {
		return fmt.Errorf("subscribing presentation identity projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "User.*", trustProvider.IdentityProjector.Handle); err != nil {
		return fmt.Errorf("subscribing trust identity projector: %w", err)
	}

	// Schema events → Credential, Issuer
	if _, err := bus.Subscribe(ctx, "Schema.*", credentialProvider.SchemaProjector.Handle); err != nil {
		return fmt.Errorf("subscribing credential schema projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "Schema.*", issuerProvider.SchemaProjector.Handle); err != nil {
		return fmt.Errorf("subscribing issuer schema projector: %w", err)
	}

	// Credential events → Presentation, Profile, Trust
	if _, err := bus.Subscribe(ctx, "Credential.*", presentationProvider.CredentialProjector.Handle); err != nil {
		return fmt.Errorf("subscribing presentation credential projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "Credential.*", profileProvider.CredentialProjector.Handle); err != nil {
		return fmt.Errorf("subscribing profile credential projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "Credential.*", trustProvider.CredentialProjector.Handle); err != nil {
		return fmt.Errorf("subscribing trust credential projector: %w", err)
	}

	// Organization events → Trust, Issuer
	if _, err := bus.Subscribe(ctx, "Organization.*", trustProvider.OrganizationProjector.Handle); err != nil {
		return fmt.Errorf("subscribing trust organization projector: %w", err)
	}
	if _, err := bus.Subscribe(ctx, "Organization.*", issuerProvider.OrganizationProjector.Handle); err != nil {
		return fmt.Errorf("subscribing issuer organization projector: %w", err)
	}

	// Issuer events → Schema
	if _, err := bus.Subscribe(ctx, "Issuer.*", schemaProvider.IssuerProjector.Handle); err != nil {
		return fmt.Errorf("subscribing schema issuer projector: %w", err)
	}

	logger.Info("cross-context projectors subscribed")

	// ========================================================================
	// HTTP Router
	// ========================================================================

	router := chi.NewRouter()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(obs.ComponentLogger("http")))
	router.Use(middleware.Recovery())

	// Health checks
	healthChecker := health.NewChecker()
	healthChecker.Register("database", health.Custom("database", func(ctx context.Context) error {
		return db.Ping(ctx)
	}))
	healthHandler := health.NewHandler(healthChecker)
	healthHandler.RegisterChiRoutes(router)

	// Auth middleware (validates session tokens via Identity context)
	authMiddleware := newAuthMiddleware(identityProvider.SessionLookup, obs.ComponentLogger("auth"))

	// Routes
	router.Route("/api/v1", func(r chi.Router) {
		identityProvider.RegisterRoutes(r, authMiddleware)
	})
	schemaProvider.RegisterRoutes(router)
	credentialProvider.RegisterRoutes(router)
	verificationProvider.RegisterRoutes(router)
	walletProvider.RegisterRoutes(router)
	organizationProvider.RegisterRoutes(router)
	profileProvider.RegisterRoutes(router)
	notificationProvider.RegisterRoutes(router)
	presentationProvider.RegisterRoutes(router)
	trustProvider.RegisterRoutes(router)
	integrationProvider.RegisterRoutes(router)
	issuerProvider.RegisterRoutes(router)
	ledgerProvider.RegisterRoutes(router)

	logger.Info("routes registered")

	// ========================================================================
	// HTTP Server
	// ========================================================================

	serverConfig := pkghttp.DefaultServerConfig()
	serverConfig.Port = port
	serverConfig.Logger = logger

	server := pkghttp.NewServer(router, serverConfig)

	healthChecker.MarkStarted()
	logger.Info("server starting", log.String("address", server.Address()))
	return server.ListenAndServe()
}

// ============================================================================
// Environment Helpers
// ============================================================================

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
