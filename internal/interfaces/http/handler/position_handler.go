package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// PositionHandler gerencia endpoints relacionados a posições
type PositionHandler struct {
	savePositionUC     *usecase.SaveUserPositionUseCase
	findNearbyUC       *usecase.FindNearbyUsersUseCase
	getUsersInSectorUC *usecase.GetUsersInSectorUseCase
	logger             logger.Logger
}

// NewPositionHandler cria uma nova instância do handler
func NewPositionHandler(
	savePositionUC *usecase.SaveUserPositionUseCase,
	findNearbyUC *usecase.FindNearbyUsersUseCase,
	getUsersInSectorUC *usecase.GetUsersInSectorUseCase,
	logger logger.Logger,
) *PositionHandler {
	return &PositionHandler{
		savePositionUC:     savePositionUC,
		findNearbyUC:       findNearbyUC,
		getUsersInSectorUC: getUsersInSectorUC,
		logger:             logger,
	}
}

// SavePositionRequest representa o payload para salvar posição
type SavePositionRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

// SavePosition salva a posição de um usuário
// POST /api/v1/positions
func (h *PositionHandler) SavePosition(c *gin.Context) {
	var req SavePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request payload", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Converter para use case request
	ucRequest := usecase.SaveUserPositionRequest{
		UserID:    req.UserID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timestamp: time.Now(),
	}

	// Executar use case
	response, err := h.savePositionUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		h.logger.Error("Failed to save position",
			"user_id", req.UserID,
			"latitude", req.Latitude,
			"longitude", req.Longitude,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save position",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Position saved successfully",
		"user_id", req.UserID,
		"position_id", response.PositionID,
		"sector_id", response.SectorID,
	)

	c.JSON(http.StatusCreated, response)
}

// FindNearbyRequest representa o payload para buscar usuários próximos
type FindNearbyRequest struct {
	Latitude   float64 `form:"latitude" binding:"required,min=-90,max=90"`
	Longitude  float64 `form:"longitude" binding:"required,min=-180,max=180"`
	RadiusM    float64 `form:"radius_meters" binding:"required,min=1,max=50000"`
	MaxResults int     `form:"max_results"`
}

// FindNearbyUsers busca usuários próximos
// GET /api/v1/positions/nearby?user_id=X&latitude=Y&longitude=Z&radius_meters=R&max_results=N
func (h *PositionHandler) FindNearbyUsers(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	var req FindNearbyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid query parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Converter para use case request
	ucRequest := usecase.FindNearbyUsersRequest{
		UserID:     userID,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		RadiusM:    req.RadiusM,
		MaxResults: req.MaxResults,
	}

	// Executar use case
	response, err := h.findNearbyUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		h.logger.Error("Failed to find nearby users",
			"user_id", userID,
			"latitude", req.Latitude,
			"longitude", req.Longitude,
			"radius", req.RadiusM,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to find nearby users",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Nearby users search completed",
		"user_id", userID,
		"total_found", response.TotalFound,
	)

	c.JSON(http.StatusOK, response)
}

// GetUsersInSectorRequest representa o payload para buscar usuários no setor
type GetUsersInSectorRequest struct {
	Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"longitude" binding:"required,min=-180,max=180"`
}

// GetUsersInSector busca usuários no mesmo setor
// GET /api/v1/positions/sector?user_id=X&latitude=Y&longitude=Z
func (h *PositionHandler) GetUsersInSector(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	var req GetUsersInSectorRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid query parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Converter para use case request
	ucRequest := usecase.GetUsersInSectorRequest{
		UserID:    userID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	// Executar use case
	response, err := h.getUsersInSectorUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		h.logger.Error("Failed to get users in sector",
			"user_id", userID,
			"latitude", req.Latitude,
			"longitude", req.Longitude,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get users in sector",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Sector users search completed",
		"user_id", userID,
		"sector_id", response.SectorID,
		"total_found", response.TotalFound,
	)

	c.JSON(http.StatusOK, response)
}
