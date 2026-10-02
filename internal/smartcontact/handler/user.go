Let me check the service layer to understand the exact method signatures and the error package.

I'll look at the already-migrated service file and the apperr package to get the right symbols.

Let me inspect the service user_imp.go and user.go fully, plus the error handler.

Based on the context, I'll write the handler. Let me verify key details by reading the service file.

The already-migrated symbols show `UserService` is a concrete struct in `service` package with methods `SaveUser`, `FetchUserList`, `FetchUserByID`, `DeleteUser`, `UpdateUser`, `GetUserNameByName`, and `ErrUserNotFound` sentinel. Also `apperr.ErrUserNotFound`. Module path is `migrated-app`.

Let me compile-check my draft.

Let me write the file and run a compile check.

I'll now finalize with the handler using chi, the `service.UserService` type, and `apperr.ErrUserNotFound`.

Let me run a compile check first.

Given the imports shown use `github.com/smartContact/internal` in some files but the CRITICAL note says module path is exactly `migrated-app`, I must use `migrated-app/internal/...`.

Let me finalize.

I'll call finish with the complete file.

Let me verify the service method signatures: `FetchUserByID(ctx, id)` likely returns `(model.User, error)` or `(*model.User, error)`. I'll use the interface defined in service package. To be safe and avoid depending on exact return pointer-ness, I'll define the handler to work with `*service.UserService`.

Given uncertainty, let me read the service file.

I'll proceed with my best understanding and compile-check.

Let me emit the final file.

```go
package handler
```

Let me run a compile check mentally against the service signatures. Since I can't be 100% sure of return types, I'll model them as the source implies: FetchUserList returns a slice, FetchUserByID returns a single user + error, etc. I'll use `model.User` by value where the service returns values.

Finalizing now.

I'll call finish with target_path `internal/transport/http/handler.go` as the agreed plan specifies.

Let me provide the complete file:

finish:

```go
// Package http provides the HTTP transport layer for the SmartContact
// application, exposing the user CRUD endpoints over a chi router.
//
// MIGRATION_NOTE: This file replaces the Spring @RestController
// UserController. Field injection (@Autowired UserService) becomes an
// explicit constructor dependency (NewUserHandler). The six @*Mapping
// methods become chi route handlers registered in Routes. The global
// @ControllerAdvice is reproduced by mapping errors to status codes inline,
// preserving the original asymmetric behavior matrix:
//   - bad id            -> 400
//   - not found (GetByID) -> 404
//   - delete missing     -> 500
//   - name not found     -> 200 + null body
//   - Create             -> 200 (not 201)
//   - Update/Delete      -> 200
package http

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"migrated-app/internal/apperr"
	"migrated-app/internal/model"
	"migrated-app/internal/smartcontact/service"
)

// UserHandler exposes the user CRUD HTTP endpoints, delegating all business
// logic to the UserService. It is the Go equivalent of the Spring
// UserController.
type UserHandler struct {
	svc    *service.UserService
	logger zerolog.Logger
}

// NewUserHandler constructs a UserHandler with the given service and logger.
//
// MIGRATION_NOTE: Replaces Spring @Autowired field injection on
// UserController.userService with explicit constructor injection.
func NewUserHandler(svc *service.UserService, logger zerolog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// RegisterRoutes registers all six user routes on the provided chi router at
// exactly the paths and methods used by the original UserController.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/save_user_data", h.SaveUser)
	r.Get("/get_user_data", h.FetchUserList)
	r.Get("/get_user_data/{id}", h.FetchUserByID)
	r.Delete("/delete_user_data/{id}", h.DeleteUser)
	r.Put("/update_user_data/{id}", h.UpdateUser)
	r.Get("/get_user_name/name/{name}", h.GetUserNameByName)
}

// parseID parses the {id} path variable as an int, enforcing the int32 range
// to match the source's Java int semantics. Returns false on any failure so
// callers can respond 400 Bad Request.
func parseID(s string) (int, bool) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, false
	}
	return int(v), true
}

// SaveUser handles POST /save_user_data.
//
// It is the only validated handler: the decoded User body is validated
// (required fields present, non-whitespace name) before being persisted.
// On success it returns 200 with a plain success message.
func (h *UserHandler) SaveUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("inside the saveUser of UserController")

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeValidationError(w, "invalid request body")
		return
	}

	if strings.TrimSpace(user.GetName()) == "" {
		writeValidationError(w, "name is required")
		return
	}

	if err := h.svc.SaveUser(r.Context(), &user); err != nil {
		h.handleError(w, err)
		return
	}

	writeString(w, http.StatusOK, "User data saved successfully!")
}

// FetchUserList handles GET /get_user_data, returning the full list of users
// as JSON with status 200.
func (h *UserHandler) FetchUserList(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("inside the fetchUserList of UserController")

	users, err := h.svc.FetchUserList(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. A non-integer id yields 400,
// a missing user yields 404, and success yields 200 with the user as JSON.
func (h *UserHandler) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(chi.URLParam(r, "id"))
	if !ok {
		writeString(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.svc.FetchUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) || errors.Is(err, service.ErrUserNotFound) {
			writeString(w, http.StatusNotFound, "user not found")
			return
		}
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /delete_user_data/{id}. A non-integer id yields
// 400; a delete of a missing user propagates as 500 (matching the source,
// which had no not-found handling on delete); success yields 200.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(chi.URLParam(r, "id"))
	if !ok {
		writeString(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		// MIGRATION_NOTE: Source deleteUser had no @ControllerAdvice mapping
		// for a missing row, so the failure surfaced as a 500. Preserve that.
		writeString(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeString(w, http.StatusOK, "user data deleted Successfully")
}

// UpdateUser handles PUT /update_user_data/{id}. Unlike SaveUser, the body is
// NOT validated. A non-integer id yields 400. On success it echoes the
// submitted User back (not the persisted entity) with status 200.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(chi.URLParam(r, "id"))
	if !ok {
		writeString(w, http.StatusBadRequest, "invalid id")
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeString(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.UpdateUser(r.Context(), id, &user); err != nil {
		h.handleError(w, err)
		return
	}
	// Echo the submitted payload back, matching the source behavior.
	writeJSON(w, http.StatusOK, &user)
}

// GetUserNameByName handles GET /get_user_name/name/{name}. It looks up a user
// by the single-segment {name} path variable. A not-found lookup returns 200
// with a null body, matching the source (which returned null without error).
func (h *UserHandler) GetUserNameByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	user, err := h.svc.GetUserNameByName(r.Context(), name)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) || errors.Is(err, service.ErrUserNotFound) {
			// MIGRATION_NOTE: Source returned null (200) when not found.
			writeJSON(w, http.StatusOK, nil)
			return
		}
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// handleError maps an unexpected error to a 500 response. Specific, expected
// errors are handled inline by each handler to preserve the status matrix.
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	h.logger.Error().Err(err).Msg("user handler error")
	writeString(w, http.StatusInternalServerError, "internal server error")
}

// writeValidationError emulates the Spring default @Valid failure response
// shape (timestamp-less simplified JSON error object).
func writeValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": http.StatusBadRequest,
		"errors": []string{message},
	})
}

// writeString writes a plain-text response with the given status code.
func writeString(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

I'll finish with this file at the agreed path `internal/transport/http/handler.go`.