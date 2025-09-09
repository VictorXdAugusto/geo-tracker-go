package events

import (
	"context"
	"sync"

	"github.com/vitao/geolocation-tracker/internal/domain/events"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/cache"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// EventService gerencia publishers e consumers de eventos
type EventService struct {
	publisher *RedisStreamPublisher
	consumer  *RedisStreamConsumer
	logger    logger.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewEventService cria um novo service de eventos
func NewEventService(redis *cache.Redis, logger logger.Logger) *EventService {
	ctx, cancel := context.WithCancel(context.Background())

	publisher := NewRedisStreamPublisher(redis.Client(), logger)
	consumer := NewRedisStreamConsumer(redis.Client(), logger)

	return &EventService{
		publisher: publisher,
		consumer:  consumer,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start inicializa os streams e consumers
func (s *EventService) Start() error {
	s.logger.Info("Starting Event Service...")

	// 1. Inicializar streams no Redis
	if err := s.publisher.InitializeStreams(s.ctx); err != nil {
		return err
	}

	// 2. Registrar handlers
	s.registerEventHandlers()

	// 3. Iniciar consumers
	s.startConsumers()

	s.logger.Info("Event Service started successfully")
	return nil
}

// Stop para o service de eventos
func (s *EventService) Stop() {
	s.logger.Info("Stopping Event Service...")

	s.cancel()  // Cancela o contexto
	s.wg.Wait() // Aguarda todas as goroutines terminarem

	s.logger.Info("Event Service stopped")
}

// Publisher retorna o publisher para uso em use cases
func (s *EventService) Publisher() events.Publisher {
	return s.publisher
}

// registerEventHandlers registra todos os handlers de eventos
func (s *EventService) registerEventHandlers() {
	// Handlers para notificações
	notificationHandler := NewNotificationHandler(s.logger)
	s.consumer.RegisterHandler(events.EventTypePositionChanged, notificationHandler)
	s.consumer.RegisterHandler(events.EventTypeUserEnteredSector, notificationHandler)
	s.consumer.RegisterHandler(events.EventTypeUserLeftSector, notificationHandler)

	// Handlers para analytics
	analyticsHandler := NewAnalyticsHandler(s.logger)
	s.consumer.RegisterHandler(events.EventTypePositionChanged, analyticsHandler)

	// Handlers para tempo real
	realtimeHandler := NewRealtimeHandler(s.logger)
	s.consumer.RegisterHandler(events.EventTypePositionChanged, realtimeHandler)

	s.logger.Info("Event handlers registered",
		"notification_types", 3,
		"analytics_types", 1,
		"realtime_types", 1,
	)
}

// startConsumers inicia todos os consumers necessários
func (s *EventService) startConsumers() {
	// Consumer para notificações
	s.startConsumer(
		events.StreamPositionEvents,
		events.ConsumerGroupNotifications,
		"notification-worker-1",
	)

	// Consumer para analytics
	s.startConsumer(
		events.StreamPositionEvents,
		events.ConsumerGroupAnalytics,
		"analytics-worker-1",
	)

	// Consumer para tempo real
	s.startConsumer(
		events.StreamPositionEvents,
		events.ConsumerGroupRealtime,
		"realtime-worker-1",
	)
}

// startConsumer inicia um consumer específico
func (s *EventService) startConsumer(streamName, consumerGroup, consumerName string) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		s.logger.Info("Starting consumer",
			"stream", streamName,
			"group", consumerGroup,
			"consumer", consumerName,
		)

		// Subscribe ao stream
		eventChan, err := s.consumer.Subscribe(s.ctx, streamName, consumerGroup, consumerName)
		if err != nil {
			s.logger.Error("Failed to subscribe consumer",
				"stream", streamName,
				"group", consumerGroup,
				"consumer", consumerName,
				"error", err,
			)
			return
		}

		// Processar eventos
		s.consumer.ProcessEvents(s.ctx, eventChan, streamName, consumerGroup)

		s.logger.Info("Consumer stopped",
			"stream", streamName,
			"group", consumerGroup,
			"consumer", consumerName,
		)
	}()
}

// GetStats retorna estatísticas dos streams
func (s *EventService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Estatísticas do stream de posições
	positionLen, err := s.publisher.client.XLen(ctx, events.StreamPositionEvents).Result()
	if err != nil {
		return nil, err
	}

	// Lista básica dos consumer groups conhecidos
	consumerGroups := []string{
		events.ConsumerGroupNotifications,
		events.ConsumerGroupAnalytics,
		events.ConsumerGroupRealtime,
	}

	stats["streams"] = map[string]interface{}{
		events.StreamPositionEvents: map[string]interface{}{
			"length": positionLen,
			"groups": len(consumerGroups),
		},
	}

	stats["consumer_groups"] = make(map[string]interface{})
	for _, groupName := range consumerGroups {
		// Para cada grupo, tentamos obter informações básicas
		stats["consumer_groups"].(map[string]interface{})[groupName] = map[string]interface{}{
			"name":   groupName,
			"active": true, // Simplificado - assumimos que estão ativos
		}
	}

	// Adicionar timestamp da consulta
	stats["generated_at"] = ctx.Value("timestamp")
	if stats["generated_at"] == nil {
		stats["generated_at"] = "now"
	}

	return stats, nil
}
