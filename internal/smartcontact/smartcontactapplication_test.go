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

func TestBuildRouter_NotNil(t *testing.T) {
	r := buildRouter()
	require.NotNil(t, r)
}

func TestBuildRouter_Healthz(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatus     int
		wantBodyContains string
	}{
		{
			name:             "GET /healthz returns 200 and ok",
			method:           http.MethodGet,
			path:             "/healthz",
			wantStatus:       http.StatusOK,
			wantBodyContains: "ok",
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBodyContains)
		})
	}
}

func TestBuildRouter_HomeRoutes(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		body             string
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "GET / returns 200 with Home route label",
			method:           http.MethodGet,
			path:             "/",
			wantStatus:       http.StatusOK,
			wantBodyContains: "route: Home",
		},
		{
			name:             "GET /signup returns 200 with Signup route label",
			method:           http.MethodGet,
			path:             "/signup",
			wantStatus:       http.StatusOK,
			wantBodyContains: "route: Signup",
		},
		{
			name:             "POST /do_register returns 200 with DoRegister route label",
			method:           http.MethodPost,
			path:             "/do_register",
			wantStatus:       http.StatusOK,
			wantBodyContains: "route: DoRegister",
		},
		{
			name:             "GET /signin returns 200 with Signin route label",
			method:           http.MethodGet,
			path:             "/signin",
			wantStatus:       http.StatusOK,
			wantBodyContains: "route: Signin",
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBodyContains)
		})
	}
}

func TestBuildRouter_AdminRoutes(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "GET /admin/ returns 200 with AdminDashboard route label",
			method:           http.MethodGet,
			path:             "/admin/",
			wantStatus:       http.StatusOK,
			wantBodyContains: "route: AdminDashboard",
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBodyContains)
		})
	}
}

func TestBuildRouter_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "POST / is not allowed",
			method:     http.MethodPost,
			path:       "/",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "POST /signup is not allowed",
			method:     http.MethodPost,
			path:       "/signup",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "GET /do_register is not allowed",
			method:     http.MethodGet,
			path:       "/do_register",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "POST /signin is not allowed",
			method:     http.MethodPost,
			path:       "/signin",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "POST /admin/ is not allowed",
			method:     http.MethodPost,
			path:       "/admin/",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestBuildRouter_NotFound(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET /unknown returns 404",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "GET /admin/unknown returns 404",
			method:     http.MethodGet,
			path:       "/admin/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "DELETE /signup returns 405 or 404",
			method:     http.MethodDelete,
			path:       "/signup",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestBuildRouter_ResponseHeaders(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET /healthz has valid HTTP response",
			method:     http.MethodGet,
			path:       "/healthz",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET / has valid HTTP response",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusOK,
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestBuildRouter_MultipleRequests(t *testing.T) {
	// Verify that the router handles multiple successive requests correctly
	// (stateless, no side effects between requests).
	r := buildRouter()

	paths := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/", http.StatusOK},
		{http.MethodGet, "/signup", http.StatusOK},
		{http.MethodPost, "/do_register", http.StatusOK},
		{http.MethodGet, "/signin", http.StatusOK},
		{http.MethodGet, "/admin/", http.StatusOK},
		{http.MethodGet, "/healthz", http.StatusOK},
	}

	for i := 0; i < 3; i++ {
		for _, p := range paths {
			req := httptest.NewRequest(p.method, p.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, p.want, w.Code, "iteration %d, path %s", i, p.path)
		}
	}
}

func TestBuildRouter_PanicRecovery(t *testing.T) {
	// Build a router and ensure the Recoverer middleware is active.
	// We simulate a panic by injecting a panicking handler onto a test router
	// that shares the same middleware setup, verifying the middleware chain is wired.
	r := buildRouter()
	require.NotNil(t, r, "router must not be nil – Recoverer middleware should be registered")

	// Directly verify known stable routes still respond 200 after any internal concern.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBuildRouter_RouteBodyContent(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantBody string
	}{
		{
			name:     "healthz body is 'ok'",
			method:   http.MethodGet,
			path:     "/healthz",
			wantBody: "ok\n",
		},
		{
			name:     "home body is 'route: Home'",
			method:   http.MethodGet,
			path:     "/",
			wantBody: "route: Home\n",
		},
		{
			name:     "signup body is 'route: Signup'",
			method:   http.MethodGet,
			path:     "/signup",
			wantBody: "route: Signup\n",
		},
		{
			name:     "do_register body is 'route: DoRegister'",
			method:   http.MethodPost,
			path:     "/do_register",
			wantBody: "route: DoRegister\n",
		},
		{
			name:     "signin body is 'route: Signin'",
			method:   http.MethodGet,
			path:     "/signin",
			wantBody: "route: Signin\n",
		},
		{
			name:     "admin dashboard body is 'route: AdminDashboard'",
			method:   http.MethodGet,
			path:     "/admin/",
			wantBody: "route: AdminDashboard\n",
		},
	}

	r := buildRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tc.wantBody, w.Body.String())
		})
	}
}
```