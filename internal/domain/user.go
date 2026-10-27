// Package domain - define User
package domain

import (
	"time"
	"uuid"
)

// User struct
type User struct {
	ID           string    // `json:"id"`
	Email        string    // `json:"email"`
	PasswordHash string    // `json:"-"`
	Name         string    // `json:"name"`
	CreatedAt    time.Time // `json:"created_at"`
	UpdatedAt    time.Time // `json:"updated_at"`
}

// NewUser Crea un nuevo usuario
func NewUser(email, name, passwordHash string) *User {
	now := time.Now()
	newUser := User{
		ID:           uuid.NewV7().String(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return &newUser
}

// IsValid verifica validez de valores de usuario
func (u *User) IsValid() bool {
	return u.Email != "" && u.Name != ""
}
