package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	apperr "migrated-app/internal/smartcontact/error/apperror"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/service"
)

// UserHandler exposes the user CRUD HTTP endpoints, delegating all business
// logic to a service.UserService. It replaces the Spring UserController.
type UserHandler struct {
	service service.UserService
}

// NewUserHandler constructs a UserHandler with the given UserService.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// RegisterRoutes registers every user endpoint on the provided mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /save_user_data", h.SaveUser)
	mux.HandleFunc("GET /get_user_data", h.FetchUserList)
	mux.HandleFunc("GET /get_user_data/{id}", h.FetchUserById)
	mux.HandleFunc("DELETE /delete_user_data/{id}", h.DeleteUser)
	mux.HandleFunc("PUT /update_user_data/{id}", h.UpdateUser)
	mux.HandleFunc("GET /get_user_name/name/{name}", h.GetUserNameByName)
}

func (h *UserHandler) SaveUser(w http.ResponseWriter, r *http.Request) {
	log.Print("inside the saveUser of UserController")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, fmt.Errorf("invalid request body"))
		return
	}

	if err := user.Validate(); err != nil {
		writeError(w, err)
		return
	}

	if _, err := h.service.SaveUser(r.Context(), &user); err != nil {
		writeError(w, err)
		return
	}

	writeString(w, http.StatusOK, "User data saved successfully!")
}

func (h *UserHandler) FetchUserList(w http.ResponseWriter, r *http.Request) {
	log.Print("inside the fetchUserList of UserController")

	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) FetchUserById(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	user, err := h.service.FetchUserById(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.service.DeleteUser(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeString(w, http.StatusOK, "user data deleted Successfully")
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, fmt.Errorf("invalid request body"))
		return
	}

	if err := h.service.UpdateUser(r.Context(), id, &user); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetUserNameByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	user, err := h.service.GetUserNameByName(r.Context(), name)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func pathID(r *http.Request) (int, error) {
	raw := r.PathValue("id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("id must be a valid integer")
	}
	return id, nil
}

func writeError(w http.ResponseWriter, err error) {
	status, body := apperr.MapError(err)
	writeJSON(w, status, body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeString(w http.ResponseWriter, status int, s string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(s)); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}