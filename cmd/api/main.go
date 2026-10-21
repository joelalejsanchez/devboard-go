package main

import (
	"log"

	"devboard/internal/handler"
	"devboard/internal/server"
)

func main() {
	healthHandler := handler.NewHealthHandler()
	server := server.New("localhost:8000")
	server.RegisterRoute("GET /1", healthHandler)
	server.RegisterRoute("GET /2", healthHandler)

	log.Println("Iniciando en http://localhost:8000")
	err := server.Start()
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

}
