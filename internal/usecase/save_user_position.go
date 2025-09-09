package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/events"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// SaveUserPositionRequest representa os dados de entrada para salvar posição
type SaveUserPositionRequest struct {
	UserID    string    `json:"user_id" validate:"required,uuid"`
	Latitude  float64   `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64   `json:"longitude" validate:"required,min=-180,max=180"`
	Timestamp time.Time `json:"timestamp"`
}

// SaveUserPositionResponse representa a resposta
type SaveUserPositionResponse struct {
	PositionID string `json:"position_id"`
	SectorID   string `json:"sector_id"`
	Message    string `json:"message"`
}

// SaveUserPositionUseCase implementa a lógica de negócio para salvar posições
type SaveUserPositionUseCase struct {
	userRepo       repository.UserRepository
	positionRepo   repository.PositionRepository
	eventPublisher events.Publisher
	logger         logger.Logger
}

// NewSaveUserPositionUseCase cria uma nova instância do use case
func NewSaveUserPositionUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	eventPublisher events.Publisher,
	logger logger.Logger,
) *SaveUserPositionUseCase {
	return &SaveUserPositionUseCase{
		userRepo:       userRepo,
		positionRepo:   positionRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Execute executa o use case de salvar posição do usuário
func (uc *SaveUserPositionUseCase) Execute(ctx context.Context, req SaveUserPositionRequest) (*SaveUserPositionResponse, error) {
	// 1. Criar UserID e validar se o usuário existe
	userIDPtr, err := entity.NewUserID(req.UserID)
	if err != nil {
		uc.logger.Error("Invalid user ID", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	userID := *userIDPtr // Desreferencia o ponteiro
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error("User not found", map[string]interface{}{
			"user_id": req.UserID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Criar coordenada e validar
	coordinate, err := valueobject.NewCoordinate(req.Latitude, req.Longitude)
	if err != nil {
		uc.logger.Error("Invalid coordinates", map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	// 3. Usar timestamp atual se não fornecido
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// 4. Criar nova posição
	positionID := uuid.New().String()
	position, err := entity.NewPosition(
		positionID,
		user.ID(),
		coordinate.Latitude(),
		coordinate.Longitude(),
		timestamp,
	)
	if err != nil {
		uc.logger.Error("Failed to create position", map[string]interface{}{
			"user_id": user.ID(),
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	// 5. Buscar posição anterior para comparação (para eventos)
	var previousPosition *entity.Position
	previousPosition, _ = uc.positionRepo.FindCurrentByUserID(ctx, userID)
	// Não retornamos erro se não encontrar posição anterior (usuário novo)

	// 6. Salvar posição no repositório
	if err := uc.positionRepo.Save(ctx, position); err != nil {
		uc.logger.Error("Failed to save position", map[string]interface{}{
			"position_id": position.ID(),
			"user_id":     user.ID(),
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to save position: %w", err)
	}

	// 7. Publicar evento de mudança de posição
	if err := uc.publishPositionChangedEvent(ctx, user, position, previousPosition); err != nil {
		// Log error mas não falha a operação (evento é secundário)
		uc.logger.Error("Failed to publish position changed event",
			"position_id", position.ID(),
			"user_id", user.ID(),
			"error", err.Error(),
		)
	}

	// 8. Log de sucesso
	uc.logger.Info("Position saved successfully", map[string]interface{}{
		"position_id": position.ID(),
		"user_id":     user.ID(),
		"sector":      position.Sector().ID(),
		"latitude":    coordinate.Latitude(),
		"longitude":   coordinate.Longitude(),
	})

	// 9. Retornar resposta
	positionIDEntity := position.ID()
	return &SaveUserPositionResponse{
		PositionID: positionIDEntity.String(),
		SectorID:   position.Sector().ID(),
		Message:    "Position saved successfully",
	}, nil
}

// publishPositionChangedEvent publica evento quando posição do usuário muda
func (uc *SaveUserPositionUseCase) publishPositionChangedEvent(
	ctx context.Context,
	user *entity.User,
	newPosition *entity.Position,
	previousPosition *entity.Position,
) error {
	// Preparar dados do evento
	var previousLat, previousLng float64
	var previousSector string
	var distanceMoved float64

	if previousPosition != nil {
		previousLat = previousPosition.Coordinate().Latitude()
		previousLng = previousPosition.Coordinate().Longitude()
		previousSector = previousPosition.Sector().ID()

		// Calcular distância movida
		distanceMoved = valueobject.CalculateDistance(
			previousLat, previousLng,
			newPosition.Coordinate().Latitude(), newPosition.Coordinate().Longitude(),
		)
	}

	// Criar dados do evento
	positionID := newPosition.ID()
	userID := user.ID()

	eventData := events.PositionChangedData{
		PositionID:     positionID.String(),
		PreviousLat:    previousLat,
		PreviousLng:    previousLng,
		NewLat:         newPosition.Coordinate().Latitude(),
		NewLng:         newPosition.Coordinate().Longitude(),
		PreviousSector: previousSector,
		NewSector:      newPosition.Sector().ID(),
		DistanceMoved:  distanceMoved,
	}

	// Criar evento
	event := events.NewPositionChangedEvent(
		userID.String(),
		"default-event", // TODO: pegar do contexto do evento
		eventData,
	)

	// Publicar evento
	return uc.eventPublisher.PublishPositionChanged(ctx, event)
}
