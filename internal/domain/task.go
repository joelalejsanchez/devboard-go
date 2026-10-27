package domain

import (
	"errors"
	"time"
)

// Task representa una tarea del tablero
type Task struct {
	ID          string     // `json:"id"`
	ProjectID   string     // `json:"project_id"`
	Title       string     // `json:"title"`
	Description string     // `json:"description"`
	Status      TaskStatus // `json:"status"`
	AssigneeID  *string    // `json:"assinee_id"` // no obliga a crear desde un inicio la referencia
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsValid Valida los valores
func (t *Task) IsValid() bool {
	return t.Title != "" && t.ProjectID != ""
}

// AssignTo asigna una tarea a un usuario
func (t *Task) AssignTo(userID string) error {
	if t.Status == TaskStatusDone {
		return ErrTaskNotAssignable
	}
	t.AssigneeID = &userID
	t.UpdatedAt = time.Now()

	return nil
}

// MarkAsDone marca la tarea como completada
func (t *Task) MarkAsDone() error {
	if t.Status == TaskStatusDone {
		return ErrTaskAlreadyDone
	}
	t.Status = TaskStatusDone
	t.UpdatedAt = time.Now()
	return nil
}

// MoveToStatus cambia el estado de una tarea
func (t *Task) MoveToStatus(status TaskStatus) error {
	if !status.IsValid() {
		return ErrTaskNewStatusInvalid
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

// NewTask crea una nueva tarea
func NewTask(projectID, title, createdBy string) *Task {
	now := time.Now()
	return &Task{
		ProjectID: projectID,
		Title:     title,
		Status:    TaskStatusTodo,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

var (
	// ErrTaskNewStatusInvalid estado invalido
	ErrTaskNewStatusInvalid = errors.New("el nuevo estado es invalido")
	// ErrTaskAlreadyDone tarea completada
	ErrTaskAlreadyDone = errors.New("la tarea ya está completada")
	// ErrTaskAlreadyAssignable no se puede asignar en estado dnne
	ErrTaskNotAssignable = errors.New("la tarea no puede asignarse en este estado")
)
