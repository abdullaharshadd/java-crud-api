package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/service"
)

type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New()}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", h.SaveUser)
	r.Get("/get_user_data", h.FetchUserList)
	r.Get("/get_user_data/{id}", h.FetchUserByID)
	r.Delete("/delete_user_data/{id}", h.DeleteUser)
	r.Put("/update_user_data/{id}", h.UpdateUser)
	r.Get("/get_user_name/name/{name}", h.GetUserNameByName)
}

func (h *UserHandler) SaveUser(w http.ResponseWriter, req *http.Request) {
	slog.Info("inside the SaveUser of UserHandler")

	var user model.User
	if err := decodeJSON(req, &user); err != nil {
		writeMalformedBody(w, req, err)
		return
	}

	if err := user.Validate(h.validate); err != nil {
		writeValidationError(w, err)
		return
	}

	if _, err := h.svc.Save(req.Context(), user); err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "User data saved successfully!")
}

func (h *UserHandler) FetchUserList(w http.ResponseWriter, req *http.Request) {
	slog.Info("inside the FetchUserList of UserHandler")

	users, err := h.svc.List(req.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) FetchUserByID(w http.ResponseWriter, req *http.Request) {
	id, err := parseIDParam(req, "id")
	if err != nil {
		writeValidationError(w, err)
		return
	}

	user, err := h.svc.GetByID(req.Context(), id)
	if err != nil {
		h.writeUserLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, req *http.Request) {
	id, err := parseIDParam(req, "id")
	if err != nil {
		writeValidationError(w, err)
		return
	}

	if err := h.svc.Delete(req.Context(), id); err != nil {
		h.writeUserLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "user data deleted Successfully")
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, req *http.Request) {
	id, err := parseIDParam(req, "id")
	if err != nil {
		writeValidationError(w, err)
		return
	}

	var user model.User
	if err := decodeJSON(req, &user); err != nil {
		writeMalformedBody(w, req, err)
		return
	}

	if err := h.svc.Update(req.Context(), id, user); err != nil {
		h.writeUserLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &user)
}

func (h *UserHandler) GetUserNameByName(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")

	user, err := h.svc.GetByName(req.Context(), name)
	if err != nil {
		h.writeUserLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) writeUserLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, smartcontacterror.ErrUserNotFound) {
		writeJSON(w, http.StatusNotFound, model.NewErrorMessage(http.StatusNotFound, err.Error()))
		return
	}
	writeInternalError(w, err)
}

func parseIDParam(req *http.Request, key string) (int, error) {
	raw := chi.URLParam(req, key)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func decodeJSON(req *http.Request, dst any) error {
	dec := json.NewDecoder(req.Body)
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", slog.Any("error", err))
	}
}

func writeValidationError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, model.NewErrorMessage(http.StatusBadRequest, err.Error()))
}

func writeInternalError(w http.ResponseWriter, err error) {
	slog.Error("internal server error", slog.Any("error", err))
	writeJSON(w, http.StatusInternalServerError, model.NewErrorMessage(http.StatusInternalServerError, "internal server error"))
}

func writeMalformedBody(w http.ResponseWriter, req *http.Request, err error) {
	body := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    http.StatusBadRequest,
		"error":     http.StatusText(http.StatusBadRequest),
		"message":   err.Error(),
		"path":      req.URL.Path,
	}
	writeJSON(w, http.StatusBadRequest, body)
}

var _ = context.Background