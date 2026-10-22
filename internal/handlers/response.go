// Package handlers - utilidades para errores
package handlers

import (
	"devboard/internal/domain"
	"devboard/internal/validator"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// errorResponse struct
type errorResponse struct {
	Error   string                      `json:"error"`
	Details []validator.ValidationError `json:"details,omitempty"`
}

// ProblemDetails struct
type ProblemDetails struct {
	Type     string                      `json:"type"`
	Title    string                      `json:"title"`
	Status   int                         `json:"status"`
	Detail   string                      `json:"detail"`
	Instance string                      `json:"instance"`
	Errors   []validator.ValidationError `json:"errors,omitempty"`
}

const problemBaseURL = "https://github.com/joelalejsanchez"

// RespondError centralizado
func RespondError(w http.ResponseWriter, r *http.Request, l *slog.Logger, err error) {
	var problem ProblemDetails
	problem.Instance = r.URL.Path

	switch {
	case errors.Is(err, domain.ErrNotFound):
		problem.Type = problemBaseURL + "/not-found"
		problem.Title = "Recurso no encontrado"
		problem.Status = http.StatusNotFound
		problem.Detail = err.Error()
	case errors.Is(err, domain.ErrAlreadyExists):
		problem.Type = problemBaseURL + "/conflict"
		problem.Title = "El recurso ya existe"
		problem.Status = http.StatusConflict
		problem.Detail = err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		problem.Type = problemBaseURL + "/unauthorized"
		problem.Title = "No autorizado"
		problem.Status = http.StatusUnauthorized
		problem.Detail = err.Error()
	case errors.Is(err, domain.ErrForbidden):
		problem.Type = problemBaseURL + "/forbidden"
		problem.Title = "Acceso prohibido"
		problem.Status = http.StatusForbidden
		problem.Detail = err.Error()
	case errors.Is(err, domain.ErrInternal):
		problem.Type = problemBaseURL + "/internal-server-error"
		problem.Title = "Error en el servidor"
		problem.Status = http.StatusInternalServerError
		problem.Detail = err.Error()
	case errors.Is(err, domain.ErrInvalidInput):
		problem.Type = problemBaseURL + "/bad-request"
		problem.Title = "Request invalida"
		problem.Status = http.StatusBadRequest
		problem.Detail = err.Error()
	default:
		problem.Type = problemBaseURL + "/internal-server-error"
		problem.Title = "Error en el servidor"
		problem.Status = http.StatusInternalServerError
		problem.Detail = "error interno del servidor"

		l.ErrorContext(r.Context(), "error no clasificado",
			slog.Any("err", err),
			slog.String("path", r.URL.Path),
		)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	jsonErr := json.NewEncoder(w).Encode(problem)
	if jsonErr != nil {
		http.Error(w, domain.ErrInternal.Error(), http.StatusInternalServerError)
	}
}

// RespondValidationError responde errores de validacion
func RespondValidationError(w http.ResponseWriter, errs []validator.ValidationError) {
	respondJSON(w, http.StatusBadRequest, errorResponse{
		Error:   "datos de entrada invalidos",
		Details: errs,
	})
}

// ResponseJSON envia una respuesta exitosa en JSON
func ResponseJSON(w http.ResponseWriter, status int, data any) {
	respondJSON(w, status, data)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, domain.ErrInternal.Error(), http.StatusInternalServerError)
	}
}
