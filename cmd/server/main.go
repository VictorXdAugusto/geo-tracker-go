package main

import (
	"log"

	"github.com/vitao/geolocation-tracker/internal/interfaces/http/routes"
	"github.com/vitao/geolocation-tracker/internal/wire"
	"github.com/vitao/geolocation-tracker/pkg/config"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

func main() {
	// 1. Inicializar container via Wire
	container, err := wire.InitializeContainer()
	if err != nil {
		log.Fatal("Failed to initialize container:", err)
	}

	// 2. Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 3. Criar logger
	logger := logger.NewLogger()
	defer logger.Sync()

	// 4. Setup HTTP routes usando o container
	router := routes.SetupRoutes(
		container.CreateUser,
		container.SaveUserPosition,
		container.FindNearbyUsers,
		container.GetUsersInSector,
		container.GetCurrentPosition,
		container.GetPositionHistory,
		logger,
	)

	// 5. Start server
	logger.Info("Starting geolocation tracker server with Wire DI", "port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
