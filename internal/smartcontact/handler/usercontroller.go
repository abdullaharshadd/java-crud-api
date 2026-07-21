// Package handler contains the SmartContact service's HTTP transport layer.
// It is the Go equivalent of the Java package com.smartContact.Controller.
//
// MIGRATION_NOTE: The source UserController.java was a Spring @RestController
// that used declarative routing (@GetMapping/@PostMapping/etc.), field
// injection (@Autowired), @RequestBody/@PathVariable binding, @Valid bean
// validation, and ResponseEntity for status control. Spring's global
// exception handling (@ControllerAdvice) translated UserNotFoundException
// into HTTP responses.
//
// In idiomatic Go these responsibilities are made explicit:
//   - The UserService dependency is injected via NewUserController.
//   - Routes are registered explicitly against a chi router.
//   - JSON encoding/decoding is done with encoding/json.
//   - Path parameters are extracted with chi.URLParam.
//   - Validation is invoked explicitly (model.Validate).
//   - Error translation is delegated to the shared WriteError helper from the
//     error/restresponseentityexceptionhandling package, mirroring the Java
//     @ControllerAdvice behaviour.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	exceptionhandling "internal/smartcontact/error/restresponseentityexceptionhandling"
	"internal/smartcontact/model"
	"internal/smartcontact/service"
)

// UserController exposes CRUD HTTP endpoints for managing users, delegating
// all business logic to the injected UserService.
//
// MIGRATION_NOTE: The Java controller relied on @Autowired field injection.
// Here the dependency is provided explicitly through NewUserController.
type UserController struct {
	userService service.UserService
	logger      *slog.Logger
}

// NewUserController constructs a UserController with the given UserService.
// If logger is nil, the default slog logger is used.
func NewUserController(userService service.UserService, logger *slog.Logger) *UserController {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserController{
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registers all user endpoints on the provided chi router.
//
// MIGRATION_NOTE: Each route corresponds directly to a Spring mapping
// annotation on the original controller. Paths and HTTP methods are preserved
// exactly.
func (c *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", c.SaveUser)
	r.Get("/get_user_data", c.FetchUserList)
	r.Get("/get_user_data/{id}", c.FetchUserByID)
	r.Delete("/delete_user_data/{id}", c.DeleteUser)
	r.Put("/update_user_data/{id}", c.UpdateUser)
	r.Get("/get_user_name/name/{name}", c.GetUserByName)
}

// SaveUser handles POST /save_user_data. It decodes and validates a user from
// the JSON request body, persists it, and returns a success message.
func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the SaveUser of UserController")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		exceptionhandling.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	// MIGRATION_NOTE: Replaces Spring's @Valid bean validation with an
	// explicit call to the model's Validate method.
	if err := user.Validate(); err != nil {
		exceptionhandling.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := c.userService.SaveUser(r.Context(), &user); err != nil {
		exceptionhandling.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns the full list of users.
func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the FetchUserList of UserController")

	users, err := c.userService.FetchUserList(r.Context())
	if err != nil {
		exceptionhandling.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// integer ID.
//
// MIGRATION_NOTE: The Java method declared "throws UserNotFoundException",
// which Spring resolved through its exception handler. Here the error is
// translated explicitly via WriteError.
func (c *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		exceptionhandling.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	user, err := c.userService.FetchUserByID(r.Context(), id)
	if err != nil {
		exceptionhandling.WriteError(w, r, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by ID
// and returns a success message.
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		exceptionhandling.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := c.userService.DeleteUser(r.Context(), id); err != nil {
		exceptionhandling.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. It updates the user with the
// given ID using the JSON request body and echoes back the input user.
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		exceptionhandling.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		exceptionhandling.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := c.userService.UpdateUser(r.Context(), id, &user); err != nil {
		exceptionhandling.WriteError(w, r, http.StatusInternalServerError, err)
		return
	}

	// MIGRATION_NOTE: Mirrors the Java controller, which returned the input
	// user object rather than the persisted result.
	writeJSON(w, http.StatusOK, user)
}

// GetUserByName handles GET /get_user_name/name/{name}. It returns a user
// looked up by name.
func (c *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := c.userService.GetUserByName(r.Context(), name)
	if err != nil {
		exceptionhandling.WriteError(w, r, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// parseIDParam extracts and parses an integer path parameter from the request.
func parseIDParam(r *http.Request, key string) (int, error) {
	return strconv.Atoi(chi.URLParam(r, key))
}

// writeJSON writes the given value as a JSON response with the provided status
// code. Encoding errors are logged implicitly via the default handler.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
