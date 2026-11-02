// Package main - entrada al programa
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "devboard/docs"
	"devboard/internal/cache"
	"devboard/internal/handlers"
	_ "devboard/internal/handlers/utils"
	"devboard/internal/loggers"
	"devboard/internal/middlewares"
	"devboard/internal/repository/postgres"
	"devboard/internal/server"
	"devboard/internal/usecases"
	"devboard/internal/validator"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title DEVBOARD API
// @version 0.1
// @description REST API for project management.
// @contact.name Soporte Devboard Joel
// @host localhost:8000
// @BasePath /api/v1
func main() {

	/*
	 * ========================
	 * VARIABLES DE ENTORNO
	 * ========================
	 */

	env := os.Getenv("GO_ENV")
	pgURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_ADDR")

	/*
	 * ========================
	 * VALORES GLOBALES
	 * ========================
	 */
	var lgr *slog.Logger
	if env == "production" {
		lgr = loggers.New(loggers.ProductionConfig())
	} else {
		lgr = loggers.New(loggers.DefaultConfig())
	}

	vlt := validator.New()
	// Postgres connection
	postgresPool, err := postgres.NewPool(context.Background(), postgres.Config{
		URL:             pgURL,
		MaxConns:        20,
		MinConns:        5,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	})
	if err != nil {
		lgr.Error("no se pudo conectar a base de datos", slog.Any("error", err))
		os.Exit(1)
	}
	defer postgresPool.Close()
	lgr.Info("conexion a base de datos iniciada")

	redisClient, err := cache.NewClient(context.Background(), cache.Config{
		Addr:    redisURL,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		lgr.Error("no se pudo conectar a redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			lgr.Error("no se pudo cerrar conexion a redis",
				slog.Any("error", err),
			)
		}
	}()
	lgr.Info("conexion a redis iniciada")

	/*
	 * ========================
	 * SERVER Y MIDDLWARES
	 * ========================
	 */

	svr := server.New(
		"localhost:8000",
		server.WithLogger(lgr),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(30*time.Second),
	)
	svr.UseChain(
		middlewares.Recovery(lgr),
		middlewares.Logger(lgr),
	)

	/*
	 * ========================
	 * DEPENDENCIAS
	 * ========================
	 */

	userRepo := postgres.NewUserRepository(postgresPool)
	uUC := usecases.NewUserUseCase(userRepo)
	userHandler := handlers.NewUserHandler(uUC, vlt, lgr)

	taskRepo := postgres.NewTaskRepository(postgresPool)
	tUC := usecases.NewTaskUseCase(taskRepo, taskRepo)
	taskHandler := handlers.NewTaskHandler(tUC, vlt, lgr)

	/*
	 * ========================
	 * RUTAS Y HANDLERS
	 * ========================
	 */
	svr.RegisterRoute("GET  /api/v1/docs/", httpSwagger.WrapHandler)
	svr.RegisterRoute("GET  /api/v1/health", &handlers.HealthHandler{})
	svr.RegisterRoute("GET /api/v1/users", http.HandlerFunc(userHandler.ListAll))
	svr.RegisterRoute("POST /api/v1/users", http.HandlerFunc(userHandler.Create))
	svr.RegisterRoute("GET  /api/v1/users/{id}", http.HandlerFunc(userHandler.Get))
	svr.RegisterRoute("POST /api/v1/tasks", http.HandlerFunc(taskHandler.Create))
	svr.RegisterRoute("GET 	/api/v1/tasks/{id}", http.HandlerFunc(taskHandler.GetByID))
	svr.RegisterRoute("GET 	/api/v1/projects/{id}/tasks", http.HandlerFunc(taskHandler.ListByProject))
	svr.RegisterRoute("PUT  /api/v1/tasks/{id}/assign", http.HandlerFunc(taskHandler.Assign))
	svr.RegisterRoute("PUT 	/api/v1/tasks/{id}/status", http.HandlerFunc(taskHandler.UpdateStatus))

	/*
	 * ========================
	 * INICIO DE SERVIDOR
	 * ========================
	 */
	if err := svr.Start(); err != nil {
		lgr.Error("error fatal", "error", err)
		os.Exit(1)
	}
}
