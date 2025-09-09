package wire

import (
	"github.com/google/wire"
	"github.com/vitao/geolocation-tracker/internal/domain/events"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/cache"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/database"
	infraEvents "github.com/vitao/geolocation-tracker/internal/infrastructure/events"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/pkg/config"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// Infrastructure Providers
var InfrastructureSet = wire.NewSet(
	// Config and Logger
	config.Load,
	logger.NewLogger,

	// Database
	database.New,
	database.NewUserRepository,
	database.NewPositionRepository,

	// Redis and Events
	cache.NewRedis,
	NewCacheInterface,
	NewRedisEventPublisher,
)

// UseCase Providers
var UseCaseSet = wire.NewSet(
	usecase.NewCreateUserUseCase,
	usecase.NewSaveUserPositionUseCase,
	usecase.NewFindNearbyUsersUseCase,
	usecase.NewGetUsersInSectorUseCase,
	usecase.NewGetCurrentPositionUseCase,
	usecase.NewGetPositionHistoryUseCase,
)

// Complete Application Set
var ApplicationSet = wire.NewSet(
	InfrastructureSet,
	UseCaseSet,
)

// NewRedisEventPublisher cria um novo publisher usando Redis client
func NewRedisEventPublisher(redis *cache.Redis, logger logger.Logger) events.Publisher {
	return infraEvents.NewRedisStreamPublisher(redis.Client(), logger)
}

// NewCacheInterface converte *cache.Redis para usecase.CacheInterface
func NewCacheInterface(redis *cache.Redis) usecase.CacheInterface {
	return redis
}
