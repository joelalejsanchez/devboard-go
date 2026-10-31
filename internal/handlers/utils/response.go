// Package utils provides HTTP handler utilities.
package utils

import (
	"devboard/internal/domain"
	"devboard/internal/validator"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// errorResponse represents a validation error response.
type errorResponse struct {
	Error   string                      `json:"error"`
	Details []validator.ValidationError `json:"details,omitempty"`
} // @name ErrorResponse

// ProblemDetails represents an HTTP problem response.
type ProblemDetails struct {
	Type     string                      `json:"type"`
	Title    string                      `json:"title"`
	Status   int                         `json:"status"`
	Detail   string                      `json:"detail"`
	Instance string                      `json:"instance"`
	Errors   []validator.ValidationError `json:"errors,omitempty"`
}

const problemBaseURL = "https://github.com/joelalejsanchez"

// RespondError writes a problem response for err and logs unexpected errors.
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
		l.Error("error al escribir problem details",
			slog.String("error", jsonErr.Error()),
		)
	}
}

// RespondValidationError writes a bad request response for validation errors.
func RespondValidationError(w http.ResponseWriter, errs []validator.ValidationError) error {
	return respondJSON(w, http.StatusBadRequest, errorResponse{
		Error:   "datos de entrada invalidos",
		Details: errs,
	})
}

// RespondJSON writes data as a JSON response with the supplied status code.
func RespondJSON(w http.ResponseWriter, status int, data any) error {
	return respondJSON(w, status, data)
}

func respondJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
