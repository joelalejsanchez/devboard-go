// Package main - entrada al programa
package main

import (
	"os"

	_ "devboard/docs"
	"devboard/internal/handlers"
	"devboard/internal/loggers"
	"devboard/internal/middlewares"
	"devboard/internal/server"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title DEVBOARD API
// @version 0.1
// @description API REST para gestión de proyectos
// @contact.name Soporte Devboard Joel
// @host localhost:8000
// @BasePath /api/v1
func main() {
	lgr := loggers.New(loggers.DefaultConfig())
	svr := server.New(
		"localhost:8000",
		server.WithLogger(lgr),
	)

	// Middlewares
	svr.Use(middlewares.Recovery(lgr))
	svr.Use(middlewares.Logger(lgr))

	// Rutas
	svr.RegisterRoute("GET /docs/", httpSwagger.WrapHandler)
	svr.RegisterRoute("GET /1", &handlers.HealthHandler{})
	svr.RegisterRoute("GET /users/one", &handlers.GetUserHandler{})
	svr.RegisterRoute("POST /users", &handlers.PostUserHandler{Logger: lgr})

	if err := svr.Start(); err != nil {
		lgr.Error("error fatal", "error", err)
		os.Exit(1)
	}

}
