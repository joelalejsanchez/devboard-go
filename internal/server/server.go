// Package server configures and runs the HTTP server.
package server

import (
	"context"
	"devboard/internal/middlewares"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server serves HTTP requests and manages the server lifecycle.
type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	logger     *slog.Logger
}

// Start runs the HTTP server and shuts it down gracefully when the context is canceled.
func (s *Server) Start() error {
	// Crear canal para notificar cuando se llama a apagar el proceso.
	// SIGTERM cierre pasivo
	// SIGINT  cierre forzoso, Ctrl+C,
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(quit)

	servErr := make(chan error, 1)
	go func() {
		s.logger.Info("Iniciando", slog.String("addr", "http://localhost:8000"))
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			servErr <- err
		}
	}()

	select {
	case err := <-servErr:
		return err
	case sig := <-quit:
		s.logger.Info("signal received, starting shutdown",
			slog.String("signal", sig.String()),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		s.logger.Error("error en shutdown", slog.Any("error", err))
	}
	s.logger.Info("server stopped")

	return nil
}

// RegisterRoute registers an HTTP handler for the given pattern.
func (s *Server) RegisterRoute(pattern string, h http.Handler) {
	s.mux.Handle(pattern, h)
}

// ServeHTTP dispatches a request to the registered handlers.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Use adds a middleware to the server.
func (s *Server) Use(middleware func(http.Handler) http.Handler) {
	s.httpServer.Handler = middleware(s.httpServer.Handler)
}

// UseChain adds multiple middlewares to the server.
func (s *Server) UseChain(m ...middlewares.Middleware) {
	s.httpServer.Handler = middlewares.Chain(s.httpServer.Handler, m...)
}

// New creates an HTTP server with the supplied address and options.
func New(addr string, opts ...Option) *Server {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,

		ReadTimeout:       cfg.readTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
		ReadHeaderTimeout: cfg.readHeaderTimeout,
	}

	return &Server{
		httpServer: httpServer,
		mux:        mux,
		logger:     &cfg.logger,
	}
}
