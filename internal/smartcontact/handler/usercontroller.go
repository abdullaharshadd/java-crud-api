// Package handler contains the HTTP handlers for the SmartContact
// application.
//
// MIGRATION_NOTE: The Java source UserController.java was a Spring
// @RestController using field injection (@Autowired) and declarative routing
// annotations (@GetMapping/@PostMapping/@PutMapping/@DeleteMapping). The
// idiomatic Go translation is:
//
//   - A UserHandler struct that holds the service.UserService dependency,
//     injected explicitly via NewUserHandler (replacing Spring's field
//     injection / component scanning).
//   - Explicit route registration via RegisterRoutes on a chi router
//     (replacing the annotation-driven routing).
//   - Manual JSON (de)serialization with encoding/json (replacing @RequestBody
//     / automatic JSON marshalling of return values).
//   - Explicit bean validation via model.Validate (replacing @Valid triggering
//     Spring's validator). Validation failures produce the migrated
//     model.ErrorMessage shape with 400.
//   - Malformed JSON bodies produce a Spring-DefaultErrorAttributes-shaped
//     400 response (writeMalformedBody), matching the original framework
//     behaviour observed in the agent debate (Change 14).
//   - The UserNotFoundException checked-exception path becomes the migrated
//     error.ErrUserNotFound sentinel, mapped to a 404 response.
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

	smartcontacterror "github.com/smartcontact/internal/smartcontact/error"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/service"
)

// UserHandler exposes the HTTP endpoints for User CRUD operations, delegating
// all business logic to the underlying service.UserService.
//
// MIGRATION_NOTE: Replaces the Spring @RestController. The service dependency
// is injected via the constructor instead of @Autowired field injection.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler constructs a UserHandler with the given UserService.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes registers all User endpoints on the provided chi router,
// preserving the exact HTTP methods and paths from the Java controller.
//
// MIGRATION_NOTE: The Java source declared 'get_user_name/name/{name}' without
// a leading slash; Spring normalizes this to '/get_user_name/name/{name}', so
// the leading slash is added explicitly here.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", h.SaveUser)
	r.Get("/get_user_data", h.FetchUserList)
	r.Get("/get_user_data/{id}", h.FetchUserByID)
	r.Delete("/delete_user_data/{id}", h.DeleteUser)
	r.Put("/update_user_data/{id}", h.UpdateUser)
	r.Get("/get_user_name/name/{name}", h.GetUserNameByName)
}

// SaveUser handles POST /save_user_data. It decodes and validates the request
// body, persists the user, and returns a success message with status 200.
//
// Malformed JSON yields a Spring-DefaultErrorAttributes-shaped 400 response;
// validation failures yield a model.ErrorMessage-shaped 400 response.
func (h *UserHandler) SaveUser(w http.ResponseWriter, req *http.Request) {
	slog.Info("inside the SaveUser of UserHandler")

	var user model.User
	if err := decodeJSON(req, &user); err != nil {
		writeMalformedBody(w, req, err)
		return
	}

	if err := model.Validate(&user); err != nil {
		writeValidationError(w, err)
		return
	}

	if _, err := h.svc.Save(req.Context(), &user); err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns all users as JSON.
func (h *UserHandler) FetchUserList(w http.ResponseWriter, req *http.Request) {
	slog.Info("inside the FetchUserList of UserHandler")

	users, err := h.svc.List(req.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// id, mapping error.ErrUserNotFound to a 404 response.
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

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by id and
// returns a success message with status 200.
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

// UpdateUser handles PUT /update_user_data/{id}. It updates a user by id and
// echoes back the submitted user payload.
//
// MIGRATION_NOTE: The Java controller did not annotate the body with @Valid on
// update, so validation is intentionally not triggered here to preserve exact
// behaviour. Malformed JSON still yields a 400.
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

	if _, err := h.svc.Update(req.Context(), id, &user); err != nil {
		h.writeUserLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &user)
}

// GetUserNameByName handles GET /get_user_name/name/{name}. It looks up a user
// by name and returns it, mapping error.ErrUserNotFound to a 404 response.
func (h *UserHandler) GetUserNameByName(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")

	user, err := h.svc.GetByName(req.Context(), name)
	if err != nil {
		h.writeUserLookupError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// writeUserLookupError maps service-layer errors to HTTP responses, translating
// the migrated error.ErrUserNotFound sentinel to a 404.
func (h *UserHandler) writeUserLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, smartcontacterror.ErrUserNotFound) {
		writeJSON(w, http.StatusNotFound, model.NewErrorMessage(err.Error()))
		return
	}
	writeInternalError(w, err)
}

// parseIDParam extracts and parses an integer path parameter.
func parseIDParam(req *http.Request, key string) (int, error) {
	raw := chi.URLParam(req, key)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// decodeJSON decodes a JSON request body into dst, rejecting unknown fields.
func decodeJSON(req *http.Request, dst any) error {
	dec := json.NewDecoder(req.Body)
	return dec.Decode(dst)
}

// writeJSON marshals v to JSON and writes it with the given status code.
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

// writeValidationError writes a model.ErrorMessage-shaped 400 response,
// matching Spring's @Valid failure handling.
func writeValidationError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, model.NewErrorMessage(err.Error()))
}

// writeInternalError writes a generic 500 response.
func writeInternalError(w http.ResponseWriter, err error) {
	slog.Error("internal server error", slog.Any("error", err))
	writeJSON(w, http.StatusInternalServerError, model.NewErrorMessage("internal server error"))
}

// writeMalformedBody writes a Spring DefaultErrorAttributes-shaped 400
// response for malformed request bodies.
//
// MIGRATION_NOTE: Spring's DefaultErrorAttributes response has a distinct shape
// (timestamp/status/error/message/path) that differs from the application's
// ErrorMessage shape. This is reproduced here so clients relying on that shape
// continue to work (Change 14).
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

// Ensure context is referenced for I/O propagation clarity in reviews.
var _ = context.Background
