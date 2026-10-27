package repository

import (
	"context"
	"devboard/internal/domain"
	"errors"
	"testing"
	"uuid"
)

func TestUserRepository_Create(t *testing.T) {
	var err error
	repo := NewUserRepository()
	ctx := context.Background()

	user := &domain.User{
		Name:  "Ana",
		Email: "ana@devboard.dev",
	}

	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create() devolvio un error inesperado: %v", err)
	}

	if user.ID == "" {
		t.Fatalf("Create() no asigno un id al usuario")
	}

	_, err = uuid.Parse(user.ID)
	if err != nil {
		t.Fatalf("Create() no asigno un uuid correcto. id: %s", user.ID)
	}
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	var err error
	repo := NewUserRepository()
	ctx := context.Background()

	_, err = repo.GetByID(ctx, "1")
	if err == nil {
		t.Fatal("GetByID() debia devolver un error")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID() devolvio %v: esperaba ErrNotFound", err)
	}
}
