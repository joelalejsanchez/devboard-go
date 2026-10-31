package domain

import (
	"errors"
	"time"
)

// Task represents a task on the board.
type Task struct {
	ID          string     // `json:"id"`
	ProjectID   string     // `json:"project_id"`
	Title       string     // `json:"title"`
	Description string     // `json:"description"`
	Status      TaskStatus // `json:"status"`
	AssigneeID  *string    // The assignee is optional when a task is created.
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsValid reports whether the task has valid values.
func (t *Task) IsValid() bool {
	return t.Title != "" && t.ProjectID != ""
}

// AssignTo assigns the task to a user.
func (t *Task) AssignTo(userID string) error {
	if t.Status == TaskStatusDone {
		return ErrTaskNotAssignable
	}
	t.AssigneeID = &userID
	t.UpdatedAt = time.Now()

	return nil
}

// MarkAsDone marks the task as complete.
func (t *Task) MarkAsDone() error {
	if t.Status == TaskStatusDone {
		return ErrTaskAlreadyDone
	}
	t.Status = TaskStatusDone
	t.UpdatedAt = time.Now()
	return nil
}

// MoveToStatus changes the task status.
func (t *Task) MoveToStatus(status TaskStatus) error {
	if !status.IsValid() {
		return ErrTaskNewStatusInvalid
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

// NewTask creates a task with the supplied project, title, description, and creator.
func NewTask(projectID, title, description, createdBy string) *Task {
	now := time.Now()
	return &Task{
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      TaskStatusTodo,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

var (
	// ErrTaskNewStatusInvalid indicates that a requested task status is invalid.
	ErrTaskNewStatusInvalid = errors.New("el nuevo estado es invalido")
	// ErrTaskAlreadyDone indicates that the task is already complete.
	ErrTaskAlreadyDone = errors.New("la tarea ya está completada")
	// ErrTaskNotAssignable indicates that the task cannot be assigned in its current state.
	ErrTaskNotAssignable = errors.New("la tarea no puede asignarse en este estado")
)
