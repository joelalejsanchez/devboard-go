package main

import (
	"log"
	"log/slog"
	"os"

	"devboard/internal/handlers"
	"devboard/internal/middlewares"
	"devboard/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	server := server.New("localhost:8000")

	server.Use(middlewares.Logger(logger))

	healthHandler := handlers.NewHealthHandler()
	server.RegisterRoute("GET /1", healthHandler)
	server.RegisterRoute("GET /2", healthHandler)

	logger.Info("Iniciando", slog.String("addr", "http://localhost:8000"))
	err := server.Start()
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

}
