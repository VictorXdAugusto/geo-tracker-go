package wire

import (
	"github.com/google/wire"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/database"
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
