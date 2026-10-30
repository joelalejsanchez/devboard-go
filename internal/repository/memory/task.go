// Package memory - repositorio de task
package memory

import (
	"context"
	"devboard/internal/domain"
	"fmt"
	"sync"
	"uuid"
)

// TaskRepository struct
type TaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]*domain.Task
}

// NewTaskRepository constructor
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

// Create crea tarea
func (r *TaskRepository) Create(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = uuid.NewV7().String()
	r.tasks[task.ID] = task

	return nil
}

// Update actualiza tarea
func (r *TaskRepository) Update(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID] = task
	return nil
}

// GetByID obtiene tarea
func (r *TaskRepository) GetByID(_ context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("tarea %s: %w", id, domain.ErrNotFound)
	}

	return task, nil
}

// ListByProject obtiene tareas de proyecto
func (r *TaskRepository) ListByProject(_ context.Context, projectID string) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, task := range r.tasks {
		if task.ProjectID == projectID {
			result = append(result, task)
		}
	}
	return result, nil
}
