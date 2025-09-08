package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// RequestLogger middleware para logging estruturado de requisições
func RequestLogger(logger logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}

// CORS middleware para configurar headers CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimiter middleware básico para rate limiting (placeholder)
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implementar rate limiting com Redis
		// Por enquanto, apenas pass-through
		c.Next()
	}
}

// SecurityHeaders middleware para adicionar headers de segurança
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// Timeout middleware para timeout de requisições
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Criar contexto com timeout
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Substituir contexto da requisição
		c.Request = c.Request.WithContext(ctx)

		// Channel para verificar se a requisição completou
		done := make(chan struct{})

		go func() {
			defer close(done)
			c.Next()
		}()

		select {
		case <-done:
			// Requisição completou normalmente
			return
		case <-ctx.Done():
			// Timeout ocorreu
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
				"code":  "TIMEOUT",
			})
			c.Abort()
		}
	}
}

// ErrorHandler middleware para tratamento centralizado de erros
func ErrorHandler(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Verificar se houve erro
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			logger.Error("Request error",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)

			// Retornar erro formatado
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
		}
	}
}
