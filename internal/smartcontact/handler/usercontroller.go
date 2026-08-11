// Package handler provides the HTTP transport layer for the Smart Contact
// service. It is the Go equivalent of the source project's
// com.smartContact.Controller package.
//
// MIGRATION_NOTE: The Java source was a Spring @RestController (UserController)
// using @Autowired field injection of a UserService and annotation-driven
// routing (@PostMapping, @GetMapping, ...). Spring bound path variables and
// JSON request bodies automatically, ran bean validation via @Valid, and
// converted return values / thrown exceptions into HTTP responses.
//
// In idiomatic Go there is no annotation-driven routing or DI container. We
// therefore:
//
//   - Define a Handler struct holding the UserService dependency, constructed
//     explicitly via NewHandler (constructor injection instead of @Autowired).
//   - Register every route explicitly on a chi router in RegisterRoutes.
//   - Decode JSON bodies with encoding/json and run model validation manually.
//   - Translate errors into HTTP responses via the shared error handler
//     (restresponseentityexceptionhandling.WriteError), the Go analogue of
//     Spring's @ControllerAdvice exception handler.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	errhandler "github.com/smartcontact/internal/smartcontact/error/restresponseentityexceptionhandling"
	"github.com/smartcontact/internal/smartcontact/model"
	"github.com/smartcontact/internal/smartcontact/service"
)

// Handler exposes the User CRUD HTTP endpoints, delegating all business logic
// to the injected UserService. It is the Go equivalent of the Spring
// UserController @RestController.
type Handler struct {
	userService service.UserService
	logger      *slog.Logger
}

// NewHandler constructs a Handler with the given UserService and logger.
//
// MIGRATION_NOTE: This replaces Spring's @Autowired field injection with
// explicit constructor injection. If logger is nil, the default slog logger is
// used so the Handler is always safe to use.
func NewHandler(userService service.UserService, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registers every User endpoint on the given chi router at the
// exact HTTP method/path pairs exposed by the original Spring controller.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", h.SaveUser)
	r.Get("/get_user_data", h.FetchUserList)
	r.Get("/get_user_data/{id}", h.FetchUserByID)
	r.Delete("/delete_user_data/{id}", h.DeleteUser)
	r.Put("/update_user_data/{id}", h.UpdateUser)
	r.Get("/get_user_name/name/{name}", h.GetUserByName)
}

// SaveUser handles POST /save_user_data. It decodes and validates the request
// body, persists the user, and returns 200 with a success message.
func (h *Handler) SaveUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("inside the saveUser of UserController")

	// MIGRATION_NOTE: DisallowUnknownFields is intentionally NOT set (CHANGE 19)
	// to mirror Spring's lenient JSON binding.
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	// @Valid equivalent: run model validation explicitly.
	if err := user.Validate(); err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	if err := h.userService.SaveUser(r.Context(), &user); err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns the full list of users
// as JSON.
func (h *Handler) FetchUserList(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("inside the fetchUserList of UserController")

	users, err := h.userService.FetchUserList(r.Context())
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// id, or an error (translated to a 404) if the user does not exist.
func (h *Handler) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	user, err := h.userService.FetchUserByID(r.Context(), id)
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes a user by id and
// returns 200 with a success message.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	writeText(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. It updates a user by id and
// returns the submitted user payload.
//
// MIGRATION_NOTE: The id is parsed before the request body is decoded
// (CHANGE 13). The source returned the submitted payload; the migrated handler
// preserves that behaviour by echoing the decoded user back as JSON.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	if err := h.userService.UpdateUser(r.Context(), id, &user); err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, user)
}

// GetUserByName handles GET /get_user_name/name/{name}. It fetches a user by
// name.
func (h *Handler) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := h.userService.GetUserByName(r.Context(), name)
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, user)
}

// parseIDParam extracts and parses an integer URL parameter from the request.
func parseIDParam(r *http.Request, key string) (int, error) {
	raw := chi.URLParam(r, key)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// writeText writes a plain-text response with the given status code.
func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

// writeJSON serialises v as JSON with the given status code, delegating to the
// shared error handler if encoding fails.
func writeJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	buf, err := json.Marshal(v)
	if err != nil {
		errhandler.WriteError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf)
}
