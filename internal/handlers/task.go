// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"devboard/internal/domain"
	"devboard/internal/handlers/utils"
	"devboard/internal/usecases"
	"devboard/internal/validator"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// TaskHandler handles task-related HTTP requests.
type TaskHandler struct {
	usecase   *usecases.TaskUseCase
	validator *validator.Validator
	logger    *slog.Logger
}

// NewTaskHandler creates a task handler with its dependencies.
func NewTaskHandler(uc *usecases.TaskUseCase, v *validator.Validator, l *slog.Logger) *TaskHandler {
	return &TaskHandler{
		usecase:   uc,
		validator: v,
		logger:    l,
	}
}

// createTaskRequest contains the fields required to create a task.
type createTaskRequest struct {
	ProjectID   string `json:"project_id" validate:"required"`
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"max=2000"`
} // @name CreateTaskRequest

type taskResponse struct {
	ID          string            `json:"id"`
	ProjectID   string            `json:"project_id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      domain.TaskStatus `json:"status"`
	AssineeID   *string           `json:"assinee_id,omitempty"`
	CreatedBy   string            `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
} // @name TaskResponse

// toTaskResponse converts a domain task into its API response representation.
func toTaskResponse(task *domain.Task) taskResponse {
	return taskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		AssineeID:   task.AssigneeID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		CreatedBy:   task.CreatedBy,
	}
}

// Create creates a task.
// @Summary Create Task
// @Description Creates a task.
// @Tags tasks
// @Accepts json
// @Produce json
// @Param request body createTaskRequest true "Task data"
// @Success 201 {object} TaskResponse
// @Failure 400 {object} utils.ProblemDetails
// @Router /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		if err := utils.RespondValidationError(w, errs); err != nil {
			h.logger.ErrorContext(r.Context(), "error al escribir respuesta de error de validacion")
		}
	}

	task, err := h.usecase.CreateTask(r.Context(), req.ProjectID, req.Title, req.Description, "system")
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
	}

	if err := utils.RespondJSON(w, http.StatusCreated, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

// GetByID returns a task by its ID.
// @Summary Get Task
// @Description Returns a task by its ID.
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 404 {object} utils.ProblemDetails
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := h.usecase.GetTask(r.Context(), id)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	if err := utils.RespondJSON(w, http.StatusOK, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

// ListByProject returns the tasks for a project.
// @Summary Get Tasks from Project
// @Description Returns the tasks that belong to a project.
// @Tags tasks
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} utils.PaginatedResponse[taskResponse]
// @Failure 404 {object} utils.ProblemDetails
// @Router /projects/{id}/tasks [get]
func (h *TaskHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	tasks, err := h.usecase.ListProjectTasks(r.Context(), projectID)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	response := make([]taskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = toTaskResponse(task)
	}

	if err := utils.RespondJSON(w, http.StatusOK, utils.PaginatedResponse[taskResponse]{
		Data:    response,
		HasMore: false,
	}); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

type assignTaskRequest struct {
	AssigneeID string `json:"assignee_id" validate:"required,uuid7"`
} // @name AssignTaskRequest

// Assign assigns a task to a user.
// @Summary Assign a task to a user
// @Description Assigns a task to a user.
// @Tags tasks
// @Accepts json
// @Produce json
// @Param request body assignTaskRequest true "Assignee data"
// @Param id path string true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 404 {object} utils.ProblemDetails
// @Router /tasks/{id}/assign [put]
func (h *TaskHandler) Assign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req assignTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		if err := utils.RespondValidationError(w, errs); err != nil {
			h.logger.ErrorContext(r.Context(), "error al escribir respuesta de error de validacion")
		}
		return
	}

	task, err := h.usecase.AssignTask(r.Context(), id, req.AssigneeID)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	if err := utils.RespondJSON(w, http.StatusOK, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

type updateStatusRequest struct {
	Status domain.TaskStatus `json:"status" validate:"required"`
} // @name UpdateStatusRequest

// UpdateStatus changes a task's status.
// @Summary Update a task status
// @Description Changes the status of a task.
// @Tags tasks
// @Accepts json
// @Produce json
// @Param request body updateStatusRequest true "Task status"
// @Param id path string true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 404 {object} utils.ProblemDetails
// @Router /tasks/{id}/status [put]
func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req updateStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		if err := utils.RespondValidationError(w, errs); err != nil {
			h.logger.ErrorContext(r.Context(), "error al escribir respuesta de error de validacion")
		}
		return
	}

	task, err := h.usecase.UpdateStatus(r.Context(), id, domain.TaskStatus(req.Status))
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	if err := utils.RespondJSON(w, http.StatusOK, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}
