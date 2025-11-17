package di

import (
	"context"

	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"google.golang.org/grpc"
)

// InfrastructureContainer holds core infrastructure dependencies.
// Domains should depend on this, not the other way around.
type InfrastructureContainer struct {
	Logger     logger.Logger
	EventBus   events.EventBus
	GRPCServer *grpc.Server
}

// Close gracefully shuts down infrastructure.
func (c *InfrastructureContainer) Close(ctx context.Context) error {
	// Stop gRPC server gracefully
	if c.GRPCServer != nil {
		c.Logger.Info("Stopping gRPC server...")
		c.GRPCServer.GracefulStop()
	}

	// Close event bus
	if c.EventBus != nil {
		if err := c.EventBus.Close(); err != nil {
			c.Logger.Error("Failed to close event bus", logger.Err(err))
		}
	}

	// Sync logger last
	if c.Logger != nil {
		c.Logger.Sync()
	}

	return nil
}
