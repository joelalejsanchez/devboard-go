// Package memory provides in-memory repository implementations.
package memory

import (
	"context"
	"devboard/internal/domain"
	"fmt"
	"sync"
	"uuid"
)

// TaskRepository stores tasks in memory.
type TaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]*domain.Task
}

// NewTaskRepository creates an empty in-memory task repository.
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

// Create stores a task in the repository.
func (r *TaskRepository) Create(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = uuid.NewV7().String()
	r.tasks[task.ID] = task

	return nil
}

// Update replaces a task in the repository.
func (r *TaskRepository) Update(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID] = task
	return nil
}

// GetByID returns a task by ID.
func (r *TaskRepository) GetByID(_ context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("tarea %s: %w", id, domain.ErrNotFound)
	}

	return task, nil
}

// ListByProject returns all tasks belonging to a project.
func (r *TaskRepository) ListByProject(_ context.Context, projectID string) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, task := range r.tasks {
		if task.ProjectID == projectID {
			result = append(result, task)
		}
	}
	return result, nil
}
