// Package auth contains auth and crypto functions
package auth

import (
	"crypto/rand"
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

// generateToken returns new token with configuration
func (m *JWTManager) generateToken(userID, email, role string, duration time.Duration) (string, error) {
	now := time.Now()
	Claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateTokenID(),
			Issuer:    "devboard",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	signed, err := token.SignedString(m.config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("firmando token: %w", err)
	}

	return signed, nil
}

// generateTokenID return a new random ID
func generateTokenID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
