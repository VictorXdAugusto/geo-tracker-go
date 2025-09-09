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
	cache        CacheInterface
	logger       logger.Logger
}

// NewFindNearbyUsersUseCase cria uma nova instância do use case
func NewFindNearbyUsersUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	cache CacheInterface,
	logger logger.Logger,
) *FindNearbyUsersUseCase {
	return &FindNearbyUsersUseCase{
		userRepo:     userRepo,
		positionRepo: positionRepo,
		cache:        cache,
		logger:       logger,
	}
}

// Execute executa o use case de buscar usuários próximos
func (uc *FindNearbyUsersUseCase) Execute(ctx context.Context, req FindNearbyUsersRequest) (*FindNearbyUsersResponse, error) {
	// 1. Tentar buscar no cache primeiro (apenas para coordenadas fixas, sem considerar user_id)
	var cachedResponse FindNearbyUsersResponse
	if err := uc.cache.GetCachedNearbyUsers(ctx, req.Latitude, req.Longitude, req.RadiusM, &cachedResponse); err == nil {
		// Ajustar o search center para o usuário atual se ele estiver nos resultados
		searchCenter, nearbyUsers := uc.adjustSearchCenterFromCache(cachedResponse, req.UserID)

		response := &FindNearbyUsersResponse{
			SearchCenter: searchCenter,
			NearbyUsers:  nearbyUsers,
			TotalFound:   len(nearbyUsers),
			Message:      fmt.Sprintf("Found %d users within %.0fm radius", len(nearbyUsers), req.RadiusM),
		}

		uc.logger.Info("Cache hit for nearby users search", map[string]interface{}{
			"user_id":     req.UserID,
			"latitude":    req.Latitude,
			"longitude":   req.Longitude,
			"radius":      req.RadiusM,
			"total_found": len(nearbyUsers),
			"source":      "cache",
		})

		return response, nil
	}

	// 2. Cache miss - executar busca completa
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

	// 3. Validar coordenadas de busca
	searchCoordinate, err := valueobject.NewCoordinate(req.Latitude, req.Longitude)
	if err != nil {
		uc.logger.Error("Invalid search coordinates", map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("invalid search coordinates: %w", err)
	}

	// 4. Definir valores padrão
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 20 // Padrão: 20 resultados
	}

	// 5. Buscar posições próximas
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

	// 6. Processar resultados
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

	// 7. Limitar resultados
	if len(nearbyUsers) > maxResults {
		nearbyUsers = nearbyUsers[:maxResults]
	}

	// 8. Preparar resposta para cache
	response := &FindNearbyUsersResponse{
		SearchCenter: searchCenter,
		NearbyUsers:  nearbyUsers,
		TotalFound:   len(nearbyUsers),
		Message:      fmt.Sprintf("Found %d users within %.0fm radius", len(nearbyUsers), req.RadiusM),
	}

	// 9. Salvar no cache (sem o search center específico, para reutilização)
	cacheableResponse := FindNearbyUsersResponse{
		NearbyUsers: append(nearbyUsers, searchCenter), // Incluir todos os usuários
		TotalFound:  len(nearbyUsers) + 1,
		Message:     response.Message,
	}
	if cacheErr := uc.cache.CacheNearbyUsers(ctx, req.Latitude, req.Longitude, req.RadiusM, cacheableResponse); cacheErr != nil {
		uc.logger.Error("Failed to cache nearby users", map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"radius":    req.RadiusM,
			"error":     cacheErr.Error(),
		})
		// Não falhar a operação por erro de cache
	}

	// 10. Log de sucesso
	uc.logger.Info("Nearby users search completed from database", map[string]interface{}{
		"user_id":     req.UserID,
		"latitude":    req.Latitude,
		"longitude":   req.Longitude,
		"radius":      req.RadiusM,
		"total_found": len(nearbyUsers),
		"has_center":  searchCenterSet,
		"source":      "database",
	})

	return response, nil
}

// adjustSearchCenterFromCache ajusta o search center baseado no usuário atual
func (uc *FindNearbyUsersUseCase) adjustSearchCenterFromCache(cachedResponse FindNearbyUsersResponse, userID string) (NearbyUserResponse, []NearbyUserResponse) {
	var searchCenter NearbyUserResponse
	var nearbyUsers []NearbyUserResponse

	// Procurar o usuário nos resultados cached
	for _, user := range cachedResponse.NearbyUsers {
		if user.UserID == userID {
			searchCenter = user
		} else {
			nearbyUsers = append(nearbyUsers, user)
		}
	}

	return searchCenter, nearbyUsers
}
