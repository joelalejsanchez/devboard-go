// Package main - entrada al programa
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "devboard/docs"
	"devboard/internal/auth"
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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Escribe el token con el formato: Bearer {token}
func main() {

	/*
	 * ========================
	 * VARIABLES DE ENTORNO
	 * ========================
	 */

	env := os.Getenv("GO_ENV")
	pgURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_ADDR")
	jwtSecret := os.Getenv("JWT_SECRET")

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
	 * AUTH INFRASTRUCTURE
	 * ========================
	 */

	refreshTokenDuration := 7 * 24 * time.Hour
	jwtManager, err := auth.NewJWTManager(auth.JWTConfig{
		SecretKey:            []byte(jwtSecret),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: refreshTokenDuration,
	})

	if err != nil {
		lgr.Error("Error configurando JWT", slog.Any("error", err))
		os.Exit(1)
	}

	tokenStore := auth.NewTokenStore(redisClient)

	/*
	 * ========================
	 * SERVER
	 * ========================
	 */

	svr := server.New(
		"localhost:8000",
		server.WithLogger(lgr),
		server.WithReadTimeout(15*time.Second),
		server.WithWriteTimeout(30*time.Second),
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

	authUC := usecases.NewAuthUseCase(userRepo, jwtManager, tokenStore, refreshTokenDuration)
	authHandler := handlers.NewAuthHandler(authUC, vlt, lgr)

	/*
	 * ========================
	 * RUTAS Y HANDLERS
	 * ========================
	 */

	authMW := middlewares.RequireAuth(jwtManager, lgr)

	// Rutas Publicas
	svr.RegisterRoute("GET  /api/v1/docs/", httpSwagger.WrapHandler)
	svr.RegisterRoute("GET  /api/v1/health", &handlers.HealthHandler{})
	svr.RegisterRoute("POST /api/v1/auth/register", http.HandlerFunc(authHandler.Register))
	svr.RegisterRoute("POST /api/v1/auth/login", http.HandlerFunc(authHandler.Login))
	svr.RegisterRoute("POST /api/v1/auth/refresh", http.HandlerFunc(authHandler.Refresh))
	svr.RegisterRoute("POST /api/v1/auth/logout", http.HandlerFunc(authHandler.Logout))

	// Rutas Protegidas
	svr.RegisterRoute("GET  /api/v1/users",
		middlewares.Chain(
			http.HandlerFunc(userHandler.ListAll),
			authMW,
		))
	svr.RegisterRoute("POST /api/v1/users",
		middlewares.Chain(
			http.HandlerFunc(userHandler.Create),
			authMW,
		))
	svr.RegisterRoute("GET  /api/v1/users/{id}",
		middlewares.Chain(
			http.HandlerFunc(userHandler.Get),
			authMW,
		))
	svr.RegisterRoute("POST /api/v1/tasks",
		middlewares.Chain(
			http.HandlerFunc(taskHandler.Create),
			authMW,
		))
	svr.RegisterRoute("GET 	/api/v1/tasks/{id}",
		middlewares.Chain(
			http.HandlerFunc(taskHandler.GetByID),
			authMW,
		))
	svr.RegisterRoute("GET 	/api/v1/projects/{id}/tasks",
		middlewares.Chain(
			http.HandlerFunc(taskHandler.ListByProject),
			authMW,
		))
	svr.RegisterRoute("PUT  /api/v1/tasks/{id}/assign",
		middlewares.Chain(
			http.HandlerFunc(taskHandler.Assign),
			authMW,
		))
	svr.RegisterRoute("PUT 	/api/v1/tasks/{id}/status",
		middlewares.Chain(
			http.HandlerFunc(taskHandler.UpdateStatus),
			authMW,
		))

	/*
	 * ========================
	 * MIDDLEWARE GLOBALES
	 * ========================
	 */

	svr.UseChain(
		middlewares.Recovery(lgr),
		middlewares.Logger(lgr),
		middlewares.SecurityHeaders(middlewares.SecurityConfig{
			AllowedOrigins: []string{
				"http://localhost:3000",
				"http://localhost:5173",
			},
		}),
	)

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
