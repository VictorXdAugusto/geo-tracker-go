package events

import (
	"time"
)

// EventType representa os tipos de eventos no sistema
type EventType string

const (
	// PositionChanged quando um usuário move sua posição
	EventTypePositionChanged EventType = "position.changed"

	// UserEnteredSector quando usuário entra em novo setor
	EventTypeUserEnteredSector EventType = "sector.user_entered"

	// UserLeftSector quando usuário sai de um setor
	EventTypeUserLeftSector EventType = "sector.user_left"

	// UserNearby quando usuários ficam próximos
	EventTypeUserNearby EventType = "proximity.user_nearby"
)

// Event representa a estrutura base de um evento
type Event struct {
	ID        string                 `json:"id"`        // UUID único do evento
	Type      EventType              `json:"type"`      // Tipo do evento
	StreamID  string                 `json:"stream_id"` // ID no Redis Stream
	UserID    string                 `json:"user_id"`   // Usuário que gerou o evento
	EventID   string                 `json:"event_id"`  // ID do evento (contexto)
	Timestamp time.Time              `json:"timestamp"` // Quando aconteceu
	Data      map[string]interface{} `json:"data"`      // Dados específicos do evento
	Metadata  EventMetadata          `json:"metadata"`  // Metadados adicionais
}

// EventMetadata contém informações adicionais sobre o evento
type EventMetadata struct {
	Source    string `json:"source"`     // De onde veio (API, worker, etc)
	Version   string `json:"version"`    // Versão do schema do evento
	RequestID string `json:"request_id"` // ID da requisição que gerou
}

// PositionChangedData dados específicos do evento de mudança de posição
type PositionChangedData struct {
	PositionID     string  `json:"position_id"`     // ID da nova posição
	PreviousLat    float64 `json:"previous_lat"`    // Latitude anterior (pode ser 0)
	PreviousLng    float64 `json:"previous_lng"`    // Longitude anterior (pode ser 0)
	NewLat         float64 `json:"new_lat"`         // Nova latitude
	NewLng         float64 `json:"new_lng"`         // Nova longitude
	PreviousSector string  `json:"previous_sector"` // Setor anterior (pode ser vazio)
	NewSector      string  `json:"new_sector"`      // Novo setor
	DistanceMoved  float64 `json:"distance_moved"`  // Distância movida em metros
}

// SectorChangedData dados específicos de mudança de setor
type SectorChangedData struct {
	SectorX       int     `json:"sector_x"`        // Coordenada X do setor
	SectorY       int     `json:"sector_y"`        // Coordenada Y do setor
	SectorID      string  `json:"sector_id"`       // ID do setor (ex: "sector_100_200")
	Latitude      float64 `json:"latitude"`        // Lat do usuário no setor
	Longitude     float64 `json:"longitude"`       // Lng do usuário no setor
	UsersInSector int     `json:"users_in_sector"` // Quantos usuários no setor agora
}

// ProximityData dados específicos de proximidade entre usuários
type ProximityData struct {
	NearUserID   string  `json:"near_user_id"`   // ID do usuário próximo
	NearUserName string  `json:"near_user_name"` // Nome do usuário próximo
	Distance     float64 `json:"distance"`       // Distância entre eles em metros
	MaxDistance  float64 `json:"max_distance"`   // Distância máxima configurada
	IsEntering   bool    `json:"is_entering"`    // true=entrando no raio, false=saindo
}

// NewPositionChangedEvent cria um novo evento de mudança de posição
func NewPositionChangedEvent(userID, eventID string, data PositionChangedData) *Event {
	return &Event{
		Type:      EventTypePositionChanged,
		UserID:    userID,
		EventID:   eventID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"position_id":     data.PositionID,
			"previous_lat":    data.PreviousLat,
			"previous_lng":    data.PreviousLng,
			"new_lat":         data.NewLat,
			"new_lng":         data.NewLng,
			"previous_sector": data.PreviousSector,
			"new_sector":      data.NewSector,
			"distance_moved":  data.DistanceMoved,
		},
		Metadata: EventMetadata{
			Source:  "position-api",
			Version: "1.0",
		},
	}
}

// NewSectorChangedEvent cria um novo evento de mudança de setor
func NewSectorChangedEvent(userID, eventID string, eventType EventType, data SectorChangedData) *Event {
	return &Event{
		Type:      eventType, // EventTypeUserEnteredSector ou EventTypeUserLeftSector
		UserID:    userID,
		EventID:   eventID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sector_x":        data.SectorX,
			"sector_y":        data.SectorY,
			"sector_id":       data.SectorID,
			"latitude":        data.Latitude,
			"longitude":       data.Longitude,
			"users_in_sector": data.UsersInSector,
		},
		Metadata: EventMetadata{
			Source:  "position-api",
			Version: "1.0",
		},
	}
}
