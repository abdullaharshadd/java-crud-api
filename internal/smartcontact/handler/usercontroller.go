// Package handler contains the HTTP transport layer for the smartcontact
// application. It exposes the User CRUD endpoints and delegates all business
// logic to the service layer.
//
// MIGRATION_NOTE: The Java source, UserController, was a Spring @RestController
// using field injection (@Autowired), declarative routing (@GetMapping etc.)
// and ResponseEntity for status control. In Go these annotations have no
// runtime equivalent, so routing is registered explicitly (RegisterRoutes),
// dependencies are injected via a constructor (NewUserController), and status
// codes / serialization are written directly to the http.ResponseWriter.
//
// MIGRATION_NOTE: Spring's @Valid bean validation is replaced by an explicit
// call to the model's Validate method. Spring's global exception handler
// (@ControllerAdvice) is replaced by explicit error inspection in each handler,
// delegating to the error package's WriteDomainError / WriteError helpers.
//
// MIGRATION_NOTE (BLOCKER RESOLUTION): The Java saveUser returned
// ResponseEntity<String>("User data saved successfully!", HttpStatus.OK) —
// i.e. Variant A: a plain-text body. This is preserved here via writeText with
// the textContentType constant. Likewise deleteUser returns a plain-text
// message. All User/list responses use writeJSON. If a running-app curl -i
// capture later shows a different content type, adjust textContentType.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	smartcontacterror "migrated-app/internal/smartcontact/error"
	"migrated-app/internal/smartcontact/model"
	"migrated-app/internal/smartcontact/service"
)

// textContentType is the Content-Type used for the plain-text mutation
// responses (saveUser / deleteUser), preserving the Java
// ResponseEntity<String> behaviour.
//
// MIGRATION_NOTE: Spring's ResponseEntity<String> defaults to
// "text/plain;charset=UTF-8". Confirm against a `curl -i` capture from the
// running Java app if byte-for-byte parity matters.
const textContentType = "text/plain; charset=utf-8"

// UserController exposes HTTP handlers for the User resource, delegating all
// business logic to a UserService.
type UserController struct {
	userService service.UserService
}

// NewUserController constructs a UserController with its required dependencies.
//
// MIGRATION_NOTE: Replaces Spring field injection (@Autowired UserService)
// with explicit constructor injection.
func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// RegisterRoutes wires the controller's handlers onto the given ServeMux at the
// exact paths defined by the original Spring mappings.
//
// MIGRATION_NOTE: Uses net/http's method-aware pattern routing (Go 1.22+).
// Path variables ({id}, {name}) are read via r.PathValue. The original Java
// mapping for get_user_name lacked a leading slash; the canonical, reachable
// path "/get_user_name/name/{name}" is registered here.
func (c *UserController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /save_user_data", c.SaveUser)
	mux.HandleFunc("GET /get_user_data", c.FetchUserList)
	mux.HandleFunc("GET /get_user_data/{id}", c.FetchUserByID)
	mux.HandleFunc("DELETE /delete_user_data/{id}", c.DeleteUser)
	mux.HandleFunc("PUT /update_user_data/{id}", c.UpdateUser)
	mux.HandleFunc("GET /get_user_name/name/{name}", c.GetUserByName)
}

// SaveUser handles POST /save_user_data. It validates and persists a new user,
// returning a plain-text success message with 200 OK.
func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	log.Println("inside the saveUser of UserController")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		smartcontacterror.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// MIGRATION_NOTE: Replaces Spring @Valid bean validation.
	if err := user.Validate(); err != nil {
		smartcontacterror.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := c.userService.SaveUser(r.Context(), &user); err != nil {
		smartcontacterror.WriteDomainError(w, err)
		return
	}

	writeText(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns the full list of users
// as JSON.
func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	log.Println("inside the fetchUserList of UserController")

	users, err := c.userService.FetchUserList(r.Context())
	if err != nil {
		smartcontacterror.WriteDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// ID, or a 404 if no such user exists.
func (c *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		smartcontacterror.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := c.userService.FetchUserByID(r.Context(), id)
	if err != nil {
		// MIGRATION_NOTE: Replaces the global @ControllerAdvice that mapped
		// UserNotFoundException -> 404. WriteDomainError performs that mapping.
		smartcontacterror.WriteDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by ID and
// returns a plain-text success message with 200 OK.
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		smartcontacterror.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := c.userService.DeleteUser(r.Context(), id); err != nil {
		smartcontacterror.WriteDomainError(w, err)
		return
	}

	writeText(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. It updates the user with the
// given ID and echoes back the submitted payload as JSON.
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		smartcontacterror.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		smartcontacterror.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if _, err := c.userService.UpdateUser(r.Context(), id, &user); err != nil {
		smartcontacterror.WriteDomainError(w, err)
		return
	}

	// MIGRATION_NOTE: The Java handler echoed back the submitted payload
	// (return user), not the persisted entity. This behaviour is preserved.
	writeJSON(w, http.StatusOK, &user)
}

// GetUserByName handles GET /get_user_name/name/{name}. It fetches a user by
// name.
func (c *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	user, err := c.userService.GetUserByName(r.Context(), name)
	if err != nil {
		smartcontacterror.WriteDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// writeJSON serializes v to JSON and writes it with the given status code.
// It uses json.Marshal (no trailing newline) to match the canonical response
// format used across the migrated application.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		smartcontacterror.WriteError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// writeText writes a plain-text response body with the given status code,
// preserving the Java ResponseEntity<String> behaviour for mutation endpoints.
func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", textContentType)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

// ensure the error sentinel import is referenced so intent is explicit even if
// future refactors remove direct WriteDomainError usage.
var _ = errors.Is
var _ = smartcontacterror.ErrUserNotFound
