package handlers

/*
 * Pruebas propias
 */
import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type User struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
}
type GetUserHandler struct{}

func (h *GetUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := User{"joel", "sanchez"}
	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type PostUserHandler struct {
	Logger *slog.Logger
}

func (h *PostUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		h.Logger.Error("error al decodificar body", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func NewPostUserHandler() *PostUserHandler {
	return &PostUserHandler{}
}
func NewGetUserHandler() *GetUserHandler {
	return &GetUserHandler{}
}
