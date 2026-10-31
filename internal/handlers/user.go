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
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	usecase   *usecases.UserUseCase
	validator *validator.Validator
	logger    *slog.Logger
}

// NewUserHandler creates a user handler with its dependencies.
func NewUserHandler(uc *usecases.UserUseCase, v *validator.Validator, l *slog.Logger) *UserHandler {
	return &UserHandler{
		usecase:   uc,
		validator: v,
		logger:    l,
	}
}

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
} // @name CreateUserRequest

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
} // @name UserResponse

// Create creates a user.
// @Summary Create User
// @Description Creates a user.
// @Tags users
// @Accepts json
// @Produce json
// @Param request body createUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} utils.ProblemDetails
// @Router /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, r, h.logger, domain.ErrInvalidInput)
		return
	}
	if errs := h.validator.Validate(req); errs != nil {
		if err := utils.RespondValidationError(w, errs); err != nil {
			h.logger.ErrorContext(r.Context(), "error al escribir respuesta de error de validacion")
		}
		return
	}

	// TODO ImplementarHasshing
	passwordHash := "TEMP_" + req.Password
	user, err := h.usecase.CreateUser(r.Context(), req.Email, req.Name, passwordHash)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	if err := utils.RespondJSON(w, http.StatusCreated, userResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}

// Get returns a user by ID.
//
// @Summary Get User
// @Description Returns a user by ID.
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} utils.ProblemDetails
// @Router /users/{id} [get]
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, err := h.usecase.GetUser(r.Context(), id)
	if err != nil {
		utils.RespondError(w, r, h.logger, err)
		return
	}

	if err := utils.RespondJSON(w, http.StatusOK, userResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}); err != nil {
		h.logger.ErrorContext(r.Context(), "error al enviar respuesta")
	}
}
