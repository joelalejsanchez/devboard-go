// Package auth contains auth and crypto functions
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrTokenRevoked token revoked
var ErrTokenRevoked = errors.New("token revocado")

// ErrTokenNotFound token not found
var ErrTokenNotFound = errors.New("token no encontrado")

// TokenStore struct
type TokenStore struct {
	redis *redis.Client
}

// NewTokenStore constructor
func NewTokenStore(redisClient *redis.Client) *TokenStore {
	return &TokenStore{redis: redisClient}
}

func refreshKey(tokenID string) string {
	return "refresh:" + tokenID
}

func revokedKey(tokenID string) string {
	return "revoked:" + tokenID
}

// StoreRefreshToken stores refresh token
func (s *TokenStore) StoreRefreshToken(ctx context.Context, tokenID, userID string, ttl time.Duration) error {
	err := s.redis.Set(
		ctx,
		refreshKey(tokenID),
		userID,
		ttl,
	).Err()

	if err != nil {
		return fmt.Errorf("no se puedo guardar el token: %w", err)
	}
	return nil
}

// ValidateRefreshToken verifica que un token no sea revocado
func (s *TokenStore) ValidateRefreshToken(ctx context.Context, tokenID string) (string, error) {
	revoked, err := s.redis.Exists(ctx, revokedKey(tokenID)).Result()
	if err != nil {
		return "", fmt.Errorf("verificando revocacion: %w", err)
	}
	if revoked > 0 {
		return "", ErrTokenRevoked
	}

	userID, err := s.redis.Get(ctx, refreshKey(tokenID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTokenNotFound
		}
		return "", fmt.Errorf("obteniendo refresh token: %w", err)
	}

	return userID, nil
}

// RevokeRefreshToken stores revoked token and removes refresh token
func (s *TokenStore) RevokeRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, revokedKey(tokenID), "1", ttl)
	pipe.Del(ctx, refreshKey(tokenID))

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("revocando refresh token: %w", err)
	}
	return nil
}
