// Package usecases - caso de uso para task
package usecases

import (
	"context"
	"devboard/internal/domain"
	"fmt"
)

// TaskRepository contrato
type TaskRepository interface {
	Create(ctx *context.Context, task *domain.Task) error
	GetById(ctx *context.Context, id string) (*domain.Task, error)
}

// TaskUseCase struct
type TaskUseCase struct {
	repo TaskRepository
}

// CreateTask crea una tarea
func (uc *TaskUseCase) CreateTask(ctx *context.Context, task *domain.Task) (*domain.Task, error) {
	err := uc.repo.Create(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("creando tarea: %w", err)
	}
	return task, nil
}

// GetTask obtiene tarea
func (uc *TaskUseCase) GetTask(ctx *context.Context, id string) (*domain.Task, error) {
	task, err := uc.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo tarea: %w", err)
	}

	return task, nil
}
