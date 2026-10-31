// Package server configures and runs the HTTP server.
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

// defaultConfig returns the default server configuration.
func defaultConfig() config {
	return config{
		readTimeout:       10 * time.Second,
		writeTimeout:      30 * time.Second,
		idleTimeout:       60 * time.Second,
		readHeaderTimeout: 5 * time.Second,
		logger:            *slog.Default(),
	}
}

// Option configures a server.
type Option func(*config)

// WithReadHeaderTimeout sets the maximum time to read request headers.
func WithReadHeaderTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.readHeaderTimeout = duration
	}
}

// WithReadTimeout sets the maximum duration for reading the entire request.
func WithReadTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.readTimeout = duration
	}
}

// WithWriteTimeout sets the maximum duration before a response write times out.
func WithWriteTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.writeTimeout = duration
	}
}

// WithIdleTimeout sets the maximum time to wait for the next request on a connection.
func WithIdleTimeout(duration time.Duration) Option {
	return func(c *config) {
		c.idleTimeout = duration
	}
}

// WithLogger sets the logger used by the server.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = *logger
	}
}
