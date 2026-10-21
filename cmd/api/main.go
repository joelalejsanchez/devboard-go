package main

import (
	"os"

	"devboard/internal/handlers"
	"devboard/internal/loggers"
	"devboard/internal/middlewares"
	"devboard/internal/server"
)

func main() {
	lgr := loggers.New(loggers.DefaultConfig())
	svr := server.New("localhost:8000", lgr)

	// Middlewares
	svr.Use(middlewares.Recovery(lgr))
	svr.Use(middlewares.Logger(lgr))

	// Rutas
	svr.RegisterRoute("GET /1", handlers.NewHealthHandler())

	if err := svr.Start(); err != nil {
		lgr.Error("error fatal", "error", err)
		os.Exit(1)
	}

}
