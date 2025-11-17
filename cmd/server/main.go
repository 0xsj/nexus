// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/0xsj/nexus/pkg/config"
)

// Config holds server configuration
type Config struct {
	GRPCPort string
	HTTPPort string
	Env      string
}

func main() {
	// Load configuration
	cfg := loadConfig()

	// Setup logging
	logger := log.New(os.Stdout, "[nexus] ", log.LstdFlags|log.Lshortfile)
	logger.Printf("Starting Nexus server in %s mode", cfg.Env)

	// Create gRPC server
	grpcServer := createGRPCServer(logger)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service (for grpcurl)
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	// Start server in goroutine
	go func() {
		logger.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mark health check as not serving
	healthServer.Shutdown()

	// Stop accepting new connections and wait for existing ones to finish
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-shutdownCtx.Done():
		logger.Println("Shutdown timeout, forcing stop...")
		grpcServer.Stop()
	case <-stopped:
		logger.Println("Server stopped gracefully")
	}
}

func loadConfig() Config {
	return Config{
		GRPCPort: config.GetEnv("GRPC_PORT", "9090"),
		HTTPPort: config.GetEnv("HTTP_PORT", "8080"),
		Env:      config.GetEnv("ENV", "development"),
	}
}

func createGRPCServer(logger *log.Logger) *grpc.Server {
	// Create server with interceptors for logging, auth, etc.
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(logger),
		),
	}

	return grpc.NewServer(opts...)
}

// loggingInterceptor logs all gRPC requests
func loggingInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.Printf("RPC: %s | Duration: %v | Error: %v", info.FullMethod, duration, err)
		} else {
			logger.Printf("RPC: %s | Duration: %v | Success", info.FullMethod, duration)
		}

		return resp, err
	}
}
