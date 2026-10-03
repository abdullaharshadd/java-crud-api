package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"migrated-app/internal/model"
	"migrated-app/internal/service"
)

// Fixed plain-text responses produced by the source UserController.
const (
	saveUserSuccessMessage   = "User data saved successfully!"
	deleteUserSuccessMessage = "user data deleted Successfully"
	textPlainUTF8            = "text/plain;charset=UTF-8"
	applicationJSON          = "application/json"
)

// UserHandler exposes the User CRUD endpoints over HTTP. It is the Go
// counterpart of the source's @RestController UserController and delegates
// all business logic to a service.UserService.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler builds a UserHandler backed by svc (replaces @Autowired
// field injection).
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register mounts every UserController route on mux using Go 1.22 method
// patterns. GET patterns also answer HEAD, like Spring MVC.
func (h *UserHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /save_user_data", h.SaveUser)
	mux.HandleFunc("GET /get_user_data", h.FetchUserList)
	mux.HandleFunc("GET /get_user_data/{id}", h.FetchUserByID)
	mux.HandleFunc("DELETE /delete_user_data/{id}", h.DeleteUser)
	mux.HandleFunc("PUT /update_user_data/{id}", h.UpdateUser)
	mux.HandleFunc("GET /get_user_name/name/{name}", h.GetUserNameByName)
}

// NewRouter builds the complete HTTP handler for the user API, including the
// global Spring-compatible error handling chain.
func NewRouter(svc service.UserService) http.Handler {
	mux := http.NewServeMux()
	NewUserHandler(svc).Register(mux)
	return WrapMux(mux)
}

// SaveUser handles POST /save_user_data: validates the JSON body, saves the
// user and replies with a fixed plain-text message.
func (h *UserHandler) SaveUser(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("inside the saveUser of UserController ")
	u, err := decodeUserBody(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if err := u.Validate(); err != nil {
		WriteError(w, r, fmt.Errorf("save user: %w: %w", ErrValidation, err))
		return
	}
	if _, err := h.svc.SaveUser(r.Context(), u); err != nil {
		WriteError(w, r, err)
		return
	}
	writeText(w, http.StatusOK, saveUserSuccessMessage)
}

// FetchUserList handles GET /get_user_data and returns all users as a JSON
// array ([] when empty).
func (h *UserHandler) FetchUserList(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("inside the fetchUserList of UserController ")
	users, err := h.svc.FetchUserList(r.Context())
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if users == nil {
		users = []model.User{}
	}
	respondJSON(w, r, http.StatusOK, users)
}

// FetchUserByID handles GET /get_user_data/{id}. A miss yields the 404
// ErrorMessage body via WriteError.
func (h *UserHandler) FetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	u, err := h.svc.FetchUserByID(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	respondJSON(w, r, http.StatusOK, u)
}

// DeleteUser handles DELETE /delete_user_data/{id}. Deleting a missing id
// surfaces the repository error as a 500, as in the source.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		WriteError(w, r, err)
		return
	}
	writeText(w, http.StatusOK, deleteUserSuccessMessage)
}

// UpdateUser handles PUT /update_user_data/{id}. The body is NOT validated
// (the source has no @Valid); the request user, with the path id applied by
// the service, is echoed back.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	u, err := decodeUserBody(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if _, err := h.svc.UpdateUser(r.Context(), id, u); err != nil {
		WriteError(w, r, err)
		return
	}
	respondJSON(w, r, http.StatusOK, u)
}

// GetUserNameByName handles GET /get_user_name/name/{name} and returns the
// full matching User. No match yields 200 with an empty body and no
// Content-Type (Spring writing a null @ResponseBody).
func (h *UserHandler) GetUserNameByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	u, err := h.svc.GetUserNameByName(r.Context(), name)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	respondJSON(w, r, http.StatusOK, u)
}

// pathID parses the {id} path variable as a 32-bit int, matching Spring's
// String-to-int conversion (failure → 400).
func pathID(r *http.Request) (int32, error) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("path variable id %q: %w: %w", raw, ErrInvalidPathParam, err)
	}
	return int32(id), nil
}

// decodeUserBody enforces a JSON Content-Type (415 otherwise) and decodes a
// non-null User, ignoring unknown properties like Jackson in Spring Boot.
func decodeUserBody(r *http.Request) (*model.User, error) {
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		return nil, fmt.Errorf("content type %q: %w", r.Header.Get("Content-Type"), ErrUnsupportedMediaType)
	}
	var u *model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("required request body is missing: %w", ErrMalformedBody)
		}
		return nil, fmt.Errorf("decode user: %w: %w", ErrMalformedBody, err)
	}
	if u == nil {
		// MIGRATION_NOTE: a literal JSON null body fails Spring's required
		// @RequestBody check (HttpMessageNotReadableException → 400).
		return nil, fmt.Errorf("request body is null: %w", ErrMalformedBody)
	}
	return u, nil
}

// isJSONContentType reports whether ct is application/json or application/*+json,
// the media types Spring's Jackson converter accepts.
func isJSONContentType(ct string) bool {
	if ct == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return mt == applicationJSON || (strings.HasPrefix(mt, "application/") && strings.HasSuffix(mt, "+json"))
}

// writeText writes a text/plain;charset=UTF-8 response.
func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", textPlainUTF8)
	w.WriteHeader(status)
	if _, err := io.WriteString(w, body); err != nil {
		log.Debug().Err(err).Msg("write text response")
	}
}

// respondJSON encodes v as a JSON success response; encoding failures become
// a 500 Boot error.
func respondJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		WriteError(w, r, fmt.Errorf("encode response: %w", err))
		return
	}
	w.Header().Set("Content-Type", applicationJSON)
	w.WriteHeader(status)
	if _, err := w.Write(payload); err != nil {
		log.Debug().Err(err).Msg("write json response")
	}
}
