package handler

// Package handler contains the HTTP transport layer for the SmartContact
// application. This file corresponds to the original Spring Java class
// com.smartContact.Controller.UserController.
//
// MIGRATION_NOTE: The Java source used Spring MVC annotations
// (@RestController, @GetMapping/@PostMapping/@PutMapping/@DeleteMapping,
// @PathVariable, @RequestBody, @Valid) plus field injection (@Autowired
// UserService). In idiomatic Go we replace all of that with:
//   - an explicit constructor (NewUserController) that receives its
//     collaborators (the service.UserService interface),
//   - explicit JSON encoding/decoding via encoding/json,
//   - explicit route registration on a chi router (RegisterRoutes),
//   - explicit HTTP status codes written to the ResponseWriter,
//   - explicit error translation (mirroring the behaviour of Spring's
//     @ControllerAdvice by delegating to the already-migrated
//     restresponseentityexceptionhandling helpers).
//
// Spring's @Valid bean validation is replaced by an explicit call to the
// already-migrated model.Validate. Per the migration notes, updateUser does
// NOT run validation (mirroring the original, which never called validate),
// and it returns the submitted user object rather than the persisted result.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	smartErr "internal/smartcontact/error"
	"internal/smartcontact/error/restresponseentityexceptionhandling"
	"internal/smartcontact/model"
	"internal/smartcontact/service"
)

// UserController exposes CRUD HTTP endpoints for User entities, delegating all
// business logic to a service.UserService. It is the Go equivalent of the
// Spring @RestController UserController.
type UserController struct {
	userService service.UserService
}

// NewUserController constructs a UserController wired to the given
// service.UserService. This replaces Spring's @Autowired field injection with
// explicit constructor injection.
func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// RegisterRoutes registers every HTTP route owned by this controller on the
// supplied chi router. All routes from the original Spring controller are
// registered here at their exact method and path.
//
// MIGRATION_NOTE: The original getUserNameByName mapping was written without a
// leading slash ("get_user_name/name/{name}"); Spring normalizes this to a
// leading-slash path, so we register it as "/get_user_name/name/{name}".
func (c *UserController) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", c.SaveUser)
	r.Get("/get_user_data", c.FetchUserList)
	r.Get("/get_user_data/{id}", c.FetchUserByID)
	r.Delete("/delete_user_data/{id}", c.DeleteUser)
	r.Put("/update_user_data/{id}", c.UpdateUser)
	r.Get("/get_user_name/name/{name}", c.GetUserNameByName)
}

// SaveUser handles POST /save_user_data. It validates and saves a new user
// decoded from the JSON request body and, on success, returns a plain-text
// success message with 200 OK.
func (c *UserController) SaveUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := decodeUser(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	// Replaces Spring's @Valid bean validation trigger.
	if err := model.Validate(user); err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := c.userService.SaveUser(ctx, user); err != nil {
		c.writeServiceError(ctx, w, err)
		return
	}

	writePlainText(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data. It returns the full list of users
// as a JSON array. An empty table yields [] (never null), matching the
// repository-layer GAP J fix.
func (c *UserController) FetchUserList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := c.userService.FetchUserList(ctx)
	if err != nil {
		c.writeServiceError(ctx, w, err)
		return
	}

	// Guarantee a JSON array rather than null on an empty result set.
	if users == nil {
		users = []model.User{}
	}

	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. It returns a single user by
// integer id, or an error response (404 for a missing user) mirroring Spring's
// UserNotFoundException handling.
func (c *UserController) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	user, err := c.userService.FetchUserByID(ctx, id)
	if err != nil {
		c.writeServiceError(ctx, w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. It deletes the user with
// the given integer id and returns a plain-text success message with 200 OK.
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := c.userService.DeleteUser(ctx, id); err != nil {
		c.writeServiceError(ctx, w, err)
		return
	}

	writePlainText(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. It updates the user with the
// given id from the JSON request body and returns the SUBMITTED user object
// (not the persisted result), matching the original Spring behaviour.
//
// MIGRATION_NOTE: As in the Java source, no validation is run on the update
// path — model.Validate is intentionally NOT called here.
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	user, err := decodeUser(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	if err := c.userService.UpdateUser(ctx, id, user); err != nil {
		c.writeServiceError(ctx, w, err)
		return
	}

	// Return the submitted user object, exactly as the original controller did.
	writeJSON(w, http.StatusOK, user)
}

// GetUserNameByName handles GET /get_user_name/name/{name}. It fetches a user
// by name string.
func (c *UserController) GetUserNameByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	name := chi.URLParam(r, "name")

	user, err := c.userService.GetUserByName(ctx, name)
	if err != nil {
		c.writeServiceError(ctx, w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// decodeUser decodes a model.User from the JSON request body.
func decodeUser(r *http.Request) (model.User, error) {
	var user model.User
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&user); err != nil {
		return model.User{}, err
	}
	return user, nil
}

// parseID extracts and parses the {id} path parameter as an integer.
func parseID(r *http.Request) (int, error) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// writeServiceError translates service-layer errors into HTTP responses. A
// missing-user error maps to 404 via the already-migrated exception handler
// (mirroring Spring's @ControllerAdvice); all other errors become 500.
func (c *UserController) writeServiceError(ctx context.Context, w http.ResponseWriter, err error) {
	if errors.Is(err, smartErr.ErrUserNotFound) {
		restresponseentityexceptionhandling.HandleUserNotFound(w, err)
		return
	}
	restresponseentityexceptionhandling.WriteError(w, http.StatusInternalServerError, err)
}

// writeBadRequest writes a 400 Bad Request response with a structured error
// message body.
func writeBadRequest(w http.ResponseWriter, err error) {
	restresponseentityexceptionhandling.WriteError(w, http.StatusBadRequest, err)
}

// writeJSON serialises v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writePlainText writes a plain-text body with the given status code.
func writePlainText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}
