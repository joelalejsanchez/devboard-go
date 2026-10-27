// Package usecase - casos de uso de usuario
package usecase

import (
	"context"
	"devboard/internal/domain"
	"errors"
	"fmt"
)

// UserRepository contrato para obtener usuarios
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetById(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

// UserUseCase orquesta las operaciones de negocio sobre usuario
type UserUseCase struct {
	repo UserRepository
}

// CreateUser orquesta la creacion de un usuario
func (uc *UserUseCase) CreateUser(ctx context.Context, email, name, passwordHash string) (*domain.User, error) {
	existing, err := uc.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("verificando email existente: %w", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("email %s: %w", email, domain.ErrAlreadyExists)
	}

	user := domain.NewUser(email, name, passwordHash)
	if !user.IsValid() {
		return nil, domain.ErrInvalidInput
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creando usuario: %w", err)
	}
	return user, nil
}

// GetUser obtiene usuario por id
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo usuario %s: %w", id, err)
	}
	return user, nil
}

// NewUserUseCase constructor
func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}
