package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler gerencia o endpoint de health check
type HealthHandler struct {
	// Futuramente adicionaremos dependências para DB e Redis
}

// HealthResponse representa a resposta do health check
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// NewHealthHandler cria uma nova instância do handler de health check
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check verifica o status da aplicação e suas dependências
func (h *HealthHandler) Check(c *gin.Context) {
	// Por enquanto, vamos retornar apenas o status da API
	// Nas próximas etapas adicionaremos verificação de DB e Redis

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services: map[string]string{
			"api":      "healthy",
			"database": "not_configured", // Será implementado na próxima etapa
			"cache":    "not_configured", // Será implementado na próxima etapa
			"events":   "not_configured", // Será implementado na próxima etapa
		},
	}

	c.JSON(http.StatusOK, response)
}
