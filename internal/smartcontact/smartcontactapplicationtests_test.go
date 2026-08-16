```go
package smartcontact

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockUserService is a test double that satisfies whatever UserService
// interface BuildRouterWith expects, without touching a real database.
type mockUserService struct {
	healthy bool
}

// Ensure mockUserService implements UserService at compile time.
// If the interface changes this will surface immediately.
var _ UserService = (*mockUserService)(nil)

// ---------------------------------------------------------------------------
// Smoke-test table
// ---------------------------------------------------------------------------

func TestContextLoads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         Config
		svc         UserService
		wantRouter  bool   // router must be non-nil
		wantPanic   bool   // BuildRouterWith must not panic
		description string // mirrors the spec scenario
	}{
		{
			name: "valid_configuration_and_service_wires_successfully",
			cfg: Config{
				Port:   "8080",
				DBPath: ":memory:",
			},
			svc:         &mockUserService{healthy: true},
			wantRouter:  true,
			wantPanic:   false,
			description: "the application is started in a test environment with a valid configuration",
		},
		{
			name: "empty_port_still_produces_router",
			cfg: Config{
				Port:   "",
				DBPath: ":memory:",
			},
			svc:         &mockUserService{healthy: true},
			wantRouter:  true,
			wantPanic:   false,
			description: "edge-case: missing port does not prevent context load",
		},
		{
			name: "zero_value_config_does_not_panic",
			cfg:         Config{},
			svc:         &mockUserService{healthy: true},
			wantRouter:  true,
			wantPanic:   false,
			description: "zero-value Config represents bare-minimum misconfiguration that should not panic",
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Verify the smoke-test constant is stable (mirrors the Java test name).
			assert.Equal(t, "SmartContactApplicationContextLoads", smokeTestName,
				"smokeTestName constant must match the origin Java test")

			// Guard against panics during dependency-injection wiring.
			if tc.wantPanic {
				assert.Panics(t, func() {
					BuildRouterWith(tc.cfg, tc.svc)
				}, "expected BuildRouterWith to panic for scenario: %s", tc.description)
				return
			}

			assert.NotPanics(t, func() {
				router := BuildRouterWith(tc.cfg, tc.svc)
				if tc.wantRouter {
					assert.NotNil(t, router,
						"router must be non-nil after successful wiring: %s", tc.description)
				}
			}, "BuildRouterWith must not panic for scenario: %s", tc.description)
		})
	}
}

// TestContextLoads_RouterServesHTTP verifies that the wired router actually
// handles HTTP requests — the equivalent of "all beans are resolvable and the
// embedded server is functional".
func TestContextLoads_RouterServesHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cfg            Config
		svc            UserService
		method         string
		path           string
		wantStatusMin  int // inclusive lower bound
		wantStatusMax  int // inclusive upper bound
		description    string
	}{
		{
			name: "health_endpoint_is_reachable",
			cfg: Config{
				Port:   "8080",
				DBPath: ":memory:",
			},
			svc:          &mockUserService{healthy: true},
			method:       http.MethodGet,
			path:         "/health",
			wantStatusMin: http.StatusOK,
			wantStatusMax: http.StatusOK,
			description:  "wired router must respond to health checks",
		},
		{
			name: "unknown_path_returns_not_found",
			cfg: Config{
				Port:   "8080",
				DBPath: ":memory:",
			},
			svc:          &mockUserService{healthy: true},
			method:       http.MethodGet,
			path:         "/no-such-route-xyzzy",
			wantStatusMin: http.StatusNotFound,
			wantStatusMax: http.StatusNotFound,
			description:  "wired router must return 404 for unregistered paths",
		},
		{
			name: "root_path_does_not_panic",
			cfg: Config{
				Port:   "8080",
				DBPath: ":memory:",
			},
			svc:          &mockUserService{healthy: true},
			method:       http.MethodGet,
			path:         "/",
			wantStatusMin: http.StatusOK,
			wantStatusMax: http.StatusNotFound, // either is acceptable for root
			description:  "wired router must not panic on root path",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			router := BuildRouterWith(tc.cfg, tc.svc)
			assert.NotNil(t, router,
				"router must not be nil before issuing test request")

			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				router.ServeHTTP(w, req)
			}, "router.ServeHTTP must not panic: %s", tc.description)

			res := w.Result()
			assert.GreaterOrEqual(t, res.StatusCode, tc.wantStatusMin,
				"status code too low for: %s", tc.description)
			assert.LessOrEqual(t, res.StatusCode, tc.wantStatusMax,
				"status code too high for: %s", tc.description)
		})
	}
}

// TestContextLoads_AllBeansResolvable mirrors the Spring invariant
// "all required beans must be resolvable during context initialization".
// In Go terms this means BuildRouterWith must complete without error when
// all dependencies are provided.
func TestContextLoads_AllBeansResolvable(t *testing.T) {
	t.Parallel()

	type beanSet struct {
		name string
		cfg  Config
		svc  UserService
	}

	tests := []struct {
		name     string
		beans    beanSet
		wantFail bool // true when we expect wiring to be impossible
	}{
		{
			name: "full_dependency_set_resolves",
			beans: beanSet{
				name: "full",
				cfg:  Config{Port: "9090", DBPath: ":memory:"},
				svc:  &mockUserService{healthy: true},
			},
			wantFail: false,
		},
		{
			name: "nil_service_represents_missing_required_bean",
			beans: beanSet{
				name: "missing_svc",
				cfg:  Config{Port: "9090", DBPath: ":memory:"},
				svc:  nil,
			},
			// BuildRouterWith may panic or return nil when svc is nil —
			// either counts as "bean not resolvable".
			wantFail: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantFail {
				// Acceptable outcomes: panic OR nil router.
				var router http.Handler
				panicked := func() (p bool) {
					defer func() {
						if r := recover(); r != nil {
							p = true
						}
					}()
					router = BuildRouterWith(tc.beans.cfg, tc.beans.svc)
					return false
				}()

				if !panicked {
					assert.Nil(t, router,
						"nil service must yield nil router (bean not resolvable)")
				}
				return
			}

			// Happy path: no panic, non-nil router.
			var router http.Handler
			assert.NotPanics(t, func() {
				router = BuildRouterWith(tc.beans.cfg, tc.beans.svc)
			})
			assert.NotNil(t, router,
				"all beans resolvable → router must be non-nil")
		})
	}
}

// TestSmokeTestName validates the package-level constant that documents the
// migration origin.
func TestSmokeTestName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "SmartContactApplicationContextLoads", smokeTestName,
		"constant must preserve the Java test name for traceability")
	assert.NotEmpty(t, smokeTestName,
		"smokeTestName must never be empty")
}
```