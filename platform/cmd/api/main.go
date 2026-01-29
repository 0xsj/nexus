package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/credential"
	"github.com/0xsj/nexus/platform/internal/identity"
	identityquery "github.com/0xsj/nexus/platform/internal/identity/application/query"
	"github.com/0xsj/nexus/platform/internal/verification"
	"github.com/0xsj/nexus/platform/internal/verification/infrastructure/oauth"
	"github.com/0xsj/nexus/platform/internal/wallet"
	"github.com/0xsj/nexus/platform/pkg/config"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/database"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/middleware"
	"github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability"
	"github.com/0xsj/nexus/platform/pkg/observability/health"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

var (
	version = "0.1.0"
)

func main() {
	// Load .env file if exists
	_ = config.LoadEnvFileIfExists(".env")

	// Initialize observability provider
	obs := observability.NewBuilder().
		WithLogLevel(config.GetEnv("LOG_LEVEL", "debug")).
		WithColorize(true).
		WithCaller(true).
		WithComponent("api-server").
		Build()

	logger := obs.Logger()

	ctx := context.Background()
	if err := obs.Start(ctx); err != nil {
		logger.Fatal("failed to start observability", log.Err(err))
	}

	// Initialize health checker
	checker := health.NewChecker().
		WithVersion(version).
		WithDefaultTimeout(5 * time.Second)
	checker.Register("self", health.AlwaysUp())

	// Initialize database
	var db *postgres.DB

	if shouldConnectDatabase() {
		logger.Info("connecting to PostgreSQL...")

		dbConfig := buildDatabaseConfig()
		logger.Debug("database config",
			log.String("host", dbConfig.Host),
			log.Int("port", dbConfig.Port),
			log.String("database", dbConfig.Database),
			log.String("user", dbConfig.Username),
		)

		var err error
		db, err = postgres.New(ctx, dbConfig)
		if err != nil {
			logger.Fatal("failed to connect to database", log.Err(err))
		}
		defer db.Close()

		checker.Register("postgres", postgresHealthCheck(db))
		logger.Info("connected to PostgreSQL",
			log.String("host", dbConfig.Host),
			log.Int("port", dbConfig.Port),
			log.String("database", dbConfig.Database),
		)
	} else {
		logger.Info("running without database")
	}

	// ========================================================================
	// Initialize Modules
	// ========================================================================

	// Credential module
	credentialModule, err := credential.NewModuleWithConfig(logger, credential.ModuleConfig{
		EnableSigning: true,
		Database:      db,
	})
	if err != nil {
		logger.Fatal("failed to initialize credential module", log.Err(err))
	}

	// Identity module (requires database)
	var identityModule *identity.Module
	if db != nil {
		identityModule, err = identity.NewModuleWithConfig(logger, identity.ModuleConfig{
			Database:      db,
			Domain:        config.GetEnv("AUTH_DOMAIN", "nexus.io"),
			URI:           config.GetEnv("AUTH_URI", "https://nexus.io"),
			TokenIssuer:   config.GetEnv("TOKEN_ISSUER", "https://nexus.io"),
			TokenAudience: []string{config.GetEnv("TOKEN_AUDIENCE", "https://nexus.io")},
		})
		if err != nil {
			logger.Fatal("failed to initialize identity module", log.Err(err))
		}
	}

	// Wallet module (requires database)
	var walletModule *wallet.Module
	if db != nil {
		var authMiddleware func(http.Handler) http.Handler
		if identityModule != nil {
			authMiddleware = createAuthMiddleware(identityModule, logger)
		}

		walletModule, err = wallet.NewModuleWithConfig(logger, wallet.ModuleConfig{
			Database:       db,
			Domain:         config.GetEnv("AUTH_DOMAIN", "nexus.io"),
			URI:            config.GetEnv("AUTH_URI", "https://nexus.io"),
			ChallengeTTL:   config.GetEnvInt("WALLET_CHALLENGE_TTL", 600),
			NonceTTL:       config.GetEnvInt("WALLET_NONCE_TTL", 600),
			AuthMiddleware: authMiddleware,
		})
		if err != nil {
			logger.Fatal("failed to initialize wallet module", log.Err(err))
		}
		defer walletModule.Stop()
	}

	// Verification module (requires database, credential module, and identity module)
	var verificationModule *verification.Module
	if db != nil && credentialModule != nil && identityModule != nil {
		oauthConfig := buildOAuthConfig()

		if len(oauthConfig.EnabledProviders()) > 0 {
			var authMiddleware func(http.Handler) http.Handler
			authMiddleware = createAuthMiddleware(identityModule, logger)

			userDIDResolver := NewIdentityUserDIDResolver(identityModule.QueryBus())

			verificationModule, err = verification.NewModuleWithConfig(logger, verification.ModuleConfig{
				Database:             db,
				OAuth:                oauthConfig,
				CredentialCommandBus: credentialModule.CommandBus(),
				IssuerDID:            credentialModule.IssuerDID().String(),
				UserDIDResolver:      userDIDResolver,
				VerificationTTL:      config.GetEnvDuration("VERIFICATION_TTL", 15*time.Minute),
				AuthMiddleware:       authMiddleware,
			})
			if err != nil {
				logger.Fatal("failed to initialize verification module", log.Err(err))
			}
			defer verificationModule.Stop()

			logger.Info("verification module initialized",
				log.Int("providers", len(oauthConfig.EnabledProviders())),
			)
		} else {
			logger.Warn("verification module disabled: no OAuth providers configured")
		}
	}

	// ========================================================================
	// Create Router
	// ========================================================================

	router := chi.NewRouter()

	// Global middleware
	router.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.Logger(logger),
		middleware.CORSAllowAll(),
		middleware.Timeout(30*time.Second),
	)

	// Health routes
	healthHandler := health.NewHandler(checker)
	router.Get("/health", healthHandler.Health)
	router.Get("/healthz", healthHandler.Liveness)
	router.Get("/readyz", healthHandler.Readiness)
	router.Get("/startupz", healthHandler.Startup)

	// ========================================================================
	// Mount API Routes
	// ========================================================================

	router.Route("/api/v1", func(r chi.Router) {
		// Root info
		r.Get("/", handleRoot(logger))

		// Credentials: /api/v1/credentials/*
		r.Mount("/credentials", credentialModule.Routes())

		// Identity: /api/v1/auth/*, /api/v1/users/*, /api/v1/sessions/*, etc.
		if identityModule != nil {
			mountIdentityRoutes(r, identityModule)
		}

		// Wallet public: /api/v1/wallet/*
		// Wallet protected: /api/v1/wallets/*
		if walletModule != nil {
			r.Mount("/wallet", walletModule.PublicRoutes())
			r.Mount("/wallets", walletModule.ProtectedRoutes())
		}

		// Verifications: /api/v1/verifications/*
		if verificationModule != nil {
			r.Mount("/verifications", verificationModule.Routes())
		}
	})

	// ========================================================================
	// Start Server
	// ========================================================================

	port := config.GetEnvInt("PORT", 8090)

	serverConfig := pkghttp.ServerConfig{
		Port:            port,
		ReadTimeout:     config.GetEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    config.GetEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     config.GetEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: config.GetEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		Logger:          logger,
	}

	server := pkghttp.NewServer(router, serverConfig)
	checker.MarkStarted()

	storageMode := "none"
	if db != nil {
		storageMode = "PostgreSQL"
	}

	logger.Info("application ready",
		log.Int("port", port),
		log.String("version", version),
		log.String("storage", storageMode),
	)

	printRoutes(logger, verificationModule != nil)

	if err := server.ListenAndServe(); err != nil {
		logger.Error("server error", log.Err(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := obs.Stop(shutdownCtx); err != nil {
		logger.Error("observability shutdown error", log.Err(err))
	}

	logger.Info("server stopped gracefully")
}

// ============================================================================
// Route Mounting Helpers
// ============================================================================

func mountIdentityRoutes(r chi.Router, identityModule *identity.Module) {
	r.Mount("/", identityModule.Routes())
}

// ============================================================================
// Auth Middleware
// ============================================================================

func createAuthMiddleware(identityModule *identity.Module, logger log.Logger) func(http.Handler) http.Handler {
	tokenService := identityModule.TokenService()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, response.ErrUnauthorized("missing authorization header"))
				return
			}

			const bearerPrefix = "Bearer "
			if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				response.Unauthorized(w, response.ErrUnauthorized("invalid authorization header format"))
				return
			}

			token := authHeader[len(bearerPrefix):]

			claims, err := tokenService.ValidateAccessToken(r.Context(), token)
			if err != nil {
				logger.Debug("token validation failed",
					log.Err(err),
					log.String("path", r.URL.Path),
				)
				response.Unauthorized(w, response.ErrUnauthorized("invalid or expired token"))
				return
			}

			// Use the shared log context helper
			ctx := log.ContextWithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ============================================================================
// OAuth Configuration
// ============================================================================

func buildOAuthConfig() oauth.Config {
	cfg := oauth.DefaultConfig()

	cfg.CallbackBaseURL = config.GetEnv("OAUTH_CALLBACK_BASE_URL", "http://localhost:8090")

	// GitHub
	cfg.GitHub.Enabled = config.GetEnv("GITHUB_CLIENT_ID", "") != ""
	cfg.GitHub.ClientID = config.GetEnv("GITHUB_CLIENT_ID", "")
	cfg.GitHub.ClientSecret = config.GetEnv("GITHUB_CLIENT_SECRET", "")

	// LinkedIn
	cfg.LinkedIn.Enabled = config.GetEnv("LINKEDIN_CLIENT_ID", "") != ""
	cfg.LinkedIn.ClientID = config.GetEnv("LINKEDIN_CLIENT_ID", "")
	cfg.LinkedIn.ClientSecret = config.GetEnv("LINKEDIN_CLIENT_SECRET", "")

	return cfg
}

// ============================================================================
// User DID Resolver
// ============================================================================

// IdentityUserDIDResolver implements verification.domain.UserDIDResolver
// by querying the identity module.
type IdentityUserDIDResolver struct {
	queryBus cqrs.QueryBus
}

// NewIdentityUserDIDResolver creates a new resolver.
func NewIdentityUserDIDResolver(queryBus cqrs.QueryBus) *IdentityUserDIDResolver {
	return &IdentityUserDIDResolver{queryBus: queryBus}
}

// ResolvePrimaryDID resolves the user's primary DID.
func (r *IdentityUserDIDResolver) ResolvePrimaryDID(ctx context.Context, userID string) (string, error) {
	result, err := r.queryBus.Dispatch(ctx, &identityquery.GetUser{UserID: userID})
	if err != nil {
		return "", err
	}

	userView, ok := result.(*identityquery.UserView)
	if !ok || userView == nil {
		return "", &UserDIDNotFoundError{UserID: userID}
	}

	if userView.PrimaryDID == "" {
		return "", &UserDIDNotFoundError{UserID: userID}
	}

	return userView.PrimaryDID, nil
}

// ResolveOrCreateDID resolves or creates a DID for a user.
func (r *IdentityUserDIDResolver) ResolveOrCreateDID(ctx context.Context, userID string) (string, error) {
	did, err := r.ResolvePrimaryDID(ctx, userID)
	if err == nil {
		return did, nil
	}
	return "", &UserDIDNotFoundError{UserID: userID}
}

// UserDIDNotFoundError indicates the user's DID could not be resolved.
type UserDIDNotFoundError struct {
	UserID string
}

func (e *UserDIDNotFoundError) Error() string {
	return "could not resolve DID for user: " + e.UserID
}

// ============================================================================
// Database Configuration
// ============================================================================

func shouldConnectDatabase() bool {
	return config.GetEnv("DATABASE_HOST", "") != ""
}

func buildDatabaseConfig() postgres.Config {
	baseConfig := database.DefaultConfig().
		WithHost(config.GetEnv("DATABASE_HOST", "localhost")).
		WithPort(config.GetEnvInt("DATABASE_PORT", 5432)).
		WithDatabase(config.GetEnv("DATABASE_NAME", "nexus")).
		WithCredentials(
			config.GetEnv("DATABASE_USER", "nexus"),
			config.GetEnv("DATABASE_PASSWORD", "nexus"),
		).
		WithSSLMode(config.GetEnv("DATABASE_SSLMODE", "disable"))

	return postgres.NewConfig(baseConfig).
		WithApplicationName("nexus-api")
}

func postgresHealthCheck(db *postgres.DB) health.Check {
	return func(ctx context.Context) *health.Result {
		start := time.Now()
		status := db.HealthCheck(ctx)
		duration := time.Since(start)

		if status.Healthy {
			return health.Up().
				WithDuration(duration).
				WithMessage(status.Message).
				WithDetail("latency_ms", status.LatencyMs).
				WithDetail("open_connections", status.Stats.OpenConnections).
				WithDetail("in_use", status.Stats.InUse).
				WithDetail("idle", status.Stats.Idle)
		}

		return health.Down(nil).
			WithDuration(duration).
			WithMessage(status.Message)
	}
}

// ============================================================================
// Handlers
// ============================================================================

func handleRoot(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{
			"service": "nexus",
			"status":  "running",
			"version": version,
		})
	}
}

func printRoutes(logger log.Logger, verificationEnabled bool) {
	fields := []log.Field{
		log.String("health", "/health, /healthz, /readyz, /startupz"),
		log.String("credentials", "/api/v1/credentials/*"),
		log.String("auth", "/api/v1/auth/*"),
		log.String("users", "/api/v1/users/*"),
		log.String("sessions", "/api/v1/sessions/*"),
		log.String("api-keys", "/api/v1/api-keys/*"),
		log.String("wallet-public", "/api/v1/wallet/*"),
		log.String("wallet-protected", "/api/v1/wallets/*"),
	}

	if verificationEnabled {
		fields = append(fields, log.String("verifications", "/api/v1/verifications/*"))
	}

	logger.Info("routes registered", fields...)
}
