// Package usecases coordinates application workflows.
package usecases

import (
	"context"
	"devboard/internal/domain"
	"errors"
	"fmt"
)

// UserRepository defines user persistence operations required by the use case.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

// UserUseCase coordinates user-related business workflows.
type UserUseCase struct {
	repo UserRepository
}

// CreateUser creates a user and stores it through the repository.
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

// GetUser returns a user by ID.
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo usuario %s: %w", id, err)
	}
	return user, nil
}

// NewUserUseCase creates a user use case with its repository.
func NewUserUseCase(repo UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}
