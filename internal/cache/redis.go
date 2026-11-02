// Package cache coordinates cache storage
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config redis config
type Config struct {
	Addr     string
	Password string
	DB       int
	Timeout  time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Timeout:  5 * time.Second,
	}
}

// NewClient connects and returns a redis client
func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})
	pingCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("verificando conexion a redis: %w", err)
	}
	return client, nil
}
