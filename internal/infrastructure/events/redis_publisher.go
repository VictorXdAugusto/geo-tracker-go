package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	domainEvents "github.com/vitao/geolocation-tracker/internal/domain/events"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// RedisStreamPublisher implementa Publisher usando Redis Streams
type RedisStreamPublisher struct {
	client *redis.Client
	logger logger.Logger
}

// NewRedisStreamPublisher cria uma nova instância do publisher
func NewRedisStreamPublisher(client *redis.Client, logger logger.Logger) *RedisStreamPublisher {
	return &RedisStreamPublisher{
		client: client,
		logger: logger,
	}
}

// Publish publica um evento no stream especificado
func (p *RedisStreamPublisher) Publish(ctx context.Context, streamName string, event *domainEvents.Event) error {
	// Gerar ID único se não tiver
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Serializar os dados do evento para JSON
	eventDataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	// Preparar os campos para o Redis Stream
	// Redis Streams armazena como key-value pairs
	fields := map[string]interface{}{
		"event_id":  event.ID,
		"type":      string(event.Type),
		"user_id":   event.UserID,
		"event_ctx": event.EventID, // Renomeado para evitar confusão
		"timestamp": event.Timestamp.Format(time.RFC3339Nano),
		"data":      string(eventDataJSON),
		"metadata":  string(metadataJSON),
	}

	// Publicar no Redis Stream
	// XADD stream_name * field1 value1 field2 value2 ...
	result := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		ID:     "*", // Deixar o Redis gerar o ID automaticamente
		Values: fields,
	})

	if result.Err() != nil {
		p.logger.Error("Failed to publish event to Redis Stream",
			"stream", streamName,
			"event_type", event.Type,
			"event_id", event.ID,
			"error", result.Err(),
		)
		return fmt.Errorf("failed to publish to stream %s: %w", streamName, result.Err())
	}

	// Guardar o ID do stream no evento para referência
	event.StreamID = result.Val()

	p.logger.Info("Event published successfully to Redis Stream",
		"stream", streamName,
		"event_type", event.Type,
		"event_id", event.ID,
		"stream_id", event.StreamID,
		"user_id", event.UserID,
	)

	return nil
}

// PublishPositionChanged publica evento de mudança de posição
func (p *RedisStreamPublisher) PublishPositionChanged(ctx context.Context, event *domainEvents.Event) error {
	return p.Publish(ctx, domainEvents.StreamPositionEvents, event)
}

// PublishSectorChanged publica evento de mudança de setor
func (p *RedisStreamPublisher) PublishSectorChanged(ctx context.Context, event *domainEvents.Event) error {
	return p.Publish(ctx, domainEvents.StreamSectorEvents, event)
}

// Close fecha a conexão (não precisamos fazer nada aqui pois o Redis client é compartilhado)
func (p *RedisStreamPublisher) Close() error {
	return nil
}

// ensureStreamExists garante que o stream existe e cria consumer groups se necessário
func (p *RedisStreamPublisher) ensureStreamExists(ctx context.Context, streamName string) error {
	// Tentar criar o stream - se já existir, isso não fará nada
	// Criar um evento dummy para garantir que o stream existe
	dummyID, err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		ID:     "*",
		Values: map[string]interface{}{
			"init":      "true",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to ensure stream %s exists: %w", streamName, err)
	}

	// Remover o evento dummy
	p.client.XDel(ctx, streamName, dummyID)

	p.logger.Info("Stream ensured to exist", "stream", streamName)

	// Criar consumer groups se não existirem
	groups := []string{
		domainEvents.ConsumerGroupNotifications,
		domainEvents.ConsumerGroupAnalytics,
		domainEvents.ConsumerGroupRealtime,
	}

	for _, group := range groups {
		// XGROUP CREATE stream group $ MKSTREAM
		err = p.client.XGroupCreate(ctx, streamName, group, "$").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			p.logger.Error("Failed to create consumer group",
				"stream", streamName,
				"group", group,
				"error", err,
			)
		} else if err == nil {
			p.logger.Info("Created consumer group",
				"stream", streamName,
				"group", group,
			)
		}
	}

	return nil
}

// InitializeStreams inicializa todos os streams necessários
func (p *RedisStreamPublisher) InitializeStreams(ctx context.Context) error {
	streams := []string{
		domainEvents.StreamPositionEvents,
		domainEvents.StreamSectorEvents,
		domainEvents.StreamProximityEvents,
	}

	for _, stream := range streams {
		if err := p.ensureStreamExists(ctx, stream); err != nil {
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
	}

	p.logger.Info("All Redis Streams initialized successfully")
	return nil
}
