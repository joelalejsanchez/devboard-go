// Package usecases coordinates application workflows.
package usecases

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/domain"
	"errors"
	"fmt"
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

// NewAuthUseCase constructor
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

// Register registers user
func (uc *AuthUseCase) Register(ctx context.Context, email, name, password string) (*AuthTokens, error) {
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := domain.NewUser(email, name, passwordHash)
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creando usuario: %w", err)
	}

	tokens, err := uc.generateTokenPair(ctx, user.ID, user.Email, "member")
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// generateTokenPair returns access and refresh token
func (uc *AuthUseCase) generateTokenPair(ctx context.Context, id, email, role string) (*AuthTokens, error) {
	accessToken, err := uc.jwtManager.GenerateAccessToken(id, email, role)
	if err != nil {
		return nil, fmt.Errorf("generando access token: %w", err)
	}

	refreshToken, err := uc.jwtManager.GenerateRefreshToken(id, email, role)
	if err != nil {
		return nil, fmt.Errorf("generando refresh token: %w", err)
	}

	refreshClaims, err := uc.jwtManager.VerifyToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("extrayendo claims refresh token: %w", err)
	}

	if err := uc.tokenStore.StoreRefreshToken(
		ctx,
		refreshClaims.ID,
		id,
		uc.refreshTokenDuration,
	); err != nil {
		return nil, fmt.Errorf("guardando refresh token: %w", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * time.Minute,
	}, nil
}

// Login verifies credentials and returns tokens
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*AuthTokens, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("credenciales invalidas: %w", err)
		}
		return nil, fmt.Errorf("buscando usuario: %w", err)
	}

	if err := auth.VerifyPassword(password, user.PasswordHash); err != nil {
		if errors.Is(err, auth.ErrPasswordMissmatch) {
			return nil, fmt.Errorf("credenciales invalidas: %w", domain.ErrUnauthorized)
		}
		return nil, fmt.Errorf("verificando password: %w", domain.ErrUnauthorized)
	}

	tokens, err := uc.generateTokenPair(ctx, user.ID, user.Email, "member")
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// RefreshTokens verifies and returns new pair of tokens
func (uc *AuthUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	claims, err := uc.jwtManager.VerifyToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token invalido: %w", domain.ErrUnauthorized)
	}

	_, err = uc.tokenStore.ValidateRefreshToken(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, auth.ErrTokenRevoked) || errors.Is(err, auth.ErrTokenNotFound) {
			return nil, fmt.Errorf("refresh token invalido: %w", domain.ErrUnauthorized)
		}
		return nil, fmt.Errorf("validando refresh token: %w", err)
	}

	if err := uc.tokenStore.RevokeRefreshToken(ctx, claims.ID, uc.refreshTokenDuration); err != nil {
		return nil, fmt.Errorf("revocando token anterior: %w", err)
	}

	tokens, err := uc.generateTokenPair(ctx, claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Logout revoke refresh tokens
func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {

	claims, err := uc.jwtManager.VerifyToken(refreshToken)
	if err != nil {
		return fmt.Errorf("refresh token invalido: %w", domain.ErrUnauthorized)
	}

	if err := uc.tokenStore.RevokeRefreshToken(ctx, claims.ID, uc.refreshTokenDuration); err != nil {
		return fmt.Errorf("revocando token: %w", err)
	}

	return nil
}
