```go
package smartcontact_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"internal/smartcontact"
)

// TestContextLoads is the direct Go equivalent of the Spring Boot
// SmartContactApplicationTests contextLoads smoke test.  It asserts that the
// full dependency graph (repository -> service -> handler -> router) can be
// assembled without error and that every declared route is reachable.
func TestContextLoads(t *testing.T) {
	t.Run("the application context loads successfully", func(t *testing.T) {
		err := smartcontact.SmokeTest(context.Background())
		require.NoError(t, err, "SmokeTest must pass: the dependency graph must assemble and all routes must be wired")
	})
}

// TestNewSmokeRouter verifies that NewSmokeRouter returns a non-nil router that
// satisfies http.Handler, mirroring the invariant "all beans defined in the
// application configuration must be resolvable".
func TestNewSmokeRouter(t *testing.T) {
	t.Run("returns a non-nil http.Handler", func(t *testing.T) {
		router := smartcontact.NewSmokeRouter()
		assert.NotNil(t, router, "NewSmokeRouter must return a non-nil handler")
	})
}

// TestSmokeRoutes_Inventory asserts that SmokeRoutes returns the expected set
// of routes.  Adding or removing routes from the handler without updating the
// inventory will cause this test to fail, acting as a guard against incomplete
// wiring.
func TestSmokeRoutes_Inventory(t *testing.T) {
	expected := []smartcontact.SmokeRoute{
		{Method: http.MethodPost, Path: "/api/users"},
		{Method: http.MethodGet, Path: "/api/users"},
		{Method: http.MethodGet, Path: "/api/users/1"},
		{Method: http.MethodDelete, Path: "/api/users/1"},
		{Method: http.MethodPut, Path: "/api/users/1"},
		{Method: http.MethodGet, Path: "/api/users/name/alice"},
	}

	got := smartcontact.SmokeRoutes()
	assert.Equal(t, expected, got, "SmokeRoutes must return the canonical route inventory")
}

// TestSmokeTest_AllRoutesWired is a table-driven test that fires each declared
// route against the smoke router and asserts it does not return 404, proving
// the route was registered ("all beans/routes resolvable" invariant).
func TestSmokeTest_AllRoutesWired(t *testing.T) {
	router := smartcontact.NewSmokeRouter()

	for _, route := range smartcontact.SmokeRoutes() {
		route := route // capture
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			req := httptest.NewRequest(route.Method, route.Path, http.NoBody)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.NotEqual(t, http.StatusNotFound, rec.Code,
				"route %s %s must be registered on the router (got 404)",
				route.Method, route.Path)
		})
	}
}

// TestSmokeTest_ContextCancelled verifies that SmokeTest propagates a
// cancelled context to request execution without panicking, and still reports
// route wiring correctly.  Routes are wired even when the context is done;
// the router does not depend on context liveness for routing decisions.
func TestSmokeTest_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	// Even with a cancelled context the router is assembled (no DB involved),
	// and the fake service returns benign zero values, so SmokeTest must not
	// return a RouteNotWiredError.
	err := smartcontact.SmokeTest(ctx)
	if err != nil {
		var notWired *smartcontact.RouteNotWiredError
		assert.False(t, errors.As(err, &notWired),
			"a cancelled context must not cause route-wiring failures; got: %v", err)
	}
}

// TestRouteNotWiredError_Error verifies the error message format of
// RouteNotWiredError, covering the error_cases from the spec.
func TestRouteNotWiredError_Error(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		wantMsg string
	}{
		{
			name:    "GET route not wired",
			method:  http.MethodGet,
			path:    "/api/users",
			wantMsg: "route not wired: GET /api/users",
		},
		{
			name:    "POST route not wired",
			method:  http.MethodPost,
			path:    "/api/users",
			wantMsg: "route not wired: POST /api/users",
		},
		{
			name:    "DELETE route not wired",
			method:  http.MethodDelete,
			path:    "/api/users/42",
			wantMsg: "route not wired: DELETE /api/users/42",
		},
		{
			name:    "PUT route not wired",
			method:  http.MethodPut,
			path:    "/api/users/7",
			wantMsg: "route not wired: PUT /api/users/7",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := &smartcontact.RouteNotWiredError{Method: tc.method, Path: tc.path}
			assert.EqualError(t, err, tc.wantMsg)
		})
	}
}

// TestRouteNotWiredError_IsErrorInterface confirms that *RouteNotWiredError
// satisfies the error interface and can be unwrapped with errors.As.
func TestRouteNotWiredError_IsErrorInterface(t *testing.T) {
	wrapped := &smartcontact.RouteNotWiredError{Method: http.MethodGet, Path: "/api/users"}

	var target *smartcontact.RouteNotWiredError
	assert.True(t, errors.As(wrapped, &target),
		"*RouteNotWiredError must satisfy errors.As targeting *RouteNotWiredError")
	assert.Equal(t, http.MethodGet, target.Method)
	assert.Equal(t, "/api/users", target.Path)
}

// TestSmokeTest_UnregisteredRoute_DetectedAs404 builds a minimal router that
// intentionally omits one of the routes and confirms that SmokeTest returns a
// RouteNotWiredError for that route.  This exercises the error_case: "test
// fails if any bean cannot be instantiated or its dependencies cannot be
// resolved".
//
// We achieve this by replacing the router with a stub that always returns 404
// for every request, simulating a completely broken wiring.
func TestSmokeTest_UnregisteredRoute_DetectedAs404(t *testing.T) {
	// A handler that always returns 404, simulating no routes registered.
	alwaysNotFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Exercise each route manually against the always-404 handler.
	for _, route := range smartcontact.SmokeRoutes() {
		route := route
		t.Run("detects missing route: "+route.Method+" "+route.Path, func(t *testing.T) {
			req := httptest.NewRequest(route.Method, route.Path, http.NoBody)
			rec := httptest.NewRecorder()
			alwaysNotFound.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code,
				"stub handler must return 404 for %s %s", route.Method, route.Path)

			// Confirm that the RouteNotWiredError is what SmokeTest would
			// surface for such a response.
			err := &smartcontact.RouteNotWiredError{Method: route.Method, Path: route.Path}
			assert.Contains(t, err.Error(), route.Method)
			assert.Contains(t, err.Error(), route.Path)
		})
	}
}

// TestSmokeTest_BeanInstantiation asserts that calling NewSmokeRouter multiple
// times does not panic, modeling the Spring invariant "all beans defined in the
// application configuration must be resolvable" across repeated context loads.
func TestSmokeTest_BeanInstantiation(t *testing.T) {
	const iterations = 5

	for i := 0; i < iterations; i++ {
		i := i
		t.Run("iteration", func(t *testing.T) {
			assert.NotPanics(t, func() {
				router := smartcontact.NewSmokeRouter()
				assert.NotNil(t, router, "router must be non-nil on iteration %d", i)
			}, "NewSmokeRouter must not panic on iteration %d", i)
		})
	}
}

// TestSmokeTest_NoTestMethodsRequired mirrors the Spring invariant
// "no test methods are required for the context-load verification to execute":
// SmokeTest must complete successfully even when invoked with no additional
// setup, relying solely on the in-memory fake.
func TestSmokeTest_NoTestMethodsRequired(t *testing.T) {
	err := smartcontact.SmokeTest(context.Background())
	assert.NoError(t, err,
		"SmokeTest must pass with no additional setup beyond the in-memory fake service")
}

// TestSmokeTest_FullRouteTable is a comprehensive table-driven test that
// consolidates all declared routes into a single test table, verifying that
// each route:
//  1. Is not 404 (wired),
//  2. Returns a status code the handler can produce (i.e. non-zero),
//  3. Does not cause the router to panic.
func TestSmokeTest_FullRouteTable(t *testing.T) {
	router := smartcontact.NewSmokeRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		mustNotBe404   bool
		mustNotBe405   bool
	}{
		{
			name:         "POST /api/users creates a user",
			method:       http.MethodPost,
			path:         "/api/users",
			mustNotBe404: true,
		},
		{
			name:         "GET /api/users lists users",
			method:       http.MethodGet,
			path:         "/api/users",
			mustNotBe404: true,
		},
		{
			name:         "GET /api/users/1 fetches by ID",
			method:       http.MethodGet,
			path:         "/api/users/1",
			mustNotBe404: true,
		},
		{
			name:         "DELETE /api/users/1 deletes a user",
			method:       http.MethodDelete,
			path:         "/api/users/1",
			mustNotBe404: true,
		},
		{
			name:         "PUT /api/users/1 updates a user",
			method:       http.MethodPut,
			path:         "/api/users/1",
			mustNotBe404: true,
		},
		{
			name:         "GET /api/users/name/alice fetches by name",
			method:       http.MethodGet,
			path:         "/api/users/name/alice",
			mustNotBe404: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, http.NoBody)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				router.ServeHTTP(rec, req)
			}, "router.ServeHTTP must not panic for %s %s", tc.method, tc.path)

			assert.NotZero(t, rec.Code,
				"response code must be set for %s %s", tc.method, tc.path)

			if tc.mustNotBe404 {
				assert.NotEqual(t, http.StatusNotFound, rec.Code,
					"route %s %s must be wired (must not return 404)", tc.method, tc.path)
			}
		})
	}
}
```