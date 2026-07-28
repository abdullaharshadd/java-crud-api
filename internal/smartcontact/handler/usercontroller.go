// Package handler contains the HTTP handlers for the smartContact
// application. It is the presentation layer that translates HTTP requests into
// service-layer calls and marshals the results back into HTTP responses.
//
// MIGRATION_NOTE: The Java source was a Spring @RestController (UserController)
// that relied on annotation-driven routing (@GetMapping / @PostMapping / etc.),
// automatic JSON (de)serialization, @Valid bean validation, @PathVariable /
// @RequestBody binding, and a global @ControllerAdvice for exception mapping.
// Go has no annotation-driven routing or AOP, so:
//   - Routing is registered explicitly on a chi router in RegisterRoutes.
//   - Field injection (@Autowired UserService) becomes explicit constructor
//     injection via NewUserController.
//   - JSON (de)serialization is done explicitly with encoding/json.
//   - Bean validation (@Valid) is done explicitly (see validateUser); there is
//     no equivalent runtime validation framework wired in here.
//   - The global exception handler becomes the RecoverMiddleware +
//     WriteError helper from the error package; UserNotFoundError is written
//     out as a structured 404.
//
// MIGRATION_NOTE: Per the migration debate, every user-returning endpoint
// (fetchUserById, updateUser, getUserNameByName, and the list endpoint) emits a
// model.UserResponse rather than model.User so that nullable columns serialize
// as JSON null consistently.
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

// UserService is the subset of the service-layer contract that the
// UserController depends on. It is declared here (in the consuming package) so
// the handler is decoupled from any concrete implementation.
//
// MIGRATION_NOTE: This mirrors the methods of the migrated
// service.UserService interface. It is redeclared locally as a consumer-side
// interface (idiomatic Go) rather than importing the service interface, which
// keeps the handler package free of a hard dependency on the service package's
// concrete surface. Any type satisfying service.UserService also satisfies
// this interface.
type UserService interface {
	SaveUser(ctx context.Context, user model.User) error
	FetchUserList(ctx context.Context) ([]model.UserResponse, error)
	FetchUserByID(ctx context.Context, id int) (model.UserResponse, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, id int, user model.User) error
	GetUserByName(ctx context.Context, name string) (model.UserResponse, error)
}

// UserController exposes the CRUD HTTP endpoints for User entities. It
// delegates all business logic to the injected UserService.
type UserController struct {
	service UserService
	logger  *slog.Logger
}

// NewUserController constructs a UserController with the given service
// dependency. If logger is nil, the default slog logger is used.
func NewUserController(service UserService, logger *slog.Logger) *UserController {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserController{service: service, logger: logger}
}

// RegisterRoutes wires all UserController endpoints onto the provided chi
// router at their exact paths and HTTP methods.
//
// MIGRATION_NOTE: In the Java source the last mapping ("get_user_name/name/{name}")
// lacked a leading slash; Spring normalized it to "/get_user_name/name/{name}".
// We register the normalized, leading-slash form here.
func (c *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", c.SaveUser)
	r.Get("/get_user_data", c.FetchUserList)
	r.Get("/get_user_data/{id}", c.FetchUserByID)
	r.Delete("/delete_user_data/{id}", c.DeleteUser)
	r.Put("/update_user_data/{id}", c.UpdateUser)
	r.Get("/get_user_name/name/{name}", c.GetUserByName)
}

// SaveUser handles POST /save_user_data. It decodes and validates a User from
// the JSON request body, persists it, and returns a success message with 200
// OK.
func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the saveUser of UserController")

	var user model.User
	if err := decodeJSON(r, &user); err != nil {
		writeBadRequest(w, err)
		return
	}

	// MIGRATION_NOTE: Replaces Spring's @Valid bean validation. There is no
	// validation framework wired in, so mandatory-field checks are explicit.
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

// FetchUserList handles GET /get_user_data. It returns the full list of users
// as JSON.
func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the fetchUserList of UserController")

	users, err := c.service.FetchUserList(r.Context())
	if err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	// Emit an empty JSON array rather than null when there are no users.
	if users == nil {
		users = []model.UserResponse{}
	}

	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// numeric id, or a 404 response if no such user exists.
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

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by
// numeric id and returns a success message with 200 OK.
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

// UpdateUser handles PUT /update_user_data/{id}. It updates the user with the
// given numeric id using the provided body, then echoes back the updated user
// as a UserResponse.
//
// MIGRATION_NOTE: The Java source echoed back the raw request User object.
// Per the migration debate we echo a model.UserResponse (built from the
// decoded body plus the path id) so nullable columns serialize consistently.
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

	writeJSON(w, http.StatusOK, model.UserResponse(user))
}

// GetUserByName handles GET /get_user_name/name/{name}. It fetches a user by
// name and returns it as JSON.
func (c *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := c.service.GetUserByName(r.Context(), name)
	if err != nil {
		c.handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleServiceError maps a service-layer error onto an HTTP response. A
// UserNotFoundError becomes a structured 404 (delegated to WriteError);
// anything else becomes a 500.
func (c *UserController) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var notFound *apperr.UserNotFoundError
	if errors.As(err, &notFound) {
		restresponseentityexceptionhandling.WriteError(w, r, err)
		return
	}

	c.logger.Error("unexpected error handling request", slog.String("error", err.Error()))
	writeJSON(w, http.StatusInternalServerError,
		model.NewErrorMessage(http.StatusInternalServerError, model.StatusText(http.StatusInternalServerError), err.Error()))
}

// pathID extracts and parses the {id} path parameter as an int.
func pathID(r *http.Request) (int, error) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("invalid id path parameter: must be an integer")
	}
	return id, nil
}

// decodeJSON decodes the request body into dst, rejecting unknown fields and
// empty bodies.
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return errors.New("invalid request body: " + err.Error())
	}
	return nil
}

// validateUser performs minimal mandatory-field validation, replacing Spring's
// @Valid.
//
// MIGRATION_NOTE: The Java model relied on javax.validation constraint
// annotations that are not visible in this file. The concrete constraints must
// be confirmed against model.User and this function updated accordingly.
func validateUser(user model.User) error {
	if user.Name == "" {
		return errors.New("validation failed: name must not be empty")
	}
	return nil
}

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeText writes a plain-text response with the given status code.
func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

// writeBadRequest writes a structured 400 response for client-side errors.
func writeBadRequest(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest,
		model.NewErrorMessage(http.StatusBadRequest, model.StatusText(http.StatusBadRequest), err.Error()))
}
