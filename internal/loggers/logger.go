// Package loggers - Configura loggers de aplicacion
package loggers

import (
	"io"
	"log/slog"
	"os"
)

// Config struct
type Config struct {
	Level  slog.Level
	JSON   bool
	Output io.Writer
}

// DefaultConfig retorna configuracion de desarrollo
func DefaultConfig() Config {
	return Config{
		Level:  slog.LevelDebug,
		JSON:   false,
		Output: os.Stdout,
	}
}

// ProductionConfig retorna configuracion de produccion
func ProductionConfig() Config {
	return Config{
		Level:  slog.LevelInfo,
		JSON:   true,
		Output: os.Stdout,
	}
}

// New retorna nuevo slog/logger
func New(config Config) *slog.Logger {
	out := config.Output
	if out == nil {
		out = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.Level == slog.LevelDebug,
	}

	var handler slog.Handler
	if config.JSON {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}
