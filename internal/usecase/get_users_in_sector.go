package usecase

import (
	"context"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// GetUsersInSectorRequest representa os dados de entrada
type GetUsersInSectorRequest struct {
	UserID    string  `json:"user_id" validate:"required,uuid"`
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// SectorUserResponse representa um usuário no setor
type SectorUserResponse struct {
	UserID     string  `json:"user_id"`
	UserName   string  `json:"user_name"`
	PositionID string  `json:"position_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Age        string  `json:"age"` // Ex: "5m30s"
}

// GetUsersInSectorResponse representa a resposta
type GetUsersInSectorResponse struct {
	SectorID      string               `json:"sector_id"`
	SectorBounds  SectorBounds         `json:"sector_bounds"`
	RequestedBy   SectorUserResponse   `json:"requested_by"`
	UsersInSector []SectorUserResponse `json:"users_in_sector"`
	TotalFound    int                  `json:"total_found"`
	Message       string               `json:"message"`
}

// SectorBounds representa os limites do setor
type SectorBounds struct {
	MinLatitude  float64 `json:"min_latitude"`
	MaxLatitude  float64 `json:"max_latitude"`
	MinLongitude float64 `json:"min_longitude"`
	MaxLongitude float64 `json:"max_longitude"`
}

// GetUsersInSectorUseCase implementa a busca de usuários no mesmo setor
type GetUsersInSectorUseCase struct {
	userRepo     repository.UserRepository
	positionRepo repository.PositionRepository
	cache        CacheInterface
	logger       logger.Logger
}

// NewGetUsersInSectorUseCase cria uma nova instância do use case
func NewGetUsersInSectorUseCase(
	userRepo repository.UserRepository,
	positionRepo repository.PositionRepository,
	cache CacheInterface,
	logger logger.Logger,
) *GetUsersInSectorUseCase {
	return &GetUsersInSectorUseCase{
		userRepo:     userRepo,
		positionRepo: positionRepo,
		cache:        cache,
		logger:       logger,
	}
}

// Execute executa o use case de buscar usuários no mesmo setor
func (uc *GetUsersInSectorUseCase) Execute(ctx context.Context, req GetUsersInSectorRequest) (*GetUsersInSectorResponse, error) {
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

	// 2. Validar coordenadas e calcular setor
	coordinate, err := valueobject.NewCoordinate(req.Latitude, req.Longitude)
	if err != nil {
		uc.logger.Error("Invalid coordinates", map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	// 3. Calcular setor a partir das coordenadas
	sector, err := valueobject.NewSectorFromCoordinate(coordinate)
	if err != nil {
		uc.logger.Error("Failed to create sector", map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to create sector: %w", err)
	}

	// 4. Buscar todas as posições no setor
	sectorPositions, err := uc.positionRepo.FindInSector(ctx, sector)
	if err != nil {
		uc.logger.Error("Failed to find positions in sector", map[string]interface{}{
			"sector_id": sector.ID(),
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to find positions in sector: %w", err)
	}

	// 5. Processar resultados
	var usersInSector []SectorUserResponse
	var requestedBy SectorUserResponse
	requestedBySet := false

	for _, position := range sectorPositions {
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

		// Criar resposta do usuário
		positionCoordinate := position.Coordinate()
		userIDValue := positionUser.ID()
		positionIDValue := position.ID()
		sectorUser := SectorUserResponse{
			UserID:     userIDValue.String(),
			UserName:   positionUser.Name(),
			PositionID: positionIDValue.String(),
			Latitude:   positionCoordinate.Latitude(),
			Longitude:  positionCoordinate.Longitude(),
			Age:        position.Age().String(),
		}

		// Se é o usuário que fez a requisição
		positionUserID := position.UserID()
		if positionUserID.Equals(&userID) && !requestedBySet {
			requestedBy = sectorUser
			requestedBySet = true
		} else {
			usersInSector = append(usersInSector, sectorUser)
		}
	}

	// 6. Calcular bounds do setor
	bounds := uc.calculateSectorBounds(sector)

	// 7. Log de sucesso
	uc.logger.Info("Sector users search completed", map[string]interface{}{
		"user_id":          req.UserID,
		"sector_id":        sector.ID(),
		"total_found":      len(usersInSector),
		"requested_by_set": requestedBySet,
	})

	// 8. Retornar resposta
	return &GetUsersInSectorResponse{
		SectorID:      sector.ID(),
		SectorBounds:  bounds,
		RequestedBy:   requestedBy,
		UsersInSector: usersInSector,
		TotalFound:    len(usersInSector),
		Message:       fmt.Sprintf("Found %d users in sector %s", len(usersInSector), sector.ID()),
	}, nil
}

// calculateSectorBounds calcula os limites geográficos do setor
func (uc *GetUsersInSectorUseCase) calculateSectorBounds(sector *valueobject.Sector) SectorBounds {
	// Cada setor representa um quadrado de 100x100 metros
	// Aqui calculamos os bounds aproximados baseados no centro do setor

	// Para simplificar, vamos usar uma aproximação
	// 1 grau de latitude ≈ 111.000 metros
	// 1 grau de longitude ≈ 111.000 * cos(latitude) metros

	deltaLat := 50.0 / 111000.0 // 50 metros em graus (raio do setor)
	deltaLng := 50.0 / 111000.0 // Aproximação simples

	// Coordenadas do centro do setor (aproximadas)
	centerLat := float64(sector.Y()) * 0.001 // Conversão simplificada
	centerLng := float64(sector.X()) * 0.001

	return SectorBounds{
		MinLatitude:  centerLat - deltaLat,
		MaxLatitude:  centerLat + deltaLat,
		MinLongitude: centerLng - deltaLng,
		MaxLongitude: centerLng + deltaLng,
	}
}
