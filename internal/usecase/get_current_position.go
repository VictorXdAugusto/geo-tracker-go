package usecase

import (
	"context"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// GetCurrentPositionRequest representa os dados de entrada
type GetCurrentPositionRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// GetCurrentPositionResponse representa a resposta
type GetCurrentPositionResponse struct {
	UserID     string  `json:"user_id"`
	UserName   string  `json:"user_name"`
	PositionID string  `json:"position_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	SectorID   string  `json:"sector_id"`
	Age        string  `json:"age"`
	Message    string  `json:"message"`
}

// GetCurrentPositionUseCase implementa a busca da posição atual do usuário
type GetCurrentPositionUseCase struct {
	userRepo     repository.UserRepository
	positionRepo repository.PositionRepository
	cache        CacheInterface
	logger       logger.Logger
}

// NewGetCurrentPositionUseCase cria uma nova instância do use case
func NewGetCurrentPositionUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	cache CacheInterface,
	logger logger.Logger,
) *GetCurrentPositionUseCase {
	return &GetCurrentPositionUseCase{
		userRepo:     userRepo,
		positionRepo: positionRepo,
		cache:        cache,
		logger:       logger,
	}
}

// Execute executa o use case de buscar posição atual do usuário
func (uc *GetCurrentPositionUseCase) Execute(ctx context.Context, req GetCurrentPositionRequest) (*GetCurrentPositionResponse, error) {
	// 1. Tentar buscar no cache primeiro
	var cachedResponse GetCurrentPositionResponse
	if err := uc.cache.GetCachedUserPosition(ctx, req.UserID, &cachedResponse); err == nil {
		uc.logger.Info("Cache hit for current position", map[string]interface{}{
			"user_id":     req.UserID,
			"position_id": cachedResponse.PositionID,
			"source":      "cache",
		})
		return &cachedResponse, nil
	}

	// 2. Cache miss - buscar dados completos
	userIDPtr, err := entity.NewUserID(req.UserID)
	if err != nil {
		uc.logger.Error("Invalid user ID", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	userID := *userIDPtr
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error("User not found", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 3. Buscar posição atual do usuário
	currentPosition, err := uc.positionRepo.FindCurrentByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Current position not found", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("current position not found: %w", err)
	}

	// 4. Preparar resposta
	coordinate := currentPosition.Coordinate()
	userIDValue := user.ID()
	positionIDValue := currentPosition.ID()

	response := &GetCurrentPositionResponse{
		UserID:     userIDValue.String(),
		UserName:   user.Name(),
		PositionID: positionIDValue.String(),
		Latitude:   coordinate.Latitude(),
		Longitude:  coordinate.Longitude(),
		SectorID:   currentPosition.Sector().ID(),
		Age:        currentPosition.Age().String(),
		Message:    "Current position retrieved successfully",
	}

	// 5. Salvar no cache para próximas consultas
	if cacheErr := uc.cache.CacheUserPosition(ctx, req.UserID, response); cacheErr != nil {
		uc.logger.Error("Failed to cache user position", map[string]interface{}{
			"user_id": req.UserID,
			"error":   cacheErr.Error(),
		})
		// Não falhar a operação por erro de cache
	}

	// 6. Log de sucesso
	uc.logger.Info("Current position retrieved from database", map[string]interface{}{
		"user_id":     req.UserID,
		"position_id": response.PositionID,
		"sector_id":   response.SectorID,
		"source":      "database",
	})

	return response, nil
}
