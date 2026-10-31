// Package usecases coordinates application workflows.
package usecases

import (
	"context"
	"devboard/internal/domain"
	"fmt"
)

// TaskCommandRepository defines write operations for tasks.
type TaskCommandRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	Update(ctx context.Context, task *domain.Task) error
}

// TaskQueryRepository defines read operations for tasks.
type TaskQueryRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	ListByProject(ctx context.Context, projectID string) ([]*domain.Task, error)
}

// TaskUseCase coordinates task-related business workflows.
type TaskUseCase struct {
	commands TaskCommandRepository
	queries  TaskQueryRepository
}

// NewTaskUseCase creates a task use case with command and query repositories.
func NewTaskUseCase(commands TaskCommandRepository, queries TaskQueryRepository) *TaskUseCase {
	return &TaskUseCase{
		commands: commands,
		queries:  queries,
	}
}

// CreateTask creates a task and stores it through the command repository.
func (uc *TaskUseCase) CreateTask(ctx context.Context, projectID, title, description, createdBy string) (*domain.Task, error) {
	task := domain.NewTask(projectID, title, description, createdBy)

	if err := uc.commands.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("creando tarea: %w", err)
	}

	return task, nil
}

// ListProjectTasks returns all tasks for a project.
func (uc *TaskUseCase) ListProjectTasks(ctx context.Context, projectID string) ([]*domain.Task, error) {
	tasks, err := uc.queries.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo tareas de proyecto: %w", err)
	}
	return tasks, nil
}

// AssignTask assigns a task to a user.
func (uc *TaskUseCase) AssignTask(ctx context.Context, taskID, userID string) (*domain.Task, error) {
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

// UpdateStatus changes a task's status.
func (uc *TaskUseCase) UpdateStatus(ctx context.Context, taskID string, status domain.TaskStatus) (*domain.Task, error) {
	task, err := uc.queries.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo tarea a asignar: %w", err)
	}

	if err := task.MoveToStatus(status); err != nil {
		return nil, fmt.Errorf("actualizando estado de tarea: %w", err)
	}

	if err := uc.commands.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("actualizando tarea: %w", err)

	}

	return task, nil
}

// GetTask returns a task by ID.
func (uc *TaskUseCase) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	task, err := uc.queries.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo tarea: %w", err)
	}

	return task, nil
}
