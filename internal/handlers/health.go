// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler reports the health of the service.
type HealthHandler struct{}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
} // @name HealthResponse

// ServeHTTP writes the health status as a JSON response.
// @Summary Health check
// @Description Reports whether the server is running and responding.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(time.Millisecond * 200)
	res := healthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Optional because encoding/json writes a status if none is set.

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

// NewHealthHandler creates a health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}
