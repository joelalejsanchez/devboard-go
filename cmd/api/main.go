// Package main - entrada al programa
package main

import (
	"log/slog"
	"os"
	"time"

	_ "devboard/docs"
	"devboard/internal/handlers"
	"devboard/internal/loggers"
	"devboard/internal/middlewares"
	"devboard/internal/notification"
	"devboard/internal/server"
	"devboard/internal/validator"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title DEVBOARD API
// @version 0.1
// @description API REST para gestión de proyectos
// @contact.name Soporte Devboard Joel
// @host localhost:8000
// @BasePath /api/v1
func main() {

	// 1. Configuracion del entorno
	env := os.Getenv("GO_ENV")

	// 2. Logger
	var lgr *slog.Logger
	if env == "production" {
		lgr = loggers.New(loggers.ProductionConfig())
	} else {
		lgr = loggers.New(loggers.DefaultConfig())
	}

	// 3. Dependencias compartidas
	validatorInstance := validator.New()
	_ = validatorInstance
	notifierInstance := notification.NewLogNotifier(lgr)
	_ = notifierInstance

	// 4. Configuracion de servidor
	svr := server.New(
		"localhost:8000",
		server.WithLogger(lgr),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(30*time.Second),
	)
	// 6. handlers si existen
	// 7. rutas
	svr.RegisterRoute("GET /docs/", httpSwagger.WrapHandler)
	svr.RegisterRoute("GET /1", &handlers.HealthHandler{})
	svr.RegisterRoute("GET /users/one", &handlers.GetUserHandler{})
	svr.RegisterRoute("POST /users", &handlers.PostUserHandler{Logger: lgr})

	// 8. middlewares , esto funciona dado que se autoinyecta el handler
	svr.UseChain(
		middlewares.Recovery(lgr),
		middlewares.Logger(lgr),
	)

	// 9. inicio del servidor

	if err := svr.Start(); err != nil {
		lgr.Error("error fatal", "error", err)
		os.Exit(1)
	}

}
