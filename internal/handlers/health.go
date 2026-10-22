// Package handlers - endpoints de health
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler struct
type HealthHandler struct{}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ServeHTTP function
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(time.Millisecond * 200)
	res := healthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // opcional

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

// NewHealthHandler retorna HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}
