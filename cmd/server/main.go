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

	contentv1 "github.com/0xsj/nexus/api/content/v1"
	usersv1 "github.com/0xsj/nexus/api/users/v1"
	"github.com/0xsj/nexus/pkg/observability/logger"
)

func main() {
	ctx := context.Background()

	// Initialize container with Wire
	container, err := InitializeContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer container.Close(ctx)

	log := container.Infrastructure.Logger
	grpcServer := container.Infrastructure.GRPCServer

	log.Info("Nexus server starting...")

	// Register gRPC services
	usersv1.RegisterUserServiceServer(grpcServer, container.UserHandler)
	contentv1.RegisterContentServiceServer(grpcServer, container.ContentHandler)

	log.Info("gRPC services registered")

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatal("Failed to listen",
			logger.Err(err),
			logger.String("port", grpcPort),
		)
	}

	// Start server in goroutine
	go func() {
		log.Info("gRPC server listening",
			logger.String("port", grpcPort),
		)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("Failed to serve gRPC",
				logger.Err(err),
			)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Warn("Shutdown timeout, forcing stop...")
		grpcServer.Stop()
	case <-stopped:
		log.Info("Server stopped gracefully")
	}
}
