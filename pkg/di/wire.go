//go:build wireinject
// +build wireinject

package di

import (
	"context"

	"github.com/google/wire"
)

// InfrastructureProviderSet provides core infrastructure dependencies.
var InfrastructureProviderSet = wire.NewSet(
	ProvideLoggerConfig,
	ProvideLogger,
	ProvideEventBus,
	ProvideGRPCServer,
)

// InitializeInfrastructure creates infrastructure container.
func InitializeInfrastructure(ctx context.Context) (*InfrastructureContainer, error) {
	wire.Build(
		InfrastructureProviderSet,
		wire.Struct(new(InfrastructureContainer), "*"),
	)
	return nil, nil
}
