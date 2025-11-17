package main

import (
	"context"

	contentgrpc "github.com/0xsj/nexus/internal/content/presentation/grpc"
	usergrpc "github.com/0xsj/nexus/internal/users/presentation/grpc"
	"github.com/0xsj/nexus/pkg/di"
)

// Container holds all application dependencies.
type Container struct {
	Infrastructure *di.InfrastructureContainer

	// Domain handlers
	UserHandler    *usergrpc.UserHandler
	ContentHandler *contentgrpc.ContentHandler
}

// Close gracefully shuts down the application.
func (c *Container) Close(ctx context.Context) error {
	return c.Infrastructure.Close(ctx)
}
