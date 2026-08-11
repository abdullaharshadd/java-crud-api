// Package handler contains the HTTP delivery layer migrated from
// com.smartContact.Controller.
//
// MIGRATION_NOTE: The Java UserController was a Spring @RestController that
// relied on annotations (@GetMapping/@PostMapping/etc.) for routing, field
// injection (@Autowired) for the UserService dependency, automatic JSON
// (de)serialization of return values/request bodies, @Valid bean validation,
// and a @ControllerAdvice for exception-to-HTTP mapping.
//
// In idiomatic Go these become:
//   - Explicit route registration on an *http.ServeMux (RegisterRoutes).
//   - Constructor injection via NewUserHandler(service.UserService).
//   - Manual encoding/json marshalling of responses / decoding of bodies.
//   - Explicit validation calls (model.User.Validate) where the source used
//     @Valid (saveUser only — updateUser deliberately omits validation to
//     preserve the source asymmetry).
//   - apperror.MapError at the HTTP boundary in place of @ControllerAdvice.
package handler

import (
	"encoding/json"
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
// This replaces Spring's @Autowired field injection with explicit
// constructor injection.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// RegisterRoutes registers every user endpoint on the provided mux at the
// exact method/path pairs the source controller exposed.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /save_user_data", h.SaveUser)
	mux.HandleFunc("GET /get_user_data", h.FetchUserList)
	mux.HandleFunc("GET /get_user_data/{id}", h.FetchUserById)
	mux.HandleFunc("DELETE /delete_user_data/{id}", h.DeleteUser)
	mux.HandleFunc("PUT /update_user_data/{id}", h.UpdateUser)
	mux.HandleFunc("GET /get_user_name/name/{name}", h.GetUserNameByName)
}

// SaveUser validates and persists a new user parsed from the JSON request
// body, then returns a success message with 200 OK.
//
// Mirrors the source @Valid behavior: the body is validated before saving.
func (h *UserHandler) SaveUser(w http.ResponseWriter, r *http.Request) {
	log.Print("inside the saveUser of UserController")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	// @Valid equivalent — validation is applied on save (asymmetric with update).
	if err := user.Validate(); err != nil {
		writeError(w, err)
		return
	}

	if err := h.service.SaveUser(r.Context(), &user); err != nil {
		writeError(w, err)
		return
	}

	writeString(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList returns a JSON list of all users.
func (h *UserHandler) FetchUserList(w http.ResponseWriter, r *http.Request) {
	log.Print("inside the fetchUserList of UserController")

	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// FetchUserById returns a single user by integer id, mapping a
// UserNotFoundError to 404 via apperror.MapError.
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

// DeleteUser deletes a user by id and returns a success message with 200 OK.
//
// MIGRATION_NOTE: the source deleteUser does not distinguish a missing row;
// any service error propagates to a generic 500 via apperror.MapError,
// preserving the original behavior.
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

// UpdateUser updates a user by id from the request body and returns the input
// user object (the request payload, not the persisted entity — matching the
// source).
//
// MIGRATION_NOTE: unlike SaveUser, the source updateUser deliberately omits
// @Valid validation; that asymmetry is preserved here.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	if err := h.service.UpdateUser(r.Context(), id, &user); err != nil {
		writeError(w, err)
		return
	}

	// Return the request payload, not the persisted entity.
	writeJSON(w, http.StatusOK, user)
}

// GetUserNameByName fetches a user by name path variable.
//
// MIGRATION_NOTE: the source getUserNameByName has no @ControllerAdvice for a
// name miss, so a nil result is serialized as JSON null with 200 OK. That
// behavior is preserved: on a name miss the service returns (nil, nil) and we
// encode null.
func (h *UserHandler) GetUserNameByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	user, err := h.service.GetUserNameByName(r.Context(), name)
	if err != nil {
		writeError(w, err)
		return
	}

	// user may be nil here — encodes as JSON null with 200, matching the source.
	writeJSON(w, http.StatusOK, user)
}

// pathID extracts and parses the {id} path variable as an int, returning a
// *apperror.ValidationError (mapped to 400) when it is non-numeric.
func pathID(r *http.Request) (int, error) {
	raw := r.PathValue("id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperr.NewValidationError("id must be a valid integer")
	}
	return id, nil
}

// writeError maps an error to its HTTP status and structured body using the
// centralized apperror.MapError (the @ControllerAdvice replacement).
func writeError(w http.ResponseWriter, err error) {
	status, body := apperr.MapError(err)
	writeJSON(w, status, body)
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// writeString serializes s as a plain-string JSON-compatible response body,
// matching Spring's ResponseEntity<String> which returns the raw string.
func writeString(w http.ResponseWriter, status int, s string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(s)); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
