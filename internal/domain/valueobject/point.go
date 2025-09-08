package valueobject

import (
	"errors"
	"fmt"
	"math"
)

// Point representa um ponto em coordenadas cartesianas (x, y)
// Usado para representar setores de 100x100 metros
type Point struct {
	x int
	y int
}

// Constantes para setores
const (
	SectorSizeMeters = 100     // Cada setor tem 100x100 metros
	MaxSectorCoord   = 100000  // Limite máximo de coordenadas (ajustável)
	MinSectorCoord   = -100000 // Limite mínimo de coordenadas
)

// Erros específicos
var (
	ErrInvalidSectorX = errors.New("sector X coordinate out of bounds")
	ErrInvalidSectorY = errors.New("sector Y coordinate out of bounds")
)

// NewPoint cria um novo ponto com validação
func NewPoint(x, y int) (*Point, error) {
	if x < MinSectorCoord || x > MaxSectorCoord {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSectorX, x)
	}

	if y < MinSectorCoord || y > MaxSectorCoord {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidSectorY, y)
	}

	return &Point{x: x, y: y}, nil
}

// Getters
func (p *Point) X() int {
	return p.x
}

func (p *Point) Y() int {
	return p.y
}

// String implementa fmt.Stringer
func (p *Point) String() string {
	return fmt.Sprintf("Point(%d, %d)", p.x, p.y)
}

// Equals compara dois pontos por valor
func (p *Point) Equals(other *Point) bool {
	if other == nil {
		return false
	}
	return p.x == other.x && p.y == other.y
}

// DistanceTo calcula distância euclidiana entre dois pontos
// Retorna distância em metros (considerando escala de setores)
func (p *Point) DistanceTo(other *Point) float64 {
	if other == nil {
		return 0
	}

	dx := float64(p.x - other.x)
	dy := float64(p.y - other.y)

	// Distância em número de setores * tamanho do setor
	sectorDistance := math.Sqrt(dx*dx + dy*dy)
	return sectorDistance * SectorSizeMeters
}

// GetNeighboringSectors retorna pontos dos setores vizinhos (8 direções + próprio)
func (p *Point) GetNeighboringSectors() []*Point {
	neighbors := make([]*Point, 0, 9)

	// Incluir o próprio setor e os 8 vizinhos
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			newX := p.x + dx
			newY := p.y + dy

			// Verificar se está dentro dos limites
			if newX >= MinSectorCoord && newX <= MaxSectorCoord &&
				newY >= MinSectorCoord && newY <= MaxSectorCoord {

				neighbor, _ := NewPoint(newX, newY) // Não deveria dar erro aqui
				if neighbor != nil {
					neighbors = append(neighbors, neighbor)
				}
			}
		}
	}

	return neighbors
}

// GetSectorsInRadius retorna todos os setores dentro de um raio específico
func (p *Point) GetSectorsInRadius(radiusMeters float64) []*Point {
	if radiusMeters <= 0 {
		return []*Point{p} // Apenas o próprio setor
	}

	// Calcular quantos setores cabem no raio
	radiusInSectors := int(math.Ceil(radiusMeters / SectorSizeMeters))

	sectors := make([]*Point, 0)

	// Iterar em uma área quadrada e filtrar por distância real
	for dx := -radiusInSectors; dx <= radiusInSectors; dx++ {
		for dy := -radiusInSectors; dy <= radiusInSectors; dy++ {
			newX := p.x + dx
			newY := p.y + dy

			// Verificar limites
			if newX >= MinSectorCoord && newX <= MaxSectorCoord &&
				newY >= MinSectorCoord && newY <= MaxSectorCoord {

				candidate, _ := NewPoint(newX, newY)
				if candidate != nil {
					// Verificar se está realmente dentro do raio
					if p.DistanceTo(candidate) <= radiusMeters {
						sectors = append(sectors, candidate)
					}
				}
			}
		}
	}

	return sectors
}

// ToSectorID gera um ID único para o setor (útil para cache)
func (p *Point) ToSectorID() string {
	return fmt.Sprintf("sector_%d_%d", p.x, p.y)
}
