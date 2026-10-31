// Package handlers - handlers de task
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

// TaskHandler struct
type TaskHandler struct {
	usecase   *usecases.TaskUseCase
	validator *validator.Validator
	logger    *slog.Logger
}

// NewTaskHandler constructor
func NewTaskHandler(uc *usecases.TaskUseCase, v *validator.Validator, l *slog.Logger) *TaskHandler {
	return &TaskHandler{
		usecase:   uc,
		validator: v,
		logger:    l,
	}
}

// createTaskRequest struct
type createTaskRequest struct {
	ProjectID string `json:"project_id" validate:"required"`
	Title     string `json:"title" validate:"required,min=3,max=200"`
}

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
}

// toTaskResponse convierte entidad de dominio a estructura de respuesta
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

// Create crear tarea
// @Summary Create Task
// @Description Crea una tarea
// @Tags tasks
// @Accepts json
// @Produce json
// @Param request body createTaskRequest true "Datos de la tarea"
// @Success 201 {object} taskResponse
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

	task, err := h.usecase.CreateTask(r.Context(), req.ProjectID, req.Title, "system")
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
	}

	if err := utils.RespondJSON(w, http.StatusCreated, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

// GetByID get task by id
// @Summary Get Task
// @Description Obtiene un task por id
// @Tags tasks
// @Produce json
// @Param id path string true "ID de la tarea"
// @Success 200 {object} taskResponse
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

// ListByProject obtiene tareas de un proyecto
// @Summary Get Tasks from Project
// @Description Obtiene las tasks de un proyecto
// @Tags tasks
// @Produce json
// @Param id path string true "ID del proyecto"
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

type taskAssignRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// Assign asigna la tarea a un usuario
// @Summary Asignar tarea a usuario
// @Description Asigna una tarea a un usuario
// @Tags tasks
// @Accepts json
// @Produce json
// @Param request body taskAssignRequest true "Datos de la tarea"
// @Param id path string true "ID de la tarea"
// @Success 200 {object}
// @Failure 404 {object} utils.ProblemDetails
// @Router /tasks/{id}/assign [put]
func (h *TaskHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req taskAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		if err := utils.RespondValidationError(w, errs); err != nil {
			h.logger.ErrorContext(r.Context(), "error al escribir respuesta de error de validacion")
		}
	}

	id := r.PathValue("id")
	task, err := h.usecase.AssignTask(r.Context(), id, req.UserID)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
	}

	if err := utils.RespondJSON(w, http.StatusOK, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

type taskUpdateStatusRequest struct {
	Status string `json:"status" validate:"in_progress|todo|done"`
}

// UpdateStatus actualiza el estado
// @Summary actualiza el estado de tarea
// @Description actualiza el estado de tarae
// @Tags tasks
// @Accepts json
// @Produce json
// @Param request body taskUpdateStatusRequest true "Datos de la tarea"
// @Param id path string true "ID de la tarea"
// @Success 200 {object} taskResponse
// @Failure 404 {object} utils.ProblemDetails
// @Router /tasks/{id}/status [put]
func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req taskUpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		if err := utils.RespondValidationError(w, errs); err != nil {
			h.logger.ErrorContext(r.Context(), "error al escribir respuesta de error de validacion")
		}
	}

	id := r.PathValue("id")
	task, err := h.usecase.UpdateStatus(r.Context(), id, domain.TaskStatus(req.Status))
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
	}

	if err := utils.RespondJSON(w, http.StatusOK, toTaskResponse(task)); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}
