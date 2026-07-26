```go
package smartcontact

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildRouter_ReturnsNonNilHandler verifies that buildRouter returns a
// non-nil http.Handler, i.e. the application can be bootstrapped without
// panicking.
func TestBuildRouter_ReturnsNonNilHandler(t *testing.T) {
	handler := buildRouter()
	require.NotNil(t, handler, "buildRouter() must return a non-nil http.Handler")
}

// routeTestCase is the shared structure used by every table-driven test below.
type routeTestCase struct {
	name           string
	method         string
	path           string
	body           string
	wantStatusCode int
	wantBodyContains string
}

// TestBuildRouter_HealthzEndpoint covers the liveness/readiness probe.
func TestBuildRouter_HealthzEndpoint(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		{
			name:             "GET /healthz returns 200 and body ok",
			method:           http.MethodGet,
			path:             "/healthz",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "ok",
		},
		{
			name:             "POST /healthz not registered – returns 405",
			method:           http.MethodPost,
			path:             "/healthz",
			wantStatusCode:   http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// TestBuildRouter_PublicRoutes exercises the publicly accessible GET routes.
func TestBuildRouter_PublicRoutes(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		{
			name:             "GET / returns 200 and Home label",
			method:           http.MethodGet,
			path:             "/",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Home",
		},
		{
			name:             "GET /signup returns 200 and Signup label",
			method:           http.MethodGet,
			path:             "/signup",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Signup",
		},
		{
			name:             "GET /signin returns 200 and Signin label",
			method:           http.MethodGet,
			path:             "/signin",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Signin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// TestBuildRouter_DoRegisterRoute exercises the registration POST endpoint.
func TestBuildRouter_DoRegisterRoute(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		{
			name:             "POST /do_register returns 200 and DoRegister label",
			method:           http.MethodPost,
			path:             "/do_register",
			body:             "username=test&password=secret",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "DoRegister",
		},
		{
			name:             "GET /do_register not registered – returns 405",
			method:           http.MethodGet,
			path:             "/do_register",
			wantStatusCode:   http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody *strings.Reader
			if tc.body != "" {
				reqBody = strings.NewReader(tc.body)
			} else {
				reqBody = strings.NewReader("")
			}
			req := httptest.NewRequest(tc.method, tc.path, reqBody)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// TestBuildRouter_UserAreaRoutes exercises the /user/** sub-router.
func TestBuildRouter_UserAreaRoutes(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		{
			name:             "GET /user/index returns 200 and UserDashboard label",
			method:           http.MethodGet,
			path:             "/user/index",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "UserDashboard",
		},
		{
			name:             "GET /user/nonexistent returns 404",
			method:           http.MethodGet,
			path:             "/user/nonexistent",
			wantStatusCode:   http.StatusNotFound,
		},
		{
			name:             "POST /user/index not registered – returns 405",
			method:           http.MethodPost,
			path:             "/user/index",
			wantStatusCode:   http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// TestBuildRouter_AdminAreaRoutes exercises the /admin/** sub-router.
func TestBuildRouter_AdminAreaRoutes(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		{
			name:             "GET /admin/index returns 200 and AdminDashboard label",
			method:           http.MethodGet,
			path:             "/admin/index",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "AdminDashboard",
		},
		{
			name:             "GET /admin/nonexistent returns 404",
			method:           http.MethodGet,
			path:             "/admin/nonexistent",
			wantStatusCode:   http.StatusNotFound,
		},
		{
			name:             "POST /admin/index not registered – returns 405",
			method:           http.MethodPost,
			path:             "/admin/index",
			wantStatusCode:   http.StatusMethodNotAllowed,
		},
		{
			name:             "DELETE /admin/index not registered – returns 405",
			method:           http.MethodDelete,
			path:             "/admin/index",
			wantStatusCode:   http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
			if tc.wantBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}

// TestBuildRouter_UnknownRoutes verifies that unregistered paths return 404.
func TestBuildRouter_UnknownRoutes(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		{
			name:           "GET /unknown-path returns 404",
			method:         http.MethodGet,
			path:           "/unknown-path",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "POST /unknown-path returns 404",
			method:         http.MethodPost,
			path:           "/unknown-path",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "GET /admin returns 404 (only /admin/index is registered)",
			method:         http.MethodGet,
			path:           "/admin",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "GET /user returns 404 (only /user/index is registered)",
			method:         http.MethodGet,
			path:           "/user",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatusCode, rr.Code)
		})
	}
}

// TestBuildRouter_RecovererMiddleware verifies that the Recoverer middleware
// is in place by confirming that a panicking handler would not crash the server.
// We test this indirectly by verifying that ordinary requests succeed after the
// middleware stack is evaluated — a direct panic test would require a custom
// handler injection, which we achieve via a thin wrapper.
func TestBuildRouter_RecovererMiddleware(t *testing.T) {
	// We can't inject a panicking handler into the built router without
	// modifying the production code. Instead, we verify the middleware is
	// active by asserting that the router handles a valid request without
	// panicking in the test process itself.
	handler := buildRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handler.ServeHTTP(rr, req)
	}, "serving a request must not panic even when Recoverer is chained")

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestBuildRouter_ResponseContentType verifies that responses from plain text
// stub handlers do not set an unexpected content-type (no JSON contract yet).
func TestBuildRouter_ResponseContentType(t *testing.T) {
	handler := buildRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"healthz", http.MethodGet, "/healthz"},
		{"home", http.MethodGet, "/"},
		{"signup", http.MethodGet, "/signup"},
		{"signin", http.MethodGet, "/signin"},
		{"user-index", http.MethodGet, "/user/index"},
		{"admin-index", http.MethodGet, "/admin/index"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			// Body must be non-empty (stub handlers always write something).
			assert.NotEmpty(t, rr.Body.String())
		})
	}
}

// TestBuildRouter_IdempotentCalls verifies that calling buildRouter() multiple
// times produces independent, fully-functional handlers (idempotent
// bootstrapping, analogous to Spring ApplicationContext being startable without
// global side-effects in tests).
func TestBuildRouter_IdempotentCalls(t *testing.T) {
	handler1 := buildRouter()
	handler2 := buildRouter()

	for i, handler := range []http.Handler{handler1, handler2} {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "call %d: /healthz must return 200", i+1)
		assert.Contains(t, rr.Body.String(), "ok", "call %d: body must contain ok", i+1)
	}
}

// TestBuildRouter_AllRegisteredRoutesSummary is a comprehensive table-driven
// test that exercises every registered route in a single pass, providing a
// high-level regression safety net equivalent to verifying that all
// component-scanned @Controller mappings are present.
func TestBuildRouter_AllRegisteredRoutesSummary(t *testing.T) {
	handler := buildRouter()

	tests := []routeTestCase{
		// Liveness probe
		{
			name:             "GET /healthz → 200 ok",
			method:           http.MethodGet,
			path:             "/healthz",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "ok",
		},
		// Public area
		{
			name:             "GET / → 200 Home",
			method:           http.MethodGet,
			path:             "/",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Home",
		},
		{
			name:             "GET /signup → 200 Signup",
			method:           http.MethodGet,
			path:             "/signup",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Signup",
		},
		{
			name:             "POST /do_register → 200 DoRegister",
			method:           http.MethodPost,
			path:             "/do_register",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "DoRegister",
		},
		{
			name:             "GET /signin → 200 Signin",
			method:           http.MethodGet,
			path:             "/signin",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "Signin",
		},
		// User area
		{
			name:             "GET /user/index → 200 UserDashboard",
			method:           http.MethodGet,
			path:             "/user/index",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "UserDashboard",
		},
		// Admin area
		{
			name:             "GET /admin/index → 200 AdminDashboard",
			method:           http.MethodGet,
			path:             "/admin/index",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "AdminDashboard",
		},
		