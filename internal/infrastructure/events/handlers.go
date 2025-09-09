package events

import (
	"context"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/events"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// NotificationHandler processa eventos para enviar notificações
type NotificationHandler struct {
	logger logger.Logger
}

// NewNotificationHandler cria um novo handler de notificações
func NewNotificationHandler(logger logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		logger: logger,
	}
}

// Handle processa eventos de posição para notificações
func (h *NotificationHandler) Handle(ctx context.Context, event *events.Event) error {
	switch event.Type {
	case events.EventTypePositionChanged:
		return h.handlePositionChanged(ctx, event)
	case events.EventTypeUserEnteredSector:
		return h.handleUserEnteredSector(ctx, event)
	case events.EventTypeUserLeftSector:
		return h.handleUserLeftSector(ctx, event)
	default:
		return fmt.Errorf("unsupported event type: %s", event.Type)
	}
}

// CanHandle verifica se pode processar este tipo de evento
func (h *NotificationHandler) CanHandle(eventType events.EventType) bool {
	return eventType == events.EventTypePositionChanged ||
		eventType == events.EventTypeUserEnteredSector ||
		eventType == events.EventTypeUserLeftSector
}

// handlePositionChanged processa eventos de mudança de posição
func (h *NotificationHandler) handlePositionChanged(ctx context.Context, event *events.Event) error {
	// Extrair dados do evento
	newLat, _ := event.Data["new_lat"].(float64)
	newLng, _ := event.Data["new_lng"].(float64)
	distanceMoved, _ := event.Data["distance_moved"].(float64)
	newSector, _ := event.Data["new_sector"].(string)
	previousSector, _ := event.Data["previous_sector"].(string)

	h.logger.Info("Position Changed Notification",
		"user_id", event.UserID,
		"event_id", event.ID,
		"new_position", fmt.Sprintf("%.6f,%.6f", newLat, newLng),
		"distance_moved_m", distanceMoved,
		"new_sector", newSector,
		"previous_sector", previousSector,
		"timestamp", event.Timestamp.Format("15:04:05"),
	)

	// Simular notificação
	if distanceMoved > 100 { // Só notificar se moveu mais de 100m
		h.logger.Info("Sending push notification",
			"user_id", event.UserID,
			"message", fmt.Sprintf("You moved %.0fm to sector %s", distanceMoved, newSector),
		)
	}

	return nil
}

// handleUserEnteredSector processa eventos de entrada em setor
func (h *NotificationHandler) handleUserEnteredSector(ctx context.Context, event *events.Event) error {
	sectorID, _ := event.Data["sector_id"].(string)
	usersInSector, _ := event.Data["users_in_sector"].(float64) // JSON numbers are float64

	h.logger.Info("User Entered Sector Notification",
		"user_id", event.UserID,
		"sector_id", sectorID,
		"users_in_sector", int(usersInSector),
		"timestamp", event.Timestamp.Format("15:04:05"),
	)

	// Notificar outros usuários no setor
	if int(usersInSector) > 1 {
		h.logger.Info("Notifying other users in sector",
			"sector_id", sectorID,
			"total_users", int(usersInSector),
		)
	}

	return nil
}

// handleUserLeftSector processa eventos de saída de setor
func (h *NotificationHandler) handleUserLeftSector(ctx context.Context, event *events.Event) error {
	sectorID, _ := event.Data["sector_id"].(string)

	h.logger.Info("User Left Sector Notification",
		"user_id", event.UserID,
		"sector_id", sectorID,
		"timestamp", event.Timestamp.Format("15:04:05"),
	)

	return nil
}

// AnalyticsHandler processa eventos para analytics e métricas
type AnalyticsHandler struct {
	logger logger.Logger
}

// NewAnalyticsHandler cria um novo handler de analytics
func NewAnalyticsHandler(logger logger.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		logger: logger,
	}
}

// Handle processa eventos para analytics
func (h *AnalyticsHandler) Handle(ctx context.Context, event *events.Event) error {
	switch event.Type {
	case events.EventTypePositionChanged:
		return h.trackPositionChange(ctx, event)
	default:
		return fmt.Errorf("unsupported event type for analytics: %s", event.Type)
	}
}

// CanHandle verifica se pode processar este tipo de evento
func (h *AnalyticsHandler) CanHandle(eventType events.EventType) bool {
	return eventType == events.EventTypePositionChanged
}

// trackPositionChange registra métricas de mudança de posição
func (h *AnalyticsHandler) trackPositionChange(ctx context.Context, event *events.Event) error {
	distanceMoved, _ := event.Data["distance_moved"].(float64)
	newSector, _ := event.Data["new_sector"].(string)
	previousSector, _ := event.Data["previous_sector"].(string)

	h.logger.Info("Analytics: Position Change",
		"user_id", event.UserID,
		"distance_moved", distanceMoved,
		"sector_changed", newSector != previousSector,
		"new_sector", newSector,
		"timestamp", event.Timestamp.Format("15:04:05"),
	)

	return nil
}

// RealtimeHandler processa eventos para atualizações em tempo real
type RealtimeHandler struct {
	logger logger.Logger
}

// NewRealtimeHandler cria um novo handler de tempo real
func NewRealtimeHandler(logger logger.Logger) *RealtimeHandler {
	return &RealtimeHandler{
		logger: logger,
	}
}

// Handle processa eventos para tempo real
func (h *RealtimeHandler) Handle(ctx context.Context, event *events.Event) error {
	switch event.Type {
	case events.EventTypePositionChanged:
		return h.broadcastPositionUpdate(ctx, event)
	default:
		return fmt.Errorf("unsupported event type for realtime: %s", event.Type)
	}
}

// CanHandle verifica se pode processar este tipo de evento
func (h *RealtimeHandler) CanHandle(eventType events.EventType) bool {
	return eventType == events.EventTypePositionChanged
}

// broadcastPositionUpdate envia atualizações via WebSocket
func (h *RealtimeHandler) broadcastPositionUpdate(ctx context.Context, event *events.Event) error {
	newLat, _ := event.Data["new_lat"].(float64)
	newLng, _ := event.Data["new_lng"].(float64)
	newSector, _ := event.Data["new_sector"].(string)

	h.logger.Info("Realtime: Broadcasting Position Update",
		"user_id", event.UserID,
		"position", fmt.Sprintf("%.6f,%.6f", newLat, newLng),
		"sector", newSector,
		"timestamp", event.Timestamp.Format("15:04:05"),
	)

	return nil
}
