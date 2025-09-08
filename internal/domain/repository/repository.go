package repository

import (
	"context"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
)

// UserRepository define operações de persistência para usuários
// Interface = contrato, não implementação
// Seguindo Repository Pattern + Dependency Inversion Principle
type UserRepository interface {
	// Save persiste um usuário (create ou update)
	Save(ctx context.Context, user *entity.User) error

	// FindByID busca usuário por ID
	FindByID(ctx context.Context, id entity.UserID) (*entity.User, error)

	// FindByEmail busca usuário por email
	FindByEmail(ctx context.Context, email entity.Email) (*entity.User, error)

	// Exists verifica se usuário existe
	Exists(ctx context.Context, id entity.UserID) (bool, error)

	// Delete remove usuário
	Delete(ctx context.Context, id entity.UserID) error

	// FindAll retorna todos os usuários (com paginação)
	FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error)
}

// PositionRepository define operações de persistência para posições
type PositionRepository interface {
	// Save persiste uma posição
	Save(ctx context.Context, position *entity.Position) error

	// FindByID busca posição por ID
	FindByID(ctx context.Context, id entity.PositionID) (*entity.Position, error)

	// FindCurrentByUserID busca posição atual de um usuário
	FindCurrentByUserID(ctx context.Context, userID entity.UserID) (*entity.Position, error)

	// FindHistoryByUserID busca histórico de posições de um usuário
	FindHistoryByUserID(ctx context.Context, userID entity.UserID, limit int) ([]*entity.Position, error)

	// FindNearby busca posições próximas a uma coordenada
	FindNearby(ctx context.Context, coord *valueobject.Coordinate, radiusMeters float64, limit int) ([]*entity.Position, error)

	// FindInSector busca posições em um setor específico
	FindInSector(ctx context.Context, sector *valueobject.Sector) ([]*entity.Position, error)

	// FindInSectors busca posições em múltiplos setores
	FindInSectors(ctx context.Context, sectors []*valueobject.Sector) ([]*entity.Position, error)

	// UpdateCurrentPosition atualiza posição atual do usuário
	UpdateCurrentPosition(ctx context.Context, position *entity.Position) error

	// DeleteOldPositions remove posições antigas (cleanup)
	DeleteOldPositions(ctx context.Context, olderThan *valueobject.Timestamp) (int, error)
}

// PositionQuery representa critérios de busca para posições
// Value Object para queries complexas
type PositionQuery struct {
	UserIDs      []entity.UserID         `json:"user_ids,omitempty"`
	Sectors      []*valueobject.Sector   `json:"sectors,omitempty"`
	Coordinate   *valueobject.Coordinate `json:"coordinate,omitempty"`
	RadiusMeters float64                 `json:"radius_meters,omitempty"`
	TimeRange    *TimeRange              `json:"time_range,omitempty"`
	Limit        int                     `json:"limit,omitempty"`
	Offset       int                     `json:"offset,omitempty"`
}

// TimeRange representa um intervalo de tempo
type TimeRange struct {
	From *valueobject.Timestamp `json:"from,omitempty"`
	To   *valueobject.Timestamp `json:"to,omitempty"`
}

// AdvancedPositionRepository define operações avançadas de consulta
type AdvancedPositionRepository interface {
	// FindByQuery busca posições usando critérios complexos
	FindByQuery(ctx context.Context, query *PositionQuery) ([]*entity.Position, error)

	// CountByQuery conta posições usando critérios
	CountByQuery(ctx context.Context, query *PositionQuery) (int, error)

	// FindUsersInRadius busca usuários únicos dentro de um raio
	FindUsersInRadius(ctx context.Context, coord *valueobject.Coordinate, radiusMeters float64) ([]entity.UserID, error)

	// GetSectorStatistics retorna estatísticas de um setor
	GetSectorStatistics(ctx context.Context, sector *valueobject.Sector) (*SectorStats, error)
}

// SectorStats representa estatísticas de um setor
type SectorStats struct {
	Sector        *valueobject.Sector    `json:"sector"`
	UserCount     int                    `json:"user_count"`
	PositionCount int                    `json:"position_count"`
	LastActivity  *valueobject.Timestamp `json:"last_activity,omitempty"`
}
