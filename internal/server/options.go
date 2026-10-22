// Package server - configuraciones de servidor
package server

import (
	"log/slog"
	"time"
)

type config struct {
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	readHeaderTimeout time.Duration
	logger            slog.Logger
}

// defaultConfig configuracion por defecto
func defaultConfig() config {
	return config{
		readTimeout:       10 * time.Second,
		writeTimeout:      30 * time.Second,
		idleTimeout:       60 * time.Second,
		readHeaderTimeout: 5 * time.Second,
		logger:            *slog.Default(),
	}
}

// Option funcion que modifica la configuracion
type Option func(*config)

// WithReadHeaderTimeout configura timeout de lectura de header
func WithReadHeaderTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.readHeaderTimeout = duration
	}
}

// WithReadTimeout configura timeout de lectura
func WithReadTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.readTimeout = duration
	}
}

// WithWriteTimeout configura timeout de escritura
func WithWriteTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.writeTimeout = duration
	}
}

// WithIdleTimeout configura timeout de inactividad de conexion
func WithIdleTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.idleTimeout = duration
	}
}

// WithLogger configura el logger del servidor
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = *logger
	}
}
