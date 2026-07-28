// Package handler contains the HTTP handlers for the smartcontact service.
// Handlers translate HTTP requests into service-layer calls and marshal the
// results back into HTTP responses.
//
// MIGRATION_NOTE: The Java source was a Spring @RestController (UserController).
// Spring provided declarative routing (@GetMapping/@PostMapping/...), automatic
// JSON (de)serialization, @PathVariable/@RequestBody binding, @Valid bean
// validation, and ResponseEntity for explicit status codes. None of these have
// runtime equivalents in Go, so they are translated into explicit chi routes,
// encoding/json calls, manual path-parameter parsing, an explicit Validate()
// call, and explicit w.WriteHeader status codes.
//
// MIGRATION_NOTE: Field injection via @Autowired is replaced by constructor
// injection (NewUserController) that accepts a service.UserService interface.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	smartErr "internal/smartcontact/error"
	"internal/smartcontact/model"
	"internal/smartcontact/service"
)

// UserController exposes CRUD HTTP endpoints for User entities, delegating all
// business logic to a UserService.
type UserController struct {
	service service.UserService
	logger  zerolog.Logger
}

// NewUserController constructs a UserController backed by the supplied
// UserService and logger.
func NewUserController(svc service.UserService, logger zerolog.Logger) *UserController {
	return &UserController{
		service: svc,
		logger:  logger,
	}
}

// RegisterRoutes wires all of the controller's endpoints onto the given router.
//
// MIGRATION_NOTE: The final Java route @GetMapping("get_user_name/name/{name}")
// is missing its leading slash. Spring normalizes this to "/get_user_name/...";
// chi requires a leading slash, so the route is registered at
// "/get_user_name/name/{name}" to preserve the effective external path.
func (c *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", c.SaveUser)
	r.Get("/get_user_data", c.FetchUserList)
	r.Get("/get_user_data/{id}", c.FetchUserByID)
	r.Delete("/delete_user_data/{id}", c.DeleteUser)
	r.Put("/update_user_data/{id}", c.UpdateUser)
	r.Get("/get_user_name/name/{name}", c.GetUserByName)
}

// SaveUser validates and persists a new user, returning a success message with
// HTTP 200 on success.
//
// Route: POST /save_user_data
func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	c.logger.Info().Msg("inside the saveUser of UserController")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		c.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// MIGRATION_NOTE: Replaces @Valid bean validation on the request body.
	if err := user.Validate(); err != nil {
		c.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.service.SaveUser(r.Context(), &user); err != nil {
		// MIGRATION_NOTE: A duplicate-email violation (MySQL 1062 in the source)
		// is mapped faithfully to HTTP 500 to match Spring's default behavior.
		// Recommended deviation: return HTTP 409 Conflict instead.
		c.logger.Error().Err(err).Msg("failed to save user")
		c.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.writeString(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList returns the full list of users as JSON.
//
// Route: GET /get_user_data
func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	c.logger.Info().Msg("inside the fetchUserList of UserController")

	users, err := c.service.FetchUserList(r.Context())
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to fetch user list")
		c.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.writeJSON(w, http.StatusOK, users)
}

// FetchUserByID returns a single user by id, responding with HTTP 404 if the
// user is not found.
//
// Route: GET /get_user_data/{id}
func (c *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := c.parseID(w, r)
	if err != nil {
		return
	}

	user, err := c.service.GetUserByID(r.Context(), id)
	if err != nil {
		// MIGRATION_NOTE: The Java method declared `throws UserNotFoundException`,
		// which Spring's exception handling surfaced as HTTP 404.
		if smartErr.IsUserNotFound(err) {
			c.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		c.logger.Error().Err(err).Msg("failed to fetch user by id")
		c.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.writeJSON(w, http.StatusOK, user)
}

// DeleteUser deletes a user by id, returning a success message with HTTP 200.
//
// Route: DELETE /delete_user_data/{id}
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := c.parseID(w, r)
	if err != nil {
		return
	}

	if err := c.service.DeleteUser(r.Context(), id); err != nil {
		if smartErr.IsUserNotFound(err) {
			c.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		c.logger.Error().Err(err).Msg("failed to delete user")
		c.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.writeString(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser updates a user by id and echoes the updated payload back to the
// caller.
//
// Route: PUT /update_user_data/{id}
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := c.parseID(w, r)
	if err != nil {
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		c.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// MIGRATION_NOTE: The Java handler had no @Valid on the update body, so no
	// validation is performed here to preserve behavior exactly.
	if err := c.service.UpdateUser(r.Context(), id, &user); err != nil {
		if smartErr.IsUserNotFound(err) {
			c.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		c.logger.Error().Err(err).Msg("failed to update user")
		c.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// The Java handler echoes the request payload back verbatim.
	c.writeJSON(w, http.StatusOK, user)
}

// GetUserByName looks up a user by name.
//
// Route: GET /get_user_name/name/{name}
//
// MIGRATION_NOTE (V4 blocker resolution): The Java UserServiceImp.getUserByName
// caught its exception and returned null, so Spring serialized a 200 response
// with a null body. To preserve that observable behavior:
//   - a nil user (not found / swallowed error) yields HTTP 200 with a JSON null
//     body, matching the Java 200-null branch;
//   - a genuine transport/DB error yields HTTP 500.
// The 404 branch does not apply here because the source never produced one for
// this endpoint.
func (c *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := c.service.GetUserByName(r.Context(), name)
	if err != nil {
		// A not-found is treated as the Java null return (HTTP 200, null body).
		if smartErr.IsUserNotFound(err) || errors.Is(err, smartErr.ErrUserNotFound) {
			c.writeJSON(w, http.StatusOK, nil)
			return
		}
		c.logger.Error().Err(err).Msg("failed to fetch user by name")
		c.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.writeJSON(w, http.StatusOK, user)
}

// parseID extracts and validates the {id} path parameter, writing a 400
// response and returning an error if it is malformed.
func (c *UserController) parseID(w http.ResponseWriter, r *http.Request) (int, error) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		c.writeError(w, http.StatusBadRequest, "invalid id path parameter")
		return 0, err
	}
	return id, nil
}

// writeJSON serializes v as JSON with the given status code.
func (c *UserController) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		c.logger.Error().Err(err).Msg("failed to encode JSON response")
	}
}

// writeString writes a plain-text response body with the given status code.
//
// MIGRATION_NOTE: The Java handlers returned ResponseEntity<String> with a
// plain string body (not JSON), so these are written as text/plain to match.
func (c *UserController) writeString(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(msg)); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response body")
	}
}

// writeError emits a structured ErrorMessage JSON payload with the given status.
func (c *UserController) writeError(w http.ResponseWriter, status int, msg string) {
	c.writeJSON(w, status, model.NewErrorMessage(status, msg))
}
