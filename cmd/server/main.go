package main

import (
	"log"

	"github.com/vitao/geolocation-tracker/internal/app"
)

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
