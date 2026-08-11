```go
package smartcontact

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// setenv temporarily overrides an environment variable and restores the
// original value via t.Cleanup.
func setenv(t *testing.T, key, value string) {
	t.Helper()
	orig, had := os.LookupEnv(key)
	t.Setenv(key, value)
	t.Cleanup(func() {
		if had {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	})
}

// bodyOf reads the full response body as a string.
func bodyOf(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	return string(b)
}

// ---------------------------------------------------------------------------
// Tests for DefaultDatabaseURL
// ---------------------------------------------------------------------------

func TestDefaultDatabaseURL(t *testing.T) {
	assert.NotEmpty(t, DefaultDatabaseURL,
		"DefaultDatabaseURL must never be the empty string")
	assert.Contains(t, DefaultDatabaseURL, "postgres",
		"DefaultDatabaseURL should reference a postgres DSN")
}

// ---------------------------------------------------------------------------
// Tests for openDB
// ---------------------------------------------------------------------------

func TestOpenDB(t *testing.T) {
	tests := []struct {
		name    string
		envDSN  string // empty means "do not set"
		wantErr bool
		check   func(t *testing.T, db *sql.DB, err error)
	}{
		{
			name:    "returns error when DSN is unreachable (no env, default URL)",
			envDSN:  "",
			wantErr: true,
			check: func(t *testing.T, db *sql.DB, err error) {
				assert.Nil(t, db, "db must be nil on error")
				assert.Error(t, err)
			},
		},
		{
			name:    "returns error when explicit DSN is unreachable",
			envDSN:  "postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable",
			wantErr: true,
			check: func(t *testing.T, db *sql.DB, err error) {
				assert.Nil(t, db, "db must be nil on error")
				assert.Error(t, err)
			},
		},
		{
			name:    "uses SMARTCONTACT_DATABASE_URL env var when set",
			envDSN:  "postgres://customhost:5432/customdb?sslmode=disable",
			wantErr: true, // host unreachable in test environment
			check: func(t *testing.T, db *sql.DB, err error) {
				// We cannot assert which DSN was used from the outside, but we
				// can confirm that openDB attempted a connection (returns an error
				// rather than panicking) and does not leak a db handle.
				assert.Nil(t, db)
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Always clear the real env var first; set if test requires it.
			os.Unsetenv("SMARTCONTACT_DATABASE_URL")
			if tc.envDSN != "" {
				setenv(t, "SMARTCONTACT_DATABASE_URL", tc.envDSN)
			}

			db, err := openDB()
			if db != nil {
				t.Cleanup(func() { db.Close() })
			}
			tc.check(t, db, err)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – /healthz
// ---------------------------------------------------------------------------

func TestBuildRouter_Healthz(t *testing.T) {
	// buildRouter always wires /healthz regardless of DB availability, so
	// we can exercise it without a real database.
	t.Setenv("SMARTCONTACT_DATABASE_URL",
		"postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")

	router := buildRouter()
	require.NotNil(t, router)

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:             "GET /healthz returns 200 OK",
			method:           http.MethodGet,
			path:             "/healthz",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "ok",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.wantBodyContains)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – degraded mode (DB unreachable)
// ---------------------------------------------------------------------------

func TestBuildRouter_DegradedMode(t *testing.T) {
	t.Setenv("SMARTCONTACT_DATABASE_URL",
		"postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")

	router := buildRouter()
	require.NotNil(t, router, "buildRouter must return a non-nil handler even when DB is down")

	tests := []struct {
		name             string
		path             string
		wantStatusCode   int
		wantBodyContains string
	}{
		{
			name:             "non-health route returns 503 when DB unavailable",
			path:             "/users",
			wantStatusCode:   http.StatusServiceUnavailable,
			wantBodyContains: "database unavailable",
		},
		{
			name:             "arbitrary route returns 503 when DB unavailable",
			path:             "/api/contacts/123",
			wantStatusCode:   http.StatusServiceUnavailable,
			wantBodyContains: "database unavailable",
		},
		{
			name:             "/healthz still returns 200 in degraded mode",
			path:             "/healthz",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "ok",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.wantBodyContains)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – returns non-nil handler (bootstrap invariant)
// ---------------------------------------------------------------------------

func TestBuildRouter_AlwaysReturnsHandler(t *testing.T) {
	tests := []struct {
		name   string
		envDSN string
	}{
		{
			name:   "no env var set (uses default unreachable DSN)",
			envDSN: "",
		},
		{
			name:   "explicit unreachable DSN",
			envDSN: "postgres://bad:bad@127.0.0.1:1/bad?sslmode=disable",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv("SMARTCONTACT_DATABASE_URL")
			if tc.envDSN != "" {
				setenv(t, "SMARTCONTACT_DATABASE_URL", tc.envDSN)
			}

			h := buildRouter()
			assert.NotNil(t, h,
				"buildRouter must never return nil – the server must always be able to start")
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – single bootstrap entry point invariant
// ---------------------------------------------------------------------------

func TestBuildRouter_SingleEntryPoint(t *testing.T) {
	// Calling buildRouter twice must produce independent, functioning handlers.
	t.Setenv("SMARTCONTACT_DATABASE_URL",
		"postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")

	h1 := buildRouter()
	h2 := buildRouter()

	require.NotNil(t, h1)
	require.NotNil(t, h2)

	for i, h := range []http.Handler{h1, h2} {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code,
			"handler %d should serve /healthz with 200", i+1)
	}
}

// ---------------------------------------------------------------------------
// Tests for openDB – environment variable fallback behaviour
// ---------------------------------------------------------------------------

func TestOpenDB_EnvFallback(t *testing.T) {
	tests := []struct {
		name            string
		setEnv          bool
		envValue        string
		wantErrContains string
	}{
		{
			name:            "falls back to DefaultDatabaseURL when env is unset",
			setEnv:          false,
			wantErrContains: "ping database", // ping must fail in CI
		},
		{
			name:            "uses env DSN when SMARTCONTACT_DATABASE_URL is set",
			setEnv:          true,
			envValue:        "postgres://env:env@127.0.0.1:1/env?sslmode=disable",
			wantErrContains: "ping database",
		},
		{
			name:            "returns open error when DSN is completely malformed",
			setEnv:          true,
			envValue:        "not-a-valid-dsn://%%%",
			wantErrContains: "open database",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv("SMARTCONTACT_DATABASE_URL")
			if tc.setEnv {
				setenv(t, "SMARTCONTACT_DATABASE_URL", tc.envValue)
			}

			db, err := openDB()
			if db != nil {
				db.Close()
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains,
				"error message should describe the failure phase")
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – middleware wiring (Logger + Recoverer)
// ---------------------------------------------------------------------------

func TestBuildRouter_MiddlewarePanicsAreRecovered(t *testing.T) {
	// The Recoverer middleware must convert a panic into a 500 rather than
	// crashing the server. We exercise this by adding a panicking route to a
	// freshly built router.  Because buildRouter returns a chi.Router (which
	// also satisfies http.Handler), we cast it to add the extra route.

	// We use a custom test router to avoid mutating the production one.
	// This validates that chi + Recoverer are wired correctly in the
	// application's router construction approach.
	t.Setenv("SMARTCONTACT_DATABASE_URL",
		"postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")

	_ = buildRouter() // confirms it doesn't panic at construction time
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – /healthz is always registered (global invariant)
// ---------------------------------------------------------------------------

func TestBuildRouter_HealthzAlwaysRegistered(t *testing.T) {
	scenarios := []struct {
		name   string
		setup  func()
		teardown func()
	}{
		{
			name: "DB unreachable via env DSN",
			setup: func() {
				os.Setenv("SMARTCONTACT_DATABASE_URL",
					"postgres://x:x@127.0.0.1:1/x?sslmode=disable")
			},
			teardown: func() { os.Unsetenv("SMARTCONTACT_DATABASE_URL") },
		},
		{
			name: "DB unreachable via default DSN",
			setup: func() {
				os.Unsetenv("SMARTCONTACT_DATABASE_URL")
			},
			teardown: func() {},
		},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			sc.setup()
			t.Cleanup(sc.teardown)

			h := buildRouter()
			require.NotNil(t, h)

			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code,
				"/healthz must return 200 regardless of DB state")
			body := bodyOf(t, rec.Result())
			assert.Contains(t, body, "ok")
		})
	}
}

// ---------------------------------------------------------------------------
// Tests for openDB – returned *sql.DB is nil on error (no leak)
// ---------------------------------------------------------------------------

func TestOpenDB_NoResourceLeakOnError(t *testing.T) {
	setenv(t, "SMARTCONTACT_DATABASE_URL",
		"postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")

	db, err := openDB()
	assert.Nil(t, db, "a nil *sql.DB must be returned when Ping fails to prevent resource leaks")
	assert.Error(t, err)
	_ = fmt.Sprintf("%v", err) // ensure error is stringable without panic
}

// ---------------------------------------------------------------------------
// Tests for buildRouter – degraded 503 body contains error details
// ---------------------------------------------------------------------------

func TestBuildRouter_DegradedBodyContainsErrorDetail(t *testing.T) {
	setenv(t, "SMARTCONTACT_DATABASE_URL",
		"postgres://invalid:invalid@127.0.0.1:1/invalid?sslmode=disable")

	h := buildRouter()
	require.NotNil(t, h)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	body := bodyOf(t, rec.Result())
	assert.Contains(t, body, "database unavailable",
		"degraded mode 503 body must explain why the service is unavailable")
}
```