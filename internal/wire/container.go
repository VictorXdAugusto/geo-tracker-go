package wire

import (
	"github.com/vitao/geolocation-tracker/internal/usecase"
)

// Container agrupa todos os use cases da aplicação
type Container struct {
	CreateUser         *usecase.CreateUserUseCase
	SaveUserPosition   *usecase.SaveUserPositionUseCase
	FindNearbyUsers    *usecase.FindNearbyUsersUseCase
	GetUsersInSector   *usecase.GetUsersInSectorUseCase
	GetCurrentPosition *usecase.GetCurrentPositionUseCase
	GetPositionHistory *usecase.GetPositionHistoryUseCase
}

// NewContainer cria um novo container com todos os use cases
func NewContainer(
	createUser *usecase.CreateUserUseCase,
	saveUserPosition *usecase.SaveUserPositionUseCase,
	findNearbyUsers *usecase.FindNearbyUsersUseCase,
	getUsersInSector *usecase.GetUsersInSectorUseCase,
	getCurrentPosition *usecase.GetCurrentPositionUseCase,
	getPositionHistory *usecase.GetPositionHistoryUseCase,
) *Container {
	return &Container{
		CreateUser:         createUser,
		SaveUserPosition:   saveUserPosition,
		FindNearbyUsers:    findNearbyUsers,
		GetUsersInSector:   getUsersInSector,
		GetCurrentPosition: getCurrentPosition,
		GetPositionHistory: getPositionHistory,
	}
}
