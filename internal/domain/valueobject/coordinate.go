package valueobject

import (
	"errors"
	"fmt"
	"math"
)

// Coordinate representa uma coordenada geográfica (latitude, longitude)
// Value Object: Imutável, auto-validação, comparação por valor
type Coordinate struct {
	latitude  float64
	longitude float64
}

// Constantes para validação de coordenadas
const (
	MinLatitude   = -90.0
	MaxLatitude   = 90.0
	MinLongitude  = -180.0
	MaxLongitude  = 180.0
	EarthRadiusKm = 6371.0 // Raio da Terra em quilômetros
)

// Erros específicos do domínio
var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90 degrees")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180 degrees")
)

// NewCoordinate cria uma nova coordenada com validação
// Factory method que garante que só coordenadas válidas são criadas
func NewCoordinate(lat, lng float64) (*Coordinate, error) {
	if lat < MinLatitude || lat > MaxLatitude {
		return nil, fmt.Errorf("%w: got %f", ErrInvalidLatitude, lat)
	}

	if lng < MinLongitude || lng > MaxLongitude {
		return nil, fmt.Errorf("%w: got %f", ErrInvalidLongitude, lng)
	}

	return &Coordinate{
		latitude:  lat,
		longitude: lng,
	}, nil
}

// Getters (Value Objects expõem seus valores de forma segura)
func (c *Coordinate) Latitude() float64 {
	return c.latitude
}

func (c *Coordinate) Longitude() float64 {
	return c.longitude
}

// String implementa fmt.Stringer para logging/debug
func (c *Coordinate) String() string {
	return fmt.Sprintf("Coordinate(%.6f, %.6f)", c.latitude, c.longitude)
}

// Equals compara duas coordenadas por valor
func (c *Coordinate) Equals(other *Coordinate) bool {
	if other == nil {
		return false
	}

	// Comparação com tolerância para problemas de ponto flutuante
	const tolerance = 1e-9
	return math.Abs(c.latitude-other.latitude) < tolerance &&
		math.Abs(c.longitude-other.longitude) < tolerance
}

// DistanceTo calcula distância entre duas coordenadas usando fórmula de Haversine
// Retorna distância em metros
func (c *Coordinate) DistanceTo(other *Coordinate) float64 {
	if other == nil {
		return 0
	}

	// Converter para radianos
	lat1Rad := degToRad(c.latitude)
	lng1Rad := degToRad(c.longitude)
	lat2Rad := degToRad(other.latitude)
	lng2Rad := degToRad(other.longitude)

	// Diferenças
	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	// Fórmula de Haversine
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	centralAngle := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distância em metros
	return EarthRadiusKm * centralAngle * 1000
}

// IsWithinRadius verifica se coordenada está dentro de um raio (em metros)
func (c *Coordinate) IsWithinRadius(other *Coordinate, radiusMeters float64) bool {
	if other == nil || radiusMeters < 0 {
		return false
	}

	return c.DistanceTo(other) <= radiusMeters
}

// ToWKT converte para formato Well-Known Text (usado no PostGIS)
func (c *Coordinate) ToWKT() string {
	return fmt.Sprintf("POINT(%f %f)", c.longitude, c.latitude)
}

// degToRad converte graus para radianos
func degToRad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

// CalculateDistance calcula distância entre duas coordenadas em metros
// Função utilitária para usar sem criar objetos Coordinate
func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	coord1, err := NewCoordinate(lat1, lng1)
	if err != nil {
		return 0
	}

	coord2, err := NewCoordinate(lat2, lng2)
	if err != nil {
		return 0
	}

	return coord1.DistanceTo(coord2)
}
