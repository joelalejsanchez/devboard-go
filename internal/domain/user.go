// Package domain defines the core domain types and business rules.
package domain

import (
	"time"
	"uuid"
)

// User represents a user in the system.
type User struct {
	ID           string    // `json:"id"`
	Email        string    // `json:"email"`
	PasswordHash string    // `json:"-"`
	Name         string    // `json:"name"`
	CreatedAt    time.Time // `json:"created_at"`
	UpdatedAt    time.Time // `json:"updated_at"`
}

// NewUser creates a user with the supplied email, name, and password hash.
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

// IsValid reports whether the user has valid values.
func (u *User) IsValid() bool {
	return u.Email != "" && u.Name != ""
}
