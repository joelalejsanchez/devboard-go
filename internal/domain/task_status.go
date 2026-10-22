package domain

import (
	"encoding/json"
	"fmt"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

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
func (status *TaskStatus) UnmarshalJSON(data []byte) error {
	var raw string
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	taskStatus := TaskStatus(raw)
	status.IsValid()
	if !taskStatus.IsValid() {
		return fmt.Errorf("estado invalido: %q", raw)
	}
	*status = taskStatus
	return nil
}
