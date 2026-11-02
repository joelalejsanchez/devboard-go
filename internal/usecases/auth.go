// Package usecases coordinates application workflows.
package usecases

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/domain"
	"time"
)

// authRepository define auth operations
type authRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// AuthTokens struct
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
}

// AuthUseCase struct
type AuthUseCase struct {
	repo                 authRepository
	jwtManager           *auth.JWTManager
	tokenStore           *auth.TokenStore
	refreshTokenDuration time.Duration
}

func NewAuthUseCase(
	repo authRepository,
	jwtManager *auth.JWTManager,
	tokenStore *auth.TokenStore,
	refreshTokenDuration time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		repo:                 repo,
		jwtManager:           jwtManager,
		tokenStore:           tokenStore,
		refreshTokenDuration: refreshTokenDuration,
	}
}
