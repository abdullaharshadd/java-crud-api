```go
package smartcontact

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// mockRouteRegistrar is a test double for routeRegistrar.
type mockRouteRegistrar struct {
	registerCalled bool
	routes         map[string]http.HandlerFunc
}

func (m *mockRouteRegistrar) RegisterRoutes(r chi.Router) {
	m.registerCalled = true
	for pattern, handler := range m.routes {
		// Register each route; we use a closure to capture handler.
		h := handler
		r.Get(pattern, h)
	}
}

// failingRouteRegistrar simulates a registrar that panics during route
// registration (to exercise the Recoverer middleware path).
type failingRouteRegistrar struct{}

func (f *failingRouteRegistrar) RegisterRoutes(r chi.Router) {
	r.Get("/panic_route", func(w http.ResponseWriter, req *http.Request) {
		panic("simulated panic")
	})
}

// ---------------------------------------------------------------------------
// Tests for notWiredHandler
// ---------------------------------------------------------------------------

func TestNotWiredHandler(t *testing.T) {
	tests := []struct {
		name           string
		routeName      string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "returns 503 for SaveUser",
			routeName:      "SaveUser",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"SaveUser\" not wired: build the router with BuildRouterWith(controller)\n",
		},
		{
			name:           "returns 503 for FetchUserList",
			routeName:      "FetchUserList",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"FetchUserList\" not wired: build the router with BuildRouterWith(controller)\n",
		},
		{
			name:           "returns 503 for FetchUserByID",
			routeName:      "FetchUserByID",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"FetchUserByID\" not wired: build the router with BuildRouterWith(controller)\n",
		},
		{
			name:           "returns 503 for DeleteUser",
			routeName:      "DeleteUser",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"DeleteUser\" not wired: build the router with BuildRouterWith(controller)\n",
		},
		{
			name:           "returns 503 for UpdateUser",
			routeName:      "UpdateUser",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"UpdateUser\" not wired: build the router with BuildRouterWith(controller)\n",
		},
		{
			name:           "returns 503 for GetUserByName",
			routeName:      "GetUserByName",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"GetUserByName\" not wired: build the router with BuildRouterWith(controller)\n",
		},
		{
			name:           "returns 503 for arbitrary route name",
			routeName:      "ArbitraryRoute",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "route \"ArbitraryRoute\" not wired: build the router with BuildRouterWith(controller)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := notWiredHandler(tt.routeName)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for buildRouter (internal, no live controller)
// ---------------------------------------------------------------------------

func TestBuildRouter_Healthz(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "healthz returns 200 ok",
			method:         http.MethodGet,
			path:           "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok\n",
		},
	}

	handler := buildRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestBuildRouter_NotWiredRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST /save_user_data returns 503",
			method:         http.MethodPost,
			path:           "/save_user_data",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "GET /get_user_data returns 503",
			method:         http.MethodGet,
			path:           "/get_user_data",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "GET /get_user_data/{id} returns 503",
			method:         http.MethodGet,
			path:           "/get_user_data/42",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "DELETE /delete_user_data/{id} returns 503",
			method:         http.MethodDelete,
			path:           "/delete_user_data/42",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "PUT /update_user_data/{id} returns 503",
			method:         http.MethodPut,
			path:           "/update_user_data/42",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "GET /get_user_name/name/{name} returns 503",
			method:         http.MethodGet,
			path:           "/get_user_name/name/alice",
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	handler := buildRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestBuildRouter_NotWiredRoutes_BodyContainsRouteName(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		expectedRoute string
	}{
		{
			name:          "save_user_data body mentions SaveUser",
			method:        http.MethodPost,
			path:          "/save_user_data",
			expectedRoute: "SaveUser",
		},
		{
			name:          "get_user_data body mentions FetchUserList",
			method:        http.MethodGet,
			path:          "/get_user_data",
			expectedRoute: "FetchUserList",
		},
		{
			name:          "get_user_data/{id} body mentions FetchUserByID",
			method:        http.MethodGet,
			path:          "/get_user_data/1",
			expectedRoute: "FetchUserByID",
		},
		{
			name:          "delete_user_data/{id} body mentions DeleteUser",
			method:        http.MethodDelete,
			path:          "/delete_user_data/1",
			expectedRoute: "DeleteUser",
		},
		{
			name:          "update_user_data/{id} body mentions UpdateUser",
			method:        http.MethodPut,
			path:          "/update_user_data/1",
			expectedRoute: "UpdateUser",
		},
		{
			name:          "get_user_name/name/{name} body mentions GetUserByName",
			method:        http.MethodGet,
			path:          "/get_user_name/name/bob",
			expectedRoute: "GetUserByName",
		},
	}

	handler := buildRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Contains(t, w.Body.String(), tt.expectedRoute)
			assert.Contains(t, w.Body.String(), "not wired")
		})
	}
}

func TestBuildRouter_UnknownRoute(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "unknown path returns 404",
			method:         http.MethodGet,
			path:           "/unknown_path",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "root path returns 404",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusNotFound,
		},
	}

	handler := buildRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for BuildRouterWith
// ---------------------------------------------------------------------------

func TestBuildRouterWith_Healthz(t *testing.T) {
	tests := []struct {
		name           string
		controller     routeRegistrar
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "healthz with nil controller returns 200",
			controller:     nil,
			method:         http.MethodGet,
			path:           "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok\n",
		},
		{
			name:           "healthz with mock controller returns 200",
			controller:     &mockRouteRegistrar{routes: map[string]http.HandlerFunc{}},
			method:         http.MethodGet,
			path:           "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := BuildRouterWith(tt.controller)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestBuildRouterWith_NilController(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "nil controller: /save_user_data returns 404 (no routes registered)",
			method:         http.MethodPost,
			path:           "/save_user_data",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "nil controller: /get_user_data returns 404",
			method:         http.MethodGet,
			path:           "/get_user_data",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "nil controller: unknown path returns 404",
			method:         http.MethodGet,
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
		},
	}

	handler := BuildRouterWith(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestBuildRouterWith_ControllerRegistersRoutes(t *testing.T) {
	tests := []struct {
		name           string
		registrarRoutes map[string]http.HandlerFunc
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "registered route is reachable and returns handler response",
			registrarRoutes: map[string]http.HandlerFunc{
				"/get_user_data": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("user list"))
				},
			},
			method:         http.MethodGet,
			path:           "/get_user_data",
			expectedStatus: http.StatusOK,
			expectedBody:   "user list",
		},
		{
			name: "registered route returns custom 201",
			registrarRoutes: map[string]http.HandlerFunc{
				"/custom_route": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("created"))
				},
			},
			method:         http.MethodGet,
			path:           "/custom_route",
			expectedStatus: http.StatusCreated,
			expectedBody:   "created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRouteRegistrar{routes: tt.registrarRoutes}
			handler := BuildRouterWith(mock)

			assert.True(t, mock.registerCalled, "RegisterRoutes should have been called")

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestBuildRouterWith_RegisterRoutesCalled(t *testing.T) {
	tests := []struct {
		name               string
		controller         routeRegistrar
		expectRegisterCall bool
	}{
		{
			name:               "non-nil controller: RegisterRoutes is called",
			controller:         &mockRouteRegistrar{routes: map[string]http.HandlerFunc{}},
			expectRegisterCall: true,
		},
		{
			name:               "nil controller: RegisterRoutes is never called",
			controller:         nil,
			expectRegisterCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.controller == nil {
				// Just ensure BuildRouterWith does not panic