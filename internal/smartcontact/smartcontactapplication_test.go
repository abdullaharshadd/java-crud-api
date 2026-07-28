```go
package smartcontact

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetDB restores the package-level DB variable after each test to prevent
// state leakage between test cases.
func resetDB(t *testing.T, original *sql.DB) {
	t.Cleanup(func() {
		DB = original
	})
}

// ---------------------------------------------------------------------------
// SetDB tests
// ---------------------------------------------------------------------------

func TestSetDB(t *testing.T) {
	// We need a non-nil *sql.DB for the "valid" cases. We open a driver-less
	// handle using the "testing" driver trick: sql.Open succeeds even without a
	// real server because the driver validation is deferred to Ping/Query.
	// A simpler approach: use a real driver that always succeeds for Open.
	// We use "sqlite3" or just capture the nil-check logic. Since we can't
	// import an actual driver here without adding deps, we create a non-nil
	// *sql.DB via sql.Open with any registered driver name. The "pgx" driver
	// might not be registered; instead we check whether any driver is
	// available, otherwise we fabricate a *sql.DB value using unsafe tricks.
	//
	// The cleanest approach for unit tests: use a DSN of "" with the
	// "testDriver" that we register ourselves, OR just verify that a non-nil
	// *sql.DB is accepted and nil is rejected, relying on the type system.
	// We do the latter by passing a pointer obtained from sql.Open with a
	// dummy driver that we know is registered: "postgres" via pq, or we skip
	// the live-driver requirement entirely by using a mock.
	//
	// Because this package has no driver import, the only registered driver
	// we can guarantee is none. We therefore use a *sql.DB that we obtain by
	// opening with an empty driver name — which returns an error — and instead
	// create the value indirectly.
	//
	// Simplest safe approach: use the fact that sql.Open returns (*sql.DB, error)
	// even for unknown drivers IF we register a minimal one. We skip the actual
	// open and just verify the nil-rejection contract using a nil pointer for
	// the negative case and a struct-level sentinel for the positive case.

	t.Run("nil DB is rejected", func(t *testing.T) {
		original := DB
		resetDB(t, original)

		err := SetDB(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not be nil")
		// Package-level DB must remain unchanged.
		assert.Equal(t, original, DB)
	})

	t.Run("non-nil DB is accepted and assigned", func(t *testing.T) {
		original := DB
		resetDB(t, original)

		// Create a non-nil *sql.DB without an actual connection.
		// sql.Open defers network I/O; it only validates the driver name.
		// We register a minimal no-op driver for this purpose.
		db := newNoOpDB(t)

		err := SetDB(db)

		require.NoError(t, err)
		assert.Equal(t, db, DB, "package-level DB should be the one we passed in")
	})

	t.Run("multiple sequential calls overwrite the previous value", func(t *testing.T) {
		original := DB
		resetDB(t, original)

		db1 := newNoOpDB(t)
		db2 := newNoOpDB(t)

		require.NoError(t, SetDB(db1))
		assert.Equal(t, db1, DB)

		require.NoError(t, SetDB(db2))
		assert.Equal(t, db2, DB, "second call should overwrite with db2")
	})
}

// ---------------------------------------------------------------------------
// buildRouter – /healthz endpoint tests
// ---------------------------------------------------------------------------

func TestBuildRouter_Healthz(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:             "GET /healthz returns 200 and ok body",
			method:           http.MethodGet,
			path:             "/healthz",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "ok",
		},
		{
			name:             "GET /healthz is reachable even when DB is nil",
			method:           http.MethodGet,
			path:             "/healthz",
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "ok",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := DB
			DB = nil // force nil to verify infrastructure route independence
			t.Cleanup(func() { DB = original })

			router := buildRouter()
			srv := httptest.NewServer(router)
			t.Cleanup(srv.Close)

			resp, err := http.Get(srv.URL + tc.path)
			require.NoError(t, err)
			t.Cleanup(func() { resp.Body.Close() })

			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), tc.wantBodyContains)
		})
	}
}

// ---------------------------------------------------------------------------
// buildRouter – DB nil / non-nil wiring tests
// ---------------------------------------------------------------------------

func TestBuildRouter_DBNilSkipsUserRoutes(t *testing.T) {
	tests := []struct {
		name           string
		dbIsNil        bool
		path           string
		method         string
		wantStatusCode int
	}{
		{
			name:           "with nil DB, /healthz still returns 200",
			dbIsNil:        true,
			path:           "/healthz",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "with nil DB, unknown routes return 404",
			dbIsNil:        true,
			path:           "/api/users",
			method:         http.MethodGet,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := DB
			t.Cleanup(func() { DB = original })

			if tc.dbIsNil {
				DB = nil
			} else {
				DB = newNoOpDB(t)
			}

			router := buildRouter()
			srv := httptest.NewServer(router)
			t.Cleanup(srv.Close)

			req, err := http.NewRequest(tc.method, srv.URL+tc.path, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { resp.Body.Close() })

			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestBuildRouter_WithNonNilDB_RouterIsConstructed(t *testing.T) {
	tests := []struct {
		name    string
		setupDB func(t *testing.T) *sql.DB
	}{
		{
			name: "non-nil DB produces a valid router",
			setupDB: func(t *testing.T) *sql.DB {
				return newNoOpDB(t)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := DB
			t.Cleanup(func() { DB = original })

			DB = tc.setupDB(t)

			// buildRouter must not panic when DB is non-nil.
			var router http.Handler
			assert.NotPanics(t, func() {
				router = buildRouter()
			})
			require.NotNil(t, router)

			// Verify that /healthz is still reachable.
			srv := httptest.NewServer(router)
			t.Cleanup(srv.Close)

			resp, err := http.Get(srv.URL + "/healthz")
			require.NoError(t, err)
			t.Cleanup(func() { resp.Body.Close() })

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// buildRouter – middleware tests (panic recovery)
// ---------------------------------------------------------------------------

func TestBuildRouter_PanicRecovery(t *testing.T) {
	tests := []struct {
		name           string
		wantStatusCode int
	}{
		{
			name:           "panicking handler returns 500 without crashing the server",
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := DB
			t.Cleanup(func() { DB = original })
			DB = nil

			// Build the router and add a panicking route for test purposes.
			r := buildRouter().(*chi.Router_if_available)
			// Since buildRouter returns an http.Handler (chi.Router implements it),
			// we wrap it and inject a panic route via a separate mux for isolation.
			// Instead, we test the Recoverer middleware indirectly by wrapping
			// the returned handler with our own panicking handler and verifying
			// chi's Recoverer middleware catches it.
			//
			// The chi.Router returned by buildRouter has Recoverer already applied.
			// We add a panicking sub-handler to verify recovery.
			_ = r // unused if type assertion fails

			// Use a direct httptest approach: create a handler that panics,
			// wrap it with chi's Recoverer middleware, and verify 500 is returned.
			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				panic("deliberate test panic")
			})

			wrapped := panicMiddlewareWrapper(panicHandler)
			srv := httptest.NewServer(wrapped)
			t.Cleanup(srv.Close)

			resp, err := http.Get(srv.URL + "/")
			require.NoError(t, err)
			t.Cleanup(func() { resp.Body.Close() })

			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)
		})
	}
}

// ---------------------------------------------------------------------------
// buildRouter – application bootstrap invariants
// ---------------------------------------------------------------------------

func TestBuildRouter_ApplicationBootstrapInvariants(t *testing.T) {
	tests := []struct {
		name    string
		dbState *sql.DB
	}{
		{
			name:    "bootstrap with nil DB produces a non-nil handler",
			dbState: nil,
		},
		{
			name:    "bootstrap with non-nil DB produces a non-nil handler",
			dbState: nil, // replaced per test
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := DB
			t.Cleanup(func() { DB = original })

			if tc.name == "bootstrap with non-nil DB produces a non-nil handler" {
				DB = newNoOpDB(t)
			} else {
				DB = tc.dbState
			}

			var handler http.Handler
			require.NotPanics(t, func() {
				handler = buildRouter()
			}, "buildRouter must never panic regardless of DB state")

			assert.NotNil(t, handler, "buildRouter must always return a non-nil handler")
		})
	}
}

// ---------------------------------------------------------------------------
// SetDB – error message format invariant
// ---------------------------------------------------------------------------

func TestSetDB_ErrorFormat(t *testing.T) {
	tests := []struct {
		name            string
		input           *sql.DB
		wantErrContains string
		wantNoErr       bool
	}{
		{
			name:            "nil input produces descriptive error",
			input:           nil,
			wantErrContains: "smartcontact",
			wantNoErr:       false,
		},
		{
			name:      "non-nil input produces no error",
			input:     nil, // overridden in test body
			wantNoErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := DB
			t.Cleanup(func() { DB = original })

			var db *sql.DB
			if tc.wantNoErr {
				db = newNoOpDB(t)
			} else {
				db = tc.input
			}

			err := SetDB(db)

			if tc.wantNoErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Healthz – content-type and body invariants
// ---------------------------------------------------------------------------

func TestHealthz_ResponseInvariants(t *testing.T) {
	tests := []struct {
		name             string
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "body contains 'ok'",
			wantStatus:       http.StatusOK,
			wantBodyContains: "ok",
		},
		{
			name:             "status is exactly 200",
			wantStatus:       http.StatusOK,
			wantBodyContains: "ok",
		},
	}

	original := DB
	DB = nil
	t.Cleanup(func() { DB = original })

	router := buildRouter()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)

			body, err := io.ReadAll(rec.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), tc.wantBodyContains)
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// noOpDriver is a minimal database/sql driver that satisfies the interface
// without requiring a real database server. It is registered once so that
// sql.Open("noop", "") returns a non-nil *sql.DB.
import (
	"database/sql/driver"
)

type noOpDriver struct{}
type noOpConn struct{}
type noOpStmt struct{}
type noOpResult struct{}
type noOpRows struct{}
type noOpTx struct{}

func (d noOpDriver) Open(_ string) (driver.Conn, error) { return noOpConn{}, nil }

func (c noOpConn) Prepare(query string) (driver.Stmt, error) { return noOpStmt{}, nil }
func (c noOpConn) Close() error                              { return nil }
func (c noOpConn) Begin() (driver.Tx, error)                 { return noOpTx{}, nil }

func (s noOpStmt) Close() error                                    { return nil }
func (s noOpStmt) NumInput() int                                    { return -1 }
func (s noOpStmt) Exec(_ []driver.Value) (driver.Result, error)    { return noOpResult{}, nil }
func (s noOpStmt) Query(_ []driver.Value) (driver.Rows, error)     { return noOpRows{}, nil }

func (r noOpResult) LastInsertId() (int64, error) { return 0, nil }
func (r noOpResult) RowsAffected() (int64, error) { return 0, nil }

func (r noOpRows) Columns() []string              { return nil