package valueobject

import (
	"fmt"
	"math"
)

// Sector representa um setor geográfico de 100x100 metros
// Combina a localização do setor (Point) com métodos específicos de conversão
type Sector struct {
	point *Point
}

// Constantes para conversão geográfica
const (
	// Aproximação: 1 grau de latitude ≈ 111.320 km
	MetersPerDegreeLat = 111320.0

	// Longitude varia com latitude, usaremos aproximação para linha do equador
	// 1 grau de longitude ≈ 111.320 km * cos(latitude)
	MetersPerDegreeLngAtEquator = 111320.0
)

// NewSector cria um novo setor
func NewSector(x, y int) (*Sector, error) {
	point, err := NewPoint(x, y)
	if err != nil {
		return nil, err
	}

	return &Sector{point: point}, nil
}

// NewSectorFromCoordinate converte coordenada geográfica para setor
// Esta é uma função crucial que mapeia o mundo real para nosso sistema de setores
func NewSectorFromCoordinate(coord *Coordinate) (*Sector, error) {
	if coord == nil {
		return nil, fmt.Errorf("coordinate cannot be nil")
	}

	// Para simplificar, vamos usar uma origem fixa (pode ser configurável)
	// Origem: (0,0) será equivalente a lat=0, lng=0 (linha do equador, meridiano de Greenwich)

	// Converter latitude para coordenada Y do setor
	// Positivo = Norte, Negativo = Sul
	latMeters := coord.Latitude() * MetersPerDegreeLat
	sectorY := int(math.Round(latMeters / SectorSizeMeters))

	// Converter longitude para coordenada X do setor
	// Ajustar por latitude para compensar convergência dos meridianos
	lngMetersPerDegree := MetersPerDegreeLngAtEquator * math.Cos(degToRad(coord.Latitude()))
	lngMeters := coord.Longitude() * lngMetersPerDegree
	sectorX := int(math.Round(lngMeters / SectorSizeMeters))

	return NewSector(sectorX, sectorY)
}

// Point retorna o ponto do setor
func (s *Sector) Point() *Point {
	return s.point
}

// X retorna coordenada X do setor
func (s *Sector) X() int {
	return s.point.X()
}

// Y retorna coordenada Y do setor
func (s *Sector) Y() int {
	return s.point.Y()
}

// ToCoordinate converte setor de volta para coordenada geográfica (centro do setor)
func (s *Sector) ToCoordinate() (*Coordinate, error) {
	// Converter X do setor para longitude
	lngMeters := float64(s.point.X()) * SectorSizeMeters
	longitude := lngMeters / MetersPerDegreeLngAtEquator

	// Converter Y do setor para latitude
	latMeters := float64(s.point.Y()) * SectorSizeMeters
	latitude := latMeters / MetersPerDegreeLat

	return NewCoordinate(latitude, longitude)
}

// GetBounds retorna as coordenadas dos cantos do setor
func (s *Sector) GetBounds() (topLeft, topRight, bottomLeft, bottomRight *Coordinate, err error) {
	center, err := s.ToCoordinate()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Calcular offset de meio setor
	halfSectorLat := (SectorSizeMeters / 2) / MetersPerDegreeLat
	halfSectorLng := (SectorSizeMeters / 2) / (MetersPerDegreeLngAtEquator * math.Cos(degToRad(center.Latitude())))

	topLeft, _ = NewCoordinate(center.Latitude()+halfSectorLat, center.Longitude()-halfSectorLng)
	topRight, _ = NewCoordinate(center.Latitude()+halfSectorLat, center.Longitude()+halfSectorLng)
	bottomLeft, _ = NewCoordinate(center.Latitude()-halfSectorLat, center.Longitude()-halfSectorLng)
	bottomRight, _ = NewCoordinate(center.Latitude()-halfSectorLat, center.Longitude()+halfSectorLng)

	return topLeft, topRight, bottomLeft, bottomRight, nil
}

// String implementa fmt.Stringer
func (s *Sector) String() string {
	return fmt.Sprintf("Sector(%d, %d)", s.point.X(), s.point.Y())
}

// Equals compara dois setores
func (s *Sector) Equals(other *Sector) bool {
	if other == nil {
		return false
	}
	return s.point.Equals(other.point)
}

// GetNeighboringSectors retorna setores vizinhos
func (s *Sector) GetNeighboringSectors() ([]*Sector, error) {
	neighborPoints := s.point.GetNeighboringSectors()
	sectors := make([]*Sector, 0, len(neighborPoints))

	for _, point := range neighborPoints {
		sector := &Sector{point: point}
		sectors = append(sectors, sector)
	}

	return sectors, nil
}

// ID retorna identificador único do setor
func (s *Sector) ID() string {
	return s.point.ToSectorID()
}
