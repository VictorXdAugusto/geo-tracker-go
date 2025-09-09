package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/vitao/geolocation-tracker/internal/interfaces/http/handler"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(
	createUserUC *usecase.CreateUserUseCase,
	savePositionUC *usecase.SaveUserPositionUseCase,
	findNearbyUC *usecase.FindNearbyUsersUseCase,
	getUsersInSectorUC *usecase.GetUsersInSectorUseCase,
	getCurrentPositionUC *usecase.GetCurrentPositionUseCase,
	getPositionHistoryUC *usecase.GetPositionHistoryUseCase,
	logger logger.Logger,
) *gin.Engine {

	// Criar router Gin
	router := gin.New()

	// Middlewares básicos
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Health check
	// @Summary Health Check
	// @Description Verifica se o serviço está funcionando corretamente
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]string "Serviço saudável"
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "geolocation-tracker",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Criar handlers
	userHandler := handler.NewUserHandler(
		createUserUC,
		getCurrentPositionUC,
		getPositionHistoryUC,
		logger,
	)

	positionHandler := handler.NewPositionHandler(
		savePositionUC,
		findNearbyUC,
		getUsersInSectorUC,
		logger,
	)

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Rotas de usuários
		api.POST("/users", userHandler.CreateUser)
		api.GET("/users/:id/position", userHandler.GetCurrentPosition)
		api.GET("/users/:id/positions/history", userHandler.GetPositionHistory)

		// Rotas de posições
		api.POST("/positions", positionHandler.SavePosition)
		api.GET("/positions/nearby", positionHandler.FindNearbyUsers)
		api.GET("/positions/sector", positionHandler.GetUsersInSector)
	}

	return router
}
