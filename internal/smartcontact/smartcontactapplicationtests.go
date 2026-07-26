package smartcontact

// MIGRATION_NOTE: The Java source SmartContactApplicationTests.java was the
// default Spring Boot generated test class. Its single implicit test
// (contextLoads) simply verifies that the Spring application context can be
// bootstrapped successfully. There is no direct Go equivalent to Spring's
// dependency-injection container, so a literal port would be meaningless.
//
// Per the migration debate, this smoke test is re-imagined as an httptest-based
// integration test that boots the real HTTP handler (the Go analogue of
// "loading the context") and asserts the API's status-code contract. The most
// interesting divergences to lock down are:
//
//   - DELETE of a missing id -> 500 (mirrors the Java NotFoundError path that
//     was not mapped to 404 by the original @ControllerAdvice for delete).
//   - Malformed JSON body -> 400 (decoder error shape).
//   - Validation failure on an otherwise well-formed body -> 400 (validator
//     error shape).
//
// This file is deliberately a non-test .go source (per the requested target
// path internal/smartcontact/smartcontactapplicationtests.go). It exposes
// NewTestServer, a reusable helper that wires an in-memory server together and
// registers every HTTP route. The corresponding _test.go file should import
// this helper and exercise the divergence cases above with the standard
// testing package (table-driven).
//
// REQUIRES MANUAL REVIEW: the concrete UserService / repository wiring is not
// present among the already-migrated symbols. The handlers below therefore
// depend on a small UserStore interface that must be satisfied by the real
// service once it is migrated. An in-memory implementation is provided for the
// smoke test.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	smartErr "internal/smartcontact/error"
	"internal/smartcontact/model"
)

// UserStore abstracts the persistence operations the HTTP layer depends on.
//
// MIGRATION_NOTE: This replaces the Spring-injected UserService/UserRepository
// beans. Using an interface (rather than a concrete type) keeps the handlers
// testable and lets the smoke test substitute an in-memory implementation.
type UserStore interface {
	// Get returns the user with the given id, or a NotFoundError if absent.
	Get(ctx context.Context, id int64) (model.User, error)
	// Create persists a new user and returns the stored representation.
	Create(ctx context.Context, u model.User) (model.User, error)
	// Delete removes the user with the given id, or returns a NotFoundError.
	Delete(ctx context.Context, id int64) error
}

// Handler bundles the HTTP handlers for the SmartContact user API together
// with their backing store.
type Handler struct {
	store UserStore
}

// NewHandler constructs a Handler backed by the supplied UserStore.
func NewHandler(store UserStore) *Handler {
	return &Handler{store: store}
}

// Routes returns an http.Handler with every SmartContact route registered at
// its exact path. This is the Go analogue of Spring's @RequestMapping wiring.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", h.handleUsersCollection)
	mux.HandleFunc("/users/", h.handleUsersItem)
	return mux
}

func (h *Handler) handleUsersCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleUsersItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	id, convErr := strconv.ParseInt(idStr, 10, 64)
	if convErr != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getUser(w, r, id)
	case http.MethodDelete:
		h.deleteUser(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var u model.User
	// Malformed JSON -> 400, distinct in shape from validation failures.
	if decodeErr := json.NewDecoder(r.Body).Decode(&u); decodeErr != nil {
		writeError(w, http.StatusBadRequest, "malformed request body")
		return
	}
	// Bean-Validation equivalent -> 400.
	if validateErr := u.Validate(); validateErr != nil {
		writeError(w, http.StatusBadRequest, validateErr.Error())
		return
	}

	created, createErr := h.store.Create(r.Context(), u)
	if createErr != nil {
		writeError(w, http.StatusInternalServerError, createErr.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request, id int64) {
	u, getErr := h.store.Get(r.Context(), id)
	if getErr != nil {
		if errors.Is(getErr, smartErr.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, model.NewErrorMessage(http.StatusNotFound, getErr.Error()))
			return
		}
		writeError(w, http.StatusInternalServerError, getErr.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request, id int64) {
	// MIGRATION_NOTE: preserving the Java divergence — deleting a missing id
	// surfaces as 500 rather than 404 because the original delete path did not
	// translate NotFoundError into a 404 response.
	if delErr := h.store.Delete(r.Context(), id); delErr != nil {
		writeError(w, http.StatusInternalServerError, delErr.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.NewErrorMessage(status, msg))
}

// inMemoryUserStore is a lightweight UserStore used by the smoke test to stand
// in for the real persistence layer. It is safe for concurrent use.
type inMemoryUserStore struct {
	mu     sync.Mutex
	nextID int64
	users  map[int64]model.User
}

// newInMemoryUserStore constructs an empty in-memory user store.
func newInMemoryUserStore() *inMemoryUserStore {
	return &inMemoryUserStore{
		nextID: 1,
		users:  make(map[int64]model.User),
	}
}

// Get implements UserStore.
func (s *inMemoryUserStore) Get(_ context.Context, id int64) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return model.User{}, smartErr.NewUserNotFound(id)
	}
	return u, nil
}

// Create implements UserStore.
func (s *inMemoryUserStore) Create(_ context.Context, u model.User) (model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	s.users[id] = u
	return u, nil
}

// Delete implements UserStore.
func (s *inMemoryUserStore) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return smartErr.NewUserNotFound(id)
	}
	delete(s.users, id)
	return nil
}

// NewTestServer builds an http.Handler wired to a fresh in-memory store,
// suitable for driving with net/http/httptest in the accompanying _test.go.
// It is the Go equivalent of Spring Boot's @SpringBootTest "contextLoads"
// smoke check: if this returns a non-nil handler, the API graph assembled
// successfully.
func NewTestServer() http.Handler {
	return NewHandler(newInMemoryUserStore()).Routes()
}
