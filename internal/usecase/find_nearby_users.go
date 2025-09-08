package usecase

import (
	"context"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// FindNearbyUsersRequest representa os dados de entrada
type FindNearbyUsersRequest struct {
	UserID     string  `json:"user_id" validate:"required,uuid"`
	Latitude   float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude  float64 `json:"longitude" validate:"required,min=-180,max=180"`
	RadiusM    float64 `json:"radius_meters" validate:"required,min=1,max=50000"` // Máximo 50km
	MaxResults int     `json:"max_results" validate:"min=1,max=100"`              // Máximo 100 resultados
}

// NearbyUserResponse representa um usuário próximo
type NearbyUserResponse struct {
	UserID     string  `json:"user_id"`
	UserName   string  `json:"user_name"`
	PositionID string  `json:"position_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	SectorID   string  `json:"sector_id"`
	DistanceM  float64 `json:"distance_meters"`
	Age        string  `json:"age"` // Ex: "5m30s"
}

// FindNearbyUsersResponse representa a resposta
type FindNearbyUsersResponse struct {
	SearchCenter NearbyUserResponse   `json:"search_center"`
	NearbyUsers  []NearbyUserResponse `json:"nearby_users"`
	TotalFound   int                  `json:"total_found"`
	Message      string               `json:"message"`
}

// FindNearbyUsersUseCase implementa a busca de usuários próximos
type FindNearbyUsersUseCase struct {
	userRepo     repository.UserRepository
	positionRepo repository.PositionRepository
	logger       logger.Logger
}

// NewFindNearbyUsersUseCase cria uma nova instância do use case
func NewFindNearbyUsersUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	logger logger.Logger,
) *FindNearbyUsersUseCase {
	return &FindNearbyUsersUseCase{
		userRepo:     userRepo,
		positionRepo: positionRepo,
		logger:       logger,
	}
}

// Execute executa o use case de buscar usuários próximos
func (uc *FindNearbyUsersUseCase) Execute(ctx context.Context, req FindNearbyUsersRequest) (*FindNearbyUsersResponse, error) {
	// 1. Validar se o usuário existe
	userIDPtr, err := entity.NewUserID(req.UserID)
	if err != nil {
		uc.logger.Error("Invalid user ID", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	userID := *userIDPtr
	_, err = uc.userRepo.FindByID(ctx, userID) // Apenas validar que existe
	if err != nil {
		uc.logger.Error("User not found", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Validar coordenadas de busca
	searchCoordinate, err := valueobject.NewCoordinate(req.Latitude, req.Longitude)
	if err != nil {
		uc.logger.Error("Invalid search coordinates", map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("invalid search coordinates: %w", err)
	}

	// 3. Definir valores padrão
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 20 // Padrão: 20 resultados
	}

	// 4. Buscar posições próximas
	nearbyPositions, err := uc.positionRepo.FindNearby(ctx, searchCoordinate, req.RadiusM, maxResults+1)
	if err != nil {
		uc.logger.Error("Failed to find nearby positions", map[string]interface{}{
			"latitude":    req.Latitude,
			"longitude":   req.Longitude,
			"radius":      req.RadiusM,
			"max_results": maxResults,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to find nearby positions: %w", err)
	}

	// 5. Processar resultados
	var nearbyUsers []NearbyUserResponse
	searchCenterSet := false
	var searchCenter NearbyUserResponse

	for _, position := range nearbyPositions {
		// Buscar dados do usuário
		positionUser, err := uc.userRepo.FindByID(ctx, position.UserID())
		if err != nil {
			positionID := position.ID()
			userIDValue := position.UserID()
			uc.logger.Error("User not found for position", map[string]interface{}{
				"position_id": positionID.String(),
				"user_id":     userIDValue.String(),
			})
			continue
		}

		// Calcular distância
		positionCoordinate := position.Coordinate()
		distance := searchCoordinate.DistanceTo(positionCoordinate)

		// Criar resposta
		userIDValue := positionUser.ID()
		positionIDValue := position.ID()
		nearbyUser := NearbyUserResponse{
			UserID:     userIDValue.String(),
			UserName:   positionUser.Name(),
			PositionID: positionIDValue.String(),
			Latitude:   positionCoordinate.Latitude(),
			Longitude:  positionCoordinate.Longitude(),
			SectorID:   position.Sector().ID(),
			DistanceM:  distance,
			Age:        position.Age().String(),
		}

		// Se é o usuário da busca, definir como centro
		positionUserID := position.UserID()
		if positionUserID.Equals(&userID) && !searchCenterSet {
			searchCenter = nearbyUser
			searchCenterSet = true
		} else {
			nearbyUsers = append(nearbyUsers, nearbyUser)
		}
	}

	// 6. Limitar resultados
	if len(nearbyUsers) > maxResults {
		nearbyUsers = nearbyUsers[:maxResults]
	}

	// 7. Log de sucesso
	uc.logger.Info("Nearby users search completed", map[string]interface{}{
		"user_id":     req.UserID,
		"latitude":    req.Latitude,
		"longitude":   req.Longitude,
		"radius":      req.RadiusM,
		"total_found": len(nearbyUsers),
		"has_center":  searchCenterSet,
	})

	// 8. Retornar resposta
	return &FindNearbyUsersResponse{
		SearchCenter: searchCenter,
		NearbyUsers:  nearbyUsers,
		TotalFound:   len(nearbyUsers),
		Message:      fmt.Sprintf("Found %d users within %.0fm radius", len(nearbyUsers), req.RadiusM),
	}, nil
}
