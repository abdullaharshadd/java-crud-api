package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"migrated-app/internal/user"
)

// Response bodies returned verbatim by the source controller.
const (
	// SaveSuccessMessage is the plain-text body returned by POST /save_user_data.
	SaveSuccessMessage = "User data saved successfully!"
	// DeleteSuccessMessage is the plain-text body returned by DELETE /delete_user_data/{id}.
	DeleteSuccessMessage = "user data deleted Successfully"
)

const (
	contentTypeText = "text/plain;charset=UTF-8"
	contentTypeJSON = "application/json"
)

// UserService is the consumer-side view of the user application service used
// by the HTTP layer. *user.ServiceImp satisfies it.
//
// MIGRATION_NOTE: the Java UserService interface is declared here, next to its
// consumer. No user.Service interface exists in the migrated user package.
type UserService interface {
	SaveUser(ctx context.Context, u *user.User) (*user.User, error)
	FetchUserList(ctx context.Context) ([]user.User, error)
	FetchUserByID(ctx context.Context, id int) (*user.User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUser(ctx context.Context, id int, u *user.User) error
	GetUserByName(ctx context.Context, name string) (*user.User, bool, error)
}

// Handler serves the User REST endpoints (source:
// com.smartContact.Controller.UserController).
type Handler struct {
	svc UserService
}

// NewHandler returns a Handler backed by svc. It replaces @Autowired field
// injection with explicit constructor injection.
func NewHandler(svc UserService) *Handler {
	return &Handler{svc: svc}
}

// Routes returns an http.Handler with every UserController endpoint
// registered. One trailing slash is stripped before routing (Spring's default
// trailing-slash matching). Unknown paths get Spring's default 404 JSON, and a
// wrong method gets a 405 with an Allow header and an empty body.
//
// Wrap the result with Recoverer when mounting it in cmd/server/router.go.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/save_user_data", methods(map[string]http.HandlerFunc{
		http.MethodPost: h.saveUser,
	}))
	mux.Handle("/get_user_data", methods(map[string]http.HandlerFunc{
		http.MethodGet: h.fetchUserList,
	}))
	mux.Handle("/get_user_data/", pathParam("/get_user_data/", methods(map[string]http.HandlerFunc{
		http.MethodGet: h.fetchUserByID,
	})))
	mux.Handle("/delete_user_data/", pathParam("/delete_user_data/", methods(map[string]http.HandlerFunc{
		http.MethodDelete: h.deleteUser,
	})))
	mux.Handle("/update_user_data/", pathParam("/update_user_data/", methods(map[string]http.HandlerFunc{
		http.MethodPut: h.updateUser,
	})))
	mux.Handle("/get_user_name/name/", pathParam("/get_user_name/name/", methods(map[string]http.HandlerFunc{
		http.MethodGet: h.getUserByName,
	})))
	mux.Handle("/", NotFoundHandler())

	return stripTrailingSlash(mux)
}

// saveUser handles POST /save_user_data.
func (h *Handler) saveUser(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("inside the saveUser of UserController ")

	u, ok := decodeUser(w, r)
	if !ok {
		return
	}
	// MIGRATION_NOTE: @Valid -> explicit validation. Failures return 400 with
	// an empty body, as the source's ResponseEntityExceptionHandler did.
	if v, isValidator := any(u).(interface{ Validate() error }); isValidator {
		if err := v.Validate(); err != nil {
			WriteEmpty(w, http.StatusBadRequest)
			return
		}
	}
	if _, err := h.svc.SaveUser(r.Context(), u); err != nil {
		HandleError(w, r, err)
		return
	}
	writeText(w, SaveSuccessMessage)
}

// fetchUserList handles GET /get_user_data.
func (h *Handler) fetchUserList(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("inside the fetchUserList of UserController ")

	users, err := h.svc.FetchUserList(r.Context())
	if err != nil {
		HandleError(w, r, err)
		return
	}
	if users == nil {
		users = []user.User{} // serialise as [] rather than null
	}
	writeJSONValue(w, r, users)
}

// fetchUserByID handles GET /get_user_data/{id}.
func (h *Handler) fetchUserByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	u, err := h.svc.FetchUserByID(r.Context(), id)
	if err != nil {
		HandleError(w, r, err) // not found -> 404 ErrorMessage
		return
	}
	writeJSONValue(w, r, u)
}

// deleteUser handles DELETE /delete_user_data/{id}.
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		// MIGRATION_NOTE: a missing id was EmptyResultDataAccessException in
		// Spring Data 2.x, which was not handled by the advice -> default 500.
		if errors.Is(err, user.ErrNotFound) || errors.Is(err, user.ErrEmptyResult) {
			WriteSpringError(w, r, http.StatusInternalServerError)
			return
		}
		HandleError(w, r, err)
		return
	}
	writeText(w, DeleteSuccessMessage)
}

// updateUser handles PUT /update_user_data/{id}. It echoes back the request
// body (with id set by the service) rather than the persisted entity.
// No bean validation is applied, matching the source (no @Valid).
func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	u, ok := decodeUser(w, r)
	if !ok {
		return
	}
	if err := h.svc.UpdateUser(r.Context(), id, u); err != nil {
		HandleError(w, r, err)
		return
	}
	writeJSONValue(w, r, u)
}

// getUserByName handles GET /get_user_name/name/{name}.
//
// MIGRATION_NOTE: source getUserNameByName returned a nullable User. Spring
// writes a 200 with an empty body for a null @ResponseBody, reproduced here.
// More than one match surfaces as an error from the repository -> 500.
func (h *Handler) getUserByName(w http.ResponseWriter, r *http.Request) {
	name := pathValue(r)
	u, found, err := h.svc.GetUserByName(r.Context(), name)
	if err != nil {
		HandleError(w, r, err)
		return
	}
	if !found || u == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	writeJSONValue(w, r, u)
}

// --- helpers ---------------------------------------------------------------

type pathValueKey struct{}

// pathParam extracts the single path segment after prefix and stores it in
// the request context. Paths with extra segments are unknown routes (404).
func pathParam(prefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, prefix)
		if rest == "" || strings.Contains(rest, "/") {
			WriteSpringError(w, r, http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), pathValueKey{}, rest)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func pathValue(r *http.Request) string {
	v, _ := r.Context().Value(pathValueKey{}).(string)
	return v
}

// methods dispatches on the request method. HEAD is served by the GET handler,
// OPTIONS returns 200 with Allow, and any other method gets a 405.
func methods(handlers map[string]http.HandlerFunc) http.Handler {
	allowed := make([]string, 0, len(handlers)+2)
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete} {
		if m == http.MethodHead {
			if _, ok := handlers[http.MethodGet]; ok {
				allowed = append(allowed, m)
			}
			continue
		}
		if _, ok := handlers[m]; ok {
			allowed = append(allowed, m)
		}
	}
	allowed = append(allowed, http.MethodOptions)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		if method == http.MethodHead {
			method = http.MethodGet
		}
		if hf, ok := handlers[method]; ok {
			hf(w, r)
			return
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			w.WriteHeader(http.StatusOK)
			return
		}
		WriteMethodNotAllowed(w, allowed...)
	})
}

// stripTrailingSlash removes a single trailing slash from the request path.
func stripTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if len(p) > 1 && strings.HasSuffix(p, "/") {
			r2 := r.Clone(r.Context())
			r2.URL.Path = strings.TrimSuffix(p, "/")
			if r2.URL.RawPath != "" {
				r2.URL.RawPath = strings.TrimSuffix(r2.URL.RawPath, "/")
			}
			r = r2
		}
		next.ServeHTTP(w, r)
	})
}

// parseID parses the {id} path segment as a 32-bit int. On failure it writes
// a 400 with an empty body (MethodArgumentTypeMismatchException).
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.ParseInt(pathValue(r), 10, 32)
	if err != nil {
		WriteEmpty(w, http.StatusBadRequest)
		return 0, false
	}
	return int(id), true
}

// decodeUser decodes the JSON request body into a User. Non-JSON content types
// get a 415, and missing or malformed bodies get a 400, both with empty bodies.
func decodeUser(w http.ResponseWriter, r *http.Request) (*user.User, bool) {
	if ct := r.Header.Get("Content-Type"); ct != "" {
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil || !(mt == contentTypeJSON || strings.HasSuffix(mt, "+json")) {
			WriteEmpty(w, http.StatusUnsupportedMediaType)
			return nil, false
		}
	}
	var u *user.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil || u == nil {
		WriteEmpty(w, http.StatusBadRequest)
		return nil, false
	}
	return u, true
}

func writeText(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", contentTypeText)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// writeJSONValue writes v as a 200 JSON response without a trailing newline,
// matching Jackson's output.
func writeJSONValue(w http.ResponseWriter, r *http.Request, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("failed to encode JSON response")
		WriteSpringError(w, r, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
