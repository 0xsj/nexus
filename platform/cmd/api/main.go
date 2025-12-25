package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// In a real app, you'd pass actual database/redis connections
	checker.Register("self", health.AlwaysUp())

	// Example: Register a TCP check (e.g., for an external service)
	// checker.Register("redis", health.TCP("localhost", 6379))

	// Example: Register a database check
	// checker.Register("database", health.Database(db))

	// Example: Register an HTTP endpoint check
	// checker.Register("external-api", health.HTTPEndpoint("https://api.example.com/health"))

	// Example: Register a non-critical check
	// checker.RegisterNonCritical("cache", health.Redis(redisClient))

	// Create health handler
	healthHandler := health.NewHandler(checker)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	mux := http.NewServeMux()

	// Register health routes
	healthHandler.RegisterRoutes(mux)

	// Application routes
	mux.HandleFunc("GET /", handleRoot(logger))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting server",
			log.String("port", port),
			log.String("version", version),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", log.Err(err))
		}
	}()

	// Mark as started after server begins listening
	// In production, you might wait for DB connections etc.
	checker.MarkStarted()
	logger.Info("application ready",
		log.String("health", "/health"),
		log.String("liveness", "/healthz"),
		log.String("readiness", "/readyz"),
		log.String("startup", "/startupz"),
	)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("received shutdown signal",
		log.String("signal", sig.String()),
	)

	// Mark as not ready for new traffic
	checker.MarkNotStarted()

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("shutting down server")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", log.Err(err))
		os.Exit(1)
	}

	// Stop observability
	if err := obs.Stop(shutdownCtx); err != nil {
		logger.Error("observability shutdown error", log.Err(err))
	}

	logger.Info("server stopped gracefully")
}

func handleRoot(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("root endpoint",
			log.String("remote_addr", r.RemoteAddr),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"nexus","status":"running"}`))
	}
}