package main

import (
	"log"

	_ "github.com/vitao/geolocation-tracker/docs" // Import docs for swagger
	"github.com/vitao/geolocation-tracker/internal/app"
)

// @title Geolocation Tracker API
// @version 1.0
// @description API para rastreamento de geolocalização de usuários em eventos
// @description Esta API permite criar usuários, salvar posições geográficas e consultar usuários próximos
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @tag.name users
// @tag.description Operações relacionadas a usuários

// @tag.name positions
// @tag.description Operações relacionadas a posições geográficas

// @tag.name health
// @tag.description Operações de health check

func main() {
	// Criar aplicação
	application, err := app.New()
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	// Iniciar aplicação
	if err := application.Start(); err != nil {
		log.Fatal("Failed to start application:", err)
	}
}
