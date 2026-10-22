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

// RespondError centralizado
func RespondError(w http.ResponseWriter, r *http.Request, l *slog.Logger, err error) {
	var status int
	var message string

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		message = err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = err.Error()
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		message = err.Error()
	case errors.Is(err, domain.ErrInternal):
		status = http.StatusInternalServerError
		message = err.Error()
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		message = err.Error()
	default:
		status = http.StatusInternalServerError
		message = "error interno del servidor"
		l.ErrorContext(r.Context(), "error no clasificado",
			slog.Any("err", err),
			slog.String("path", r.URL.Path),
		)
	}
	respondJSON(w, status, errorResponse{
		Error: message,
	})
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
