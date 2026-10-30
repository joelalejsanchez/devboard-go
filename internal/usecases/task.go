// Package usecases - caso de uso para task
package usecases

import (
	"context"
	"devboard/internal/domain"
	"fmt"
)

// TaskCommandRepository contrato de escritura CQRS
type TaskCommandRepository interface {
	Create(ctx *context.Context, task *domain.Task) error
	Update(ctx *context.Context, task *domain.Task) error
}

// TaskQueryRepository contrato de lectura CQRS
type TaskQueryRepository interface {
	GetByID(ctx *context.Context, id string) (*domain.Task, error)
	ListByProject(ctx *context.Context, projectID string) ([]*domain.Task, error)
}

// TaskUseCase struct
type TaskUseCase struct {
	commands TaskCommandRepository
	queries  TaskQueryRepository
}

// NewTaskUseCase constructor de TaskUseCase
func NewTaskUseCase(commands TaskCommandRepository, queries TaskQueryRepository) *TaskUseCase {
	return &TaskUseCase{
		commands: commands,
		queries:  queries,
	}
}

// CreateTask crea una tarea
func (uc *TaskUseCase) CreateTask(ctx *context.Context, projectID, title, createdBy string) (*domain.Task, error) {
	task := domain.NewTask(projectID, title, createdBy)

	if err := uc.commands.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("creando tarea: %w", err)
	}

	return task, nil
}

// ListProjectTasks Obtiene todas las tareas
func (uc *TaskUseCase) ListProjectTasks(ctx *context.Context, projectID string) ([]*domain.Task, error) {
	return uc.queries.ListByProject(ctx, projectID)

}

// AssignTask asigna una tarea a un usuario
func (uc *TaskUseCase) AssignTask(ctx *context.Context, taskID, userID string) (*domain.Task, error) {
	task, err := uc.queries.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo tarea a asignar: %w", err)
	}

	if err := task.AssignTo(userID); err != nil {
		return nil, fmt.Errorf("asignando tarea: %w", err)
	}

	if err := uc.commands.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("actualizando tarea: %w", err)

	}

	return task, nil
}

// GetTask obtiene tarea por id
func (uc *TaskUseCase) GetTask(ctx *context.Context, id string) (*domain.Task, error) {
	task, err := uc.queries.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo tarea: %w", err)
	}

	return task, nil
}
