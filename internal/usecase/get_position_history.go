package usecase

import (
	"context"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// GetPositionHistoryRequest representa os dados de entrada
type GetPositionHistoryRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
}

// PositionHistoryItem representa um item do histórico
type PositionHistoryItem struct {
	PositionID string  `json:"position_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	SectorID   string  `json:"sector_id"`
	Age        string  `json:"age"`
	RecordedAt string  `json:"recorded_at"`
}

// GetPositionHistoryResponse representa a resposta
type GetPositionHistoryResponse struct {
	UserID   string                `json:"user_id"`
	UserName string                `json:"user_name"`
	History  []PositionHistoryItem `json:"history"`
	Total    int                   `json:"total"`
	Message  string                `json:"message"`
}

// GetPositionHistoryUseCase implementa a busca do histórico de posições
type GetPositionHistoryUseCase struct {
	userRepo     repository.UserRepository
	positionRepo repository.PositionRepository
	logger       logger.Logger
}

// NewGetPositionHistoryUseCase cria uma nova instância do use case
func NewGetPositionHistoryUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	logger logger.Logger,
) *GetPositionHistoryUseCase {
	return &GetPositionHistoryUseCase{
		userRepo:     userRepo,
		positionRepo: positionRepo,
		logger:       logger,
	}
}

// Execute executa o use case de buscar histórico de posições
func (uc *GetPositionHistoryUseCase) Execute(ctx context.Context, req GetPositionHistoryRequest) (*GetPositionHistoryResponse, error) {
	// 1. Validar parâmetros
	if req.Limit <= 0 {
		req.Limit = 10 // Padrão: 10 posições
	}
	if req.Limit > 100 {
		req.Limit = 100 // Máximo: 100 posições
	}

	// 2. Validar se o usuário existe
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

	// 3. Buscar histórico de posições
	positions, err := uc.positionRepo.FindHistoryByUserID(ctx, userID, req.Limit)
	if err != nil {
		uc.logger.Error("Failed to get position history", map[string]interface{}{
			"user_id": req.UserID,
			"limit":   req.Limit,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to get position history: %w", err)
	}

	// 4. Converter para resposta
	var history []PositionHistoryItem
	for _, position := range positions {
		coordinate := position.Coordinate()
		positionIDValue := position.ID()
		recordedAt := position.RecordedAt()

		item := PositionHistoryItem{
			PositionID: positionIDValue.String(),
			Latitude:   coordinate.Latitude(),
			Longitude:  coordinate.Longitude(),
			SectorID:   position.Sector().ID(),
			Age:        position.Age().String(),
			RecordedAt: recordedAt.String(),
		}
		history = append(history, item)
	}

	// 5. Preparar resposta
	userIDValue := user.ID()
	response := &GetPositionHistoryResponse{
		UserID:   userIDValue.String(),
		UserName: user.Name(),
		History:  history,
		Total:    len(history),
		Message:  fmt.Sprintf("Retrieved %d position records", len(history)),
	}

	// 6. Log de sucesso
	uc.logger.Info("Position history retrieved", map[string]interface{}{
		"user_id": req.UserID,
		"total":   len(history),
		"limit":   req.Limit,
	})

	return response, nil
}
