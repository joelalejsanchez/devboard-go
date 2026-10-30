// Package handlers - handlers de task
package handlers

import (
	"devboard/internal/handlers/utils"
	"devboard/internal/usecases"
	"devboard/internal/validator"
	"encoding/json"
	"log/slog"
	"net/http"
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
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	CreatedBy string `json:"created_by"`
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

	if err := utils.RespondJSON(w, http.StatusCreated, taskResponse{
		ID:        task.ID,
		ProjectID: task.ProjectID,
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
	}); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

// GetByID obtiene un usuario por id
// @Summary Get Task
// @Description Obtiene un task por id
// @Tags tasks
// @Produce json
// @Param id path string true "ID de la tarea"
// @Success 200 {object} taskResponse
// @Failure 400 {object} utils.ProblemDetails
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := h.usecase.GetTask(r.Context(), id)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	if err := utils.RespondJSON(w, http.StatusOK, taskResponse{
		ID:        task.ID,
		ProjectID: task.ProjectID,
		Title:     task.Title,
		CreatedBy: task.CreatedBy,
	}); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}
