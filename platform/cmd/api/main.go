package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/credential"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/middleware"
	"github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability"
	"github.com/0xsj/nexus/platform/pkg/observability/health"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

var (
	version   = "0.1.0"
	buildTime = "unknown"
)

func main() {
	// Initialize observability provider
	obs := observability.NewBuilder().
		WithLogLevel("debug").
		WithColorize(true).
		WithCaller(true).
		WithComponent("api-server").
		Build()

	// Get logger
	logger := obs.Logger()

	// Start observability
	ctx := context.Background()
	if err := obs.Start(ctx); err != nil {
		logger.Fatal("failed to start observability", log.Err(err))
	}

	// Initialize health checker
	checker := health.NewChecker().
		WithVersion(version).
		WithDefaultTimeout(5 * time.Second)

	// Register health checks
	checker.Register("self", health.AlwaysUp())

	// Create health handler
	healthHandler := health.NewHandler(checker)

	// Initialize credential module
	credentialModule, err := credential.NewModule(logger)
	if err != nil {
		logger.Fatal("failed to initialize credential module", log.Err(err))
	}

	// Get port from environment
	port := getPort()

	// Create router
	router := chi.NewRouter()

	// Apply global middleware
	router.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.Logger(logger),
		middleware.CORSAllowAll(), // TODO: Configure for production
		middleware.Timeout(30*time.Second),
	)

	// Health routes
	router.Get("/health", healthHandler.Health)
	router.Get("/healthz", healthHandler.Liveness)
	router.Get("/readyz", healthHandler.Readiness)
	router.Get("/startupz", healthHandler.Startup)

	// API routes
	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/", handleRoot(logger))

			// Mount credential routes
			r.Mount("/", credentialModule.Routes())
		})
	})

	// Create server config
	serverConfig := pkghttp.ServerConfig{
		Port:            port,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		Logger:          logger,
	}

	// Create server
	server := pkghttp.NewServer(router, serverConfig)

	// Mark as started
	checker.MarkStarted()

	logger.Info("application ready",
		log.Int("port", port),
		log.String("version", version),
		log.String("health", "/health"),
		log.String("liveness", "/healthz"),
		log.String("readiness", "/readyz"),
		log.String("credentials", "/api/v1/credentials"),
	)

	// Start server with graceful shutdown
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server error", log.Err(err))
	}

	// Stop observability
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := obs.Stop(shutdownCtx); err != nil {
		logger.Error("observability shutdown error", log.Err(err))
	}

	logger.Info("server stopped gracefully")
}

// getPort returns the port from environment or default.
func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8090
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 8090
	}

	return port
}

// handleRoot handles the root endpoint.
func handleRoot(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())

		logger.Debug("root endpoint",
			log.String("request_id", requestID),
			log.String("remote_addr", r.RemoteAddr),
		)

		response.OK(w, map[string]string{
			"service": "nexus",
			"status":  "running",
			"version": version,
		})
	}
}
