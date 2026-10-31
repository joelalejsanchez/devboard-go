// Package domain defines the core domain types and business rules.
package domain

import (
	"encoding/json"
	"fmt"
)

// TaskStatus represents the current state of a task.
type TaskStatus string

const (
	// TaskStatusTodo is the initial task state.
	TaskStatusTodo TaskStatus = "todo"
	// TaskStatusInProgress indicates that work on the task is underway.
	TaskStatusInProgress TaskStatus = "in_progress"
	// TaskStatusDone indicates that the task is complete.
	TaskStatusDone TaskStatus = "done"
)

// IsValid reports whether status is one of the supported task states.
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

// UnmarshalJSON validates the status while decoding its JSON representation.
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
