package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	apperr "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/error/restresponseentityexceptionhandling"
	"migrated-app/internal/smartcontact/model"
)

type goerror = error

type UserService interface {
	SaveUser(ctx context.Context, user model.User) goerror
	FetchUserList(ctx context.Context) ([]model.UserResponse, goerror)
	FetchUserByID(ctx context.Context, id int) (model.UserResponse, goerror)
	DeleteUser(ctx context.Context, id int) goerror
	UpdateUser(ctx context.Context, id int, user model.User) goerror
	GetUserByName(ctx context.Context, name string) (model.UserResponse, goerror)
}

type UserController struct {
	service UserService
	logger  *slog.Logger
}

func NewUserController(service UserService, logger *slog.Logger) *UserController {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserController{service: service, logger: logger}
}

func (c *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", c.SaveUser)
	r.Get("/get_user_data", c.FetchUserList)
	r.Get("/get_user_data/{id}", c.FetchUserByID)
	r.Delete("/delete_user_data/{id}", c.DeleteUser)
	r.Put("/update_user_data/{id}", c.UpdateUser)
	r.Get("/get_user_name/name/{name}", c.GetUserByName)
}

func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the saveUser of UserController")

	var user model.User
	if err := decodeJSON(r, &user); err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := validateUser(user); err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := c.service.SaveUser(r.Context(), user); err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "User data saved successfully!")
}

func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the fetchUserList of UserController")

	users, err := c.service.FetchUserList(r.Context())
	if err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	if users == nil {
		users = []model.UserResponse{}
	}

	writeJSON(w, http.StatusOK, users)
}

func (c *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	user, err := c.service.FetchUserByID(r.Context(), id)
	if err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := c.service.DeleteUser(r.Context(), id); err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "user data deleted Successfully")
}

func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	var user model.User
	if err := decodeJSON(r, &user); err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := c.service.UpdateUser(r.Context(), id, user); err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, model.UserResponse{
		ID:    user.ID,
		Name:  &user.Name,
		Email: &user.Email,
		Role:  &user.Role,
		About: &user.About,
	})
}

func (c *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := c.service.GetUserByName(r.Context(), name)
	if err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (c *UserController) handleServiceError(w http.ResponseWriter, r *http.Request, err goerror) {
	var notFound *apperr.UserNotFoundError
	if errors.As(err, &notFound) {
		restresponseentityexceptionhandling.WriteError(w, r, err)
		return
	}

	c.logger.Error("unexpected error handling request", slog.String("error", err.Error()))
	writeJSON(w, http.StatusInternalServerError,
		model.NewErrorMessage(http.StatusInternalServerError, model.StatusText(http.StatusInternalServerError), err.Error()))
}

func pathID(r *http.Request) (int, goerror) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("invalid id path parameter: must be an integer")
	}
	return id, nil
}

func decodeJSON(r *http.Request, dst any) goerror {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return errors.New("invalid request body: " + err.Error())
	}
	return nil
}

func validateUser(user model.User) goerror {
	if user.Name == "" {
		return errors.New("validation failed: name must not be empty")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

func writeBadRequest(w http.ResponseWriter, err goerror) {
	writeJSON(w, http.StatusBadRequest,
		model.NewErrorMessage(http.StatusBadRequest, model.StatusText(http.StatusBadRequest), err.Error()))
}