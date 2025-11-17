//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/0xsj/nexus/internal/content"
	"github.com/0xsj/nexus/internal/users"
	"github.com/0xsj/nexus/pkg/di"
	"github.com/google/wire"
)

// InitializeContainer creates the application container.
func InitializeContainer(ctx context.Context) (*Container, error) {
	wire.Build(
		// Infrastructure
		di.InitializeInfrastructure,

		// Extract fields from InfrastructureContainer for domain providers
		wire.FieldsOf(new(*di.InfrastructureContainer), "Logger", "EventBus"),

		// Domain providers
		users.ProviderSet,
		content.ProviderSet,

		// Container
		wire.Struct(new(Container), "*"),
	)
	return nil, nil
}
