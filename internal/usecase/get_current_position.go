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
	logger       logger.Logger
}

// NewGetCurrentPositionUseCase cria uma nova instância do use case
func NewGetCurrentPositionUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	logger logger.Logger,
) *GetCurrentPositionUseCase {
	return &GetCurrentPositionUseCase{
		userRepo:     userRepo,
		positionRepo: positionRepo,
		logger:       logger,
	}
}

// Execute executa o use case de buscar posição atual do usuário
func (uc *GetCurrentPositionUseCase) Execute(ctx context.Context, req GetCurrentPositionRequest) (*GetCurrentPositionResponse, error) {
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
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error("User not found", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Buscar posição atual do usuário
	currentPosition, err := uc.positionRepo.FindCurrentByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Current position not found", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("current position not found: %w", err)
	}

	// 3. Preparar resposta
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

	// 4. Log de sucesso
	uc.logger.Info("Current position retrieved", map[string]interface{}{
		"user_id":     req.UserID,
		"position_id": response.PositionID,
		"sector_id":   response.SectorID,
	})

	return response, nil
}
