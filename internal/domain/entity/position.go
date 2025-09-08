package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
)

// Position representa uma posição geográfica de um usuário
// Entidade com regras de negócio específicas para geolocalização
type Position struct {
	id         PositionID              // Identidade única
	userID     UserID                  // Referência ao usuário
	coordinate *valueobject.Coordinate // Coordenada geográfica
	sector     *valueobject.Sector     // Setor calculado
	recordedAt *valueobject.Timestamp  // Quando foi registrada
	createdAt  *valueobject.Timestamp  // Quando foi persistida
}

// PositionID representa o identificador único da posição
type PositionID struct {
	value string
}

// Constantes de validação
const (
	MaxPositionAgeHours = 24 // Posições não podem ser muito antigas
)

// Erros específicos do domínio Position
var (
	ErrEmptyPositionID   = errors.New("position ID cannot be empty")
	ErrPositionTooOld    = errors.New("position is too old")
	ErrInvalidCoordinate = errors.New("invalid coordinate")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrFuturePosition    = errors.New("position cannot be in the future")
)

// NewPositionID cria um novo PositionID
func NewPositionID(id string) (*PositionID, error) {
	if id == "" {
		return nil, ErrEmptyPositionID
	}

	return &PositionID{value: id}, nil
}

// Value retorna o valor do PositionID
func (pid *PositionID) Value() string {
	return pid.value
}

// String implementa fmt.Stringer
func (pid *PositionID) String() string {
	return pid.value
}

// Equals compara dois PositionIDs
func (pid *PositionID) Equals(other *PositionID) bool {
	if other == nil {
		return false
	}
	return pid.value == other.value
}

// NewPosition cria uma nova posição (Factory Method)
// Aplica todas as regras de validação do domínio
func NewPosition(id string, userID UserID, lat, lng float64, recordedAt time.Time) (*Position, error) {
	// Validar PositionID
	positionID, err := NewPositionID(id)
	if err != nil {
		return nil, err
	}

	// Validar coordenada
	coordinate, err := valueobject.NewCoordinate(lat, lng)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCoordinate, err.Error())
	}

	// Validar timestamp (não pode ser futuro)
	recordedTimestamp, err := valueobject.NewTimestampNotInFuture(recordedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFuturePosition, err.Error())
	}

	// Validar idade da posição
	if err := validatePositionAge(recordedTimestamp); err != nil {
		return nil, err
	}

	// Calcular setor automaticamente
	sector, err := valueobject.NewSectorFromCoordinate(coordinate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sector: %w", err)
	}

	now := valueobject.Now()

	return &Position{
		id:         *positionID,
		userID:     userID,
		coordinate: coordinate,
		sector:     sector,
		recordedAt: recordedTimestamp,
		createdAt:  now,
	}, nil
}

// validatePositionAge valida se a posição não é muito antiga
func validatePositionAge(recordedAt *valueobject.Timestamp) error {
	maxAge := time.Duration(MaxPositionAgeHours) * time.Hour

	if recordedAt.Age() > maxAge {
		return fmt.Errorf("%w: position is %v old, max allowed is %v",
			ErrPositionTooOld, recordedAt.Age(), maxAge)
	}

	return nil
}

// Getters
func (p *Position) ID() PositionID {
	return p.id
}

func (p *Position) UserID() UserID {
	return p.userID
}

func (p *Position) Coordinate() *valueobject.Coordinate {
	return p.coordinate
}

func (p *Position) Sector() *valueobject.Sector {
	return p.sector
}

func (p *Position) RecordedAt() *valueobject.Timestamp {
	return p.recordedAt
}

func (p *Position) CreatedAt() *valueobject.Timestamp {
	return p.createdAt
}

// Latitude retorna latitude da posição
func (p *Position) Latitude() float64 {
	return p.coordinate.Latitude()
}

// Longitude retorna longitude da posição
func (p *Position) Longitude() float64 {
	return p.coordinate.Longitude()
}

// SectorX retorna coordenada X do setor
func (p *Position) SectorX() int {
	return p.sector.X()
}

// SectorY retorna coordenada Y do setor
func (p *Position) SectorY() int {
	return p.sector.Y()
}

// DistanceTo calcula distância para outra posição
func (p *Position) DistanceTo(other *Position) float64 {
	if other == nil {
		return 0
	}
	return p.coordinate.DistanceTo(other.coordinate)
}

// IsWithinRadius verifica se posição está dentro de raio de outra posição
func (p *Position) IsWithinRadius(other *Position, radiusMeters float64) bool {
	if other == nil {
		return false
	}
	return p.coordinate.IsWithinRadius(other.coordinate, radiusMeters)
}

// IsInSameSector verifica se posição está no mesmo setor que outra
func (p *Position) IsInSameSector(other *Position) bool {
	if other == nil {
		return false
	}
	return p.sector.Equals(other.sector)
}

// GetNeighboringSectors retorna setores vizinhos desta posição
func (p *Position) GetNeighboringSectors() ([]*valueobject.Sector, error) {
	return p.sector.GetNeighboringSectors()
}

// Age retorna idade da posição registrada
func (p *Position) Age() time.Duration {
	return p.recordedAt.Age()
}

// IsRecent verifica se posição foi registrada recentemente
func (p *Position) IsRecent(threshold time.Duration) bool {
	return p.recordedAt.IsWithinLast(threshold)
}

// String implementa fmt.Stringer
func (p *Position) String() string {
	return fmt.Sprintf("Position{ID: %s, UserID: %s, Lat: %.6f, Lng: %.6f, Sector: %s, Age: %v}",
		p.id.Value(), p.userID.Value(), p.Latitude(), p.Longitude(),
		p.sector.String(), p.Age().Truncate(time.Second))
}

// Equals compara duas posições pela identidade (ID)
func (p *Position) Equals(other *Position) bool {
	if other == nil {
		return false
	}
	return p.id.Equals(&other.id)
}
