package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	logger     *slog.Logger
}

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

func (s *Server) RegisterRoute(pattern string, h http.Handler) {
	s.mux.Handle(pattern, h)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Use(middleware func(http.Handler) http.Handler) {
	s.httpServer.Handler = middleware(s.httpServer.Handler)
}

func New(addr string, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,

		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		mux:        mux,
		logger:     logger,
	}
}
