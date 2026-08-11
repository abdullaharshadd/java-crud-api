// Package handler contains the HTTP transport layer for the smartcontact
// application. It maps HTTP requests to service calls and translates results
// (and errors) back into HTTP responses.
//
// MIGRATION_NOTE: The Java source was a Spring @RestController (UserController)
// exposing six CRUD endpoints and delegating all logic to a @Autowired
// UserService. Go has no annotation-driven routing or automatic JSON
// (de)serialization, so:
//   - Field injection (@Autowired) becomes explicit constructor injection via
//     NewUserController.
//   - @GetMapping/@PostMapping/etc. become explicit chi route registrations in
//     RegisterRoutes.
//   - @RequestBody becomes json.NewDecoder(r.Body).Decode(...).
//   - @PathVariable becomes chi.URLParam(...); non-numeric {id} yields 400.
//   - @Valid is preserved asymmetrically exactly as the source had it:
//     saveUser validates the decoded User; updateUser does NOT.
//   - ResponseEntity<String> becomes writing a plain-text body with 200 OK.
//   - Checked UserNotFoundException propagation is replaced by returning the
//     error to apperror.WriteError, which maps it to the correct HTTP status.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/smartContact/internal/smartcontact/error/apperror"
	"github.com/smartContact/internal/smartcontact/model"
	"github.com/smartContact/internal/smartcontact/service"
)

// UserController is the HTTP handler set for User resources. It holds its
// dependencies (the service and a validator) as unexported fields, injected via
// NewUserController rather than Spring field injection.
type UserController struct {
	userService service.UserService
	validate    *validator.Validate
}

// NewUserController constructs a UserController with the given service and
// validator. This replaces Spring's @Autowired field injection with explicit
// constructor injection.
func NewUserController(userService service.UserService, validate *validator.Validate) *UserController {
	if validate == nil {
		validate = validator.New()
	}
	return &UserController{
		userService: userService,
		validate:    validate,
	}
}

// RegisterRoutes wires all User endpoints onto the given chi.Router at exactly
// the paths and methods the Spring controller exposed.
func (h *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", h.SaveUser)
	r.Get("/get_user_data", h.FetchUserList)
	r.Get("/get_user_data/{id}", h.FetchUserByID)
	r.Delete("/delete_user_data/{id}", h.DeleteUser)
	r.Put("/update_user_data/{id}", h.UpdateUser)
	r.Get("/get_user_name/name/{name}", h.GetUserByName)
}

// SaveUser handles POST /save_user_data. It decodes and validates a User from
// the JSON request body, saves it, and returns a plain-text success message
// with 200 OK.
func (h *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		apperror.WriteError(w, r, badRequestError(err))
		return
	}

	// @Valid on saveUser is preserved: validate the decoded body.
	if err := h.validate.Struct(&user); err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	if err := h.userService.SaveUser(r.Context(), &user); err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns the full list of users
// as JSON.
func (h *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.FetchUserList(r.Context())
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// integer id. A non-numeric id yields 400; a missing user yields 404 via the
// service's UserNotFound error.
func (h *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	user, err := h.userService.FetchUserByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by id and
// returns a plain-text success message with 200 OK.
func (h *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeText(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. It updates a user by id from
// the JSON body and echoes back the submitted user object.
//
// MIGRATION_NOTE: The source's updateUser had no @Valid annotation, so no
// validation is performed here — the asymmetry with SaveUser is intentional.
func (h *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		apperror.WriteError(w, r, badRequestError(err))
		return
	}

	if err := h.userService.UpdateUser(r.Context(), id, &user); err != nil {
		apperror.WriteError(w, r, err)
		return
	}

	// Echo back the submitted user, matching the Java controller's behavior.
	writeJSON(w, http.StatusOK, user)
}

// GetUserByName handles GET /get_user_name/name/{name}. It fetches a user by
// name and returns it as JSON.
func (h *UserController) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := h.userService.GetUserByName(r.Context(), name)
	if err != nil {
		apperror.WriteError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// parseIDParam extracts and parses an integer path parameter, returning a
// 400-mapped error when the value is not a valid integer.
func parseIDParam(r *http.Request, key string) (int, error) {
	raw := chi.URLParam(r, key)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, badRequestError(err)
	}
	return id, nil
}

// badRequestError is a marker error type that apperror.StatusFor / WriteError
// map to HTTP 400 Bad Request.
//
// MIGRATION_NOTE: A non-numeric {id} in Spring produced a framework-level 400.
// We reproduce that by wrapping the parse failure in an error that the central
// error writer recognizes as a bad request.
type badRequest struct{ err error }

func badRequestError(err error) error { return &badRequest{err: err} }

// Error implements the error interface.
func (e *badRequest) Error() string {
	if e == nil || e.err == nil {
		return "bad request"
	}
	return e.err.Error()
}

// Unwrap exposes the underlying cause for errors.Is/As.
func (e *badRequest) Unwrap() error { return e.err }

// BadRequest reports whether this error should map to HTTP 400. apperror can
// call this via errors.As without importing the handler package's concrete
// type.
func (e *badRequest) BadRequest() bool { return true }

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeText writes a plain-text body with the given status code, mirroring the
// Java ResponseEntity<String> responses.
func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}
