package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	domainEvents "github.com/vitao/geolocation-tracker/internal/domain/events"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// RedisStreamConsumer implementa Consumer usando Redis Streams
type RedisStreamConsumer struct {
	client   *redis.Client
	logger   logger.Logger
	handlers map[domainEvents.EventType][]domainEvents.EventHandler
}

// NewRedisStreamConsumer cria uma nova instância do consumer
func NewRedisStreamConsumer(client *redis.Client, logger logger.Logger) *RedisStreamConsumer {
	return &RedisStreamConsumer{
		client:   client,
		logger:   logger,
		handlers: make(map[domainEvents.EventType][]domainEvents.EventHandler),
	}
}

// Subscribe se inscreve em um stream para consumir eventos
func (c *RedisStreamConsumer) Subscribe(ctx context.Context, streamName, consumerGroup, consumerName string) (<-chan *domainEvents.Event, error) {
	// Canal para enviar eventos processados
	eventChan := make(chan *domainEvents.Event, 100)

	// Criar consumer group se não existir
	err := c.client.XGroupCreate(ctx, streamName, consumerGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		c.logger.Error("Failed to create consumer group",
			"stream", streamName,
			"group", consumerGroup,
			"error", err,
		)
	}

	// Goroutine para consumir eventos continuamente
	go func() {
		defer close(eventChan)

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping consumer",
					"stream", streamName,
					"consumer", consumerName,
				)
				return

			default:
				// XREADGROUP GROUP <group> <consumer> COUNT <count> BLOCK <milliseconds> STREAMS <stream> >
				result, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    consumerGroup,
					Consumer: consumerName,
					Streams:  []string{streamName, ">"},
					Count:    10,
					Block:    1000 * time.Millisecond, // Block por 1 segundo
				}).Result()

				if err != nil {
					if err == redis.Nil {
						// Nenhuma mensagem nova, continuar
						continue
					}
					c.logger.Error("Failed to read from stream",
						"stream", streamName,
						"consumer", consumerName,
						"error", err,
					)
					time.Sleep(5 * time.Second) // Aguardar antes de tentar novamente
					continue
				}

				// Processar mensagens recebidas
				for _, stream := range result {
					for _, message := range stream.Messages {
						event, err := c.parseMessage(message)
						if err != nil {
							c.logger.Error("Failed to parse event message",
								"stream", streamName,
								"message_id", message.ID,
								"error", err,
							)
							continue
						}

						// Enviar evento pelo canal
						select {
						case eventChan <- event:
							c.logger.Debug("Event sent to channel",
								"stream", streamName,
								"event_id", event.ID,
								"event_type", event.Type,
							)
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	c.logger.Info("Consumer subscribed to stream",
		"stream", streamName,
		"consumer_group", consumerGroup,
		"consumer_name", consumerName,
	)

	return eventChan, nil
}

// parseMessage converte uma mensagem Redis Stream em Event
func (c *RedisStreamConsumer) parseMessage(message redis.XMessage) (*domainEvents.Event, error) {
	// Extrair campos da mensagem
	eventID, ok := message.Values["event_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid event_id")
	}

	eventType, ok := message.Values["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid type")
	}

	userID, ok := message.Values["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid user_id")
	}

	eventCtx, ok := message.Values["event_ctx"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid event_ctx")
	}

	timestampStr, ok := message.Values["timestamp"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid timestamp")
	}

	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Parse data JSON
	dataStr, ok := message.Values["data"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid data")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse data JSON: %w", err)
	}

	// Parse metadata JSON
	metadataStr, ok := message.Values["metadata"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid metadata")
	}

	var metadata domainEvents.EventMetadata
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	// Criar evento
	event := &domainEvents.Event{
		ID:        eventID,
		Type:      domainEvents.EventType(eventType),
		StreamID:  message.ID,
		UserID:    userID,
		EventID:   eventCtx,
		Timestamp: timestamp,
		Data:      data,
		Metadata:  metadata,
	}

	return event, nil
}

// Ack confirma o processamento de um evento
func (c *RedisStreamConsumer) Ack(ctx context.Context, streamName, consumerGroup, eventID string) error {
	err := c.client.XAck(ctx, streamName, consumerGroup, eventID).Err()
	if err != nil {
		c.logger.Error("Failed to acknowledge event",
			"stream", streamName,
			"group", consumerGroup,
			"event_id", eventID,
			"error", err,
		)
		return err
	}

	c.logger.Debug("Event acknowledged",
		"stream", streamName,
		"group", consumerGroup,
		"event_id", eventID,
	)

	return nil
}

// Close fecha a conexão do consumer
func (c *RedisStreamConsumer) Close() error {
	// Como o cliente Redis é compartilhado, não fechamos aqui
	return nil
}

// RegisterHandler registra um handler para um tipo de evento
func (c *RedisStreamConsumer) RegisterHandler(eventType domainEvents.EventType, handler domainEvents.EventHandler) {
	if c.handlers[eventType] == nil {
		c.handlers[eventType] = make([]domainEvents.EventHandler, 0)
	}
	c.handlers[eventType] = append(c.handlers[eventType], handler)

	c.logger.Info("Event handler registered",
		"event_type", eventType,
		"handler_count", len(c.handlers[eventType]),
	)
}

// ProcessEvents processa eventos usando handlers registrados
func (c *RedisStreamConsumer) ProcessEvents(ctx context.Context, eventChan <-chan *domainEvents.Event, streamName, consumerGroup string) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping event processing")
			return

		case event, ok := <-eventChan:
			if !ok {
				c.logger.Info("Event channel closed, stopping processing")
				return
			}

			// Processar evento com handlers registrados
			c.processEvent(ctx, event, streamName, consumerGroup)
		}
	}
}

// processEvent processa um evento individual
func (c *RedisStreamConsumer) processEvent(ctx context.Context, event *domainEvents.Event, streamName, consumerGroup string) {
	handlers, exists := c.handlers[event.Type]
	if !exists || len(handlers) == 0 {
		c.logger.Error("No handlers registered for event type",
			"event_type", event.Type,
			"event_id", event.ID,
		)
		// Ainda assim fazemos ACK para não reprocessar
		_ = c.Ack(ctx, streamName, consumerGroup, event.StreamID)
		return
	}

	// Executar todos os handlers para este tipo de evento
	success := true
	for _, handler := range handlers {
		if handler.CanHandle(event.Type) {
			if err := handler.Handle(ctx, event); err != nil {
				c.logger.Error("Handler failed to process event",
					"event_type", event.Type,
					"event_id", event.ID,
					"handler", fmt.Sprintf("%T", handler),
					"error", err,
				)
				success = false
			} else {
				c.logger.Debug("Handler processed event successfully",
					"event_type", event.Type,
					"event_id", event.ID,
					"handler", fmt.Sprintf("%T", handler),
				)
			}
		}
	}

	// Fazer ACK apenas se todos os handlers executaram com sucesso
	if success {
		if err := c.Ack(ctx, streamName, consumerGroup, event.StreamID); err != nil {
			c.logger.Error("Failed to acknowledge successfully processed event",
				"event_id", event.ID,
				"stream_id", event.StreamID,
			)
		}
	} else {
		c.logger.Error("Event processing failed, will be retried",
			"event_id", event.ID,
			"stream_id", event.StreamID,
		)
	}
}
