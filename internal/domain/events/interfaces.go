package events

import (
	"context"
)

// Publisher interface para publicar eventos
type Publisher interface {
	// Publish publica um evento no stream
	Publish(ctx context.Context, streamName string, event *Event) error

	// PublishPositionChanged publica evento de mudança de posição
	PublishPositionChanged(ctx context.Context, event *Event) error

	// PublishSectorChanged publica evento de mudança de setor
	PublishSectorChanged(ctx context.Context, event *Event) error

	// Close fecha a conexão do publisher
	Close() error
}

// Consumer interface para consumir eventos
type Consumer interface {
	// Subscribe se inscreve em um stream para consumir eventos
	Subscribe(ctx context.Context, streamName, consumerGroup, consumerName string) (<-chan *Event, error)

	// Ack confirma o processamento de um evento
	Ack(ctx context.Context, streamName, consumerGroup, eventID string) error

	// Close fecha a conexão do consumer
	Close() error
}

// EventHandler interface para processar eventos
type EventHandler interface {
	// Handle processa um evento específico
	Handle(ctx context.Context, event *Event) error

	// CanHandle verifica se pode processar este tipo de evento
	CanHandle(eventType EventType) bool
}

// StreamNames constantes dos nomes dos streams
const (
	StreamPositionEvents  = "geolocation:position-events"
	StreamSectorEvents    = "geolocation:sector-events"
	StreamProximityEvents = "geolocation:proximity-events"
)

// ConsumerGroups nomes dos grupos de consumidores
const (
	ConsumerGroupNotifications = "notifications"
	ConsumerGroupAnalytics     = "analytics"
	ConsumerGroupRealtime      = "realtime"
)
