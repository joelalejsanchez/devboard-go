// Package auth contains auth and crypto functions
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrTokenExpired token expired
var ErrTokenExpired = errors.New("token expirado")

// ErrTokenInvalid token invalid
var ErrTokenInvalid = errors.New("token invalido")

// Claims jwt payload
type Claims struct {
	UserID               string `json:"user_id"`
	Email                string `json:"email"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims        // campos estandar
}

// JWTConfig continene configuracion del manager
type JWTConfig struct {
	SecretKey            []byte
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// JWTManager gestiona la creacion y verificacion de tokens
type JWTManager struct {
	config JWTConfig
}

// NewJWTManager returns new jwt manager
func NewJWTManager(cfg JWTConfig) (*JWTManager, error) {
	// OWASP -> 32B HS256
	if len(cfg.SecretKey) < 32 {
		return nil, fmt.Errorf("secret key debe tener al menos 32 bytes, tiene %d", len(cfg.SecretKey))
	}
	return &JWTManager{config: cfg}, nil
}
