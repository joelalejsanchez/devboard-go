package handlers

/*
 * Pruebas propias
 */
import (
	"devboard/internal/domain"
	"encoding/json"
	"log/slog"
	"net/http"
)

// GetUserHandler struct
type GetUserHandler struct{}

// ServeHTTP function
func (h *GetUserHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	user := domain.User{
		Name:     "joel",
		Lastname: "sanchez",
	}
	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		return
	}
}

// PostUserHandler struct
type PostUserHandler struct {
	Logger *slog.Logger
}

// ServeHTTP function
func (h *PostUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, "json invalido", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("ok"))
	if err != nil {
		return
	}
}

// NewPostUserHandler retorna PostUserHandler
func NewPostUserHandler() *PostUserHandler {
	return &PostUserHandler{}
}

// NewGetUserHandler retorna GetUserHandler
func NewGetUserHandler() *GetUserHandler {
	return &GetUserHandler{}
}
