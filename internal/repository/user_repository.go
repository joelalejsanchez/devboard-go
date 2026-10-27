// Package repository - repositorio de usuarios
package repository

import (
	"context"
	"devboard/internal/domain"
	"fmt"
	"sync"
	"uuid"
)

// UserRepository adapter que implementa UserRepository , almacenamiento en memoria
type UserRepository struct {
	mu           sync.RWMutex
	users        map[string]*domain.User // [id string] -> *User
	usersByEmail map[string]*domain.User // [email string] -> *User
}

// Create crea un usuario en memoria
func (r *UserRepository) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.usersByEmail[user.Email]
	if ok {
		return fmt.Errorf("usuario con correo %s: %w", user.Email, domain.ErrAlreadyExists)
	}

	user.ID = uuid.NewV7().String()
	if !user.IsValid() {
		return fmt.Errorf("usuario invalido: %v", *user)
	}

	r.users[user.ID] = user
	r.usersByEmail[user.Email] = user

	return nil
}

// GetByID obtiene un usuario por id, desde memoria
func (r *UserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("usuario %s: %w", id, domain.ErrNotFound)
	}

	return user, nil
}

// GetByEmail obtiene un usuario por email, desde memoria
func (r *UserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[email]
	if !ok {
		return nil, fmt.Errorf("usuario %s: %w", email, domain.ErrNotFound)
	}

	return user, nil
}

// NewUserRepository crea un nuveo repositorio en memoria
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:        make(map[string]*domain.User),
		usersByEmail: make(map[string]*domain.User),
	}
}
