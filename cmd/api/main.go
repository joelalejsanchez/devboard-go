package main

import (
	"log"
	"log/slog"

	"devboard/internal/handlers"
	"devboard/internal/loggers"
	"devboard/internal/middlewares"
	"devboard/internal/server"
)

func main() {
	lgr := loggers.New(loggers.DefaultConfig())
	svr := server.New("localhost:8000")

	svr.Use(middlewares.Recovery(lgr))
	svr.Use(middlewares.Logger(lgr))

	healthHandler := handlers.NewHealthHandler()
	svr.RegisterRoute("GET /1", healthHandler)
	svr.RegisterRoute("GET /2", healthHandler)

	lgr.Info("Iniciando", slog.String("addr", "http://localhost:8000"))
	err := svr.Start()
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

}
