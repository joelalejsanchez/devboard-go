// Package main - entrada al programa
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "devboard/docs"
	"devboard/internal/handlers"
	_ "devboard/internal/handlers/utils"
	"devboard/internal/loggers"
	"devboard/internal/middlewares"
	"devboard/internal/repository/memory"
	"devboard/internal/server"
	"devboard/internal/usecases"
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

	lgr := setupLogger()
	svr := setupServer(lgr)
	setupRoutes(svr, lgr)
	if err := svr.Start(); err != nil {
		lgr.Error("error fatal", "error", err)
		os.Exit(1)
	}

}

func setupServer(lgr *slog.Logger) *server.Server {

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

	return svr
}
func setupLogger() *slog.Logger {
	env := os.Getenv("GO_ENV")

	if env == "production" {
		return loggers.New(loggers.ProductionConfig())
	}

	return loggers.New(loggers.DefaultConfig())
}

func setupRoutes(svr *server.Server, lgr *slog.Logger) {
	vlt := validator.New()

	userRepo := memory.NewUserRepository()
	uUC := usecases.NewUserUseCase(userRepo)
	userHandler := handlers.NewUserHandler(uUC, vlt, lgr)

	taskRepo := memory.NewTaskRepository()
	tUC := usecases.NewTaskUseCase(taskRepo, taskRepo)
	taskHandler := handlers.NewTaskHandler(tUC, vlt, lgr)

	svr.RegisterRoute("GET  /api/v1/docs/", httpSwagger.WrapHandler)
	svr.RegisterRoute("GET  /api/v1/health", &handlers.HealthHandler{})
	svr.RegisterRoute("POST /api/v1/users", http.HandlerFunc(userHandler.Create))
	svr.RegisterRoute("GET  /api/v1/users/{id}", http.HandlerFunc(userHandler.Get))
	svr.RegisterRoute("POST /api/v1/tasks", http.HandlerFunc(taskHandler.Create))
	svr.RegisterRoute("GET 	/api/v1/tasks/{id}", http.HandlerFunc(taskHandler.GetByID))
	svr.RegisterRoute("GET 	/api/v1/projects/{id}/tasks", http.HandlerFunc(taskHandler.ListByProject))
	svr.RegisterRoute("PUT  /api/v1/tasks/{id}/assign", http.HandlerFunc(taskHandler.Assign))
	svr.RegisterRoute("PUT 	/api/v1/tasks/{id}/status", http.HandlerFunc(taskHandler.UpdateStatus))
}
