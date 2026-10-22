package domain

import (
	"encoding/json"
	"testing"
)

type CreateTaskRequest struct {
	Title  string     `json:"title"`
	Status TaskStatus `json:"status"`
}

func TestCreateTaskRequestWithValidStatus(t *testing.T) {
	body := []byte(`{"title":"tarea","status":"todo"}`)
	var req CreateTaskRequest

	err := json.Unmarshal(body, &req)
	if err != nil {
		t.Fatalf("error en valor: %v", err)
	}

	if req.Status != TaskStatusTodo {
		t.Fatalf("se esperaba \"todo\", se recibio \"%s\"", req.Status)
	}
}
func TestCreateTaskRequestWithInvalidStatus(t *testing.T) {
	body := []byte(`{"title":"tarea","status":"hecho"}`)
	var req CreateTaskRequest

	err := json.Unmarshal(body, &req)
	if err == nil {
		t.Fatalf("esperaba error, recibio: %v", err)
	}
	expected := "estado invalido: \"hecho\""
	if err.Error() != expected {
		t.Fatalf(`esperaba %q, recibio: %q`, expected, err.Error())
	}
}
