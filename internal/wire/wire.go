//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/cache"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/database"
)

// InitializeContainer inicializa todo o container de use cases
func InitializeContainer() (*Container, error) {
	wire.Build(
		ApplicationSet,
		NewContainer,
	)
	return nil, nil
}

// InitializeDatabase inicializa apenas o banco de dados
func InitializeDatabase() (*database.DB, error) {
	wire.Build(InfrastructureSet)
	return nil, nil
}

// InitializeRedis inicializa apenas o Redis
func InitializeRedis() (*cache.Redis, error) {
	wire.Build(InfrastructureSet)
	return nil, nil
}
