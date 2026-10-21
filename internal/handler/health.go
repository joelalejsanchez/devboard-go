package handler

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *HealthHandler) S() string {
	return "hola"
}
