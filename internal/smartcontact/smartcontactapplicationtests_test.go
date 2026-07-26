```go
package smartcontact_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/smartcontact"
	smartErr "internal/smartcontact/error"
	"internal/smartcontact/model"
)

// ----------------------------------------------------------------------------
// Mock / stub UserStore implementations
// ----------------------------------------------------------------------------

// errStore always returns a configurable error for every operation.
type errStore struct {
	getErr    error
	createErr error
	deleteErr error
}

func (e *errStore) Get(_ context.Context, _ int64) (model.User, error) {
	return model.User{}, e.getErr
}
func (e *errStore) Create(_ context.Context, u model.User) (model.User, error) {
	return u, e.createErr
}
func (e *errStore) Delete(_ context.Context, _ int64) error {
	return e.deleteErr
}

// seedStore holds a fixed set of users and mirrors inMemoryUserStore semantics
// without the package-private visibility constraint.
type seedStore struct {
	users  map[int64]model.User
	nextID int64
}

func newSeedStore(initial map[int64]model.User) *seedStore {
	users := make(map[int64]model.User)
	maxID := int64(0)
	for k, v := range initial {
		users[k] = v
		if k > maxID {
			maxID = k
		}
	}
	return &seedStore{users: users, nextID: maxID + 1}
}

func (s *seedStore) Get(_ context.Context, id int64) (model.User, error) {
	u, ok := s.users[id]
	if !ok {
		return model.User{}, smartErr.NewUserNotFound(id)
	}
	return u, nil
}
func (s *seedStore) Create(_ context.Context, u model.User) (model.User, error) {
	id := s.nextID
	s.nextID++
	u2 := u
	s.users[id] = u2
	return u2, nil
}
func (s *seedStore) Delete(_ context.Context, id int64) error {
	if _, ok := s.users[id]; !ok {
		return smartErr.NewUserNotFound(id)
	}
	delete(s.users, id)
	return nil
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func buildServer(store smartcontact.UserStore) http.Handler {
	return smartcontact.NewHandler(store).Routes()
}

func doRequest(t *testing.T, handler http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func marshalUser(t *testing.T, u model.User) []byte {
	t.Helper()
	b, err := json.Marshal(u)
	require.NoError(t, err)
	return b
}

// ----------------------------------------------------------------------------
// TestContextLoads – Go analogue of Spring Boot's "contextLoads" smoke test.
// If NewTestServer returns a non-nil handler the dependency graph assembled
// correctly.
// ----------------------------------------------------------------------------

func TestContextLoads(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "application context loads successfully – NewTestServer returns non-nil handler",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := smartcontact.NewTestServer()
			assert.NotNil(t, handler, "NewTestServer must return a non-nil http.Handler")
		})
	}
}

// ----------------------------------------------------------------------------
// TestNewTestServer_Smoke – makes a real request to the wired handler to
// confirm the full request/response cycle works (analogue of "no bean wiring
// error" at start-up).
// ----------------------------------------------------------------------------

func TestNewTestServer_Smoke(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		wantStatusCode int
	}{
		{
			name:           "POST /users with valid body returns 201",
			method:         http.MethodPost,
			path:           "/users",
			body:           marshalUser(t, model.User{Name: "Alice", Email: "alice@example.com"}),
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "GET /users/9999 (missing) returns 404",
			method:         http.MethodGet,
			path:           "/users/9999",
			body:           nil,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "DELETE /users/9999 (missing) returns 500 – Java divergence preserved",
			method:         http.MethodDelete,
			path:           "/users/9999",
			body:           nil,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	handler := smartcontact.NewTestServer()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := doRequest(t, handler, tc.method, tc.path, tc.body)
			assert.Equal(t, tc.wantStatusCode, rr.Code)
		})
	}
}

// ----------------------------------------------------------------------------
// TestCreateUser – POST /users
// ----------------------------------------------------------------------------

func TestCreateUser(t *testing.T) {
	validUser := model.User{Name: "Bob", Email: "bob@example.com"}
	validBody := marshalUser(t, validUser)

	tests := []struct {
		name           string
		store          smartcontact.UserStore
		body           []byte
		rawBody        string // used when we want to bypass json.Marshal
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:           "valid user body returns 201 Created",
			store:          newSeedStore(nil),
			body:           validBody,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:             "malformed JSON body returns 400",
			store:            newSeedStore(nil),
			rawBody:          `{"name": "Bad JSON"`,
			wantStatusCode:   http.StatusBadRequest,
			wantBodyContains: "malformed",
		},
		{
			name:             "validation failure returns 400",
			store:            newSeedStore(nil),
			body:             marshalUser(t, model.User{}), // empty → fails Validate()
			wantStatusCode:   http.StatusBadRequest,
		},
		{
			name:             "store Create error returns 500",
			store:            &errStore{createErr: errors.New("db unavailable")},
			body:             validBody,
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "db unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := buildServer(tc.store)
			var bodyBytes []byte
			if tc.rawBody != "" {
				bodyBytes = []byte(tc.rawBody)
			} else {
				bodyBytes = tc.body
			}
			rr := doRequest(t, handler, http.MethodPost, "/users", bodyBytes)
			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// TestGetUser – GET /users/:id
// ----------------------------------------------------------------------------

func TestGetUser(t *testing.T) {
	existingUser := model.User{Name: "Carol", Email: "carol@example.com"}
	storeWithUser := newSeedStore(map[int64]model.User{1: existingUser})

	tests := []struct {
		name             string
		store            smartcontact.UserStore
		path             string
		wantStatusCode   int
		wantBodyContains string
	}{
		{
			name:           "existing user returns 200",
			store:          storeWithUser,
			path:           "/users/1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing user returns 404",
			store:          storeWithUser,
			path:           "/users/999",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:             "non-numeric id returns 400",
			store:            storeWithUser,
			path:             "/users/abc",
			wantStatusCode:   http.StatusBadRequest,
			wantBodyContains: "invalid user id",
		},
		{
			name:             "store Get unexpected error returns 500",
			store:            &errStore{getErr: errors.New("connection reset")},
			path:             "/users/1",
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "connection reset",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := buildServer(tc.store)
			rr := doRequest(t, handler, http.MethodGet, tc.path, nil)
			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// TestDeleteUser – DELETE /users/:id
// ----------------------------------------------------------------------------

func TestDeleteUser(t *testing.T) {
	existingUser := model.User{Name: "Dave", Email: "dave@example.com"}
	storeWithUser := func() *seedStore {
		return newSeedStore(map[int64]model.User{1: existingUser})
	}

	tests := []struct {
		name             string
		store            smartcontact.UserStore
		path             string
		wantStatusCode   int
		wantBodyContains string
	}{
		{
			name:           "existing user deleted returns 204",
			store:          storeWithUser(),
			path:           "/users/1",
			wantStatusCode: http.StatusNoContent,
		},
		{
			// MIGRATION_NOTE: Java divergence – missing id returns 500, not 404.
			name:             "missing user returns 500 (Java divergence)",
			store:            storeWithUser(),
			path:             "/users/999",
			wantStatusCode:   http.StatusInternalServerError,
		},
		{
			name:             "non-numeric id returns 400",
			store:            storeWithUser(),
			path:             "/users/xyz",
			wantStatusCode:   http.StatusBadRequest,
			wantBodyContains: "invalid user id",
		},
		{
			name:             "store Delete unexpected error returns 500",
			store:            &errStore{deleteErr: errors.New("disk full")},
			path:             "/users/1",
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "disk full",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := buildServer(tc.store)
			rr := doRequest(t, handler, http.MethodDelete, tc.path, nil)
			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// TestMethodNotAllowed – unsupported HTTP methods
// ----------------------------------------------------------------------------

func TestMethodNotAllowed(t *testing.T) {
	handler := smartcontact.NewTestServer()

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "GET /users returns 405",
			method:         http.MethodGet,
			path:           "/users",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT /users returns 405",
			method:         http.MethodPut,
			path:           "/users",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "PATCH /users/1 returns 405",
			method:         http.MethodPatch,
			path:           "/users/1",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST /users/1 returns 405",
			method:         http.MethodPost,
			path:           "/users/1",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := doRequest(t, handler, tc.method, tc.path, nil)
			assert.Equal(t, tc.wantStatusCode, rr.Code)
		})
	}
}

// ----------------------------------------------------------------------------
// TestResponseContentType – every non-204 response carries application/json
// ----------------------------------------------------------------------------

func TestResponseContentType(t *testing.T) {
	seededStore := newSeedStore(map[int64]model.User{
		1: {Name: "Eve", Email: "eve@example.com"},
	})
	handler := buildServer(seededStore)

	tests := []struct {
		name            string
		method          string
		path            string
		body            []byte
		wantContentType string
	}{
		{
			name:            "GET existing user has application/json content type",
			method:          http.MethodGet,
			path:            "/users/1",
			wantContentType: "application/json",
		},
		{
			name:            "GET missing user has application/json content type",
			method:          http.MethodGet,
			path:            "/users/999",
			wantContentType: "application/json",
		},
		{
			name:            "POST valid user has application/json content type",
			method:          http.MethodPost,
			path:            "/users",
			body:            marshalUser(t, model.User{Name: "Frank", Email: "frank@example.com"}),
			wantContentType: "application/json",
		},
		{
			name:            "POST malformed JSON has application/json content type",
			method:          http.MethodPost,
			path:            "/users",
			body:            []byte(`not-json`),
			wantContentType: "application/json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := doRequest(t, handler, tc.method, tc.path, tc.body)
			assert.Contains(t, rr.Header().Get("Content-Type"), tc.wantContentType)
		})
	}
}

// ----------------------------------------------------------------------------
// TestInMemoryUserStore – unit tests for the store returned by NewTestServer
// ----------------------------------------------------------------------------

func TestInMemoryUserStore_RoundTrip(t *testing.T) {
	// We drive the store indirectly via the