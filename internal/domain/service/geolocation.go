package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
)

// GeoLocationService contém lógica geoespacial complexa
// Domain Service = lógica que não pertence a uma entidade específica
type GeoLocationService struct {
	positionRepo repository.PositionRepository
}

// ProximityResult representa resultado de busca por proximidade
type ProximityResult struct {
	User     entity.UserID    `json:"user_id"`
	Position *entity.Position `json:"position"`
	Distance float64          `json:"distance_meters"`
}

// SectorAnalysis representa análise de um setor
type SectorAnalysis struct {
	Sector          *valueobject.Sector   `json:"sector"`
	UserCount       int                   `json:"user_count"`
	Density         float64               `json:"density_per_km2"`
	NeighborSectors []*valueobject.Sector `json:"neighbor_sectors"`
}

// Erros específicos do domain service
var (
	ErrNoPositionsFound = errors.New("no positions found")
	ErrInvalidRadius    = errors.New("invalid radius")
	ErrInvalidSector    = errors.New("invalid sector")
)

// NewGeoLocationService cria um novo serviço de geolocalização
func NewGeoLocationService(positionRepo repository.PositionRepository) *GeoLocationService {
	return &GeoLocationService{
		positionRepo: positionRepo,
	}
}

// FindNearbyUsers encontra usuários próximos a uma coordenada
func (s *GeoLocationService) FindNearbyUsers(ctx context.Context, coord *valueobject.Coordinate, radiusMeters float64) ([]*ProximityResult, error) {
	if radiusMeters <= 0 {
		return nil, fmt.Errorf("%w: radius must be positive", ErrInvalidRadius)
	}

	// Buscar posições próximas
	positions, err := s.positionRepo.FindNearby(ctx, coord, radiusMeters, 100) // Limite de 100
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby positions: %w", err)
	}

	if len(positions) == 0 {
		return []*ProximityResult{}, nil
	}

	// Calcular distâncias e criar resultados
	results := make([]*ProximityResult, 0, len(positions))
	for _, pos := range positions {
		distance := coord.DistanceTo(pos.Coordinate())

		result := &ProximityResult{
			User:     pos.UserID(),
			Position: pos,
			Distance: distance,
		}
		results = append(results, result)
	}

	// Ordenar por distância (mais próximos primeiro)
	return s.sortByDistance(results), nil
}

// FindUsersInSector encontra usuários em um setor específico
func (s *GeoLocationService) FindUsersInSector(ctx context.Context, sector *valueobject.Sector) ([]*entity.Position, error) {
	if sector == nil {
		return nil, ErrInvalidSector
	}

	positions, err := s.positionRepo.FindInSector(ctx, sector)
	if err != nil {
		return nil, fmt.Errorf("failed to find users in sector %s: %w", sector.ID(), err)
	}

	return positions, nil
}

// AnalyzeSector analisa um setor e seus vizinhos
func (s *GeoLocationService) AnalyzeSector(ctx context.Context, sector *valueobject.Sector) (*SectorAnalysis, error) {
	if sector == nil {
		return nil, ErrInvalidSector
	}

	// Buscar posições no setor
	positions, err := s.positionRepo.FindInSector(ctx, sector)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze sector %s: %w", sector.ID(), err)
	}

	// Obter setores vizinhos
	neighbors, err := sector.GetNeighboringSectors()
	if err != nil {
		return nil, fmt.Errorf("failed to get neighboring sectors: %w", err)
	}

	// Calcular densidade (usuários por km²)
	// Cada setor = 100m x 100m = 0.01 km²
	sectorAreaKm2 := 0.01
	density := float64(len(positions)) / sectorAreaKm2

	return &SectorAnalysis{
		Sector:          sector,
		UserCount:       len(positions),
		Density:         density,
		NeighborSectors: neighbors,
	}, nil
}

// FindUsersInRadius encontra usuários em múltiplos setores dentro de um raio
func (s *GeoLocationService) FindUsersInRadius(ctx context.Context, center *valueobject.Coordinate, radiusMeters float64) ([]*ProximityResult, error) {
	if radiusMeters <= 0 {
		return nil, fmt.Errorf("%w: radius must be positive", ErrInvalidRadius)
	}

	// Converter coordenada central para setor
	centralSector, err := valueobject.NewSectorFromCoordinate(center)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coordinate to sector: %w", err)
	}

	// Obter setores dentro do raio
	sectorsInRadius := centralSector.Point().GetSectorsInRadius(radiusMeters)

	// Converter pontos para setores
	sectors := make([]*valueobject.Sector, 0, len(sectorsInRadius))
	for _, point := range sectorsInRadius {
		sector, err := valueobject.NewSector(point.X(), point.Y())
		if err != nil {
			continue // Pular setores inválidos
		}
		sectors = append(sectors, sector)
	}

	// Buscar posições em todos os setores
	positions, err := s.positionRepo.FindInSectors(ctx, sectors)
	if err != nil {
		return nil, fmt.Errorf("failed to find positions in sectors: %w", err)
	}

	// Filtrar por distância real e criar resultados
	results := make([]*ProximityResult, 0)
	for _, pos := range positions {
		distance := center.DistanceTo(pos.Coordinate())

		// Só incluir se estiver realmente dentro do raio
		if distance <= radiusMeters {
			result := &ProximityResult{
				User:     pos.UserID(),
				Position: pos,
				Distance: distance,
			}
			results = append(results, result)
		}
	}

	return s.sortByDistance(results), nil
}

// CalculateOptimalSectorSize calcula tamanho ótimo de setor baseado na densidade
func (s *GeoLocationService) CalculateOptimalSectorSize(userDensityPerKm2 float64) float64 {
	// Lógica heurística: mais usuários = setores menores
	// Para 1000 usuários/km² = setores de 50m
	// Para 100 usuários/km² = setores de 100m
	// Para 10 usuários/km² = setores de 200m

	if userDensityPerKm2 >= 1000 {
		return 50.0
	} else if userDensityPerKm2 >= 100 {
		return 100.0
	} else if userDensityPerKm2 >= 10 {
		return 200.0
	} else {
		return 500.0 // Áreas com poucos usuários
	}
}

// sortByDistance ordena resultados por distância (bubble sort simples)
func (s *GeoLocationService) sortByDistance(results []*ProximityResult) []*ProximityResult {
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Distance > results[j+1].Distance {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
	return results
}
