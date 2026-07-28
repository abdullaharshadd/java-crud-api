// Package handler provides the HTTP transport layer for the SmartContact
// application. It replaces the Java Spring @RestController under
// com.smartContact.Controller.
//
// MIGRATION_NOTE: The Java source was a @RestController using field injection
// (@Autowired UserService) and declarative routing annotations
// (@GetMapping/@PostMapping/etc.). In idiomatic Go there is no component
// scanning: the dependency (the service) is injected explicitly via the
// NewUserController constructor, and routes are registered explicitly on a
// chi router in RegisterRoutes (wired at the composition root).
//
// Behavioural deviations agreed during migration:
//   - deviation 2: PUT on a missing id returns 404 (Spring's original silently
//     updated nothing / echoed the body). The service surfaces ErrUserNotFound.
//   - deviation 5: all not-found paths are consistent via the ErrUserNotFound
//     sentinel, mapped to HTTP 404 by the shared WriteError helper.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	smarterr "migrated-app/internal/smartcontact/error"
	resterr "migrated-app/internal/smartcontact/error/restresponseentityexceptionhandling"
	"migrated-app/internal/smartcontact/model"
)

// UserServicer is the narrow interface the handler depends on. Defining it at
// the point of use (rather than importing the concrete *service.UserService)
// keeps the handler decoupled and trivially mockable in tests.
//
// MIGRATION_NOTE: This mirrors the Java UserService contract, but only the
// methods this controller actually calls are declared here. Signatures match
// service.UserService exactly: context.Context (not *http.Request — the
// service layer has no transport dependency) and model.UserResponse (the
// service's password-omitting DTO, not the internal model.User).
type UserServicer interface {
	// SaveUser validates and persists a new user.
	SaveUser(ctx context.Context, req model.CreateUserRequest) (model.UserResponse, error)
	// FetchUserList returns all users.
	FetchUserList(ctx context.Context) ([]model.UserResponse, error)
	// FetchUserByID returns a single user by id, or ErrUserNotFound.
	FetchUserByID(ctx context.Context, id int) (model.UserResponse, error)
	// DeleteUser removes a user by id.
	DeleteUser(ctx context.Context, id int) error
	// UpdateUser updates a user by id and returns the updated user.
	UpdateUser(ctx context.Context, id int, req model.CreateUserRequest) (model.UserResponse, error)
	// GetUserByName returns a user matched by name, or ErrUserNotFound.
	GetUserByName(ctx context.Context, name string) (model.UserResponse, error)
}

// UserController holds the dependencies needed to serve user endpoints.
type UserController struct {
	svc    UserServicer
	logger *slog.Logger
}

// NewUserController constructs a UserController with the given service and
// logger. If logger is nil the default slog logger is used.
//
// MIGRATION_NOTE: replaces Spring field injection (@Autowired) with explicit
// constructor injection.
func NewUserController(svc UserServicer, logger *slog.Logger) *UserController {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserController{svc: svc, logger: logger}
}

// RegisterRoutes wires every user endpoint onto the provided chi router at the
// exact HTTP method and path the original Spring controller exposed.
func (c *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", c.SaveUser)
	r.Get("/get_user_data", c.FetchUserList)
	r.Get("/get_user_data/{id}", c.FetchUserByID)
	r.Delete("/delete_user_data/{id}", c.DeleteUser)
	r.Put("/update_user_data/{id}", c.UpdateUser)
	r.Get("/get_user_name/name/{name}", c.GetUserByName)
}

// SaveUser handles POST /save_user_data. It decodes and validates the request
// body, persists the user, and returns a success message with HTTP 200.
func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the saveUser of UserController")

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resterr.WriteError(w, r, http.StatusBadRequest, err)
		return
	}
	if problems := req.Validate(); len(problems) > 0 {
		resterr.WriteError(w, r, http.StatusBadRequest, errors.New(strings.Join(problems, "; ")))
		return
	}

	if _, err := c.svc.SaveUser(r.Context(), req); err != nil {
		c.mapErr(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns the full list of users
// as JSON.
func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("inside the fetchUserList of UserController")

	responses, err := c.svc.FetchUserList(r.Context())
	if err != nil {
		c.mapErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, responses)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// id, or HTTP 404 if no such user exists.
func (c *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		resterr.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	user, err := c.svc.FetchUserByID(r.Context(), id)
	if err != nil {
		c.mapErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by id and
// returns a success message with HTTP 200.
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		resterr.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	if err := c.svc.DeleteUser(r.Context(), id); err != nil {
		c.mapErr(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. It updates a user by id and
// echoes the updated user.
//
// MIGRATION_NOTE (deviation 2): unlike the original Spring handler which
// echoed the submitted body unconditionally, a PUT against a missing id now
// yields HTTP 404 via the ErrUserNotFound sentinel returned by the service.
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		resterr.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resterr.WriteError(w, r, http.StatusBadRequest, err)
		return
	}

	user, err := c.svc.UpdateUser(r.Context(), id, req)
	if err != nil {
		c.mapErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// GetUserByName handles GET /get_user_name/name/{name}. It returns a user
// matched by name, or HTTP 404 if none exists.
func (c *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := c.svc.GetUserByName(r.Context(), name)
	if err != nil {
		c.mapErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// mapErr translates a service-layer error into an HTTP response.
// ErrUserNotFound maps to 404 (deviation 5: all not-found paths consistent);
// everything else maps to 500.
func (c *UserController) mapErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, smarterr.ErrUserNotFound) {
		resterr.WriteError(w, r, http.StatusNotFound, err)
		return
	}
	c.logger.Error("unexpected error handling request", slog.String("err", err.Error()))
	resterr.WriteError(w, r, http.StatusInternalServerError, err)
}

// pathID extracts and parses the {id} path parameter.
func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

// writeJSON serialises v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeText writes a plain-text response with the given status code, matching
// the original controller's String ResponseEntity bodies.
func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}
