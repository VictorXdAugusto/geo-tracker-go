package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitao/geolocation-tracker/internal/infrastructure/events"
	"github.com/vitao/geolocation-tracker/internal/interfaces/http/routes"
	"github.com/vitao/geolocation-tracker/internal/wire"
	"github.com/vitao/geolocation-tracker/pkg/config"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

type Application struct {
	config       *config.Config
	logger       logger.Logger
	server       *http.Server
	container    *wire.Container
	eventService *events.EventService
}

// New cria uma nova instância da aplicação
func New() (*Application, error) {
	// Configurar logger estruturado
	log := logger.NewLogger()

	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Configurar Gin mode baseado no environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializar container via Wire
	container, err := wire.InitializeContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	// Inicializar Redis (extraído do container)
	redis, err := wire.InitializeRedis()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Inicializar event service
	eventService := events.NewEventService(redis, log)

	app := &Application{
		config:       cfg,
		logger:       log,
		container:    container,
		eventService: eventService,
	}

	return app, nil
}

// Start inicia a aplicação
func (a *Application) Start() error {
	a.logger.Info("Starting Geolocation Tracker Application...")

	// 1. Iniciar event service
	if err := a.eventService.Start(); err != nil {
		return fmt.Errorf("failed to start event service: %w", err)
	}

	// 2. Configurar rotas
	router := a.setupRoutes()

	// 3. Configurar servidor HTTP
	a.server = &http.Server{
		Addr:         ":" + a.config.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para capturar sinais de encerramento
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	go func() {
		a.logger.Info("Starting server",
			"port", a.config.Port,
			"environment", a.config.Environment,
		)

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Aguardar sinal de encerramento
	<-quit
	a.logger.Info("Shutting down server...")

	return a.gracefulShutdown()
}

// setupRoutes configura todas as rotas da aplicação
func (a *Application) setupRoutes() *gin.Engine {
	router := routes.SetupRoutes(
		a.container.CreateUser,
		a.container.SaveUserPosition,
		a.container.FindNearbyUsers,
		a.container.GetUsersInSector,
		a.container.GetCurrentPosition,
		a.container.GetPositionHistory,
		a.logger,
	)

	// Adicionar endpoint para estatísticas de eventos
	router.GET("/api/v1/events/stats", a.handleEventStats)

	return router
}

// handleEventStats retorna estatísticas dos eventos
func (a *Application) handleEventStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	stats, err := a.eventService.GetStats(ctx)
	if err != nil {
		a.logger.Error("Failed to get event stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get event statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// gracefulShutdown realiza o encerramento gracioso da aplicação
func (a *Application) gracefulShutdown() error {
	a.logger.Info("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Shutdown do servidor HTTP
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown", "error", err)
		return err
	}
	a.logger.Info("HTTP server stopped")

	// 2. Parar event service
	a.eventService.Stop()

	// 3. Sync dos logs pendentes
	if err := a.logger.Sync(); err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	a.logger.Info("Application shutdown completed")
	return nil
}
