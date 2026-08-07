```go
package smartcontact

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextLoads_TableDriven is the table-driven Go equivalent of the Spring
// Boot "context loads" smoke test. It verifies that the application composition
// root (NewApp / Router) can be constructed without error, mirroring Spring's
// behaviour of failing the test if the application context cannot be
// initialised.
func TestContextLoads_TableDriven(t *testing.T) {
	tests := []struct {
		name                string
		buildApp            func() *App
		wantAppNil          bool
		wantRouterNil       bool
		wantPanic           bool
		description         string
	}{
		{
			name: "valid_complete_configuration_context_loads_successfully",
			buildApp: func() *App {
				return NewApp()
			},
			wantAppNil:    false,
			wantRouterNil: false,
			wantPanic:     false,
			description:   "the application is started with a valid and complete configuration; all beans/dependencies must be resolvable",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Guard against any panic during construction — mirrors Spring
			// failing when the application context cannot be initialised.
			var app *App
			var panicked bool
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				app = tc.buildApp()
			}()

			if tc.wantPanic {
				assert.True(t, panicked,
					"[%s] expected a panic during app construction but none occurred", tc.name)
				// Nothing further to check when we expected a panic.
				return
			}

			require.False(t, panicked,
				"[%s] NewApp panicked: application composition root failed to build — %s",
				tc.name, tc.description)

			if tc.wantAppNil {
				assert.Nil(t, app,
					"[%s] expected NewApp to return nil but got a non-nil app", tc.name)
				return
			}

			// Invariant: the application object must not be nil.
			require.NotNil(t, app,
				"[%s] NewApp returned nil: application composition root failed to build — %s",
				tc.name, tc.description)

			// Instantiate the router (all handlers/routes wired up).
			var router http.Handler
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				router = app.Router()
			}()

			require.False(t, panicked,
				"[%s] app.Router() panicked: HTTP router wiring failed — %s",
				tc.name, tc.description)

			if tc.wantRouterNil {
				assert.Nil(t, router,
					"[%s] expected Router to return nil but got a non-nil handler", tc.name)
				return
			}

			// Invariant: the router must not be nil — all mandatory beans/
			// routes must be resolvable at startup.
			assert.NotNil(t, router,
				"[%s] App.Router returned nil: HTTP router failed to wire up — %s",
				tc.name, tc.description)
		})
	}
}

// TestRouterServesRequests_TableDriven is a table-driven smoke test that
// confirms the wired router can accept and dispatch HTTP requests without
// panicking. Each row exercises a different endpoint / method combination to
// verify that the transport layer is reachable.
func TestRouterServesRequests_TableDriven(t *testing.T) {
	// Build the application once; it is shared across all table rows because
	// construction is the concern under test and we want to avoid masking
	// issues that only appear on repeated use of the same instance.
	app := NewApp()
	require.NotNil(t, app, "NewApp returned nil: cannot proceed with request smoke tests")

	router := app.Router()
	require.NotNil(t, router, "App.Router returned nil: cannot proceed with request smoke tests")

	tests := []struct {
		name           string
		method         string
		path           string
		wantNonZeroCode bool
		description    string
	}{
		{
			name:           "GET_api_users_returns_well_formed_response",
			method:         http.MethodGet,
			path:           "/api/users",
			wantNonZeroCode: true,
			description:    "GET /api/users must produce any well-formed HTTP response, proving the transport layer is wired",
		},
		{
			name:           "POST_api_users_returns_well_formed_response",
			method:         http.MethodPost,
			path:           "/api/users",
			wantNonZeroCode: true,
			description:    "POST /api/users must produce any well-formed HTTP response",
		},
		{
			name:           "GET_unknown_path_returns_well_formed_response",
			method:         http.MethodGet,
			path:           "/api/unknown-endpoint-xyz",
			wantNonZeroCode: true,
			description:    "an unknown path must still yield a well-formed HTTP response (e.g. 404), not a zero code",
		},
		{
			name:           "GET_root_returns_well_formed_response",
			method:         http.MethodGet,
			path:           "/",
			wantNonZeroCode: true,
			description:    "GET / must produce any well-formed HTTP response",
		},
		{
			name:           "GET_health_or_root_api_returns_well_formed_response",
			method:         http.MethodGet,
			path:           "/api",
			wantNonZeroCode: true,
			description:    "GET /api must produce any well-formed HTTP response",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			// The router must dispatch without panicking.
			var panicked bool
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				router.ServeHTTP(rec, req)
			}()

			assert.False(t, panicked,
				"[%s] router panicked while serving %s %s — transport layer is not stable",
				tc.name, tc.method, tc.path)

			if tc.wantNonZeroCode {
				// httptest.ResponseRecorder initialises Code to 200 if
				// WriteHeader is never called, but any explicit write
				// also sets it. A code of 0 would only happen if the
				// ResponseRecorder were in some unexpected state; the
				// important assertion here is "no panic + code recorded".
				assert.NotEqual(t, 0, rec.Code,
					"[%s] router did not write any response for %s %s: transport layer is not wired — %s",
					tc.name, tc.method, tc.path, tc.description)

				// Validate that the response code is within the valid HTTP
				// status-code range, ensuring a well-formed response.
				assert.GreaterOrEqual(t, rec.Code, 100,
					"[%s] response code %d is below the valid HTTP range for %s %s",
					tc.name, rec.Code, tc.method, tc.path)
				assert.Less(t, rec.Code, 600,
					"[%s] response code %d is above the valid HTTP range for %s %s",
					tc.name, rec.Code, tc.method, tc.path)
			}
		})
	}
}

// TestContextLoads_ErrorCases validates the invariants described in the
// behavioural spec's error_cases section: the test must fail if the
// application cannot be constructed. Because Go does not have a DI container,
// we simulate misconfiguration by providing factory functions that return nil
// or panic, and assert that the test infrastructure correctly detects these.
func TestContextLoads_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		appFactory    func() *App
		expectFailure bool
		description   string
	}{
		{
			name: "nil_app_is_detected_as_failure",
			appFactory: func() *App {
				// Simulates a misconfigured composition root that returns nil,
				// analogous to a missing required bean in Spring.
				return nil
			},
			expectFailure: true,
			description:   "a nil App is the Go equivalent of a missing required bean — must be detected",
		},
		{
			name: "valid_app_is_not_a_failure",
			appFactory: func() *App {
				return NewApp()
			},
			expectFailure: false,
			description:   "a properly constructed App should not be treated as a failure",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var app *App
			var panicked bool

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				app = tc.appFactory()
			}()

			isFailure := panicked || app == nil

			if tc.expectFailure {
				assert.True(t, isFailure,
					"[%s] expected a construction failure (nil app or panic) but the app was built successfully — %s",
					tc.name, tc.description)
			} else {
				assert.False(t, isFailure,
					"[%s] did not expect a construction failure but got one (panicked=%v, app==nil=%v) — %s",
					tc.name, panicked, app == nil, tc.description)

				// Additional invariant: if the app is non-nil, its router must
				// also be non-nil — all mandatory routes must be resolvable.
				if app != nil {
					router := app.Router()
					assert.NotNil(t, router,
						"[%s] Router() returned nil even though NewApp succeeded — mandatory routes are not wired",
						tc.name)
				}
			}
		})
	}
}

// TestApplicationContextInvariant_NoExplicitAssertionsBeyondLoading verifies
// the spec invariant: "no explicit assertions are performed beyond successful
// context loading". The test simply builds the app; if NewApp or Router
// panics or returns nil, the test fails — exactly as stated.
func TestApplicationContextInvariant_NoExplicitAssertionsBeyondLoading(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "application_bootstraps_full_context_without_errors",
			description: "the application must be able to bootstrap its full context; all mandatory config and beans must be resolvable at startup",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var (
				app    *App
				router http.Handler
			)

			require.NotPanics(t, func() {
				app = NewApp()
			}, "[%s] NewApp must not panic — %s", tc.name, tc.description)

			require.NotNil(t, app,
				"[%s] NewApp must return a non-nil App — %s", tc.name, tc.description)

			require.NotPanics(t, func() {
				router = app.Router()
			}, "[%s] app.Router must not panic — %s", tc.name, tc.description)

			require.NotNil(t, router,
				"[%s] app.Router must return a non-nil http.Handler — %s", tc.name, tc.description)

			// Spec invariant: the test passes if and only if the entire
			// context initialises without errors. No further business logic
			// assertions are made here.
		})
	}
}
```