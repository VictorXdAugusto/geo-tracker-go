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
	"github.com/vitao/geolocation-tracker/internal/interfaces/http/routes"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/pkg/config"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

type Application struct {
	config *config.Config
	logger logger.Logger
	server *http.Server
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

	app := &Application{
		config: cfg,
		logger: log,
	}

	return app, nil
}

// Start inicia a aplicação
func (a *Application) Start() error {
	// Configurar rotas
	router := a.setupRoutes()

	// Configurar servidor HTTP
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
	router := gin.New()

	// Middlewares básicos do Gin
	router.Use(gin.Recovery())

	// Create use case instances
	createUserUseCase := &usecase.CreateUserUseCase{}
	saveUserPositionUseCase := &usecase.SaveUserPositionUseCase{}
	findNearbyUsersUseCase := &usecase.FindNearbyUsersUseCase{}
	getUsersInSectorUseCase := &usecase.GetUsersInSectorUseCase{}
	getCurrentPositionUseCase := &usecase.GetCurrentPositionUseCase{}
	getPositionHistoryUseCase := &usecase.GetPositionHistoryUseCase{}

	// Configurar todas as rotas (middlewares incluídos internamente)
	routes.SetupRoutes(createUserUseCase, saveUserPositionUseCase, findNearbyUsersUseCase, getUsersInSectorUseCase, getCurrentPositionUseCase, getPositionHistoryUseCase, a.logger)

	return router
}

// gracefulShutdown realiza o encerramento gracioso da aplicação
func (a *Application) gracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown do servidor HTTP
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	// Sync dos logs pendentes
	if err := a.logger.Sync(); err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	a.logger.Info("Server exited gracefully")
	return nil
}
