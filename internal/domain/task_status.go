// Package domain - define task_status
package domain

import (
	"encoding/json"
	"fmt"
)

// TaskStatus type
type TaskStatus string

const (
	// TaskStatusTodo estado pendiente
	TaskStatusTodo TaskStatus = "todo"
	// TaskStatusInProgress estado en progreso
	TaskStatusInProgress TaskStatus = "in_progress"
	// TaskStatusDone estado hecho
	TaskStatusDone TaskStatus = "done"
)

// IsValid metodo de validacion de valor
func (status TaskStatus) IsValid() bool {
	switch status {
	case
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusDone:

		return true

	default:
		return false
	}
}

// Este metodo se usa para validaciones anidadas o custom al hacer un UnmarshalJSON
// Valida Status al desserializar

// UnmarshalJSON valida al desserializar el valor
func (status *TaskStatus) UnmarshalJSON(data []byte) error {
	var raw string
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	taskStatus := TaskStatus(raw)
	if !taskStatus.IsValid() {
		return fmt.Errorf("estado invalido: %q", raw)
	}
	*status = taskStatus
	return nil
}
