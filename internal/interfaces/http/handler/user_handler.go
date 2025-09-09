package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// UserHandler gerencia endpoints relacionados a usuários
type UserHandler struct {
	createUserUC         *usecase.CreateUserUseCase
	getCurrentPositionUC *usecase.GetCurrentPositionUseCase
	getPositionHistoryUC *usecase.GetPositionHistoryUseCase
	logger               logger.Logger
}

// NewUserHandler cria uma nova instância do handler
func NewUserHandler(
	createUserUC *usecase.CreateUserUseCase,
	getCurrentPositionUC *usecase.GetCurrentPositionUseCase,
	getPositionHistoryUC *usecase.GetPositionHistoryUseCase,
	logger logger.Logger,
) *UserHandler {
	return &UserHandler{
		createUserUC:         createUserUC,
		getCurrentPositionUC: getCurrentPositionUC,
		getPositionHistoryUC: getPositionHistoryUC,
		logger:               logger,
	}
}

// CreateUser cria um novo usuário
// @Summary Criar um novo usuário
// @Description Cria um novo usuário no sistema para participar de um evento
// @Tags users
// @Accept json
// @Produce json
// @Param request body usecase.CreateUserRequest true "Dados do usuário"
// @Success 201 {object} usecase.CreateUserResponse "Usuário criado com sucesso"
// @Failure 400 {object} map[string]interface{} "Erro de validação"
// @Failure 409 {object} map[string]interface{} "Usuário já existe"
// @Failure 500 {object} map[string]interface{} "Erro interno do servidor"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req usecase.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request payload for create user", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Executar use case
	response, err := h.createUserUC.Execute(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create user", map[string]interface{}{
			"user_id": req.ID,
			"error":   err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User created successfully", map[string]interface{}{
		"user_id": response.UserID,
		"name":    response.Name,
	})

	c.JSON(http.StatusCreated, response)
}

// GetCurrentPosition retorna a posição atual do usuário
// @Summary Obter posição atual do usuário
// @Description Retorna a posição geográfica atual de um usuário específico
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID do usuário"
// @Success 200 {object} usecase.GetCurrentPositionResponse "Posição atual do usuário"
// @Failure 400 {object} map[string]interface{} "ID do usuário inválido"
// @Failure 404 {object} map[string]interface{} "Usuário não encontrado"
// @Failure 500 {object} map[string]interface{} "Erro interno do servidor"
// @Router /users/{id}/position [get]
func (h *UserHandler) GetCurrentPosition(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	// Converter para use case request
	ucRequest := usecase.GetCurrentPositionRequest{
		UserID: userID,
	}

	// Executar use case
	response, err := h.getCurrentPositionUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		h.logger.Error("Failed to get current position",
			"user_id", userID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get current position",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Current position retrieved successfully",
		"user_id", userID,
		"position_id", response.PositionID,
	)

	c.JSON(http.StatusOK, response)
}

// GetPositionHistory retorna o histórico de posições do usuário
// @Summary Obter histórico de posições do usuário
// @Description Retorna o histórico de posições geográficas de um usuário com limite configurável
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID do usuário"
// @Param limit query int false "Número máximo de posições a retornar (padrão: 10, máximo: 100)"
// @Success 200 {object} usecase.GetPositionHistoryResponse "Histórico de posições do usuário"
// @Failure 400 {object} map[string]interface{} "ID do usuário inválido"
// @Failure 404 {object} map[string]interface{} "Usuário não encontrado"
// @Failure 500 {object} map[string]interface{} "Erro interno do servidor"
// @Router /users/{id}/positions/history [get]
func (h *UserHandler) GetPositionHistory(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	// Parse do parâmetro limit
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // Valor padrão
	}
	if limit > 100 {
		limit = 100 // Máximo permitido
	}

	// Converter para use case request
	ucRequest := usecase.GetPositionHistoryRequest{
		UserID: userID,
		Limit:  limit,
	}

	// Executar use case
	response, err := h.getPositionHistoryUC.Execute(c.Request.Context(), ucRequest)
	if err != nil {
		h.logger.Error("Failed to get position history",
			"user_id", userID,
			"limit", limit,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get position history",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Position history retrieved successfully",
		"user_id", userID,
		"total", response.Total,
		"limit", limit,
	)

	c.JSON(http.StatusOK, response)
}
